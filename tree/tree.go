package tree

import (
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
