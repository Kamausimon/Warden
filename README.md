# Warden

A container runtime built from scratch in Go using raw Linux syscalls — no Docker, no containerd, no abstractions. Warden implements the four core primitives that make containers work: process isolation, resource limits, filesystem isolation, and container networking.

This is what Docker does under the hood.

## How it works

Warden implements four layers of isolation:

**Layer 1 — Process isolation (Linux namespaces)**

When you run a container, Warden calls `clone()` with namespace flags directly via Go's `syscall` package. The container process gets its own PID tree (it sees itself as PID 1), its own hostname, its own mount table, and its own network stack — completely isolated from the host.

The binary uses a re-execution pattern: `warden run` is the parent that sets up namespaces, then re-executes the same binary as `warden child` inside those namespaces. The child never knows it's a re-execution.

**Layer 2 — Resource limits (cgroups v2)**

Warden writes directly to `/sys/fs/cgroup/` to enforce hard limits on the container process:
- Memory: 100MB cap (`memory.max`)
- CPU: 10% quota (`cpu.max`)

The kernel enforces these limits. No application inside the container can bypass them.

**Layer 3 — Filesystem isolation (pivot_root + Alpine Linux)**

Warden uses `pivot_root` (not `chroot`) to give the container its own root filesystem backed by an Alpine Linux minirootfs. The old host root is unmounted entirely after pivoting — there is no escape path back to the host filesystem.

Before any mounts, the mount namespace is made private (`MS_PRIVATE|MS_REC`) to prevent mount propagation leaking back to the host. Without this step, mounting `/proc` inside the container would overwrite the host's `/proc` and crash the system.

**Layer 4 — Container networking (veth pairs + bridge + NAT)**

Warden creates a virtual network topology:

```
HOST                              CONTAINER
eth0/wlp2s0 (internet)
      |
warden0 bridge (172.20.0.1)  <==veth==>  veth1 (172.20.0.10)
```

A veth pair acts as a virtual ethernet cable between the host bridge and the container's isolated network namespace. iptables MASQUERADE handles NAT so the container can reach the network using the host's IP address.

A sync pipe coordinates timing: the child process waits on a file descriptor until the parent has finished setting up the network before proceeding. This prevents a race condition where the container tries to configure `veth1` before the parent has moved it into the container's namespace.

## Prerequisites

- Linux kernel 4.6+ (cgroups v2)
- Go 1.21+
- Root privileges (`sudo`)
- `iptables`
- An Alpine Linux minirootfs extracted to `/home/<user>/warden-rootfs`

```
mkdir ~/warden-rootfs
wget https://dl-cdn.alpinelinux.org/alpine/v3.23/releases/x86_64/alpine-minirootfs-3.23.4-x86_64.tar.gz
tar -xzf alpine-minirootfs-3.23.4-x86_64.tar.gz -C ~/warden-rootfs
```

## Build and run

```
git clone https://github.com/kamausimon/warden
cd warden
go build -o warden .
sudo ./warden run
```

## What you will see

```
run: spawning container process
Container network setup complete with bridge warden0 and veth pair veth0 <-> veth1
The PID of the current running child 1
bash-5.3#
```

Inside the container:

```
bash-5.3# hostname
container

bash-5.3# ps aux
USER   PID  COMMAND
root     1  /bin/bash
root    12  ps aux

bash-5.3# cat /etc/os-release
NAME="Alpine Linux"
VERSION_ID=3.23.4

bash-5.3# ping 192.168.0.1
64 bytes from 192.168.0.1: icmp_seq=1 ttl=63 time=4.75 ms
```

## Known limitations

- The rootfs path is hardcoded. Running on a different machine requires updating the path in `child.go`.
- Only one container can run at a time. Concurrent containers would conflict on the cgroup name, bridge, and veth pair names.
- The NAT rule accumulates duplicates across runs. Cleanup is not yet automated.
- No user namespace (`CLONE_NEWUSER`) — the container runs as root mapped to host root.
- No image format. The rootfs must be pre-extracted manually.

## Project structure

```
main.go      — argument routing only
run.go       — parent: namespace setup, cgroup, network, re-execution
child.go     — child: pivot_root, proc mount, network config, exec
cgroups.go   — cgroups v2 resource limits
network.go   — veth pair, bridge, NAT setup via netlink
```
