package example

import (
	"fmt"
	"log"
	zzz "strings" // not needed for mocks
	"sync"        // not needed
	"sync/atomic" // not needed for mocks
)

type privateSoNotMocked struct {
	Nothing bool
}

type Person struct {
	log  *log.Logger
	lock sync.Mutex
	name string
	age  uint32
}

func (p *Person) ShowName(other string, maybe bool) error {
	return zzz.Join([]string{p.name, other}, ",")
}

func (p *Person) GuessAge(tmpl fmt.Stringer, guesses map[string]uint32, hitEvent chan<- bool) ([][]byte, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for name, guess := range guesses {
		if p.age == guess {
			log.Printf("matched %q", name)
			hitEvent <- true
		}
	}

	return nil, fmt.Errorf("ouch: %+v", guesses)
}

func (p *Person) GetOlder(triplePtr ***uint32) (*log.Logger, *uint32, <-chan bool) {
	result := atomic.AddUint32(&p.age, 1)
	result = atomic.AddUint32((**triplePtr), result)

	return p.log, &result, make(<-chan bool)
}

func (p *Person) notBeingMocked() {
	// no-op
}
