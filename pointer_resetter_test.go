package gcnotifier

import (
	"runtime"
	"testing"
)

type ptr[T any] struct {
	ch chan struct{}
}

func (p *ptr[T]) Store(v *T) {
	if v == nil {
		close(p.ch)
	} else {
		panic(v)
	}
}

func TestPointerResetters(t *testing.T) {
	for i := 0; i < 1000; i++ {
		p := new(ptr[int])
		p.ch = make(chan struct{})
		Add[int](p)
		select {
		case <-p.ch:
			t.Fatal("closed early")
		default:
		}
		runtime.GC()
		<-p.ch
		Remove[int](p)
	}
}
