package shared

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

func CPUWatcher() {
	t, err := os.Create("cpu.prof")
	fmt.Printf("err: %v\n", err)
	runtime.SetCPUProfileRate(100)

	pprof.StartCPUProfile(io.Writer(t))

	go func() {
		<-time.After(60 * time.Second)
		pprof.StopCPUProfile()
		t.Sync()
		t.Close()
	}()
}

func MemoryWatcher() {
	go func() {
		for i := 0; i < 30; i++ {
			t2, err2 := os.Create(fmt.Sprintf("heap_%02d.prof", i))
			fmt.Printf("err2: %v\n", err2)
			pprof.WriteHeapProfile(t2)
			t2.Sync()
			t2.Close()
			<-time.After(3 * time.Second)
		}
	}()
}
