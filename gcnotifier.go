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
// terminates (either because the program terminates before a GC cycle completes,
// or because GC itself is disabled).
package gcnotifier

import (
	"runtime"
	"sync/atomic"
)

// GCNotifier allows your code to control and receive notifications every time
// the garbage collector runs.
type GCNotifier struct {
	close func()
	ch    chan struct{}
}

// AfterGC returns the channel that will receive a notification after every GC
// run. No further notifications will be sent until the previous notification
// has been consumed. To stop notifications immediately call the Close() method.
// Otherwise notifications will continue until the GCNotifier object itself is
// garbage collected. Note that the channel returned by AfterGC will be closed
// only when GCNotifier is garbage collected.
// The channel is unique to a single GCNotifier object: use dedicated
// GCNotifiers if you need to listen for GC notifications in multiple receivers
// at the same time.
func (n *GCNotifier) AfterGC() <-chan struct{} {
	return n.ch
}

// Close will stop and release all resources associated with the GCNotifier. It
// is not required to call Close explicitly: when the GCNotifier object is
// garbage collected Close is called implicitly.
// If you don't call Close explicitly make sure not to accidently maintain the
// GCNotifier object alive.
func (n *GCNotifier) Close() {
	n.close()
}

// New creates and arms a new GCNotifier.
func New() *GCNotifier {
	n := &GCNotifier{
		ch: make(chan struct{}, 1),
	}
	n.close = AfterGC(func() {
		select {
		case n.ch <- struct{}{}:
		default:
		}
	})
	runtime.SetFinalizer(n, func(n *GCNotifier) {
		n.Close()
	})
	return n
}

// AfterGC executes the provided function after each GC cycle.
//
// This is a low-level interface, so the provided function must
// follow the same considerations that apply to finalizers passed
// to runtime.SetFinalizer. If the provided function can block
// or take significant time to execute, the provided function
// should start a goroutine and execute the code inside of it.
// If the provided function panics the whole process will be
// terminated.
//
// The returned function must be called to stop the notifications.
// If the returned function is not called, the provided function
// will continue to be called until the process exits.
//
// For a safer alternative, use GCNotifier instead.
func AfterGC(fn func()) func() {
	stop := uint32(0)

	var finalizer func(*[16]byte)
	finalizer = func(s *[16]byte) {
		if atomic.LoadUint32(&stop) != 0 {
			return
		}
		runtime.SetFinalizer(s, finalizer)
		fn()
	}
	runtime.SetFinalizer(new([16]byte), finalizer)

	return func() {
		atomic.StoreUint32(&stop, 1)
	}
}
