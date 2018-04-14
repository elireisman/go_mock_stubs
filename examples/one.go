package examples

import (
	"fmt"
	"net/http"
)

type Fruit struct {
	Tasty bool
	Name  string
}

func (f *Fruit) Handler(w http.ResponseWriter, r *http.Rrequest) {
	return http.Error(w, fmt.Sprintf("Sorry, tasty %s costs money!", f.Name), http.PaymentRequired)
}
