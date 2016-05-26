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
	// we get here only if the channel was not closed
	runtime.SetFinalizer(s, finalizer)
}

// AfterGCUntilReachable is like AfterGC, but the channel will be closed when
// the object supplied as argument is garbage collected. No finalizer should be
// set on the object before or after calling this method. Pay attention to not
// inadvertently keep the object alive (e.g. by referencing it in a callback or
// goroutine) or the object may never be collected.
func AfterGCUntilCollected(obj interface{}) <-chan struct {
  gcCh := AfterGC()
	runtime.SetFinalizer(obj, func() { close(goCh) })
	return goCh
}
