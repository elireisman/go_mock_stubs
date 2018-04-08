package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
)

const (
  GlobalScope = ""

  MockTemplate = `
package {{.Pkg}}

import (
  ## TODO ##
)

## TODO: PRINT STRUCTS FROM {{.Funcs}} MAP KEYS AS interface'S + FUNC LISTS ##

## TODO: PRINT FUNCS W/RECEIVERS AS NO-OP STUBS ##
`
)

var (
	SourceFile string
	Data       *CompilationUnit
)

type CompilationUnit struct {
	Pkg     string
	Imports []string
	Funcs   map[string][]Signature
}

type Signature struct {
	Name     string
	Receiver string
	Args     []string
	Returns  []string
}

func main() {
	flag.StringVar(&SourceFile, "source-file", "example.go", "the Golang file to parse")
	flag.Parse()
	//destFilePath := buildDest()

	fileSet := token.NewFileSet()
	node, err := parser.ParseFile(fileSet, SourceFile, nil, parser.ParseComments)
	if err != nil {
		panic(fmt.Sprintf("failed to parse source file %q into Golang AST: %s", err))
	}

	unit := &CompilationUnit{
		Pkg:     node.Name.Name,
		Imports: []string{},
		Funcs:   map[string][]Signature{},
	}

	for _, impt := range node.Imports {
		unit.Imports = append(unit.Imports, impt.Path.Value)
	}

	// TODO: find public struct & function defs
	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if len(fn.Name.Name) > 0 && fn.Name.IsExported() {
				rcvr := GlobalScope
				if len(fn.Recv.List) > 0 && len(fn.Recv.List[0].Names) > 0 {
					rcvr = fn.Recv.List[0].Names[0].Name
				}
				sig := Signature{
					Name:     fn.Name.Name,
					Receiver: rcvr,
					Args:     formatArgs(fn.Type.Params),
					Returns:  formatArgs(fn.Type.Results),
				}
				unit.Funcs[rcvr] = append(unit.Funcs[rcvr], sig)
			}
		}

		return true
	})

        fmt.Printf("CAPTURED AST: %+v\n", unit)
	fmt.Println()

        out, err := render(unit); err != nil {
          panic(err)
        }
        fmt.Println(out)
}

func render(unit *CompilationUnit) (string, error) {
	tmpl, err := template.New("mock").Parse(MockTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %s", err)
	}

	var output bytes.Buffer
	if err := tmpl.Execute(&output, *conf); err != nil {
		return "", fmt.Errorf("failed to resolve output string from template: %s", err)
	}

        // TODO: maybe output.Bytes() instead?
	return output.String(), nil
}

func formatArgs(args *ast.FieldList) []string {
	var out []string
	for _, f := range args.List {
		if len(f.Names) > 0 {
			out = append(out, fmt.Sprintf("%s %s", f.Names[0], f.Type))
		} else {
			out = append(out, fmt.Sprintf("%s", f.Type))
		}
	}

	return out
}

func buildDest() string {
	if len(SourceFile) == 0 || SourceFile[len(SourceFile)-3:] != ".go" {
		panic(fmt.Sprintf("illegal argument to --source, got: %q", SourceFile))
	}

	mockFile := filepath.Base(SourceFile)[:len(SourceFile)-3] + "_mock.go"
	return filepath.Dir(SourceFile) + mockFile
}
