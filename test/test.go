package main

import (
	"github.com/cafxx/gcnotifier"
	"log"
	"runtime"
	"time"
)

func main() {
	go func() {
		M := &runtime.MemStats{}
		NumGC := 0
		for range gcnotifier.AfterGC() {
			runtime.ReadMemStats(M)
			NumGC += 1
			log.Printf("AfterGC %d %d\n", M.NumGC, NumGC)
		}
	}()
	for {
		b := make([]byte, 1<<20)
		b[0] = 1
		// wait some time to make sure we don't skip notifications
		<-time.After(10 * time.Millisecond)
	}
}
