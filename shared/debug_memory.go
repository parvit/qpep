package shared

import (
	"io"
	"os"
	"runtime/pprof"
	"time"
)

func DumpMemoryHeapWatcher() {
	go func() {
		home := os.Getenv("HOME")
		for {
			t, _ := os.OpenFile(home+"/dump_"+time.Now().Format(time.RFC3339Nano)+".prof", os.O_RDWR, 0777)
			pprof.WriteHeapProfile(io.Writer(t))
			t.Sync()
			t.Close()
			<-time.After(10 * time.Second)
		}
	}()
}
