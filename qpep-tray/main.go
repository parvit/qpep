package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/getlantern/systray"
	"github.com/parvit/qpep/shared"
)

func main() {
	defer func() {
		// clear the proxy in case a orphaned client cannot
		shared.SetSystemProxy(false)
	}()

	// note: channel is never dequeued as to stop the ctrl-c signal from stopping also
	// this process and only the child client / server
	interruptListener := make(chan os.Signal, 1)
	signal.Notify(interruptListener, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// open the log file in current directory
	ExeDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))

	f, err := os.OpenFile(filepath.Join(ExeDir, "qpep-tray.log"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	wrt := io.MultiWriter(os.Stdout, f)
	log.SetOutput(wrt)

	log.SetFlags(log.Ltime | log.Lmicroseconds)

	// read configuration
	if err := shared.ReadConfiguration(true); err != nil {
		ErrorMsg("Could not load configuration file, please edit: %v", err)
	}

	systray.Run(onReady, onExit)
}
