package main

import (
	"fmt"
	"os"
)

func SetupCgroup(pid int) error {

	if err := os.MkdirAll("/sys/fs/cgroup/warden", 0755); err != nil {
		return fmt.Errorf("error creating cgroup directory: %v", err)
	}

	if err := os.WriteFile("/sys/fs/cgroup/cgroup.subtree_control", []byte("+memory +cpu"), 0644); err != nil {
		return fmt.Errorf("error enabling cgroup controllers: %v", err)
	}

	if err := os.WriteFile("/sys/fs/cgroup/warden/memory.max", []byte("104857600"), 0644); err != nil {
		return fmt.Errorf("error setting memory limit: %v", err)
	}

	if err := os.WriteFile("/sys/fs/cgroup/warden/cpu.max", []byte("10000 100000"), 0644); err != nil {
		return fmt.Errorf("error setting cpu limit: %v", err)
	}

	if err := os.WriteFile("/sys/fs/cgroup/warden/cgroup.procs", []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
		return fmt.Errorf("error adding process to cgroup: %v", err)
	}
	return nil
}
