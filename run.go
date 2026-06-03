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
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Printf("run: error starting command: %v\n", err)
		return
	}
	if err := SetupCgroup(cmd.Process.Pid); err != nil {
		fmt.Printf("run: error setting up cgroup: %v\n", err)
	}

	if err := setupContainerNetwork(cmd.Process.Pid); err != nil {
		fmt.Printf("run: error setting up network: %v\n", err)
	}

	if err := cmd.Wait(); err != nil {
		fmt.Printf("run: error running command: %v\n", err)
	}
}
