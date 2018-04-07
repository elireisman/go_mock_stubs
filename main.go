package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
)

var (
	SourceFile string
)

func init() {
	flag.StringVar(&SourceFile, "source-file", "example.go", "the Golang file to parse")
}

func main() {
	flag.Parse()
	destFile := buildDest()

	fileSet := NewFileSet()
	node, err := parser.ParseFile(fset, SourceFile, nil, parser.ParseComments)
	if err != nil {
		panic(fmt.Sprintf("failed to parse source file %q into Golang AST: %s", err))
	}

	funcs := []*ast.FuncDecl{}
	ast.Inspect(node, func(n ast.Node) bool {
		switch elem := n.(type) {
		case *ast.FuncDecl:
			panic(fmt.Sprintf("%#v", *elem))
			funcs := append(funcs, elem)

		default:
			return true
		}
	})
}

func buildDest() string {
	if len(SourceFile) == 0 || SourceFile[len(SourceFile-3):] != ".go" {
		panic(fmt.Sprintf("illegal argument to --source, got: %q", SourceFile))
	}

	mockFile := Base(SourceFile)[:len(SourceFile-3)] + "_mock.go"
	return Dir(SourceFile) + mockFile
}
