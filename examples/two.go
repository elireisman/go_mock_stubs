package examples

import (
	"fmt"
	"log"
	"sync"        // not needed
	"sync/atomic" // not needed for mocks
)

// a bunch of random methods to exercise generator for:
// 1. a wider range of data types
// 2. removal of imports that aren't present in method signatures

type privateSoNotMocked struct {
	Nothing bool
}

type Person struct {
	log  *log.Logger
	lock sync.Mutex
	name string
	age  uint32
}

func (p *Person) GuessAge(tmpl fmt.Stringer, guesses map[string]uint32, hitEvent chan<- bool) ([50][20]byte, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for name, guess := range guesses {
		if p.age == guess {
			log.Printf("matched %q", name)
			hitEvent <- true
		}
	}

	return [50][20]byte{}, fmt.Errorf("ouch: %+v", guesses)
}

func (p *Person) GetOlder(triplePtr ***uint32) (*log.Logger, *uint32, <-chan bool) {
	result := atomic.AddUint32(&p.age, 1)
	result = atomic.AddUint32((**triplePtr), result)

	return p.log, &result, make(<-chan bool)
}

// private methods of Person should be stripped from output
func (p *Person) notBeingMocked() {}
