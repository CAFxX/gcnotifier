// Package gcnotifier provides a way to receive notifications after every
// garbage collection (GC) cycle. This can be useful, in long-running programs,
// to instruct your code to free additional memory resources that you may be
// using.
//
// A common use case for this is when you have custom data structures (e.g.
// buffers, caches, rings, trees, pools, ...): instead of setting a maximum size
// to your data structure you can leave it unbounded and then drop all (or some)
// of the allocated-but-unused slots after every GC run (e.g. sync.Pool drops
// all allocated-but-unused objects in the pool during GC).
//
// To minimize the load on the GC the code that runs after receiving the
// notification should try to avoid allocations as much as possible, or at the
// very least make sure that the amount of new memory allocated is significantly
// smaller than the amount of memory that has been "freed" in response to the
// notification.
//
// GCNotifier guarantees to send a notification after every GC cycle completes.
// Note that the Go runtime does not guarantee that the GC will run:
// specifically there is no guarantee that a GC will run before the program
// terminates.
package gcnotifier

import "runtime"

// GCNotifier allows your code to control and receive notifications every time
// the garbage collector runs.
type GCNotifier struct {
	n *gcnotifier
}

type gcnotifier struct {
	doneCh chan struct{}
	gcCh   chan struct{}
}

type sentinel gcnotifier

// AfterGC returns the channel that will receive a notification after every GC
// run. No further notifications will be sent until the previous notification
// has been consumed. To stop the notifications call the Close() method. The
// channel is unique to a single GCNotifier object: use dedicated GCNotifiers
// if you need to listen for GC notifications in multiple receivers at the same
// time.
func (n *GCNotifier) AfterGC() <-chan struct{} {
	return n.n.gcCh
}

// Close will stop and release all resources associated with the GCNotifier. It
// is not required to call Close explicitly: when the GCNotifier object is
// garbage collected Close is called implicitly.
// If you don't call Close explicitly make sure not to accidently maintain the
// GCNotifier object alive.
func (n *GCNotifier) Close() {
	autoclose(n.n)
}

// autoclose is both called explicitely via Close or when the GCNotifier is
// garbage collected
func autoclose(obj interface{}) {
	select {
	case obj.(*gcnotifier).doneCh <- struct{}{}:
	default:
	}
}

// New creates and arms a new GCNotifier
func New() *GCNotifier {
	n := &gcnotifier{
		gcCh:   make(chan struct{}, 1),
		doneCh: make(chan struct{}, 1),
	}
	runtime.SetFinalizer(&sentinel{gcCh: n.gcCh, doneCh: n.doneCh}, finalizer)
	runtime.SetFinalizer(n, autoclose)
	return &GCNotifier{n: n}
}

func finalizer(obj interface{}) {
	s := obj.(*sentinel)

	// check if we have to shutdown
	select {
	case <-s.doneCh:
		close(s.gcCh)
		return
	default:
	}

	// send the notification
	select {
	case s.gcCh <- struct{}{}:
	default:
		// drop it if there's already an unread notification in gcCh
	}

	// re-arm the finalizer
	runtime.SetFinalizer(obj, finalizer)
}
