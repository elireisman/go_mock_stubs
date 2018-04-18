package examples

import (
	"fmt" // should be stripped from output
	"net/http"
)

// should not be present in output
type Kiwi interface{}

// MockBanana + BananaIface should be generated in output
type Banana struct {
	Tasty bool
	Name  string
}

// methods with struct receiver are captured for Banana
func (b Banana) IsTasty() bool {
	return b.Tasty
}

// methods with pointer receivers are captured for Banana
func (b *Banana) Handler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, fmt.Sprintf("We have no %s today", b.Name), http.StatusPaymentRequired)
}
