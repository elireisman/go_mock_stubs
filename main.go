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

	MockTemplate = `
package {{.Pkg}}

{{.FormatImports}}

{{range $rcvr, $sigs := .Funcs}}
type {{$x := index $sigs 0}}{{$x.Receiver.ToMock}} struct { }

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
	SourceFile string
	StdOut     bool
)

func main() {
	flag.StringVar(&SourceFile, "source-file", "example.go", "the Golang file to parse")
	flag.BoolVar(&StdOut, "stdout", false, "stream code to stdout rather than written to a *_mock.go file")
	flag.Parse()

	fileSet := token.NewFileSet()
	node, err := parser.ParseFile(fileSet, SourceFile, nil, parser.ParseComments)
	if err != nil {
		panic(fmt.Sprintf("failed to parse source file %q into Golang AST: %s", err))
	}

	unit := &tree.CompilationUnit{
		Pkg:      node.Name.Name,
		Imports:  []tree.Import{},
		Prefixes: map[string]bool{},
		Funcs:    map[string][]tree.Signature{},
	}

	for _, impt := range node.Imports {
		alias := ""
		if impt.Name != nil {
			alias = impt.Name.Name
		}
		path := impt.Path.Value[1 : len(impt.Path.Value)-1]
		next := tree.Import{Alias: alias, Path: path}
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
					Args:     utils.FormatArgs(unit, fn.Type.Params),
					Returns:  utils.FormatArgs(unit, fn.Type.Results),
				}

				unit.Funcs[rType] = append(unit.Funcs[rType], sig)
			}
		}

		return true
	})

	out, err := utils.Render(unit, MockTemplate)
	if err != nil {
		panic(err)
	}
	if StdOut {
		fmt.Println(out.String())
	} else {
		destFilePath := utils.BuildDest(SourceFile)
		if ioutil.WriteFile(destFilePath, out.Bytes(), os.FileMode(0664)); err != nil {
			panic(fmt.Sprintf("failed to write output to %q, error: %s", destFilePath, err))
		}
	}

	fmt.Println()
}
