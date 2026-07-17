package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/seccomp/libseccomp-golang"
	"golang.org/x/sys/unix"
)

const (
	stageEnv        = "_AIRGAP_STAGE"
	keepLoopbackEnv = "_AIRGAP_KEEP_LOOPBACK"
	logExternalEnv  = "_AIRGAP_LOG_EXTERNAL"
)

func main() {
	if os.Getenv(stageEnv) == "1" {
		runIsolated(os.Getenv(keepLoopbackEnv) == "1", os.Getenv(logExternalEnv) == "1")
		return
	}

	keepLoopback := flag.Bool("keep-loopback", false, "keep the loopback interface (127.0.0.1) up and reachable")
	logExternal := flag.Bool("log-external", false, "log blocked network syscalls to stderr")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [--keep-loopback] [--log-external] <program> [args...]\n", os.Args[0])
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	self, err := os.Executable()
	if err != nil {
		panic(err)
	}

	cmd := exec.Command(self, flag.Args()...)
	cmd.Env = append(os.Environ(), stageEnv+"=1")
	if *keepLoopback {
		cmd.Env = append(cmd.Env, keepLoopbackEnv+"=1")
	}
	if *logExternal {
		cmd.Env = append(cmd.Env, logExternalEnv+"=1")
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUSER |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWUTS,

		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},

		GidMappingsEnableSetgroups: false,
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
		},
	}

	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runIsolated(keepLoopback, logExternal bool) {
	// Prevent privilege escalation through setuid/setcap binaries.
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		panic(err)
	}

	if keepLoopback {
		if err := bringUpLoopback(); err != nil {
			panic(err)
		}
	}

	dropCapabilities()

	installSeccomp(keepLoopback, logExternal)

	if len(os.Args) < 2 {
		os.Exit(1)
	}

	target := exec.Command(os.Args[1], os.Args[2:]...)
	target.Stdin = os.Stdin
	target.Stdout = os.Stdout
	target.Stderr = os.Stderr

	if err := target.Run(); err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			os.Exit(e.ExitCode())
		}

		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// bringUpLoopback sets the "lo" interface UP inside the new network
// namespace. The kernel creates it disabled by default; the namespace still
// has no other interfaces or routes, so this only enables 127.0.0.1/::1
// traffic, not external access.
func bringUpLoopback() error {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
	if err != nil {
		return err
	}
	defer unix.Close(fd)

	ifr, err := unix.NewIfreq("lo")
	if err != nil {
		return err
	}
	if err := unix.IoctlIfreq(fd, unix.SIOCGIFFLAGS, ifr); err != nil {
		return err
	}
	ifr.SetUint16(ifr.Uint16() | unix.IFF_UP | unix.IFF_RUNNING)
	return unix.IoctlIfreq(fd, unix.SIOCSIFFLAGS, ifr)
}

func dropCapabilities() {
	// Remove all inheritable/effective capabilities.
	for cap := 0; cap < 64; cap++ {
		_ = unix.Prctl(
			unix.PR_CAPBSET_DROP,
			uintptr(cap),
			0,
			0,
			0,
		)
	}

	// Clear ambient capabilities.
	_ = unix.Prctl(
		unix.PR_CAP_AMBIENT,
		unix.PR_CAP_AMBIENT_CLEAR_ALL,
		0,
		0,
		0,
	)
}

func installSeccomp(keepLoopback, logExternal bool) {
	filter, err := seccomp.NewFilter(seccomp.ActAllow)
	if err != nil {
		panic(err)
	}

	action := seccomp.ActErrno.SetReturnCode(int16(syscall.EPERM))
	if logExternal {
		// Route blocked calls through userspace so they can be logged
		// before being denied, instead of just returning EPERM.
		action = seccomp.ActNotify
	}

	socketCall, err := seccomp.GetSyscallFromName("socket")
	if err == nil {
		blockSocketFamily := func(family int) {
			cond, err := seccomp.MakeCondition(
				0,
				seccomp.CompareEqual,
				uint64(family),
			)
			if err != nil {
				panic(err)
			}

			if err := filter.AddRuleConditional(socketCall, action, []seccomp.ScmpCondition{cond}); err != nil {
				panic(err)
			}
		}

		// AF_PACKET (raw sockets) has no legitimate loopback use, so it stays
		// blocked even with --keep-loopback.
		blockSocketFamily(unix.AF_PACKET)
		if !keepLoopback {
			blockSocketFamily(unix.AF_INET)
			blockSocketFamily(unix.AF_INET6)
		}
	}

	// With --keep-loopback the namespace still has no interface other than
	// "lo" and no routes, so connect/bind/etc. are left unblocked: external
	// addresses fail at the routing layer regardless, only 127.0.0.1/::1
	// traffic actually works.
	if !keepLoopback {
		for _, name := range []string{
			"connect",
			"bind",
			"listen",
			"accept",
			"accept4",
			"sendto",
			"sendmsg",
			"recvfrom",
			"recvmsg",
		} {
			syscallID, err := seccomp.GetSyscallFromName(name)
			if err != nil {
				continue
			}

			if err := filter.AddRule(syscallID, action); err != nil {
				panic(err)
			}
		}
	}

	if err := filter.Load(); err != nil {
		panic(err)
	}

	if logExternal {
		notifFd, err := filter.GetNotifFd()
		if err != nil {
			panic(err)
		}
		go logBlockedSyscalls(notifFd)
	}
}

// logBlockedSyscalls services ActNotify notifications: it logs each blocked
// syscall to stderr, then denies it with EPERM, same as the non-logging
// path. It returns once the notify fd is closed (target process exited).
//
// ponytail: logs syscall name + pid only, not the resolved destination
// address (would need to read the target's memory for sockaddr args) — add
// if per-connection detail is needed.
func logBlockedSyscalls(fd seccomp.ScmpFd) {
	for {
		req, err := seccomp.NotifReceive(fd)
		if err != nil {
			return
		}

		name, err := req.Data.Syscall.GetName()
		if err != nil {
			name = fmt.Sprintf("syscall#%d", req.Data.Syscall)
		}
		fmt.Fprintf(os.Stderr, "offline: blocked %s (pid %d)\n", name, req.Pid)

		_ = seccomp.NotifRespond(fd, &seccomp.ScmpNotifResp{
			ID:    req.ID,
			Error: int32(syscall.EPERM),
			Val:   0,
		})
	}
}
