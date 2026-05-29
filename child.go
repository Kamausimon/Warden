package main

import (
	"fmt"
	"os"
	"syscall"
)

func child() {
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	syscall.Mount("proc", "/proc", "proc", 0, "")
	syscall.Sethostname([]byte("container"))

	fmt.Printf("The PID of the current running child %v", os.Getpid())

	if err := syscall.Exec("/bin/bash", []string{"/bin/bash"}, os.Environ()); err != nil {
		fmt.Printf("child: error executing command: %v\n", err)
	}
}
