# gcnotifier

gcnotifier provides a way to receive notifications after every run of the
garbage collector (GC). Knowing when GC runs is useful to instruct your code to
free additional memory resources that it may be using.

## Why?
The common use case for gcnotifier is when you have a custom pool of objects:
instead of setting a maximum size to your pool you can leave it unbounded and
then drop all (or some) of them after every GC run (e.g. in the standard library
`sync.Pool` drops all objects in the pool during GC).

To minimize the load on the GC the code that runs after receiving the
notification should try to avoid allocations as much as possible, or at the
very least make sure that the amount of new memory allocated is significantly
smaller than the amount of memory that has been "freed" by your code.

## Example
For a simple example of how to use it have a look at `Example()` in
[gcnotifier_test.go](gcnotifier_test.go). For details have a look at the
[documentation](https://godoc.org/github.com/CAFxX/gcnotifier).

## How it works
gcnotifier uses [finalizers](https://golang.org/pkg/runtime/#SetFinalizer) to
know when a GC run has completed.

Finalizers are run when "[the garbage collector finds an unreachable block with
an associated finalizer](https://golang.org/pkg/runtime/#SetFinalizer)".

The SetFinalizer documentation notes that "[there] is no guarantee that
finalizers will run before a program exits". Finalizers won't run only when the
runtime shuts down (GC doesn't run in this case because the whole process will
die soon and all associated resources will be freed by the OS anyway) so
gcnotifier correctly does not notify of GC in this case. Finalizers can also not
run for other reasons (e.g. zero-sized or package-level objects) that don't
apply to gcnotifier.

The only other case in which a notification will not be sent by gcnotifier is if
your code hasn't consumed a previously-sent notification.

The test in [gcnotifier_test.go](gcnotifier_test.go) generates garbage in a loop
and makes sure that we receive exactly one notification for each of the first
500 GC runs. In my testing I haven't found a way yet to make gcnotifier fail to
notify of a GC run short of shutting down the process or failing to receive the
notification. If you manage to make it fail in any other way please file a
[GitHub issue](https://github.com/CAFxX/gcnotifier/issues/new).

# License
[MIT](LICENSE)

# Author
Carlo Alberto Ferraris ([@cafxx](https://twitter.com/cafxx))
