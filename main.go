package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"

	"github.com/elireisman/go_mock_stubs/tree"
	"github.com/elireisman/go_mock_stubs/utils"
)

const (
	GlobalScope = ""

	MockTemplate = `{{$unit := .}}

package {{.Pkg}}

{{.FormatImports}}

{{range $rcvr, $sigs := .Funcs}}

{{$isLocal := $rcvr | $unit.IsDeclaredHere}}{{if $isLocal}}
type {{$x := index $sigs 0}}{{$x.Receiver.ToMock}} struct { }

type {{$rcvr}}Iface interface {
{{range $sig := $sigs}}  {{$sig.Name}}({{$sig.ListArgs}}){{$sig.ListReturns}}
{{end}}
}
{{end}}

{{end}}

{{range $rcvr, $sigs := .Funcs}}{{$isLocal := $rcvr | $unit.IsDeclaredHere}}{{if $isLocal}}
{{range $sig := $sigs}}func ({{$sig.Receiver.Name}} *{{$sig.Receiver.ToMock}}) {{$sig.Name}}({{$sig.ListArgs}}){{$sig.ListReturns}} {
  panic("mock: stub method not implemented")
}
{{end}}
{{end}}

{{end}}
`
)

var (
	SourceDir   string
	StdOut      bool
	MultiLineWS *regexp.Regexp
)

func init() {
	flag.StringVar(&SourceDir, "source-dir", "example", "the directory under which all Golang source files will be parsed")
	flag.BoolVar(&StdOut, "stdout", false, "stream output code to stdout rather than writing to *_mock.go file under the source's dir")

	MultiLineWS = regexp.MustCompile(`(\r?\n)(\r?\n)+`)
}

func main() {
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

	// maintain mapping of filename to file render data for printing
	outFiles := map[string]tree.CompilationUnit{}

	// iterate over each package, all files in each package.
	for _, pkg := range inPkgs {

		pkgName := pkg.Name
		if _, ok := outPkgs[pkgName]; !ok {
			outPkgs[pkgName] = map[string][]tree.Signature{}
		}

		for fileName, node := range pkg.Files {

			unit := tree.CompilationUnit{
				Pkg:      pkgName,
				Imports:  []tree.Import{},
				DeclHere: map[string]bool{},
				Prefixes: map[string]bool{},
				Funcs:    outPkgs[pkgName],
			}

			for _, impt := range node.Imports {
				alias := ""
				if impt.Name != nil {
					alias = impt.Name.Name
				}
				path, _ := strconv.Unquote(impt.Path.Value)
				next := tree.Import{Alias: alias, Path: path}
				unit.Imports = append(unit.Imports, next)
			}

			// extract metadata from any methods with a pointer to a struct as a receiver
			ast.Inspect(node, func(n ast.Node) bool {
				if ts, ok := n.(*ast.TypeSpec); ok {
					// collect mapping of public structs declared in THIS file
					if _, ok := ts.Type.(*ast.StructType); ok {
						if ts.Name != nil && ts.Name.IsExported() {
							unit.DeclHere[ts.Name.Name] = true
						}
					}
				} else if fn, ok := n.(*ast.FuncDecl); ok {
					// collect all public struct method decls across files in pkg
					// so that we can generate mock stubs with full API at struct decl site
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

			// record this compilation unit (file) in the output mapping
			outFiles[fileName] = unit
		}
	}

	// now that we've collected all the state across all the packages,
	// we have the complete picture and can render the output files
	for fileName, unit := range outFiles {
		raw, err := utils.Render(&unit, MockTemplate)
		if err != nil {
			panic(err)
		}
		// empty output buffer means don't print the file, only print
		// mock if source file contains struct declarations we're mocking
		if len(raw.Bytes()) == 0 {
			continue
		}

		// all that whitespace for readability in MockTemplate comes at cost...
		out := MultiLineWS.ReplaceAllString(raw.String(), "\n")

		if StdOut {
			fmt.Println()
			fmt.Println(out)
		} else {
			destFilePath := utils.BuildDest(fileName)
			if ioutil.WriteFile(destFilePath, []byte(out), os.FileMode(0664)); err != nil {
				panic(fmt.Sprintf("failed to write output to %q, error: %s", destFilePath, err))
			}
			fmt.Printf("[INFO] wrote: %q\n", destFilePath)
		}
	}
	fmt.Println()
}
