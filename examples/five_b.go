package examples

import "net/http"

// TODO: store imports at method-receiver scope so imports for
// struct methods defined across multiple files works :(
func (mul *Multi) Goofy(r *http.Request) string {
	return "goofy"
}
