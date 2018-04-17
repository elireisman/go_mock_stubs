package tree

import (
	"bytes"
	"fmt"
	"go/ast"
	"path"
	"path/filepath"
	"strings"

	"github.com/elireisman/go_mock_stubs/utils"
)

type Package struct {
	Name    string
	Methods map[string][]Signature
	Imports map[Import]bool
}

type CompilationUnit struct {
	// input file this struct was populated from
	Source string

	// the parent package for this compilation unit
	Pkg *Package

	// set of all public structs declared in this file
	// avoids redfining mock structs when methods are
	// declared across multiple files
	DeclHere map[string]bool
}

func (cu *CompilationUnit) Render() (bytes.Buffer, error) {
	var output bytes.Buffer

	// if this compilation unit (file) contains struct decls, we print the
	// mock struct, API stubs, and public interface in a*_mock.go file
	if len(cu.DeclHere) > 0 {
		//fmt.Printf("[DEBUG] imports in scope for %q: %+v", cu.Source, cu.Imports)
		if err := utils.Compiled.Execute(&output, cu); err != nil {
			return output, fmt.Errorf("failed to resolve output string from template: %s", err)
		}
	}

	return output, nil
}

func (cu *CompilationUnit) FormatImports() string {
	found := map[Import]bool{}
	prefixes := cu.extractPrefixes()

	// TODO: there are still corner cases we don't handle but they are obscure.
	// ex: multiple files where struct's methods are defined include same import
	// aliased several different ways...
	for imp := range cu.Pkg.Imports {
		_, pkg := path.Split(imp.Path)
		if _, ok := prefixes[imp.Alias]; ok {
			found[imp] = true
		} else if _, ok := prefixes[pkg]; ok {
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

func (cu *CompilationUnit) Dest() string {
	if len(cu.Source) == 0 || filepath.Ext(cu.Source) != ".go" {
		panic(fmt.Sprintf("illegal argument to --source, got: %q", cu.Source))
	}

	dir := filepath.Dir(cu.Source)
	base := filepath.Base(cu.Source)
	mockFile := base[:len(base)-3] + "_mock.go"

	return dir + "/" + mockFile
}

// expensive, but we want to get this right, so KISS
func (cu *CompilationUnit) extractPrefixes() map[string]bool {
	out := map[string]bool{}
	for decl := range cu.DeclHere {
		for rcvr, sigs := range cu.Pkg.Methods {
			if decl == rcvr {
				for _, sig := range sigs {
					for _, field := range sig.Args {
						cu.extractPkg(out, field.Type)
					}
					for _, field := range sig.Returns {
						cu.extractPkg(out, field.Type)
					}
				}
			}
		}
	}

	return out
}

func (cu *CompilationUnit) extractPkg(out map[string]bool, path []string) {
	for ndx, elem := range path {
		if elem == `.` {
			out[path[ndx-1]] = true
		}
	}
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
		if s.Returns[0].Name != "" {
			return " (" + s.Returns[0].Render() + ")"
		} else {
			return " " + s.Returns[0].Render()
		}

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

func (s *Signature) ProcessArgs(args *ast.FieldList) {
	s.Args = s.processFields(args)
}

func (s *Signature) ProcessReturns(rets *ast.FieldList) {
	s.Returns = s.processFields(rets)
}

func (s *Signature) processFields(args *ast.FieldList) []Field {
	fields := []Field{}
	if args != nil {
		for _, f := range args.List {
			field := NewField(f)
			fields = append(fields, field)
		}
	}

	return fields
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
