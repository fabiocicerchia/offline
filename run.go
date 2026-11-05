package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

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
