# Getting Started

## Prerequisites

- Linux kernel with user/network/mount/PID/IPC/UTS namespaces and seccomp.
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

See the [README](../README.md) for the full flag reference, the security
model, and troubleshooting.
