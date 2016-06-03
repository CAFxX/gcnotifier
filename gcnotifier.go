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
	gcCh   chan struct{}
	doneCh chan struct{}
}

// AfterGC returns a channel that will receive a notification after every GC
// run. No further notifications will be sent until the previous notification
// has been consumed. To stop the notifications call the Close() method.
func (n *GCNotifier) AfterGC() <-chan struct{} {
	return n.gcCh
}

// Close will stop and release all resources associated with the GCNotifier
func (n *GCNotifier) Close() {
	select {
	case n.doneCh <- struct{}{}:
	default:
	}
}

// New creates and starts a new GCNotifier
func New() *GCNotifier {
	gcCh := make(chan struct{}, 1)
	doneCh := make(chan struct{}, 1)
	runtime.SetFinalizer(&sentinel{gcCh: gcCh, doneCh: doneCh}, finalizer)
	return &GCNotifier{gcCh: gcCh, doneCh: doneCh}
}

type sentinel GCNotifier

func finalizer(obj interface{}) {
	// check if we have to shutdown
	select {
	case <-obj.(*sentinel).doneCh:
		close(obj.(*sentinel).gcCh)
		return
	default:
	}

	// send the notification
	select {
	case obj.(*sentinel).gcCh <- struct{}{}:
	default:
	}

	// re-arm the finalizer
	runtime.SetFinalizer(obj, finalizer)
}
