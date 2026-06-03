package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func configureChildNetworkInside() {

	if err := exec.Command("ip", "link", "set", "lo", "up").Run(); err != nil {
		fmt.Printf("child: error bringing up loopback: %v\n", err)
	}

	if err := exec.Command("ip", "addr", "add", "172.20.0.10/24", "dev", "veth1").Run(); err != nil {
		fmt.Printf("child: error assigning IP to veth1: %v\n", err)
		return
	}

	if err := exec.Command("ip", "link", "set", "veth1", "up").Run(); err != nil {
		fmt.Printf("child: error bringing veth1 up: %v\n", err)
		return
	}

	if err := exec.Command("ip", "route", "add", "default", "via", "172.20.0.1").Run(); err != nil {
		fmt.Printf("child: error setting default gateway: %v\n", err)
		return
	}
}

func child() {
	syncPipe := os.NewFile(3, "sync_pipe")
	buf := make([]byte, 1)
	_, _ = syncPipe.Read(buf)
	syncPipe.Close()
	newroot := "/home/kamausimon/warden-rootfs"

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
	configureChildNetworkInside()

	fmt.Printf("The PID of the current running child %v", os.Getpid())

	if err := syscall.Exec("/bin/bash", []string{"/bin/bash"}, os.Environ()); err != nil {
		fmt.Printf("child: error executing command: %v\n", err)
	}
}
