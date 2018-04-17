package tree

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"

	"github.com/elireisman/go_mock_stubs/utils"
)

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

func (cu *CompilationUnit) Render(targets []string) (bytes.Buffer, error) {
	// if the --targets CSV list is populated, filter for those declarations
	cu.filterTargets(targets)

	// if this compilation unit (file) contains struct decls, we print the
	// mock struct, API stubs, and public interface in a*_mock.go file
	var output bytes.Buffer
	if len(cu.DeclHere) > 0 {
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

// side effect onto this compilation unit the pruning
// of all unit.DeclHere's not in the targets CSV list
func (cu *CompilationUnit) filterTargets(targets []string) {
	if targets != nil {
		removals := []string{}
		for decl := range cu.DeclHere {
			declared := cu.Pkg.Name + `.` + decl
			found := false
			for _, target := range targets {
				if target == declared {
					found = true
					break
				}
			}
			if !found {
				removals = append(removals, decl)
			}
		}
		for _, r := range removals {
			delete(cu.DeclHere, r)
		}
	}
}
