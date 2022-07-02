package gcnotifier

import (
	"sync"
)

type ptrrst struct {
	mu  sync.Mutex
	m   map[interface{}]func()
	gcn *GCNotifier
}

func (p *ptrrst) run(n <-chan struct{}) {
	for range n {
		p.mu.Lock()
		for _, rst := range p.m {
			rst()
		}
		p.mu.Unlock()
	}
}

var gpr ptrrst

type atomicPointer[T any] interface {
	Store(*T)
}

func Add[T any](ptr atomicPointer[T]) {
	if ptr == nil {
		panic("nil pointer")
	}

	gpr.mu.Lock()
	defer gpr.mu.Unlock()

	if gpr.gcn == nil {
		gpr.gcn = New()
		go gpr.run(gpr.gcn.AfterGC())

		gpr.m = make(map[interface{}]func())
	}

	if _, exists := gpr.m[ptr]; exists {
		panic("pointer already added")
	}

	gpr.m[ptr] = func() { ptr.Store(nil) }
}

func Remove[T any](ptr atomicPointer[T]) {
	if ptr == nil {
		panic("nil pointer")
	}

	gpr.mu.Lock()
	defer gpr.mu.Unlock()

	if _, exists := gpr.m[ptr]; !exists {
		panic("unknown pointer")
	}

	delete(gpr.m, ptr)

	if len(gpr.m) == 0 {
		gpr.gcn.Close() // will cause gpr.run() to return
		gpr.gcn = nil

		gpr.m = nil
	}
}