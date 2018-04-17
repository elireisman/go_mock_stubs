package tree

type Package struct {
	Name    string
	Methods map[string][]Signature
	Imports map[Import]bool
}
