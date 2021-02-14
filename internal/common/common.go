package common

// Code used in both wasm and server
import "sync"

type State struct {
	A float64
	B float64
}

var SMutex sync.Mutex

func (sp *State) Get() *State {
	SMutex.Lock()
	defer SMutex.Unlock()
	safeCopy := *sp
	return &safeCopy
}

func (sp *State) DirectUpdate(f func(p *State)) {
	SMutex.Lock()
	defer SMutex.Unlock()
	f(sp)
}
