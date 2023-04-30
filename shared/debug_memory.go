package shared

import (
	"fmt"
	"io"
	"os"
	"runtime/pprof"
	"time"
)

func DumpWatcher() {
	t, _ := os.OpenFile("cpu.prof", os.O_CREATE|os.O_RDWR, 0777)
	pprof.StartCPUProfile(io.Writer(t))

	go func() {
		defer func() {
			t.Sync()
			t.Close()
		}()
		for i := 0; i < 20; i++ {
			t2, _ := os.OpenFile(fmt.Sprintf("heap_%02d.prof", i), os.O_CREATE|os.O_RDWR, 0777)
			pprof.WriteHeapProfile(t)
			t2.Sync()
			t2.Close()
			<-time.After(10 * time.Second)
		}
		pprof.StopCPUProfile()
	}()
}
