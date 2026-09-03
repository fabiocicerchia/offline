// offline — run a program with its network access fully isolated.
//
// A small, self-contained sandbox, small enough to read in one sitting. One
// binary, two stages:
//
//	default            build the namespace set and re-exec self inside it
//	_AIRGAP_STAGE=1    (runs INSIDE the namespaces) drop privileges, install
//	                   the seccomp filter, then exec the wrapped program
//
// The stages exist because a running Go program cannot enter a user namespace:
// unshare(CLONE_NEWUSER) refuses a multithreaded caller, and the runtime is
// multithreaded before main() starts. Only a fresh child can be cloned into
// the set.
//
// Isolation is layered, so a gap in one layer is not a gap in the sandbox:
//
//	namespaces     no interfaces, no routes, no host IPC/mounts
//	capabilities   the bounding and ambient sets are emptied
//	seccomp        socket(AF_INET/AF_INET6/AF_PACKET) and the connect family
//
// Build:  go build -o offline offline.go
// Usage:  offline [--keep-loopback] [--log-external] <program> [args...]
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"syscall"

	seccomp "github.com/seccomp/libseccomp-golang"
	"golang.org/x/sys/unix"
)

// --------------------------------------------------------------------------- //
// Config — the isolation model lives in these constants. Review them.
// --------------------------------------------------------------------------- //

// The re-exec channel. The outer stage cannot put itself inside the new
// namespaces, so it re-executes this same binary and hands the parsed flags
// over through the environment rather than through argv: argv belongs to the
// wrapped program and must reach it untouched.
const (
	stageEnv        = "_AIRGAP_STAGE"
	keepLoopbackEnv = "_AIRGAP_KEEP_LOOPBACK"
	logExternalEnv  = "_AIRGAP_LOG_EXTERNAL"

	envOn = "1" // the only value any of the three ever carries
)

// The exit codes offline produces of its own accord. Anything else a caller
// sees is the wrapped program's own status, passed through untouched.
const (
	// exitFailure is offline failing: a bad invocation, or a program that
	// never started. A program that started and exited non-zero is not this.
	exitFailure = 1

	// exitSignalBase is the shells' 128+signal convention, used for a wrapped
	// program killed by a signal — it has no exit code of its own.
	exitSignalBase = 128
)

// Positions and flag values that would otherwise be bare numbers at the call
// site: which argv slot the wrapped program starts at, which argument of
// socket(2) carries the address family, the ID a mapping makes root inside the
// user namespace, and the "on" value PR_SET_NO_NEW_PRIVS expects.
const (
	targetArgv      = 1
	socketFamilyArg = 0
	namespaceRoot   = 0
	noNewPrivsOn    = 1
)

// jailNamespaces is the set the child is cloned into. Network is the point of
// the tool; the other four are what stop the network being reached around it —
// a shared mount namespace re-exposes host sockets, a shared IPC namespace lets
// the child talk to processes that still have one.
//
// CLONE_NEWPID is deliberately absent. It renumbers processes without giving
// them a matching procfs, and mounting one needs privileges the host may not
// grant: Ubuntu's kernel.apparmor_restrict_unprivileged_userns confines anyone
// who creates an unprivileged user namespace to a profile that refuses every
// mount, EACCES. The result is a child that is PID 1 while /proc still
// describes the host, so anything reading /proc/<getpid()> — Node and Bun
// runtimes do, on startup — reads an unrelated host process and dies. The PID
// namespace was never what blocked the network, so it buys nothing worth that.
const jailNamespaces = syscall.CLONE_NEWUSER |
	syscall.CLONE_NEWNET |
	syscall.CLONE_NEWNS |
	syscall.CLONE_NEWIPC |
	syscall.CLONE_NEWUTS

// capBoundLast is the highest capability number the bounding-set sweep drops.
// The bounding set is a 64-bit mask, so 63 is the ceiling by construction.
// Sweeping the whole range rather than stopping at today's CAP_LAST_CAP is
// what keeps this correct on a kernel that adds one: the extra calls simply
// return EINVAL and are ignored.
const capBoundLast = 63

// The address families whose socket(2) is refused, split by whether
// --keep-loopback releases them. AF_PACKET (raw sockets) has no legitimate
// loopback use and stays blocked either way; the two IP families have to be
// allowed when loopback is wanted, since 127.0.0.1/::1 are IP addresses.
var (
	alwaysBlockedSocketFamilies = []int{unix.AF_PACKET}
	ipSocketFamilies            = []int{unix.AF_INET, unix.AF_INET6}
)

