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

	// if this compilation unit (file) contains struct decls, we print the
	// mock struct, API stubs, and public interface in a*_mock.go file
	if len(unit.DeclHere) > 0 {
		fmt.Printf("[DEBUG] imports in scope for %q: %+v", unit.Source, unit.Imports)
		if err := tmpl.Execute(&output, unit); err != nil {
			return output, fmt.Errorf("failed to resolve output string from template: %s", err)
		}
	}

	return output, nil
}

func IsPublicMethod(unit *tree.CompilationUnit, fn *ast.FuncDecl) bool {
	return len(fn.Name.Name) > 0 &&
		fn.Name.IsExported() &&
		fn.Recv != nil &&
		len(fn.Recv.List) > 0 &&
		len(fn.Recv.List[0].Names[0].Name) > 0
}

func ProcessFields(unit *tree.CompilationUnit, args *ast.FieldList) []tree.Field {
	fields := []tree.Field{}
	if args != nil {
		for _, f := range args.List {
			field := tree.NewField(f)
			fields = append(fields, field)
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
