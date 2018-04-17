package tree

import (
	"fmt"
	"go/ast"
	"strings"
)

type Field struct {
	Name string
	Type []string
}

func NewField(f *ast.Field) Field {
	rcvr := ""
	if len(f.Names) > 0 {
		rcvr = f.Names[0].Name
	}
	_, path := parseType(f.Type, []string{})

	return Field{
		Name: rcvr,
		Type: path,
	}
}

func (f Field) GetType() string {
	return strings.Join(f.Type, "")
}

func (f Field) Render() string {
	if len(f.Name) > 0 {
		return fmt.Sprintf("%s %s", f.Name, f.GetType())
	}

	return fmt.Sprintf("%s", f.GetType())
}

func (f Field) ToMock() string {
	return "Mock" + f.Type[len(f.Type)-1]
}

func (f Field) Ptr() string {
	var ndx int
	for ndx = 0; ndx < len(f.Type) && f.Type[ndx] == `*`; ndx++ {
	}

	return strings.Join(f.Type[:ndx], "")
}

func parseType(t interface{}, path []string) (interface{}, []string) {
	switch elem := t.(type) {
	case *ast.Ident:
		path = append(path, elem.Name)

	case *ast.StarExpr:
		path = append(path, `*`)
		t, path = parseType(elem.X, path)

	case *ast.SelectorExpr:
		t, path = parseType(elem.X, path)
		path = append(path, `.`)
		t, path = parseType(elem.Sel, path)

	case *ast.Ellipsis:
		path = append(path, `...`)
		t, path = parseType(elem.Elt, path)

	case *ast.InterfaceType:
		// TODO: deal with elem.Methods here?
		path = append(path, `interface{}`)

	case *ast.ArrayType:
		path = append(path, `[`)
		if elem.Len != nil {
			if lit, ok := elem.Len.(*ast.BasicLit); ok {
				path = append(path, lit.Value)
			}
		}
		path = append(path, `]`)
		t, path = parseType(elem.Elt, path)

	case *ast.MapType:
		path = append(path, `map`, `[`)
		t, path = parseType(elem.Key, path)
		path = append(path, `]`)
		t, path = parseType(elem.Value, path)

	case *ast.ChanType:
		if elem.Dir == 2 {
			path = append(path, `<-`)
		}
		path = append(path, `chan`)
		if elem.Dir == 1 {
			path = append(path, `<-`)
		}
		path = append(path, ` `)
		t, path = parseType(elem.Value, path)

	default:
		panic(fmt.Sprintf("unknown child of *ast.Type (%T) in traversal: %+v", elem, elem))
	}

	return t, path
}
