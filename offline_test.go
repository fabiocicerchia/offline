// Tests for offline.
//
// What is reachable from a test process and what is not:
//
//	reachable      flag parsing and the usage/exit contract, the capability
//	               ceiling, and building the seccomp filter
//	unreachable    everything after the clone — the namespace re-exec,
//	               dropCapabilities and filter.Load() all need a user
//	               namespace, and a host that refuses one (unshare -Urn →
//	               EPERM/ENOSPC) skips straight past them
//
// So the tests here assert on what a function returns or on the rule set it
// built, never on the state of this process.
package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"golang.org/x/sys/unix"
)

// The socket-family policy is the security guarantee in one function, so it is
// the one thing worth pinning: AF_PACKET is refused whatever the flags say,
// and --keep-loopback is the only thing that releases the IP families.
func TestBlockedFamilies(t *testing.T) {
	for _, s := range []sandbox{{}, {keepLoopback: true}, {logExternal: true}} {
		if !slices.Contains(s.blockedFamilies(), unix.AF_PACKET) {
			t.Errorf("%+v: AF_PACKET must stay blocked", s)
		}
	}

	for _, family := range ipSocketFamilies {
		if !slices.Contains(sandbox{}.blockedFamilies(), family) {
			t.Errorf("family %d must be blocked by default", family)
		}
		if slices.Contains(sandbox{keepLoopback: true}.blockedFamilies(), family) {
			t.Errorf("family %d must be usable with --keep-loopback", family)
		}
	}
}

// A wrapped program's status is offline's status: an exit code passes through
// as itself, and a signal death as the shells' 128+signal.
func TestExitCode(t *testing.T) {
	for _, tc := range []struct {
		script string
		want   int
	}{
		{"exit 7", 7},
		{"exit 1", exitFailure},
		{"kill -TERM $$", exitSignalBase + int(syscall.SIGTERM)},
	} {
		err := exec.Command("sh", "-c", tc.script).Run()

		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("sh -c %q: want an ExitError, got %v", tc.script, err)
		}
		if got := exitCode(exitErr); got != tc.want {
			t.Errorf("sh -c %q: exitCode = %d, want %d", tc.script, got, tc.want)
		}
	}
}

// The flags survive the re-exec: what env() writes, sandboxFromEnv() reads
// back. A break here silently drops --keep-loopback or --log-external on the
// far side of the exec, where nothing else would notice.
func TestSandboxSurvivesTheEnvRoundTrip(t *testing.T) {
	for _, want := range []sandbox{
		{},
		{keepLoopback: true},
		{logExternal: true},
		{keepLoopback: true, logExternal: true},
	} {
		// Read the marker off env()'s return, not off the process: t.Setenv
		// leaks across iterations, so an assertion on os.Getenv would still
		// pass from the second one onwards if env() stopped emitting it.
		entries := want.env()
		if !slices.Contains(entries, stageEnv+"="+envOn) {
			t.Errorf("%+v: env() = %q, must carry the stage marker", want, entries)
		}

		t.Setenv(keepLoopbackEnv, "")
		t.Setenv(logExternalEnv, "")
		for _, entry := range entries {
			key, value, _ := strings.Cut(entry, "=")
			t.Setenv(key, value)
		}

		if got := sandboxFromEnv(); got != want {
			t.Errorf("sandboxFromEnv() = %+v, want %+v", got, want)
		}
	}
}

// TestCommandLine - Checks the usage and exit-code contract of stage 1, which
// is reached before any namespace is created and is therefore the one part of
// main() a test can run. The binary is built here rather than re-implemented,
// so the flags asserted on are the flags a user gets.
func TestCommandLine(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "offline")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}

	// The "\n  -flag\n" forms come from flag.PrintDefaults and so name the
	// flags that are really registered; the bracketed form comes from the
	// hand-written usage line. Asserting both is what catches the two
	// drifting apart.
	usage := []string{
		"usage:",
		"[--keep-loopback] [--log-external] <program> [args...]",
		"\n  -keep-loopback\n",
		"\n  -log-external\n",
	}

	cases := []struct {
		name     string
		args     []string
		wantCode int
		wantErr  []string
	}{
		{
			name:     "no program is a usage error",
			args:     nil,
			wantCode: 1,
			wantErr:  usage,
		},
		{
			name:     "help is not an error",
			args:     []string{"--help"},
			wantCode: 0,
			wantErr:  usage,
		},
		{
			name:     "an unknown flag is refused, not passed through",
			args:     []string{"--no-such-flag", "true"},
			wantCode: 2,
			wantErr:  []string{"flag provided but not defined", "usage:"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(bin, tc.args...)
			var stderr strings.Builder
			cmd.Stderr = &stderr

			// A non-nil error here is the wrapped program's exit status,
			// which is what the case is asserting on; only a process that
			// never ran leaves ProcessState nil.
			_ = cmd.Run()
			if cmd.ProcessState == nil {
				t.Fatalf("%s did not run", bin)
			}

			if code := cmd.ProcessState.ExitCode(); code != tc.wantCode {
				t.Errorf("exit code = %d, want %d (stderr: %q)", code, tc.wantCode, stderr.String())
			}
			for _, want := range tc.wantErr {
				if !strings.Contains(stderr.String(), want) {
					t.Errorf("stderr missing %q, got:\n%s", want, stderr.String())
				}
			}
		})
	}
}

