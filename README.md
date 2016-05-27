# gcnotifier

gcnotifier provides a way to receive notifications after every time
garbage collection (GC) runs. This can be useful to instruct your code to
free additional memory resources that you may be using.

A common use case for this is when you have a custom pool of objects: instead
of setting a maximum size to your pool you can leave it unbounded and then
drop all (or some) of them after every GC run (e.g. sync.Pool drops all
objects in the pool during GC).

To minimize the load on the GC the code that runs after receiving the
notification should try to avoid allocations as much as possible, or at the
very least make sure that the amount of new memory allocated is significantly
smaller than the amount of memory that has been "freed" by your code.

[Documentation](https://godoc.org/github.com/CAFxX/gcnotifier)

# License
[MIT](LICENSE)

# Author
Carlo Alberto Ferraris (@cafxx)
