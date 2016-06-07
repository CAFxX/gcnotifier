package maxheap

import(
  "math"
  "runtime"
  "time"
  "sync"
  "sync/atomic"
  "github.com/shirou/gopsutil/mem"
  "os"
  "strings"
  "strconv"
  "runtime/debug"
)

var maxheaplo uint64
var maxheaphi uint64
var l sync.RWMutex

func Set(lo, hi uint64) {
  if lo > hi {
    return
  }
  l.Lock()
  atomic.StoreUint64(&maxheaplo, lo)
  atomic.StoreUint64(&maxheaphi, hi)
  l.Unlock()
}

func Get() (lo, hi uint64) {
  l.RLock()
  lo, hi = atomic.LoadUint64(&maxheaplo), atomic.LoadUint64(&maxheaphi)
  l.RUnlock()
  return
}

func init() {
  lo, hi := parsegomaxheap(strings.TrimSpace(os.Getenv("GOMAXHEAP")))
  Set(lo, hi)
  go supervisor()
}

func parsegomaxheap(GOMAXHEAP string) (uint64, uint64) {
  var err error

  gmh := strings.Split(GOMAXHEAP, ":")
  if len(gmh) > 2 || GOMAXHEAP == "" {
    return math.MaxUint64, math.MaxUint64
  }

  lo := uint64(0)
  if gmh[0] != "" {
    lo, err = strconv.ParseUint(gmh[0], 10, 64)
    if err != nil {
      return math.MaxUint64, math.MaxUint64
    }
  }

  if len(gmh) == 1 {
    return lo, lo
  }

  hi := uint64(math.MaxUint64)
  if gmh[1] != "" {
    hi, err = strconv.ParseUint(gmh[1], 10, 64)
    if err != nil {
      return math.MaxUint64, math.MaxUint64
    }
  }

  return lo, hi
}

func supervisor() {
  var m runtime.MemStats

  for range time.Tick(250 * time.Millisecond) {
    lo, hi := Get()

    if lo == math.MaxUint64 {
      continue
    }

    runtime.ReadMemStats(&m)

    if lo == hi {
      if m.HeapAlloc > lo {
        runtime.GC()
      }
      continue
    }

    vm, err := mem.VirtualMemory()
    if err != nil {
      if m.HeapAlloc > lo {
        runtime.GC()
      }
      continue
    }

    maxheap := vm.Available + m.HeapSys
    if lo > maxheap {
      maxheap = lo
    } else if hi < maxheap {
      maxheap = hi
    }

    if m.HeapAlloc <= maxheap {
      continue
    }

    // TODO: instead of invoking the GC directly, change GOGC/gcPercent so that
    // we use background GC instead of the STW one
    if vm.Available >= m.HeapSys - m.HeapAlloc {
      runtime.GC()
    } else {
      debug.FreeOSMemory()
    }
  }
}
