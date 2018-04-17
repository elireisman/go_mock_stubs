package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/elireisman/go_mock_stubs/tree"
	"github.com/elireisman/go_mock_stubs/utils"
)

var (
	SourceDir   string
	StdOut      bool
	Targets     []string
	rawTargets  string
	MultiLineWS *regexp.Regexp
)

func init() {
	flag.StringVar(&SourceDir, "source-dir", "example", "this dir will be recursively searched for Golang source files to parse")
	flag.StringVar(&rawTargets, "targets", "", "CSV list of form `pkg.Type,pkg2.Type2,...` - only write output mock files defining these types")
	flag.BoolVar(&StdOut, "stdout", false, "stream output code to stdout rather than writing to *_mock.go file under the source's dir")

	MultiLineWS = regexp.MustCompile(`\r?\n\r?\n(\r?\n)+`)
}

func main() {
	flag.Parse()

	if rawTargets != "" {
		ts := strings.Split(rawTargets, `,`)
		for _, t := range ts {
			if t != "" {
				Targets = append(Targets, t)
			}
		}
	}

	fileSet := token.NewFileSet()
	inPkgs, err := parser.ParseDir(fileSet, SourceDir, nil, parser.ParseComments)
	if err != nil {
		panic(fmt.Sprintf("failed to parse source files in %q into Golang AST: %s", SourceDir, err))
	}

	// maintain global mapping of file imports and method defs per package.
	// this context is needed when generating output files for our target structs
	pkgDefs := map[string]*tree.Package{}

	// maintain mapping of filename to file contents for simplicity when rendering output
	outFiles := map[string]tree.CompilationUnit{}

	// iterate over all files in each package, extracting info we need to generate mock files
	for _, pkg := range inPkgs {

		pkgName := pkg.Name
		if _, ok := pkgDefs[pkgName]; !ok {
			pkgDefs[pkgName] = &tree.Package{
				Name:    pkgName,
				Methods: map[string][]tree.Signature{},
				Imports: map[tree.Import]bool{},
			}
		}

		for fileName, node := range pkg.Files {

			unit := tree.CompilationUnit{
				Pkg:      pkgDefs[pkgName],
				Source:   fileName,
				DeclHere: map[string]bool{},
			}

			for _, impt := range node.Imports {
				alias := ""
				if impt.Name != nil {
					alias = impt.Name.Name
				}
				path, _ := strconv.Unquote(impt.Path.Value)
				next := tree.Import{Alias: alias, Path: path}
				unit.Pkg.Imports[next] = true
			}

			// extract metadata from any methods with a pointer to a struct as a receiver
			ast.Inspect(node, func(n ast.Node) bool {
				if ts, ok := n.(*ast.TypeSpec); ok {
					// collect mapping of public structs declared in THIS file
					if _, ok := ts.Type.(*ast.StructType); ok {
						if ts.Name != nil && ts.Name.IsExported() {
							unit.DeclHere[ts.Name.Name] = true
						}
					}
				} else if fn, ok := n.(*ast.FuncDecl); ok {
					// collect all public struct method decls across files in pkg
					if utils.IsPublicMethod(fn) {
						sig := tree.Signature{
							Name:     fn.Name.Name,
							Receiver: tree.NewField(fn.Recv.List[0]),
						}
						sig.ProcessArgs(fn.Type.Params)
						sig.ProcessReturns(fn.Type.Results)

						rcvrType := sig.Receiver.Type[len(sig.Receiver.Type)-1]
						unit.Pkg.Methods[rcvrType] = append(unit.Pkg.Methods[rcvrType], sig)
					}
				}

				return true
			})

			// record this compilation unit (file) in the output mapping
			outFiles[fileName] = unit
		}
	}

	// now that we've collected all the state across all the packages,
	// we have the complete picture and can render the output files
	for _, unit := range outFiles {
		raw, err := unit.Render(Targets)
		if err != nil {
			panic(err)
		}
		// empty output buffer means don't print the file, only print
		// mock if source file contains struct declarations we're mocking
		if len(raw.Bytes()) == 0 {
			continue
		}

		// TODO: something less lazy here
		out := strings.Trim(MultiLineWS.ReplaceAllString(raw.String(), "\n\n"), " \t\r\n")
		destFilePath := unit.Dest()

		if StdOut {
			fmt.Println()
			fmt.Printf("\x1b[1m[.] dry run for output file: %q\x1b[0m\n", destFilePath)
			fmt.Println(out)
		} else {
			fmt.Printf("\x1b[1m[.] writing output file to: %q\x1b[0m\n", destFilePath)
			if ioutil.WriteFile(destFilePath, []byte(out), os.FileMode(0664)); err != nil {
				panic(fmt.Sprintf("failed to write output to %q, error: %s", destFilePath, err))
			}
		}
	}
	fmt.Println("\n\x1b[1m[.] code generation complete\x1b[0m\n")
}
