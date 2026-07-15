package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/seccomp/libseccomp-golang"
	"golang.org/x/sys/unix"
)

const stageEnv = "_AIRGAP_STAGE"

func main() {
	if os.Getenv(stageEnv) == "1" {
		runIsolated()
		return
	}

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <program> [args...]\n", os.Args[0])
		os.Exit(1)
	}

	self, err := os.Executable()
	if err != nil {
		panic(err)
	}

	cmd := exec.Command(self, os.Args[1:]...)
	cmd.Env = append(os.Environ(), stageEnv+"=1")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:
			syscall.CLONE_NEWUSER |
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

func runIsolated() {
	// Prevent privilege escalation through setuid/setcap binaries.
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		panic(err)
	}

	dropCapabilities()

	installSeccomp()

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

func installSeccomp() {
	filter, err := seccomp.NewFilter(seccomp.ActAllow)
	if err != nil {
		panic(err)
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

			err = filter.AddRuleConditional(
				socketCall,
				seccomp.ActErrno.SetReturnCode(int16(syscall.EPERM)),
				[]seccomp.ScmpCondition{cond},
			)
			if err != nil {
				panic(err)
			}
		}

		blockSocketFamily(unix.AF_INET)
		blockSocketFamily(unix.AF_INET6)
		blockSocketFamily(unix.AF_PACKET)
	}

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

		if err := filter.AddRule(
			syscallID,
			seccomp.ActErrno.SetReturnCode(int16(syscall.EPERM)),
		); err != nil {
			panic(err)
		}
	}

	if err := filter.Load(); err != nil {
		panic(err)
	}
}