// TestCapabilitySweepCoversKernelRange - The bounding-set sweep must reach at
// least the running kernel's highest capability number; anything above
// capBoundLast is a capability the child keeps. Lowering the constant to trim
// "wasted" iterations is exactly the change this catches.
func TestCapabilitySweepCoversKernelRange(t *testing.T) {
	raw, err := os.ReadFile("/proc/sys/kernel/cap_last_cap")
	if err != nil {
		t.Skipf("no /proc/sys/kernel/cap_last_cap on this host: %v", err)
	}

	kernelLast, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatalf("cap_last_cap is not a number: %q", raw)
	}

	if capBoundLast < kernelLast {
		t.Errorf("capBoundLast = %d, so capabilities %d..%d stay in the bounding set",
			capBoundLast, capBoundLast+1, kernelLast)
	}
}

// The two actions the filter denies with, as libseccomp's pseudo filter code
// prints them. ERRNO(1) is EPERM; SCMP_ACT_NOTIFY has no printed name in
// libseccomp 2.5, so PFC shows its raw value.
const (
	denyEPERM  = "ERRNO(1)"
	denyNotify = "0x7fc00000"
)

// The address family numbers, spelled out rather than taken from unix.AF_*, so
// this file states the policy independently of the file it is checking.
const (
	afINET   = "2"
	afINET6  = "10"
	afPACKET = "17"
)

// TestFilterDeniesTheNetwork - Reads the policy back out of the filter that
// would be installed. This is the whole reason buildSeccompFilter is separate
// from installSeccomp: loading a filter needs a user namespace this process
// does not have, but exporting one needs nothing.
//
// The expected sets below are written out by hand on purpose. Deriving them
// from alwaysBlockedSocketFamilies and blockedNetworkCalls would make the test
// agree with whatever those tables happen to say.
func TestFilterDeniesTheNetwork(t *testing.T) {
	networkCalls := []string{
		"connect", "bind", "listen", "accept", "accept4",
		"sendto", "sendmsg", "recvfrom", "recvmsg",
	}

	cases := []struct {
		name         string
		keepLoopback bool
		logExternal  bool
		deny         string
	}{
		{name: "isolated", deny: denyEPERM},
		{name: "isolated, logging", logExternal: true, deny: denyNotify},
		{name: "loopback kept", keepLoopback: true, deny: denyEPERM},
		{name: "loopback kept, logging", keepLoopback: true, logExternal: true, deny: denyNotify},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// The default stays ALLOW in every case: this filter denies by
			// exception, it is not an allowlist.
			want := map[string]string{"default": "ALLOW"}

			// AF_PACKET has no loopback use, so it is refused either way.
			want["socket("+afPACKET+")"] = tc.deny

			if !tc.keepLoopback {
				// 127.0.0.1 and ::1 are IP addresses, so the IP families can
				// only be refused when loopback is not wanted.
				want["socket("+afINET+")"] = tc.deny
				want["socket("+afINET6+")"] = tc.deny
				for _, call := range networkCalls {
					want[call] = tc.deny
				}
			}

			got := exportedPolicy(t, sandbox{keepLoopback: tc.keepLoopback, logExternal: tc.logExternal})

			for key, action := range want {
				if got[key] != action {
					t.Errorf("%s: action = %q, want %q", key, got[key], action)
				}
			}
			for key, action := range got {
				if _, expected := want[key]; !expected {
					t.Errorf("%s: unexpected rule, action %q", key, action)
				}
			}
		})
	}
}

// exportedPolicy - Builds the filter and reads its pseudo filter code back as
// a rule -> action map. Keys are a syscall name, or `socket(<family>)` for the
// conditional socket(2) rules; `default` is the fall-through action.
func exportedPolicy(t *testing.T, s sandbox) map[string]string {
	t.Helper()

	dump, err := os.Create(filepath.Join(t.TempDir(), "filter.pfc"))
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer dump.Close()

	if err := buildSeccompFilter(s).ExportPFC(dump); err != nil {
		t.Fatalf("export: %v", err)
	}

	raw, err := os.ReadFile(dump.Name())
	if err != nil {
		t.Fatalf("read back: %v", err)
	}

	policy := map[string]string{}
	rule, family := "", ""

	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "#"):
			// Every comment opens a new section, so it ends the previous
			// rule; the architecture footer must not inherit socket's name.
			rule, family = "", ""
			if _, name, ok := strings.Cut(line, `filter for syscall "`); ok {
				rule, _, _ = strings.Cut(name, `"`)
			}
			if strings.Contains(line, "default action") {
				rule = "default"
			}
		case strings.HasPrefix(line, "if ($a0.lo32 == "):
			family = strings.TrimSuffix(strings.TrimPrefix(line, "if ($a0.lo32 == "), ")")
		case strings.HasPrefix(line, "action ") && rule != "":
			key := rule
			if family != "" {
				key += "(" + family + ")"
			}
			policy[key] = strings.TrimSuffix(strings.TrimPrefix(line, "action "), ";")
		}
	}

	if len(policy) == 0 {
		t.Fatalf("no rules parsed out of:\n%s", raw)
	}
	return policy
}
