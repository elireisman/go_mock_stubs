package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
)

var (
	SourceFile string
)

func init() {
	flag.StringVar(&SourceFile, "source-file", "example.go", "the Golang file to parse")
}

func main() {
	flag.Parse()
	//destFile := buildDest()

	fileSet := token.NewFileSet()
	node, err := parser.ParseFile(fileSet, SourceFile, nil, parser.ParseComments)
	if err != nil {
		panic(fmt.Sprintf("failed to parse source file %q into Golang AST: %s", err))
	}

	//pkg := node.Name // package name of file

	// TODO: find public struct & function defs
	funcs := []*ast.FuncDecl{}
	ast.Inspect(node, func(n ast.Node) bool {

		if fn, ok := n.(*ast.FuncDecl); ok {
			if len(fn.Name.Name) > 0 && fn.Name.IsExported() {
				funcs = append(funcs, fn)
			}
		}

		return true
	})

	for _, fn := range funcs {
		fmt.Printf("%#v", *fn)
	}
	fmt.Println()
}

func buildDest() string {
	if len(SourceFile) == 0 || SourceFile[len(SourceFile)-3:] != ".go" {
		panic(fmt.Sprintf("illegal argument to --source, got: %q", SourceFile))
	}

	mockFile := filepath.Base(SourceFile)[:len(SourceFile)-3] + "_mock.go"
	return filepath.Dir(SourceFile) + mockFile
}
