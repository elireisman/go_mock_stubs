package utils

import (
	"bytes"
	"fmt"
	"go/ast"
	"path/filepath"
	"text/template"

	"github.com/elireisman/go_mock_stubs/tree"
)

func Render(unit *tree.CompilationUnit, mockTemplate string) (bytes.Buffer, error) {
	var output bytes.Buffer
	tmpl, err := template.New("mock").Parse(mockTemplate)
	if err != nil {
		return output, fmt.Errorf("failed to parse template: %s", err)
	}

	if err := tmpl.Execute(&output, unit); err != nil {
		return output, fmt.Errorf("failed to resolve output string from template: %s", err)
	}

	return output, nil
}

func ExtractPkgPrefix(unit *tree.CompilationUnit, path []string) {
	for ndx, elem := range path {
		if elem == `.` {
			unit.Prefixes[path[ndx-1]] = true
			return
		}
	}
}

func FormatArgs(unit *tree.CompilationUnit, args *ast.FieldList) []tree.Field {
	fields := []tree.Field{}
	if args != nil {
		for _, f := range args.List {
			field := tree.NewField(f)
			fields = append(fields, field)
			ExtractPkgPrefix(unit, field.Type)
		}
	}

	return fields
}

func BuildDest(sourceFile string) string {
	if len(sourceFile) == 0 || filepath.Ext(sourceFile) != ".go" {
		panic(fmt.Sprintf("illegal argument to --source, got: %q", sourceFile))
	}

	dir := filepath.Dir(sourceFile)
	base := filepath.Base(sourceFile)
	mockFile := base[:len(base)-3] + "_mock.go"

	return dir + "/" + mockFile
}
