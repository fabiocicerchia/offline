# 📴 offline

[![code-quality](https://github.com/fabiocicerchia/offline/actions/workflows/code-quality.yml/badge.svg)](https://github.com/fabiocicerchia/offline/actions/workflows/code-quality.yml)
[![security](https://github.com/fabiocicerchia/offline/actions/workflows/security.yml/badge.svg)](https://github.com/fabiocicerchia/offline/actions/workflows/security.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/fabiocicerchia/offline/badge)](https://securityscorecards.dev/viewer/?uri=github.com/fabiocicerchia/offline)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Ffabiocicerchia%2Foffline.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Ffabiocicerchia%2Foffline?ref=badge_shield)
[![Release](https://img.shields.io/github/v/release/fabiocicerchia/offline)](https://github.com/fabiocicerchia/offline/releases)

`offline` is a Linux sandbox wrapper that executes another program with its network access completely isolated.

Example:

```bash
./offline curl https://example.com
````

The wrapped process runs inside a restricted environment where it cannot access the internet.

---

## Features

### Network namespace isolation

The child process runs inside a private network namespace:

* no physical network interfaces
* no Wi-Fi/Ethernet access
* no routes
* no default gateway
* no DNS connectivity
* no inbound connections

Example:

```bash
./offline ip route
```

returns no routes.

---

### User namespace isolation

The process runs inside its own user namespace:

* separate UID/GID mapping
* reduced privileges
* no inherited capabilities

---

### Mount namespace isolation

The process receives a private mount namespace.

Mount operations performed inside the sandbox do not affect the host.

---

### PID namespace isolation

The wrapped process receives its own PID namespace.

Processes inside the sandbox cannot see host processes.

---

### IPC namespace isolation

IPC resources are isolated from the host.

---

### UTS namespace isolation

The process receives an isolated hostname namespace.

---

## Privilege hardening

### Capability dropping

`offline` removes Linux capabilities:

* no capability inheritance
* reduced privilege surface
* prevents privileged operations

### No new privileges

The wrapper enables:

```
PR_SET_NO_NEW_PRIVS
```

This prevents:

* setuid privilege escalation
* file capability escalation
* privilege gain during `exec`

---

## Seccomp filtering

`offline` applies a seccomp filter.

### Blocked socket families

The following socket families are denied:

```
AF_INET
AF_INET6
AF_PACKET
```

This prevents:

* IPv4 networking
* IPv6 networking
* raw packet access

### Allowed socket families

Allowed:

```
AF_UNIX
AF_NETLINK
```

This keeps local IPC working and allows commands like:

```bash
ip addr
ip route
```

to inspect the isolated namespace.

### Blocked network syscalls

The following syscalls are denied:

```
connect()
bind()
listen()
accept()
accept4()
sendto()
sendmsg()
recvfrom()
recvmsg()
```

---

# Requirements

Linux kernel with support for:

* user namespaces
* network namespaces
* PID namespaces
* mount namespaces
* IPC namespaces
* seccomp

Ubuntu/Debian dependencies:

```bash
sudo apt install pkg-config libseccomp-dev
```

Go dependencies are pinned in `go.mod` (`github.com/seccomp/libseccomp-golang`,
`golang.org/x/sys`) and resolved automatically by `go build`/`go mod tidy`.

---

# Build

```bash
go build -o offline offline.go
# or
make build
```

Install a `offline` command:

```bash
make install                       # drops the binary in ~/.local/bin (on your PATH)
make install BINDIR=/usr/local/bin # ...or anywhere else
```

Or the one-line installer (clones/updates a checkout under
`~/.local/share/offline` and runs `make install`):

```bash
curl -fsSL https://raw.githubusercontent.com/fabiocicerchia/offline/main/install.sh | bash
```

---

# Usage

Syntax:

```bash
./offline <program> [arguments...]
```

Examples:

## Block internet access

```bash
./offline curl https://example.com
```

Expected:

```
curl: (6) Could not resolve host: example.com
```

---

## Inspect network namespace

```bash
./offline ip addr
```

Example:

```
1: lo: <LOOPBACK>
```

No external interfaces are visible.

---

## Inspect routes

```bash
./offline ip route
```

Expected:

```
(no output)
```

---

## Run an offline shell

```bash
./offline bash
```

Inside:

```bash
curl https://example.com
```

will fail.

---

# Security model

The isolation layers:

```
                 Host
                  |
                  |
              offline
                  |
    +-------------+-------------+
    |                           |
User namespace             PID namespace
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

---

# Limitations

`offline` is designed to prevent network access.

It does not currently provide:

* full filesystem isolation
* read-only filesystem enforcement
* resource limits
* malware analysis containment
* kernel exploit protection

A kernel compromise could bypass namespace isolation.

For stronger sandboxing add:

* minimal root filesystem
* read-only mounts
* cgroups
* Landlock
* AppArmor/SELinux policies
* seccomp allowlist mode

---

# Troubleshooting

## User namespace disabled

Check:

```bash
cat /proc/sys/kernel/unprivileged_userns_clone
```

Enable:

```bash
sudo sysctl kernel.unprivileged_userns_clone=1
```

---

## libseccomp missing

Error:

```
Package 'libseccomp' not found
```

Fix:

```bash
sudo apt install libseccomp-dev pkg-config
```

---

# Verification

Run:

```bash
./offline bash
```

Inside:

```bash
ip addr
ip route
curl https://example.com
```

Expected:

* only loopback interface
* no routes
* no network access

---

# Project layout

```
offline.go       the sandbox wrapper (single file, package main)
docs/            mkdocs documentation + GitHub Pages landing page (docs/index.html)
examples/        runnable examples
.github/         CI workflows, issue/PR templates, dependabot
```

# Documentation

Full docs live in [`docs/`](docs/) (mkdocs). Runnable examples live in
[`examples/`](examples/).

# Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). By participating you agree to the
[Code of Conduct](CODE_OF_CONDUCT.md).

# Security

Found a vulnerability, especially a network-isolation bypass? See
[SECURITY.md](SECURITY.md) — please don't open a public issue.

# License

[MIT](LICENSE) © Fabio Cicerchia
