package shared

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

func DumpWatcher() {
	t, err := os.Create("cpu.prof")
	fmt.Printf("err: %v\n", err)
	pprof.StartCPUProfile(io.Writer(t))
	runtime.SetCPUProfileRate(50)

	go func() {
		defer func() {
			t.Sync()
			t.Close()
		}()
		for i := 0; i < 20; i++ {
			t2, err2 := os.Create(fmt.Sprintf("heap_%02d.prof", i))
			fmt.Printf("err2: %v\n", err2)
			pprof.WriteHeapProfile(t2)
			t2.Sync()
			t2.Close()
			<-time.After(3 * time.Second)
		}
		pprof.StopCPUProfile()
	}()
}
