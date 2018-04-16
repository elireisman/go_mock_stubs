package tree

import (
	"fmt"
	"go/ast"
	"path"
	"strings"
)

type CompilationUnit struct {
	Source   string
	Pkg      string
	Imports  []Import
	DeclHere map[string]bool
	Prefixes map[string]bool
	Funcs    map[string][]Signature
}

func (cu *CompilationUnit) FormatImports() string {
	found := map[Import]bool{}
	for _, imp := range cu.Imports {
		_, pkg := path.Split(imp.Path)
		if _, ok := cu.Prefixes[imp.Alias]; ok {
			found[imp] = true
		} else if _, ok := cu.Prefixes[pkg]; ok {
			found[imp] = true
		}
	}

	if len(found) == 0 {
		return ""
	}

	rendered := "import (\n"
	for foundImp := range found {
		rendered += "  " + foundImp.Format() + "\n"
	}
	rendered += ")\n"

	return rendered
}

func (cu *CompilationUnit) IsDeclaredHere(receiver string) bool {
	_, found := cu.DeclHere[receiver]
	return found
}

type Import struct {
	Alias string
	Path  string
}

func (i Import) Format() string {
	if len(i.Alias) > 0 {
		return i.Alias + ` "` + i.Path + `"`
	}
	return `"` + i.Path + `"`
}

type Signature struct {
	Name     string
	Receiver Field
	Args     []Field
	Returns  []Field
}

// the declared arguments for this method
func (s Signature) ListArgs() string {
	out := []string{}
	for _, arg := range s.Args {
		out = append(out, arg.Render())
	}

	return strings.Join(out, ", ")
}

// the declared return types for this method
func (s Signature) ListReturns() string {
	switch len(s.Returns) {
	case 0:
		return ""
	case 1:
		return " " + s.Returns[0].Render()
	default:
		return " (" + strings.Join(s.getDeclaredReturns(), ", ") + ")"
	}
}

func (s Signature) getDeclaredReturns() []string {
	out := []string{}
	for _, f := range s.Returns {
		out = append(out, f.Render())
	}

	return out
}

type Field struct {
	Name string
	Type []string
}

func NewField(f *ast.Field) Field {
	rcvr := ""
	if len(f.Names) > 0 {
		rcvr = f.Names[0].Name
	}
	_, path := ParseType(f.Type, []string{})

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

func ParseType(t interface{}, path []string) (interface{}, []string) {
	switch elem := t.(type) {
	case *ast.Ident:
		path = append(path, elem.Name)

	case *ast.StarExpr:
		path = append(path, `*`)
		t, path = ParseType(elem.X, path)

	case *ast.SelectorExpr:
		t, path = ParseType(elem.X, path)
		path = append(path, `.`)
		t, path = ParseType(elem.Sel, path)

	case *ast.Ellipsis:
		path = append(path, `...`)
		t, path = ParseType(elem.Elt, path)

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
		t, path = ParseType(elem.Elt, path)

	case *ast.MapType:
		path = append(path, `map`, `[`)
		t, path = ParseType(elem.Key, path)
		path = append(path, `]`)
		t, path = ParseType(elem.Value, path)

	case *ast.ChanType:
		if elem.Dir == 2 {
			path = append(path, `<-`)
		}
		path = append(path, `chan`)
		if elem.Dir == 1 {
			path = append(path, `<-`)
		}
		path = append(path, ` `)
		t, path = ParseType(elem.Value, path)

	default:
		panic(fmt.Sprintf("unknown child of *ast.Type (%T) in traversal: %+v", elem, elem))
	}

	return t, path
}
