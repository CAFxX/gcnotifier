// Package gcnotifier provides a way to receive notifications after every time
// garbage collection (GC) runs. This can be useful to instruct your code to
// free additional memory resources that you may be using.
//
// A common use case for this is when you have a custom pool of objects: instead
// of setting a maximum size to your pool you can leave it unbounded and then
// drop all (or some) of them after every GC run (e.g. sync.Pool drops all
// objects in the pool during GC).
//
// To minimize the load on the GC the code that runs after receiving the
// notification should try to avoid allocations as much as possible, or at the
// very least make sure that the amount of new memory allocated is significantly
// smaller than the amount of memory that has been "freed" by your code.
package gcnotifier

import "runtime"

// GCNotifier allows your code to control and receive notifications every time
// the garbage collector runs.
type GCNotifier struct {
	gcCh chan struct{}
}

// AfterGC returns a channel that will receive a notification after every GC
// run. If a notification is not consumed before another GC runs only one of the
// two notifications is sent. To stop the notifications call the Close() method.
func (n *GCNotifier) AfterGC() <-chan struct{} {
	return n.gcCh
}

// Close will stop and release all resources associated with the GCNotifier
func (n *GCNotifier) Close() {
	close(n.gcCh)
}

// New creates and starts a new GCNotifier
func New() *GCNotifier {
	gcCh := make(chan struct{}, 1)
	runtime.SetFinalizer(&sentinel{gcCh: gcCh}, finalizer)
	return &GCNotifier{gcCh: gcCh}
}

type sentinel struct {
	gcCh chan struct{}
}

func finalizer(obj interface{}) {
	defer recover() // writing to a closed channel will panic

	select {
	case obj.(*sentinel).gcCh <- struct{}{}:
		// notification sent. if the channel is closed the line above will panic, so
		// we'll end up in the defer, skipping the SetFinalizer below
	default:
		// the channel already contains a notification, simply drop the new one
	}

	// we get here only if the channel was not closed
	runtime.SetFinalizer(obj, finalizer)
}
