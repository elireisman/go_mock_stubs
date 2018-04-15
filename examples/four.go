package examples

import (
	"fmt"
	"time"
)

type Apple struct {
	batchID uint32
	sellBy  time.Time
}

func (a Apple) String() string {
	return fmt.Sprintf("Apple(%d)", a.batchID)
}

func (a *Apple) CanSell() bool {
	return time.Now().After(a.sellBy)
}

type Orange struct {
	juicy bool
}

func (o *Orange) Eat() (say string) {
	if o.juicy {
		say = "mmm good!"
	}
	return
}

type OtherThing interface {
	GoForIt(bool)
}

type shouldBeSkipped struct{}
