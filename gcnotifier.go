// Package gcnotifier provides a way to receive notifications after every time
// garbage collection (GC) runs. This can be useful to instruct your code to
// free additional memory resources that you may be using. To minimize the load
// on the GC the code that runs after receiving the notification should try to
// avoid allocations as much as possible.
package gcnotifier

import "runtime"

type sentinel struct {
	gcCh chan struct{}
}

// AfterGC returns a channel that will receive a notification after every GC
// run. If a notification is not consumed before another GC runs only one of the
// two notifications is sent. To stop the notifications you can safely close the
// channel.
func AfterGC() <-chan struct{} {
	s := &sentinel{gcCh: make(chan struct{})}
	runtime.SetFinalizer(s, finalizer)
	return s.gcCh
}

func finalizer(obj interface{}) {
	defer recover() // writing to a closed channel will panic
	s := obj.(*sentinel)
	select {
	case s.gcCh <- struct{}{}:
	default:
	}
	runtime.SetFinalizer(s, finalizer)
}
