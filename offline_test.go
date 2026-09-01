package main

import (
	"errors"
	"os/exec"
	"slices"
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
