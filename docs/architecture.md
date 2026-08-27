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
  → re-exec self (_AIRGAP_STAGE=1) inside clone(CLONE_NEWUSER|NEWNET|NEWNS|NEWIPC|NEWUTS)
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
docs/            mkdocs documentation + GitHub Pages landing page (docs/index.html)
examples/        runnable examples
.github/         CI workflows, issue/PR templates, dependabot
```
