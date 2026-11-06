package main

import (
	"fmt"
	"os"
	"syscall"

	seccomp "github.com/seccomp/libseccomp-golang"
)

// --------------------------------------------------------------------------- //
// Seccomp.
// --------------------------------------------------------------------------- //

// buildSeccompFilter - Assembles the network-blocking filter without
// installing it. Splitting the build from the load is what lets a test read
// the policy back — Load() needs a namespace, ExportPFC does not.
//
// The filter defaults to allow and denies by exception: the goal is to stop
// networking, not to enumerate a syscall allowlist the wrapped program would
// then trip over. Denials are EPERM, or a userspace notification when the
// caller asked to see them.
func buildSeccompFilter(s sandbox) *seccomp.ScmpFilter {
	filter, err := seccomp.NewFilter(seccomp.ActAllow)
	if err != nil {
		panic(err)
	}

	action := seccomp.ActErrno.SetReturnCode(int16(syscall.EPERM))
	if s.logExternal {
		// Route blocked calls through userspace so they can be logged
		// before being denied, instead of just returning EPERM.
		action = seccomp.ActNotify
	}

	blockSocketFamilies(filter, action, s.blockedFamilies())

	// With --keep-loopback the namespace still has no interface other than
	// "lo" and no routes, so connect/bind/etc. are left unblocked: external
	// addresses fail at the routing layer regardless, only 127.0.0.1/::1
	// traffic actually works.
	if !s.keepLoopback {
		blockNetworkCalls(filter, action)
	}

	return filter
}

// installSeccomp - Loads the network-blocking filter into the current process.
// A filter that will not load leaves the child unprotected, so it panics
// rather than carrying on.
func installSeccomp(s sandbox) {
	filter := buildSeccompFilter(s)

	if err := filter.Load(); err != nil {
		panic(err)
	}

	if s.logExternal {
		notifFd, err := filter.GetNotifFd()
		if err != nil {
			panic(err)
		}
		go logBlockedSyscalls(notifFd)
	}
}

// blockSocketFamilies - Adds one conditional rule refusing socket(2) per
// address family. Argument 0 of socket(2) is the family, hence the index.
//
// A kernel that does not know socket(2) by that name is not one this tool can
// filter; the namespace still holds, so carry on rather than abort.
func blockSocketFamilies(filter *seccomp.ScmpFilter, action seccomp.ScmpAction, families []int) {
	socketCall, err := seccomp.GetSyscallFromName("socket")
	if err != nil {
		return
	}

	for _, family := range families {
		cond, err := seccomp.MakeCondition(
			socketFamilyArg,
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
}

// blockNetworkCalls - Refuses the connect family outright. A name the running
// architecture does not have is skipped: there is nothing there to block.
func blockNetworkCalls(filter *seccomp.ScmpFilter, action seccomp.ScmpAction) {
	for _, name := range blockedNetworkCalls {
		syscallID, err := seccomp.GetSyscallFromName(name)
		if err != nil {
			continue // not on this architecture; nothing to block
		}

		if err := filter.AddRule(syscallID, action); err != nil {
			panic(err)
		}
	}
}

// logBlockedSyscalls - Services ActNotify notifications: it logs each blocked
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

		//nolint:errcheck // the supervised process is already being denied; a
		// failed response leaves it blocked, which is the safe direction.
		_ = seccomp.NotifRespond(fd, &seccomp.ScmpNotifResp{
			ID:    req.ID,
			Error: int32(syscall.EPERM),
			Val:   0,
		})
	}
}
