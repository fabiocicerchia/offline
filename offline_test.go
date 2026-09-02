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
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

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
