package shared

import (
	"io"
	"os"
	"runtime/pprof"
	"time"
)

func DumpMemoryHeapWatcher() {
	go func() {
		for {
			t, _ := os.OpenFile("dump_"+time.Now().Format(time.RFC3339Nano)+".prof", os.O_RDWR, 0777)
			defer func() {
				t.Sync()
				t.Close()
			}()
			pprof.WriteHeapProfile(io.Writer(t))
			<-time.After(10 * time.Second)
		}
	}()
}
