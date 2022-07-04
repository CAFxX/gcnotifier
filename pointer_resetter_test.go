package gcnotifier

import (
	"runtime"
	"sync/atomic"
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

type ptr2[T any] struct {
	cnt int32
}

func (p *ptr2[T]) Store(v *T) {
	if v == nil {
		atomic.AddInt32(&p.cnt, 1)
	} else {
		panic(v)
	}
}

func TestPointerResetter(t *testing.T) {
	p := new(ptr2[int])
	Add[int](p)
	defer Remove[int](p)
	for i := 0; i < 100; i++ {
		runtime.GC()
	}
	cnt := atomic.LoadInt32(&p.cnt)
	if cnt < 99 || cnt > 101 {
		t.Fatal(cnt)
	}
}

func TestPanic(t *testing.T) {
	cases := map[string]struct {
		p string
		f func()
	}{
		"Add/nil":    {"nil pointer", func() { Add[int](nil) }},
		"Add/dup":    {"duplicated pointer", func() { p := new(ptr[int]); Add[int](p); Add[int](p) }},
		"Remove/nil": {"nil pointer", func() { Remove[int](nil) }},
		"Remove/unk": {"unknown pointer", func() { p := new(ptr[int]); Remove[int](p) }},
	}
	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if r != c.p {
						t.Fatalf("expected panic %q, got %q", c.p, r)
					}
				} else {
					t.Fatal("no panic")
				}
			}()
			c.f()
		})
	}
}
