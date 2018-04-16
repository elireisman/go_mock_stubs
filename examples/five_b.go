package tmp

import "net/http"

// TODO: store imports at method-receiver scope so imports for
// struct methods defined across multiple files works :(
func (tr *Testr) Goofy(x *http.Request) string {
	return "goofy"
}
