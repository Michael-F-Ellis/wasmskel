package common

import (
	"testing"

	"github.com/go-test/deep"
)

func TestGet(t *testing.T) {
	mp := State{1, 2}
	mpCopy := *mp.Get()
	if diff := deep.Equal(mp, mpCopy); diff != nil {
		t.Errorf("%v", diff)
	}
	mp.A += 1 // change orginal
	if mpCopy.A != 1 {
		t.Errorf("expected mpCopy to be unchanged, but mpCopy.A is now %f", mpCopy.A)
	}
}

func TestDirectUpdate(t *testing.T) {
	mp := &State{1, 2}
	f := func(p *State) {
		p.B = 13
	}
	mp.DirectUpdate(f)
	p := mp.Get()
	if diff := deep.Equal(mp, p); diff != nil {
		t.Errorf("%v", diff)
	}

}
