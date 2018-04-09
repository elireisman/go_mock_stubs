package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"path/filepath"
	"strings"
)

const (
	GlobalScope = ""

	MockTemplate = `
package {{.Pkg}}

{{.FormatImports}}

{{range $rcvr, $sigs := .Funcs}}
type {{$x := index $sigs 0}}{{$x.RcvrType}} struct { }

type {{$rcvr}} interface {
{{range $sig := $sigs}}  {{$sig.Name}}({{$sig.ListArgs}}) {{$sig.ListReturns}}
{{end}}}

{{end}}

{{range $rcvr, $sigs := .Funcs}}
{{range $sig := $sigs}}func ({{$sig.RcvrName}} *{{$sig.RcvrType}}) {{$sig.Name}}({{$sig.ListArgs}}) {{$sig.ListReturns}} { {{$sig.ReturnStmt}} }
{{end}}
{{end}}
`
)

var (
	SourceFile string
	Data       *CompilationUnit
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
		parts := strings.Split(imp.Path, `.`)
		pkg := parts[len(parts)-1]
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
	Name       string
	RcvrName   string
	RcvrType   string
	Args       []string
	Returns    []string
	ReturnStmt string
}

func (s Signature) ListArgs() string {
	return strings.Join(s.Args, ", ")
}

func (s Signature) ListReturns() string {
	switch len(s.Returns) {
	case 0:
		return ""
	case 1:
		return s.Returns[0]
	default:
		return "(" + strings.Join(s.Returns, ", ") + ")"
	}
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
		Pkg:      node.Name.Name,
		Imports:  []Import{},
		Prefixes: map[string]bool{},
		Funcs:    map[string][]Signature{},
	}

	for _, impt := range node.Imports {
		alias := ""
		if impt.Name != nil {
			alias = impt.Name.Name
		}
		path := impt.Path.Value[1 : len(impt.Path.Value)-1]
		next := Import{Alias: alias, Path: path}
		unit.Imports = append(unit.Imports, next)
	}

	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if len(fn.Name.Name) > 0 && fn.Name.IsExported() {
				rName, rType := GlobalScope, GlobalScope
				if len(fn.Recv.List) > 0 {
					if ptrExpr, ok := fn.Recv.List[0].Type.(*ast.StarExpr); ok {
						astID, _ := ptrExpr.X.(*ast.Ident)
						rName, rType = fn.Recv.List[0].Names[0].Name, astID.Name
					}
				}
				sig := Signature{
					Name:       fn.Name.Name,
					RcvrName:   rName,
					RcvrType:   toMock(rType),
					Args:       formatArgs(fn.Type.Params),
					Returns:    formatArgs(fn.Type.Results),
					ReturnStmt: formatRetStmt(fn.Type.Results),
				}
				unit.Funcs[rType] = append(unit.Funcs[rType], sig)
				extractPkgPrefix(unit, rType)
			}
		}

		return true
	})

	out, err := render(unit)
	if err != nil {
		panic(err)
	}

	fmt.Println(out)
	fmt.Println()
}

func render(unit *CompilationUnit) (string, error) {
	tmpl, err := template.New("mock").Parse(MockTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %s", err)
	}

	var output bytes.Buffer
	if err := tmpl.Execute(&output, unit); err != nil {
		return "", fmt.Errorf("failed to resolve output string from template: %s", err)
	}

	// TODO: maybe output.Bytes() instead?
	return output.String(), nil
}

func toMock(t string) string {
	parts := strings.Split(t, `.`)
	last := len(parts) - 1
	sep := ""
	if last > 0 {
		sep = `.`
	}

	return strings.Join(parts[:last], `.`) + sep + "mock" + parts[last]
}

func extractPkgPrefix(unit *CompilationUnit, t string) {
	parts := strings.Split(t, `.`)
	if len(parts) > 0 {
		unit.Prefixes[parts[0]] = true
	}
}

func formatRetStmt(args *ast.FieldList) string {
	rets := []string{}
	for _, f := range args.List {
		rType := parseType(f.Type)
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
			if strings.Contains(rType, "map") || strings.Contains(rType, "[") {
				rets = append(rets, "nil")
			} else if _, ok := f.Type.(*ast.StarExpr); ok {
				// it's a pointer type
				rets = append(rets, "nil")
			} else {
				// OK, let's assume from here its a map, slice/array, or struct (...waves hands...)
				rets = append(rets, rType+"{}")
			}
		}
	}

	return "return " + strings.Join(rets, ", ")
}

func parseType(t interface{}) string {
	_, path := walkTypePath(t, []string{})
	ret := ""
	if path[0] == `*` {
		ret = "*" + strings.Join(path[1:], `.`)
	} else {
		ret = strings.Join(path, `.`)
	}
	return ret
}

func walkTypePath(t interface{}, path []string) (interface{}, []string) {
	switch elem := t.(type) {
	case *ast.Ident:
		path = append(path, elem.Name)

	case *ast.StarExpr:
		path = append(path, `*`)
		t, path = walkTypePath(elem.X, path)

	case *ast.SelectorExpr:
		t, path = walkTypePath(elem.X, path)
		t, path = walkTypePath(elem.Sel, path)

	default:
		panic(fmt.Sprintf("unknown child of *ast.Type (%T) in traversal: %+v", elem, elem))
	}

	return t, path
}

func formatArgs(args *ast.FieldList) []string {
	var out []string
	for _, f := range args.List {
		found := parseType(f.Type)
		// TODO: handle complex non-ptr types too!
		if len(f.Names) > 0 {
			out = append(out, fmt.Sprintf("%s %s", f.Names[0], found))
		} else {
			out = append(out, fmt.Sprintf("%s", found))
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
