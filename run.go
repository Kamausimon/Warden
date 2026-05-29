package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func run() {

	fmt.Printf("run: running as child process\n")
	cmd := exec.Command("/proc/self/exe", "child")

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("run: error running command: %v\n", err)
	}
}
