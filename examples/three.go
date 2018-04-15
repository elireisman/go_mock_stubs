package examples

import (
	"io"
	"time"
)

type Thing struct {
	name    string
	created time.Time
}

func (t *Thing) Since() time.Duration {
	return time.Since(t.created)
}

func (t *thing) betterNotImplementMe(shouldSkip true) bool {
	return shouldSkip
}

func (t *Thing) GetWriter() io.Writer {
	return nil
}
