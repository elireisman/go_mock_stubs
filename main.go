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
	"strings"

	"github.com/elireisman/go_mock_stubs/tree"
	"github.com/elireisman/go_mock_stubs/utils"
)

const MockTemplate = `{{$unit := .}}

package {{.Pkg}}

{{.FormatImports}}

{{range $rcvr, $sigs := $unit.Methods}}
{{$isLocal := $rcvr | $unit.IsDeclaredHere}}{{if $isLocal}}
type {{$rcvr}}Iface interface {
{{range $sig := $sigs}}  {{$sig.Name}}({{$sig.ListArgs}}){{$sig.ListReturns}}
{{end}}}
type {{$firstSig := index $sigs 0}}{{$firstSig.Receiver.ToMock}} struct { }
{{end}}{{end}}

{{range $rcvr, $sigs := $unit.Methods}}{{$isLocal := $rcvr | $unit.IsDeclaredHere}}{{if $isLocal}}
{{range $sig := $sigs}}func ({{$sig.Receiver.Name}} {{$sig.Receiver.Ptr}}{{$sig.Receiver.ToMock}}) {{$sig.Name}}({{$sig.ListArgs}}){{$sig.ListReturns}} {
  panic("mock: stub method not implemented")
}

{{end}}{{end}}

{{end}}
`

var (
	SourceDir   string
	StdOut      bool
	MultiLineWS *regexp.Regexp
)

func init() {
	flag.StringVar(&SourceDir, "source-dir", "example", "this dir will be recursively searched for Golang source files to parse")
	flag.BoolVar(&StdOut, "stdout", false, "stream output code to stdout rather than writing to *_mock.go file under the source's dir")

	MultiLineWS = regexp.MustCompile(`\r?\n\r?\n(\r?\n)+`)
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

	// maintain mapping of filename to file contents for rendering
	outFiles := map[string]tree.CompilationUnit{}

	// iterate over all files in each package, extracting info we need to generate mock files
	for _, pkg := range inPkgs {

		pkgName := pkg.Name
		if _, ok := outPkgs[pkgName]; !ok {
			outPkgs[pkgName] = map[string][]tree.Signature{}
		}

		for fileName, node := range pkg.Files {

			unit := tree.CompilationUnit{
				Pkg:      pkgName,
				Source:   fileName,
				Imports:  []tree.Import{},
				DeclHere: map[string]bool{},
				Prefixes: map[string]bool{},
				Methods:  outPkgs[pkgName],
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
					if utils.IsPublicMethod(&unit, fn) {
						sig := tree.Signature{
							Name:     fn.Name.Name,
							Receiver: tree.NewField(fn.Recv.List[0]),
							Args:     utils.ProcessFields(&unit, fn.Type.Params),
							Returns:  utils.ProcessFields(&unit, fn.Type.Results),
						}

						rcvrType := sig.Receiver.Type[len(sig.Receiver.Type)-1]
						unit.Methods[rcvrType] = append(unit.Methods[rcvrType], sig)
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

		// TODO: something less lazy here
		out := strings.Trim(MultiLineWS.ReplaceAllString(raw.String(), "\n\n"), " \t\r\n")
		destFilePath := utils.BuildDest(fileName)

		if StdOut {
			fmt.Println()
			fmt.Printf("\x1b[1m[.] dry run for output file: %q\x1b[0m\n", destFilePath)
			fmt.Println(out)
		} else {
			fmt.Printf("\x1b[1m[.] writing output file to: %q\x1b[0m\n", destFilePath)
			if ioutil.WriteFile(destFilePath, []byte(out), os.FileMode(0664)); err != nil {
				panic(fmt.Sprintf("failed to write output to %q, error: %s", destFilePath, err))
			}
		}
	}
	fmt.Println("\n\x1b[1m[.] code generation complete\x1b[0m\n")
}
