# Getting Started

## Prerequisites

- Linux kernel with user/network/mount/IPC/UTS namespaces and seccomp.
- Go 1.25+ to build.
- `pkg-config` and `libseccomp-dev` (Ubuntu/Debian: `sudo apt install
  pkg-config libseccomp-dev`).

## Setup

```sh
git clone https://github.com/fabiocicerchia/offline
cd offline
make build
```

## Run

```sh
./offline curl https://example.com   # fails: no network in the sandbox
./offline bash                       # an offline shell
```

See the [README](README.md) for the full flag reference, the security
model, and troubleshooting.

## Troubleshooting

**User namespace disabled** — check:

```bash
cat /proc/sys/kernel/unprivileged_userns_clone
```

Enable:

```bash
sudo sysctl kernel.unprivileged_userns_clone=1
```

**libseccomp missing** — error:

```
Package 'libseccomp' not found
```

Fix:

```bash
sudo apt install libseccomp-dev pkg-config
```

## Verification

```bash
./offline bash
```

Inside:

```bash
ip addr
ip route
curl https://example.com
```

Expected: only the loopback interface, no routes, no network access.

## Usage

```bash
./offline [--keep-loopback] [--log-external] <program> [arguments...]
```

- `--keep-loopback` brings `lo` up so 127.0.0.1/::1 traffic works (e.g. a
  local dev server or a database on localhost). External addresses stay
  unreachable — the namespace still has no other interface and no routes.
- `--log-external` logs each blocked network syscall (name + pid) to stderr
  before denying it.

### Block internet access

```bash
./offline curl https://example.com
```

Expected:

```
curl: (6) Could not resolve host: example.com
```

### Inspect network namespace

```bash
./offline ip addr
```

Only the loopback interface is visible:

```
1: lo: <LOOPBACK>
```

### Inspect routes

```bash
./offline ip route
```

Expected: no output — no routes.

### Run an offline shell

```bash
./offline bash
```

Inside, `curl https://example.com` will fail.
