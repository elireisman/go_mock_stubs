package utils

import (
	"bytes"
	"fmt"
	"go/ast"
	"path/filepath"
	"strings"
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

func ToMock(t string) string {
	parts := strings.Split(t, `.`)
	last := len(parts) - 1
	sep := ""
	if last > 0 {
		sep = `.`
	}

	return strings.Join(parts[:last], `.`) + sep + "mock" + parts[last]
}

func ExtractPkgPrefix(unit *tree.CompilationUnit, t string) {
	parts := strings.Split(t, `.`)
	if len(parts) > 0 && parts[0] != "" {
		unit.Prefixes[parts[0]] = true
	}
}

func FormatRetStmt(args *ast.FieldList) string {
	if args == nil {
		return "return"
	}

	rets := []string{}
	for _, f := range args.List {
		rType := ParseType(f.Type)
		switch rType {
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"float32", "float64":
			rets = append(rets, "0")
		case "string":
			rets = append(rets, `""`)
		case "bool":
			rets = append(rets, "false")
		case "rune":
			rets = append(rets, "rune(0)")
		case "complex64", "complex128":
			rets = append(rets, "complex(0, 0)")
		case "error":
			rets = append(rets, "nil")
		default:
			if strings.HasPrefix(rType, "map") || strings.HasPrefix(rType, "[") || strings.HasPrefix(rType, "chan") {
				// map, chan or array/slice type
				rets = append(rets, "nil")
			} else if strings.HasPrefix(rType, `*`) || strings.HasPrefix(rType, `...`) {
				// it's a pointer type
				rets = append(rets, "nil")
			} else if rType == `interface{}` {
				rets = append(rets, `interface{}`)
			} else {
				// OK, let's assume it's a struct (...waves hands...)
				rets = append(rets, rType+"{}")
			}
		}
	}

	return "return " + strings.Join(rets, ", ")
}

func ParseType(t interface{}) string {
	_, path := WalkTypePath(t, []string{})
	ret := ""

	// TODO: recurse here, skip dot-paths for *, ..., and [] (maybe more prefixes?) for multi-dim or double ptr etc.
	if len(path) > 0 {
		if path[0] == `*` {
			ret = `*` + strings.Join(path[1:], `.`)
		} else if path[0] == `...` {
			ret = `...` + strings.Join(path[1:], `.`)
		} else if path[0] == `[]` {
			ret = `[]` + strings.Join(path[1:], `.`)

		} else {
			ret = strings.Join(path, `.`)
		}
	}

	return ret
}

func WalkTypePath(t interface{}, path []string) (interface{}, []string) {
	switch elem := t.(type) {
	case *ast.Ident:
		path = append(path, elem.Name)

	case *ast.StarExpr:
		path = append(path, `*`)
		t, path = WalkTypePath(elem.X, path)

	case *ast.SelectorExpr:
		t, path = WalkTypePath(elem.X, path)
		t, path = WalkTypePath(elem.Sel, path)

	case *ast.Ellipsis:
		path = append(path, `...`)
		t, path = WalkTypePath(elem.Elt, path)

	case *ast.InterfaceType:
		path = append(path, `interface{}`)

	case *ast.ArrayType:
		// TODO: handle fixed size array with t.Len field "[%d]" style
		path = append(path, `[]`)
		t, path = WalkTypePath(elem.Elt, path)

		//case *ast.MapType:

		//case *ast.ChanType:

	default:
		panic(fmt.Sprintf("unknown child of *ast.Type (%T) in traversal: %+v", elem, elem))
	}

	return t, path
}

func FormatArgs(unit *tree.CompilationUnit, args *ast.FieldList) []string {
	out := []string{}
	if args != nil {
		for _, f := range args.List {
			found := ParseType(f.Type)

			split := strings.Split(found, `.`)
			if len(split) > 1 {
				ExtractPkgPrefix(unit, split[0])
			}

			if len(f.Names) > 0 {
				out = append(out, fmt.Sprintf("%s %s", f.Names[0], found))
			} else {
				out = append(out, fmt.Sprintf("%s", found))
			}
		}
	}

	return out
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
