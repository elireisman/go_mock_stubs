package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/elireisman/go_mock_stubs/tree"
	"github.com/elireisman/go_mock_stubs/utils"
)

const (
	GlobalScope = ""

	MockTemplate = `
package {{.Pkg}}

{{.FormatImports}}

{{range $rcvr, $sigs := .Funcs}}

type {{$x := index $sigs 0}}{{$x.RcvrType}} struct { }
type Iface{{$rcvr}} interface {
{{range $sig := $sigs}}  {{$sig.Name}}({{$sig.ListArgs}}) {{$sig.ListReturns}}
{{end}}}
{{end}}

{{range $rcvr, $sigs := .Funcs}}
{{range $sig := $sigs}}func ({{$sig.RcvrName}} *{{$sig.RcvrType}}) {{$sig.Name}}({{$sig.ListArgs}}) {{$sig.ListReturns}} { {{$sig.ReturnStmt}} }
{{end}}
{{end}}
`
)

var (
	SourceFile string
	Data       *tree.CompilationUnit
)

func main() {
	flag.StringVar(&SourceFile, "source-file", "example.go", "the Golang file to parse")
	flag.Parse()
	//destFilePath := utils.BuildDest()

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
					Name:       fn.Name.Name,
					RcvrName:   rName,
					RcvrType:   utils.ToMock(rType),
					Args:       utils.FormatArgs(fn.Type.Params),
					Returns:    utils.FormatArgs(fn.Type.Results),
					ReturnStmt: utils.FormatRetStmt(fn.Type.Results),
				}

				unit.Funcs[rType] = append(unit.Funcs[rType], sig)
				utils.ExtractPkgPrefix(unit, rType)
			}
		}

		return true
	})

	out, err := utils.Render(unit, MockTemplate)
	if err != nil {
		panic(err)
	}

	fmt.Println(out)
	fmt.Println()
}
