# Architecture

`offline` re-executes itself to build a sandbox, then hands off to the
wrapped program inside it.

## Overview

`offline.go` (`package main`) has two modes, selected by the `_AIRGAP_STAGE`
env var:

1. **Outer stage** (`main`, no env var set): re-execs the current binary with
   `_AIRGAP_STAGE=1` inside a `clone()` call that creates new user, network,
   mount, PID, IPC, and UTS namespaces, mapping the caller's uid/gid into the
   new user namespace.
2. **Inner stage** (`runIsolated`, env var set): running inside those fresh
   namespaces, it locks down privileges before finally exec'ing the target
   program the caller asked for.

## Components

- **`main`** — parses args, sets up `syscall.SysProcAttr.Cloneflags` and
  UID/GID mappings, re-execs itself into the sandbox.
- **`runIsolated`** — sets `PR_SET_NO_NEW_PRIVS`, drops capabilities, installs
  the seccomp filter, then `exec.Command` + `Run()`s the target program.
- **`dropCapabilities`** — `PR_CAPBSET_DROP` for every capability bit, clears
  ambient capabilities.
- **`installSeccomp`** — builds an allow-by-default `libseccomp` filter that
  denies `AF_INET`/`AF_INET6`/`AF_PACKET` sockets and the core network
  syscalls (`connect`, `bind`, `listen`, `accept`, `accept4`, `send*`,
  `recv*`).

## Data flow

```
offline <cmd> [args]
  → re-exec self (_AIRGAP_STAGE=1) inside clone(CLONE_NEWUSER|NEWNET|NEWNS|NEWPID|NEWIPC|NEWUTS)
    → runIsolated(): no-new-privs → drop caps → seccomp filter
      → exec <cmd> [args]
```

## Decisions

- Two-stage re-exec (rather than a helper process) keeps the binary
  single-file with no external dependencies beyond the two Go modules.
- Namespaces provide the isolation boundary; seccomp is defense-in-depth in
  case a namespace escape or a loophole (e.g. `AF_UNIX`/`AF_NETLINK`, which
  stay allowed for local tooling) is found.
- See the README's "Limitations" section for what this does *not* cover
  (filesystem isolation, resource limits, kernel exploit protection).