// blockedNetworkCalls are denied wholesale when loopback is down. With no
// interface and no route the namespace has nothing to reach anyway; refusing
// the syscall turns a silent connect timeout into an immediate EPERM, which is
// far easier to read in a log than a hang.
var blockedNetworkCalls = []string{
	"connect",
	"bind",
	"listen",
	"accept",
	"accept4",
	"sendto",
	"sendmsg",
	"recvfrom",
	"recvmsg",
}

// --------------------------------------------------------------------------- //
// The isolation policy, as chosen on the command line.
// --------------------------------------------------------------------------- //

// sandbox is what the two flags amount to: the parts of the isolation the
// caller asked to relax. It exists so the choice travels as one named thing
// rather than as a pair of positional booleans nobody can read at a call site.
type sandbox struct {
	keepLoopback bool
	logExternal  bool
}

// sandboxFromEnv - Reads the policy back on the far side of the re-exec.
func sandboxFromEnv() sandbox {
	return sandbox{
		keepLoopback: os.Getenv(keepLoopbackEnv) == envOn,
		logExternal:  os.Getenv(logExternalEnv) == envOn,
	}
}

// env - The policy as stage-2 environment entries, ready to append to the
// caller's own environment.
func (s sandbox) env() []string {
	env := []string{stageEnv + "=" + envOn}
	if s.keepLoopback {
		env = append(env, keepLoopbackEnv+"="+envOn)
	}
	if s.logExternal {
		env = append(env, logExternalEnv+"="+envOn)
	}
	return env
}

// blockedFamilies - The address families whose socket(2) this run refuses.
// With --keep-loopback the IP families have to stay usable, since 127.0.0.1
// and ::1 are IP addresses; AF_PACKET is refused either way.
func (s sandbox) blockedFamilies() []int {
	if s.keepLoopback {
		return alwaysBlockedSocketFamilies
	}
	return slices.Concat(alwaysBlockedSocketFamilies, ipSocketFamilies)
}

// --------------------------------------------------------------------------- //
// Stage 1 — namespace setup, on the host.
// --------------------------------------------------------------------------- //

// main - Parses the flags and re-executes this binary inside fresh namespaces,
// or runs the isolated stage when it is already inside them.
func main() {
	if os.Getenv(stageEnv) == envOn {
		runIsolated(sandboxFromEnv())
		return
	}

	keepLoopback := flag.Bool("keep-loopback", false, "keep the loopback interface (127.0.0.1) up and reachable")
	logExternal := flag.Bool("log-external", false, "log blocked network syscalls to stderr")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(exitFailure)
	}

	reexecIsolated(sandbox{keepLoopback: *keepLoopback, logExternal: *logExternal})
}

// reexecIsolated - Re-executes this binary inside fresh namespaces and exits
// with whatever the wrapped program exited with.
func reexecIsolated(s sandbox) {
	self, err := os.Executable()
	if err != nil {
		panic(err)
	}

	cmd := exec.Command(self, flag.Args()...)
	cmd.Env = append(os.Environ(), s.env()...)
	cmd.SysProcAttr = jailAttr()
	wireStdio(cmd)

	runAndExit(cmd)
}

// jailAttr - The clone(2) arguments that build the sandbox: the namespace set,
// and the ID maps that make the caller root inside the user namespace.
//
// setgroups stays off: leaving it on would let the mapped root drop
// supplementary groups it should not control.
func jailAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Cloneflags: jailNamespaces,

		UidMappings: singleIDMap(os.Getuid()),

		GidMappingsEnableSetgroups: false,
		GidMappings:                singleIDMap(os.Getgid()),
	}
}

// singleIDMap - The one-entry map making hostID root inside the user namespace
// and nobody outside it, which is what buys the other namespaces without any
// privilege on the host.
func singleIDMap(hostID int) []syscall.SysProcIDMap {
	return []syscall.SysProcIDMap{{
		ContainerID: namespaceRoot,
		HostID:      hostID,
		Size:        1,
	}}
}

