package common

// Code used in both wasm and server
import "sync"

type MonitoredParameters struct {
	A float64
	B float64
}

var MPMutex sync.Mutex

func (mp *MonitoredParameters) Get() *MonitoredParameters {
	MPMutex.Lock()
	defer MPMutex.Unlock()
	safeCopy := *mp
	return &safeCopy
}

func (mp *MonitoredParameters) DirectUpdate(f func(p *MonitoredParameters)) {
	MPMutex.Lock()
	defer MPMutex.Unlock()
	f(mp)
}
