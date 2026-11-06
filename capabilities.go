package main

import "golang.org/x/sys/unix"

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
		//nolint:errcheck // best-effort: a capability that will not drop is
		// reported by the sweep as a whole, not per bit.
		_ = unix.Prctl(
			unix.PR_CAPBSET_DROP,
			uintptr(capability),
			0,
			0,
			0,
		)
	}

	// Clear ambient capabilities.
	//nolint:errcheck // as above
	_ = unix.Prctl(
		unix.PR_CAP_AMBIENT,
		unix.PR_CAP_AMBIENT_CLEAR_ALL,
		0,
		0,
		0,
	)
}