// usage - Prints the whole interface. The flags are all of it, so --help lists
// them rather than only the usage line: a flag nobody can discover may as well
// not exist.
func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [--keep-loopback] [--log-external] <program> [args...]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Runs <program> with its network access fully isolated: private user, network,\n")
	fmt.Fprintf(os.Stderr, "mount, IPC and UTS namespaces, an emptied capability set, and a seccomp\n")
	fmt.Fprintf(os.Stderr, "filter refusing the network syscalls.\n\nFlags:\n")
	flag.PrintDefaults()
}

// --------------------------------------------------------------------------- //
// Stage 2 — inside the namespaces.
// --------------------------------------------------------------------------- //

// runIsolated - Drops what the child must not keep, installs the seccomp
// filter and runs the wrapped program, mirroring back its exit status.
//
// Order matters: no-new-privs first, so nothing later in this function can be
// undone by a setuid binary; loopback before the filter, because bringing "lo"
// up needs the very socket(2) the filter is about to refuse.
func runIsolated(s sandbox) {
	// Prevent privilege escalation through setuid/setcap binaries.
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, noNewPrivsOn, 0, 0, 0); err != nil {
		panic(err)
	}

	if s.keepLoopback {
		if err := bringUpLoopback(); err != nil {
			panic(err)
		}
	}

	dropCapabilities()

	installSeccomp(s)

	if len(os.Args) <= targetArgv {
		fmt.Fprintln(os.Stderr, "offline: no program to run")
		os.Exit(exitFailure)
	}

	target := exec.Command(os.Args[targetArgv], os.Args[targetArgv+1:]...)
	wireStdio(target)

	runAndExit(target)
}

// bringUpLoopback - Sets the "lo" interface UP inside the new network
// namespace. The kernel creates it disabled by default; the namespace still
// has no other interfaces or routes, so this only enables 127.0.0.1/::1
// traffic, not external access.
func bringUpLoopback() error {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
	if err != nil {
		return err
	}
	// Best-effort, like the capability sweep: the fd is going away with this
	// function either way, and a failed close has nothing to report to.
	defer func() { _ = unix.Close(fd) }()

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

// --------------------------------------------------------------------------- //
// Running the wrapped program. Both stages end here.
// --------------------------------------------------------------------------- //

// wireStdio - Hands the caller's own streams to cmd. offline is a wrapper, so
// the wrapped program's stdin/stdout/stderr are the caller's, untouched.
func wireStdio(cmd *exec.Cmd) {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
}

// runAndExit - Runs cmd and exits with its status, or returns if it succeeded.
//
// A program that ran and exited non-zero is not an offline failure, so its
// code is passed through and nothing is added to its stderr; only a failure to
// run the program at all is reported as offline's own. Both stages route
// through here, because the outer one collapsing the status on the way back
// out would undo the inner one's passthrough.
func runAndExit(cmd *exec.Cmd) {
	err := cmd.Run()
	if err == nil {
		return
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		os.Exit(exitCode(exitErr))
	}

	fmt.Fprintln(os.Stderr, err)
	os.Exit(exitFailure)
}

// exitCode - The status to exit with for a program that ran and failed.
// ExitCode reports -1 for a process killed by a signal, which has no exit code
// of its own; 128+signal is what a caller's shell expects there.
func exitCode(exitErr *exec.ExitError) int {
	if code := exitErr.ExitCode(); code >= 0 {
		return code
	}

	status, ok := exitErr.Sys().(syscall.WaitStatus)
	if !ok || !status.Signaled() {
		return exitFailure
	}
	return exitSignalBase + int(status.Signal())
}

// --------------------------------------------------------------------------- //
// Capabilities.
// --------------------------------------------------------------------------- //

// dropCapabilities - Empties the bounding and ambient capability sets, so no
// capability can be regained by exec'ing a file that carries one.
//
// Every call is best-effort on purpose: a number above the running kernel's
// CAP_LAST_CAP, or a set already emptied by an outer sandbox, is an EINVAL
// that says nothing went wrong.
func dropCapabilities() {
	// Remove all inheritable/effective capabilities.
	for capability := 0; capability <= capBoundLast; capability++ {
		_ = unix.Prctl(
			unix.PR_CAPBSET_DROP,
			uintptr(capability),
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

		_ = seccomp.NotifRespond(fd, &seccomp.ScmpNotifResp{
			ID:    req.ID,
			Error: int32(syscall.EPERM),
			Val:   0,
		})
	}
}
