package examples

import (
	"fmt"
	"net/http"
)

func (mul *MultiFileDef) Goofy(r *http.Request) string {
	return "goofy"
}

func (mul *MultiFileDef) Foobarz(grid [20][20]int32, printable fmt.Stringer) (e error) {
	return
}

func (mul MultiFileDef) multiNotForMocking() {}
