# 📴 offline

[![code-quality](https://github.com/fabiocicerchia/offline/actions/workflows/code-quality.yml/badge.svg)](https://github.com/fabiocicerchia/offline/actions/workflows/code-quality.yml)
[![security](https://github.com/fabiocicerchia/offline/actions/workflows/security.yml/badge.svg)](https://github.com/fabiocicerchia/offline/actions/workflows/security.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/fabiocicerchia/offline/badge)](https://securityscorecards.dev/viewer/?uri=github.com/fabiocicerchia/offline)
[![CI carbon](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/fabiocicerchia/offline/gh-pages/badge.json)](.github/workflows/carbon-badge.yml)
[![Release](https://img.shields.io/github/v/release/fabiocicerchia/offline)](https://github.com/fabiocicerchia/offline/releases)

`offline` is a Linux sandbox wrapper that executes another program with its
network access completely isolated: user/net/mount/IPC/UTS namespaces,
capability dropping, and a seccomp filter blocking network syscalls.

> **Linux only, and it isolates the network — nothing else.** The wrapped
> program keeps your filesystem and your environment; what it cannot do is
> reach anything off the machine. Unprivileged user namespaces must be
> enabled on the host.

```bash
./offline curl https://example.com
```

The wrapped process runs inside a restricted environment where it cannot
access the internet.

## How it works

One binary, two stages. The first cannot isolate itself — a running Go program
is multithreaded, and `unshare(CLONE_NEWUSER)` refuses a multithreaded caller —
so it re-executes itself and passes the flags through the environment:

```
  offline [flags] <program> [args...]
      │
      │  stage 1 — on the host
      ├─ parse flags
      └─ clone(CLONE_NEWUSER|NEWNET|NEWNS|NEWIPC|NEWUTS)
             uid/gid mapped 1:1, setgroups off
             _AIRGAP_STAGE=1 in the environment
             │
             ▼
         re-exec of the same binary
             │
             │  stage 2 — inside the namespaces
             ├─ prctl(PR_SET_NO_NEW_PRIVS)      no setuid escape hatch
             ├─ --keep-loopback ? ifup "lo"     before the filter, it needs socket(2)
             ├─ PR_CAPBSET_DROP 0..63           bounding set emptied
             ├─ PR_CAP_AMBIENT_CLEAR_ALL        ambient set emptied
             ├─ seccomp: allow by default, deny
             │     socket(AF_INET|AF_INET6|AF_PACKET)
             │     connect bind listen accept accept4 send* recv*
             └─ exec <program> [args...]
```

Three layers, so a gap in one is not a gap in the sandbox: the namespace
leaves nothing to route to, the empty capability set leaves nothing to
reconfigure it with, and the filter refuses the syscalls anyway.

## Features

- **Network namespace isolation**, so there is nothing to reach: no physical
  interfaces, no routes, no default gateway, no DNS connectivity, no inbound
  connections. `offline ip route` prints nothing.
- **User namespace isolation** with a single-entry uid/gid map — the caller is
  root inside and nobody outside, which is what buys the other namespaces
  without any privilege on the host. Caveat: it needs unprivileged user
  namespaces enabled; some hardened distros and Docker profiles turn them off.
- **Mount namespace isolation**, so mounts made inside the sandbox never reach
  the host. Caveat: the host filesystem itself is still visible and writable —
  this tool takes away the network, not the disk.
- **IPC namespace isolation**, so IPC objects are not shared with the host.
  No PID namespace: it renumbers processes without a matching `/proc`, which
  breaks runtimes that look themselves up there, and it never blocked the
  network anyway.
- **UTS namespace isolation**, so the hostname is the sandbox's own.
- **Capability dropping**, bounding *and* ambient set, so no capability can be
  regained by exec'ing a file that carries one.
- **Seccomp filtering** as the last layer, allow-by-default and deny-by
  -exception: the goal is to stop networking, not to enumerate an allowlist the
  wrapped program would trip over. Denials are `EPERM`.
- **`--keep-loopback`** for the local case — 127.0.0.1 works, external
  addresses still fail at the routing layer. Caveat: with loopback up the
  `connect`/`bind` family has to be allowed, so the filter is thinner here;
  `AF_PACKET` stays blocked either way.
- **`--log-external`** routes denials through a userspace notifier and prints
  them. Caveat: it logs the syscall name and pid, not the destination address.

## Requirements

Linux kernel with support for user, network, mount, and IPC namespaces,
plus seccomp.

Ubuntu/Debian dependencies:

```bash
sudo apt install pkg-config libseccomp-dev
```

Go dependencies are pinned in `go.mod` (`github.com/seccomp/libseccomp-golang`,
`golang.org/x/sys`) and resolved automatically by `go build`/`go mod tidy`.

## Install

**One-liner** (clones/updates a checkout under `~/.local/share/offline` and
runs `make install`):

```bash
curl -fsSL https://raw.githubusercontent.com/fabiocicerchia/offline/main/install.sh | bash
```

**Build from source:**

```bash
go build -o offline offline.go
# or
make build
```

```bash
make install                       # drops the binary in ~/.local/bin (on your PATH)
make install BINDIR=/usr/local/bin # ...or anywhere else
```

## Usage

```console
$ offline --help
usage: offline [--keep-loopback] [--log-external] <program> [args...]

Runs <program> with its network access fully isolated: private user, network,
mount, IPC and UTS namespaces, an emptied capability set, and a seccomp
filter refusing the network syscalls.

Flags:
  -keep-loopback
    	keep the loopback interface (127.0.0.1) up and reachable
  -log-external
    	log blocked network syscalls to stderr
```

A real run — the wrapped `curl` gets no further than the filter:

```console
$ offline --log-external curl -s -m 3 https://example.com
offline: blocked connect (pid 9)
offline: blocked connect (pid 9)
offline: blocked socket (pid 9)
offline: blocked socket (pid 9)
exit status 7
```

```sh
offline curl https://example.com   # fails: no network in the sandbox
offline bash                       # an offline shell
```

More in [`docs/getting-started.md`](docs/getting-started.md).

## Common errors

**`Package libseccomp was not found in the pkg-config search path.`**
The cgo binding needs the development headers, not just the runtime library:
`sudo apt install pkg-config libseccomp-dev` (or `libseccomp-devel` on
Fedora/RHEL).

**`fork/exec /usr/local/bin/offline: operation not permitted`**
The host refuses unprivileged user namespaces. Check
`sysctl kernel.unprivileged_userns_clone` and, on Ubuntu,
`sysctl kernel.apparmor_restrict_unprivileged_userns`; a container runtime can
also block the clone flags through its own seccomp profile.

**`exit status N`**
Not an offline failure — the wrapped program exited non-zero and offline
relays that on stderr. `curl` returning 7 ("failed to connect") is the sandbox
working.

## Documentation

Full docs live in [`docs/`](docs/) (mkdocs). Runnable examples live in
[`examples/`](examples/).

## References

The design is a straight reading of the manual pages; they are the reference,
not this file.

- [`user_namespaces(7)`](https://man7.org/linux/man-pages/man7/user_namespaces.7.html)
  — why the uid/gid map is a single entry and why `setgroups` stays off.
- [`network_namespaces(7)`](https://man7.org/linux/man-pages/man7/network_namespaces.7.html)
  — what a fresh netns contains (a down `lo`, and nothing else).
- [`capabilities(7)`](https://man7.org/linux/man-pages/man7/capabilities.7.html)
  — the bounding and ambient sets, and why both have to be emptied.
- [`seccomp(2)`](https://man7.org/linux/man-pages/man2/seccomp.2.html) and the
  kernel's [seccomp filter documentation](https://www.kernel.org/doc/html/latest/userspace-api/seccomp_filter.html)
  — including `SECCOMP_RET_USER_NOTIF`, which `--log-external` uses.
- [`libseccomp-golang`](https://github.com/seccomp/libseccomp-golang) — the
  binding, and the reason there is a cgo dependency at all.

## Release cycle

[Semantic Versioning](https://semver.org/). Releases are cut by
release-please from [Conventional Commits](https://www.conventionalcommits.org/),
and the tag is the only source of truth for a version.

- **Major** — a change to what is isolated, i.e. to the security guarantee.
- **Minor** — new flags or behaviour that leaves the guarantee intact.
- **Patch** — fixes; only the latest minor gets them.

Every change to the namespace flags, the capability sweep or the seccomp
filter is security-sensitive by definition and is treated as such, whatever
the version number says.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). By participating you agree to the
[Code of Conduct](CODE_OF_CONDUCT.md).

## Security

Found a vulnerability, especially a network-isolation bypass? See
[SECURITY.md](SECURITY.md) — please don't open a public issue.

## Support

Need help implementing this? [Get in touch](https://fabiocicerchia.it/contact).

## License

[MIT](LICENSE) © Fabio Cicerchia
