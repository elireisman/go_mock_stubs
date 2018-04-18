package examples

// exercise multiple mockable methods defined across several files for same struct

type MultiFileDef struct {
	omg bool
}

func (mul MultiFileDef) OMG() bool {
	return mul.omg
}
