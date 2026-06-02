package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func child() {
	newroot := "/tmp/warden-rootfs"

	err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err != nil {
		fmt.Printf("child: error making namespace private: %v\n", err)
		return
	}

	err = syscall.Mount(newroot, newroot, "", syscall.MS_BIND|syscall.MS_REC, "")
	if err != nil {
		fmt.Printf("child: error bind-mounting root filesystem: %v\n", err)
		return
	}

	err = syscall.Mount("", newroot, "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err != nil {
		fmt.Printf("child: error making mount private: %v\n", err)
		return
	}
	oldRootput := filepath.Join(newroot, ".old_root")
	if err := os.MkdirAll(oldRootput, 0700); err != nil {
		fmt.Printf("child: error creating old root directory: %v\n", err)
		return
	}

	if err := syscall.PivotRoot(newroot, oldRootput); err != nil {
		fmt.Printf("child: error pivoting root filesystem: %v\n", err)
		return
	}

	if err := syscall.Chdir("/"); err != nil {
		fmt.Printf("child: error changing working directory: %v\n", err)
		return
	}
	if err := syscall.Unmount("/.old_root", syscall.MNT_DETACH); err != nil {
		fmt.Printf("child: error unmounting old root filesystem: %v\n", err)
		return
	}
	os.RemoveAll("/.old_root")

	err = syscall.Mount("proc", "/proc", "proc", 0, "")
	if err != nil {
		fmt.Printf("child: error mounting proc filesystem: %v\n", err)
		return
	}
	syscall.Sethostname([]byte("container"))

	fmt.Printf("The PID of the current running child %v", os.Getpid())

	if err := syscall.Exec("/bin/bash", []string{"/bin/bash"}, os.Environ()); err != nil {
		fmt.Printf("child: error executing command: %v\n", err)
	}
}
