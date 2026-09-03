# Architecture

`offline` re-executes itself to build a sandbox, then hands off to the
wrapped program inside it.

## Overview

`offline.go` (`package main`) has two modes, selected by the `_AIRGAP_STAGE`
env var:

1. **Outer stage** (`main`, no env var set): re-execs the current binary with
   `_AIRGAP_STAGE=1` inside a `clone()` call that creates new user, network,
   mount, IPC, and UTS namespaces, mapping the caller's uid/gid into the
   new user namespace.
2. **Inner stage** (`runIsolated`, env var set): running inside those fresh
   namespaces, it locks down privileges before finally exec'ing the target
   program the caller asked for.

## Components

- **`sandbox`** — the two flags as one named value, and the only thing that
  crosses the re-exec: `env()` writes it into the child's environment,
  `sandboxFromEnv()` reads it back, `blockedFamilies()` turns it into the
  socket families the filter refuses.
- **`main`** — parses the flags, or dispatches to the isolated stage when the
  `_AIRGAP_STAGE` marker says it is already inside the namespaces.
- **`reexecIsolated`** / **`jailAttr`** — `jailAttr` is the `clone()` argument
  set (`Cloneflags` plus the single-entry UID/GID maps); `reexecIsolated`
  re-runs this binary under it.
- **`runIsolated`** — sets `PR_SET_NO_NEW_PRIVS`, drops capabilities, installs
  the seccomp filter, then runs the target program.
- **`dropCapabilities`** — `PR_CAPBSET_DROP` for every capability bit, clears
  ambient capabilities.
- **`installSeccomp`** — builds an allow-by-default `libseccomp` filter, then
  delegates to `blockSocketFamilies` (`AF_INET`/`AF_INET6`/`AF_PACKET`,
  by conditional rule on socket(2)'s family argument) and `blockNetworkCalls`
  (`connect`, `bind`, `listen`, `accept`, `accept4`, `send*`, `recv*`).
- **`runAndExit`** / **`exitCode`** — where both stages end. The wrapped
  program's status is offline's status; see "Exit status" below.

## Data flow

```
offline <cmd> [args]
  → re-exec self (_AIRGAP_STAGE=1) inside clone(CLONE_NEWUSER|NEWNET|NEWNS|NEWIPC|NEWUTS)
    → runIsolated(): no-new-privs → drop caps → seccomp filter
      → exec <cmd> [args]
```

## Exit status

offline is a wrapper, so the wrapped program's status is what a caller sees:

| Situation | Exit status |
|---|---|
| the program ran | its own exit code, unchanged |
| the program was killed by a signal | `128 + signal`, the shells' convention |
| offline could not run it, or was invoked wrong | `1`, with a reason on stderr |

Both stages route through `runAndExit`, because the outer stage collapsing the
status on the way back out would undo the inner stage's passthrough. Nothing is
written to stderr for a program that ran and failed — that is its own business,
and a wrapper adding a line to it would corrupt the program's output.

## Decisions

- Two-stage re-exec (rather than a helper process) keeps the binary
  single-file with no external dependencies beyond the two Go modules.
- Namespaces provide the isolation boundary; seccomp is defense-in-depth in
  case a namespace escape or a loophole (e.g. `AF_UNIX`/`AF_NETLINK`, which
  stay allowed for local tooling) is found.
- See the README's "Limitations" section for what this does *not* cover
  (filesystem isolation, resource limits, kernel exploit protection).

## Privilege hardening

### Capability dropping

`offline` removes Linux capabilities:

- no capability inheritance
- reduced privilege surface
- prevents privileged operations

### No new privileges

The wrapper enables `PR_SET_NO_NEW_PRIVS`, which prevents:

- setuid privilege escalation
- file capability escalation
- privilege gain during `exec`

## Seccomp filtering

`offline` applies a seccomp filter.

**Blocked socket families:** `AF_INET`, `AF_INET6`, `AF_PACKET` — no IPv4/IPv6
networking, no raw packet access.

**Allowed socket families:** `AF_UNIX`, `AF_NETLINK` — keeps local IPC working
and allows commands like `ip addr`/`ip route` to inspect the isolated
namespace.

**Blocked network syscalls:** `connect()`, `bind()`, `listen()`, `accept()`,
`accept4()`, `sendto()`, `sendmsg()`, `recvfrom()`, `recvmsg()`.

## Security model

The isolation layers:

```
                 Host
                  |
              offline
                  |
    +-------------+-------------+
    |                           |
User namespace             IPC namespace
    |
Network namespace
    |
No interfaces
No routes
No DNS
    |
Seccomp filter
    |
No network syscalls
```

With `--keep-loopback`, `lo` is brought up and its socket families/syscalls
are left out of the seccomp filter; the namespace still has no other
interface and no routes, so only 127.0.0.1/::1 traffic works.

## Limitations

`offline` is designed to prevent network access. It does not currently
provide:

- full filesystem isolation
- read-only filesystem enforcement
- resource limits
- malware analysis containment
- kernel exploit protection

A kernel compromise could bypass namespace isolation. For stronger sandboxing
add a minimal root filesystem, read-only mounts, cgroups, Landlock,
AppArmor/SELinux policies, or seccomp allowlist mode.

## Project layout

```
offline.go       the sandbox wrapper (single file, package main)
offline_test.go  pins the socket-family policy and the exit-status passthrough
docs/            mkdocs documentation + GitHub Pages landing page (docs/index.html)
examples/        runnable examples
.github/         CI workflows, issue/PR templates, dependabot
```
