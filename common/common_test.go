package common

import (
	"testing"

	"github.com/go-test/deep"
)

func TestGet(t *testing.T) {
	mp := State{Alpha: 1, Beta: 2, Gamma: 3}
	mpCopy := *mp.Get()
	if diff := deep.Equal(mp, mpCopy); diff != nil {
		t.Errorf("%v", diff)
	}
	mp.Alpha += 1 // change orginal
	if mpCopy.Alpha != 1 {
		t.Errorf("expected mpCopy to be unchanged, but mpCopy.A is now %f", mpCopy.Alpha)
	}
}

func TestDirectUpdate(t *testing.T) {
	mp := State{Alpha: 1, Beta: 2, Gamma: 3}
	f := func(p *State) {
		p.Beta = 13
	}
	mp.DirectUpdate(f)
	p := mp.Get()
	if diff := deep.Equal(mp, p); diff != nil {
		t.Errorf("%v", diff)
	}

}
