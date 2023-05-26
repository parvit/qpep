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
	cpuWatcher(0)
}

func cpuWatcher(idx int) {
	t, err := os.Create(fmt.Sprintf("cpu_%d.prof", idx))
	fmt.Printf("err: %v\n", err)
	runtime.SetCPUProfileRate(100)

	pprof.StartCPUProfile(io.Writer(t))

	go func() {
		<-time.After(10 * time.Second)
		pprof.StopCPUProfile()
		t.Sync()
		t.Close()

		go cpuWatcher(idx + 1)
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
