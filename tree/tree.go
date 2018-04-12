package tree

import (
	"fmt"
	"go/ast"
	"path"
	"strings"
)

type CompilationUnit struct {
	Pkg      string
	Imports  []Import
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
		return s.Returns[0].Render()
	default:
		return "(" + strings.Join(s.GetDeclaredReturns(), ", ") + ")"
	}
}

func (s Signature) GetDeclaredReturns() []string {
	out := []string{}
	for _, f := range s.Returns {
		out = append(out, f.Render())
	}

	return out
}

// these are the computed zero values used in the block of methods stubbed out
// for the mock struct created for each public interface
func (s Signature) BuildReturnStmt() string {
	if len(s.Returns) == 0 {
		return "return"
	}

	rets := []string{}
	for _, f := range s.Returns {
		switch f.Value {
		case Nil: // map, array/slice, ellipsis, chan, pointer, interface, error
			rets = append(rets, "nil")
		case Zero:
			rets = append(rets, "0")
		case Empty:
			rets = append(rets, `""`)
		case False:
			rets = append(rets, "false")

		// TODO: check these!
		case Rune:
			rets = append(rets, "rune(0)")
		case Complex:
			rets = append(rets, "complex(0, 0)")
		case Struct:
			rets = append(rets, f.GetType()+`{}`)
		default:
			// TODO: freak out here!
			fmt.Printf("[DEBUG] failed to determine stubbable return type for: %s\n", f.GetType())
		}
	}

	return "return " + strings.Join(rets, ", ")
}

type Field struct {
	Name  string
	Value Return
	Type  []string
}

func NewField(f *ast.Field) Field {
	rcvr := ""
	if len(f.Names) > 0 {
		rcvr = f.Names[0].Name
	}
	_, path := ParseType(f.Type, []string{})
	retVal := ParseReturn(f.Type, Unknown)

	return Field{
		Name:  rcvr,
		Value: retVal,
		Type:  path,
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

// only ever used on receiver which will be "Type" or "pkg.Type"
func (f Field) ToMock() string {
	if len(f.Type) != 2 {
		panic(fmt.Sprintf("failed to render mock struct from orig type: %s", f.GetType()))
	}

	return "mock" + f.Type[1]
}

type Return uint8

const (
	Unknown Return = iota
	Nil
	Zero
	False
	Empty
	Rune
	Complex
	Byte
	Struct
)

func ParseReturn(t interface{}, current Return) Return {
	switch elem := t.(type) {

	case *ast.Ident:
		fmt.Printf("[DEBUG] *ast.Ident: %+v\n", elem)
		switch elem.Name {
		case "bool":
			return False
		case "rune":
			return Rune
		case "complex64", "complex128":
			return Complex
		case "byte":
			return Byte
		case "error":
			return Nil
		case "string":
			return Empty
		case "uint8", "uint16", "uint32", "uint64",
			"int8", "int16", "int32", "int64", "int",
			"uint", "uintptr", "float32", "float64":
			return Zero

		// TODO: handle wrapped primitive types like kakfa.Offset(0) ?

		default:
			return Struct
		}

	case *ast.StarExpr, *ast.Ellipsis, *ast.ArrayType,
		*ast.MapType, *ast.ChanType, *ast.InterfaceType:
		fmt.Printf("[DEBUG] %T: %+v\n", elem, elem)
		return Nil

	case *ast.SelectorExpr:
		fmt.Printf("[DEBUG] *ast.SelectorExpr: %+v\n", elem)
		return ParseReturn(elem.Sel, current)

	default:
		panic(fmt.Sprintf("unknown child of *ast.Type (%T) in traversal: %+v\n", elem, elem))
	}
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
			// TODO: handle fixed size array with t.Len field "[%d]" style
			fmt.Printf("[DEBUG] Array Size: %+v\n", elem.Len)
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
