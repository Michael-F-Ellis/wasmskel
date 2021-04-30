// Code used in both wasm and server

package common

import "sync"

var sMutex sync.Mutex

// Get returns a pointer to a copy of a State struct. Get is concurrency-safe.
func (sp *State) Get() *State {
	sMutex.Lock()
	defer sMutex.Unlock()
	safeCopy := *sp
	return &safeCopy
}

// Set updates a state struct by applying a user supplied function that modifies
// the struct. Set is concurrency safe.
func (sp *State) DirectUpdate(f func(p *State)) {
	sMutex.Lock()
	defer sMutex.Unlock()
	f(sp)
}
