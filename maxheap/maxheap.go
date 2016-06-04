package maxheap

import(
  "math"
  "runtime"
  "time"
  "sync/atomic"
)

var maxheap uint64 = math.MaxUint64

func Set(bytes uint64) {
  atomic.StoreUint64(&maxheap, bytes)
}

func Get() uint64 {
  return atomic.LoadUint64(&maxheap)
}

func init() {
  GOMAXHEAP := os.Getenv("GOMAXHEAP")
  if bytes, err := strconv.ParseUint(GOMAXHEAP, 10, 64); err != nil {
    Set(bytes)
  }
  go supervisor()
}

func supervisor() {
  for range time.Tick(250 * time.Millisecond) {
    max := Get()
    if max == math.MaxUint64 {
      continue
    }
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    if m.HeapAlloc > max {
      runtime.GC()
    }
  }
}
