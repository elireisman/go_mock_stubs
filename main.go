package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"

	"github.com/elireisman/go_mock_stubs/tree"
	"github.com/elireisman/go_mock_stubs/utils"
)

const (
	GlobalScope = ""

	MockTemplate = `package {{.Pkg}}

{{.FormatImports}}

{{range $rcvr, $sigs := .Funcs}}type {{$x := index $sigs 0}}{{$x.Receiver.ToMock}} struct { }

type {{$rcvr}}Iface interface {
{{range $sig := $sigs}}  {{$sig.Name}}({{$sig.ListArgs}}) {{$sig.ListReturns}}
{{end}}}
{{end}}

{{range $rcvr, $sigs := .Funcs}}
{{range $sig := $sigs}}func ({{$sig.Receiver.Name}} *{{$sig.Receiver.ToMock}}) {{$sig.Name}}({{$sig.ListArgs}}) {{$sig.ListReturns}} { {{$sig.BuildReturnStmt}} }
{{end}}
{{end}}
`
)

var (
	SourceDir string
	StdOut    bool
)

func main() {
	flag.StringVar(&SourceDir, "source-dir", "example", "the directory under which all Golang source files will be parsed")
	flag.BoolVar(&StdOut, "stdout", false, "stream output code to stdout rather than writing to *_mock.go file under the source's dir")
	flag.Parse()

	fileSet := token.NewFileSet()
	inPkgs, err := parser.ParseDir(fileSet, SourceDir, nil, parser.ParseComments)
	if err != nil {
		panic(fmt.Sprintf("failed to parse source files in %q into Golang AST: %s", SourceDir, err))
	}

	// maintain global mapping of "pkg name" -> "public struct" -> "methods"
	// use this to render files in 2nd pass so that methods defined on
	// same struct across multiple source files are properly mocked
	outPkgs := map[string]map[string][]tree.Signature{}

	// maintain mapping of filename to file render data for printing each
	// compilation unit holds a ref to the global struct map for it's pkg
	outFiles := map[string]tree.CompilationUnit{}

	// iterate over each package, all files in each package.
	for _, pkg := range inPkgs {

		pkgName := pkg.Name
		if _, ok := outPkgs[pkgName]; !ok {
			outPkgs[pkgName] = map[string][]tree.Signature{}
		}

		// TODO: why no imports in node, is this an *ast.File like before or not?!?
		// TODO: need to parse out actual struct decls per file so we don't redeclare b/c global func mapping?

		for fileName, node := range pkg.Files {

			unit := tree.CompilationUnit{
				Pkg:      pkgName,
				Imports:  []tree.Import{},
				Prefixes: map[string]bool{},
				Funcs:    outPkgs[pkgName],
			}
			outFiles[fileName] = unit

			for _, impt := range node.Imports {
				alias := ""
				if impt.Name != nil {
					alias = impt.Name.Name
				}
				next := tree.Import{Alias: alias, Path: impt.Path.Value}
				unit.Imports = append(unit.Imports, next)
			}

			// extract metadata from any methods with a pointer to a struct as a receiver
			ast.Inspect(node, func(n ast.Node) bool {
				if fn, ok := n.(*ast.FuncDecl); ok {
					if len(fn.Name.Name) > 0 && fn.Name.IsExported() {
						rName, rType := GlobalScope, GlobalScope
						if fn.Recv != nil && len(fn.Recv.List) > 0 {
							if ptrExpr, ok := fn.Recv.List[0].Type.(*ast.StarExpr); ok {
								astID, _ := ptrExpr.X.(*ast.Ident)
								rName, rType = fn.Recv.List[0].Names[0].Name, astID.Name
							}
						}

						// don't pick up functions that don't have a reciever (i.e. globals)
						if rName == "" {
							return true
						}

						sig := tree.Signature{
							Name:     fn.Name.Name,
							Receiver: tree.NewField(fn.Recv.List[0]),
							Args:     utils.FormatArgs(&unit, fn.Type.Params),
							Returns:  utils.FormatArgs(&unit, fn.Type.Results),
						}

						unit.Funcs[rType] = append(unit.Funcs[rType], sig)
					}
				}

				return true
			})
		}
	}

	// now that we've collected all the state across all the packages,
	// we have the complete picture and can render the output files
	fmt.Println()
	for fileName, unit := range outFiles {
		out, err := utils.Render(&unit, MockTemplate)
		if err != nil {
			panic(err)
		}
		if StdOut {
			fmt.Println(out.String())
		} else {
			destFilePath := utils.BuildDest(fileName)
			if ioutil.WriteFile(destFilePath, out.Bytes(), os.FileMode(0664)); err != nil {
				panic(fmt.Sprintf("failed to write output to %q, error: %s", destFilePath, err))
			}
		}
		fmt.Println()
	}
}
