Incident: Mount propagation leak during container /proc remount

Date: 2026-05-29

What happened:
I ran sudo go run . run in the terminal and it resolved to the PID 1. I  then ran ps aux expecting to see 2 running processes in the private namespace that I just built. THe two running processes were there but something else happened in the background. Claude gave back "claude code terminated by signal sigabrt" error, my ubuntu terminal gave back an error that all running processes had been killed and could not be restarted

Root cause:

What really happened that ended up killing all running procceses except the 2 that were now running? I was using CLONE_NEWNS to ensure that the chid func had its own private copy of the mount table instead of my container and host sharing the same mount table. Without that a a mount proc would end up overwriting /proc on th host and we would end up with a system that sees a broken /proc which would make the system crash. I then did a syscall.Mount("proc", "/proc","proc",0,"") to mount the child in its own private namespace.So where did the issue enumarate from if everything was done correctly? When systemd starts on a modern linux system it marks the root file system mount as MS_SHARED  which means any mount done in a child namespace propagates back to the parent namespace. So when the container mounted a new /proc it leaked to the host's /proc overwriting it and showing my 2 running processes thus each proceses trying to read /proc saw a broken view
Impact:

What broke? well, the whole system crashed since each process that was running early could not read from the now altered /proc. Processes including terminal, shell and claude.
Fix:
CLONE_NEWNS gives you a new mountspace but it starts as a shared copy of the parent. You have to explicitly make the namespace private first before starting the mount. THis will ensure that your namespace does not end up overwriting the host's /proc and thus crashing your system. syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
can be used to make it private and prevent a repeat of the same. Mounting "/" on before mounting on /proc servers 3 functionalities. 1. / targets the root node of the linux filesytem subtree 2.MS_PRIVATE turns off propagation and tells linux kernel not to share any future mounts in this namespace with the host's namespace 3.MS_REC recursively applies the private rule to all existing subdirectories and direct mount about the change. THis acts as a shield and you can now mount /proc safelys

Lesson:
CLONE_NEWNS gives you a new namespace but not a private one — always remount root as MS_PRIVATE|MS_REC before any mounts inside a container. 