package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/parvit/qpep/shared"
	"github.com/skratchdot/open-golang/open"
)

func openConfigurationWithOSEditor() {
	_, baseConfigPath, _ := shared.GetConfigurationPaths()

	if err := open.Run(baseConfigPath); err != nil {
		ErrorMsg("Editor configuration failed with error: %v", err)
		return
	}
}

func openWebguiWithOSBrowser(clientMode, serverMode bool) {
	mode := "server"
	port := shared.QPepConfig.GatewayAPIPort
	if (clientMode && serverMode) || (!clientMode && !serverMode) {
		ErrorMsg("Webgui can start with just one mode between server and client!")
		return
	}
	if clientMode {
		mode = "client"
	}
	if serverMode {
		mode = "server"
	}

	guiurl := fmt.Sprintf(shared.WEBGUI_URL, port, mode, port)
	if err := open.Run(guiurl); err != nil {
		ErrorMsg("Webgui startup failed with error: %v", err)
		return
	}
}

func startReloadConfigurationWatchdog() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		_, baseConfigFile, _ := shared.GetConfigurationPaths()

		var lastModTime time.Time
		if stat, err := os.Stat(baseConfigFile); err == nil {
			lastModTime = stat.ModTime()

		} else {
			ErrorMsg("Configuration file not found, stopping")
			cancel()
			return
		}

	CHECKLOOP:
		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping configfile watchdog")
				break CHECKLOOP

			case <-time.After(10 * time.Second):
				stat, err := os.Stat(baseConfigFile)
				if err != nil {
					continue
				}
				if !stat.ModTime().After(lastModTime) {
					continue
				}
				lastModTime = stat.ModTime()
				if shared.ReadConfiguration(true) == nil {
					reloadClientIfRunning()
					reloadServerIfRunning()
				}
				continue
			}
		}
	}()

	return ctx, cancel
}
