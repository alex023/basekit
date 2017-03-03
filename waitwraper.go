package basekit

import (
	"sync"
)

// WaitWraper is a method wrapper based sync.Waitgroup
type WaitWraper struct {
	sync.WaitGroup
}

// Wrap encapsulates the method fn with sync.Waitgroup to ensure its execution.
func (w *WaitWraper) Wrap(fn func()) {
	w.Add(1)
	go func() {
		fn()
		w.Done()
	}()
}
