package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/parvit/qpep/shared"
)

const (
	EXENAME = "qpep.exe"

	CMD_SERVICE = `%s %s --service %s %s %s`
)

func getServiceCommand(start, client bool) *exec.Cmd {
	exeFile, _ := filepath.Abs(filepath.Join(ExeDir, EXENAME))

	var serviceFlag = "start"
	var clientFlag = "--client"
	var hostFlag = fmt.Sprintf("-Dlistenaddress=%s", shared.QPepConfig.ListenHost)
	var verboseFlag = "--verbose"
	if !start {
		serviceFlag = "stop"
	}
	if !client {
		verboseFlag = ""
	}
	if !shared.QPepConfig.Verbose {
		verboseFlag = ""
	}

	attr := &syscall.SysProcAttr{
		HideWindow: true,
		CmdLine:    fmt.Sprintf(CMD_SERVICE, exeFile, clientFlag, serviceFlag, hostFlag, verboseFlag),
	}

	cmd := exec.Command(exeFile)
	if cmd == nil {
		ErrorMsg("Could not create client command")
		return nil
	}
	cmd.Dir, _ = filepath.Abs(ExeDir)
	cmd.SysProcAttr = attr
	return cmd
}
