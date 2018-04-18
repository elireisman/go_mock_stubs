package examples

import (
	"fmt"
	"log"          // this should be stripped as it doesn't show up in a method signature
	abc "net/http" // this should be included with its alias
	xyz "strconv"  //  this should be stripped, not used in a method signature
)

// exercises import aliases and import stripping in general

// don't include public interfaces in generated output, just structs
type AlreadyAbstracted interface {
	DoSomething(string, int64, float64) error
}

// don't include private structs or interfaces in output
type skippedForGoodReasons struct {
	vec []float64
}

type alsoSkipThis interface {
	SomethingElse(bool) error
	GetLogger() *log.Logger
}

// this and it's methods should be mocked in output
type Implemented struct {
	info    string
	version uint
	skip    alsoSkipThis
}

func (i Implemented) String() string {
	i.skip.GetLogger().Printf("%s", i.info)
	return i.info
}

// pkg aliased to xyz will be stripped from output it doesn't show up in method sig
func (i *Implemented) Build(input string) (*abc.Request, error) {
	num, _ := xyz.Atoi(input)
	fmt.Println(num)
	return nil, nil
}

// should be stripped from output
func (i *Implemented) privateDontMock(x chan map[string]int, w abc.ResponseWriter) error {
	return nil
}
