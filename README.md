# 📴 offline

[![code-quality](https://github.com/fabiocicerchia/offline/actions/workflows/code-quality.yml/badge.svg)](https://github.com/fabiocicerchia/offline/actions/workflows/code-quality.yml)
[![security](https://github.com/fabiocicerchia/offline/actions/workflows/security.yml/badge.svg)](https://github.com/fabiocicerchia/offline/actions/workflows/security.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/fabiocicerchia/offline/badge)](https://securityscorecards.dev/viewer/?uri=github.com/fabiocicerchia/offline)
[![CI carbon](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/fabiocicerchia/offline/gh-pages/badge.json)](.github/workflows/carbon-badge.yml)
[![Release](https://img.shields.io/github/v/release/fabiocicerchia/offline)](https://github.com/fabiocicerchia/offline/releases)

`offline` is a Linux sandbox wrapper that executes another program with its
network access completely isolated: user/net/mount/PID/IPC/UTS namespaces,
capability dropping, and a seccomp filter blocking network syscalls.

```bash
./offline curl https://example.com
```

The wrapped process runs inside a restricted environment where it cannot
access the internet.

## Features

### Network namespace isolation

The child process runs inside a private network namespace:

- no physical network interfaces
- no Wi-Fi/Ethernet access
- no routes
- no default gateway
- no DNS connectivity
- no inbound connections

```bash
./offline ip route
```

returns no routes.

### User namespace isolation

The process runs inside its own user namespace:

- separate UID/GID mapping
- reduced privileges
- no inherited capabilities

### Mount namespace isolation

The process receives a private mount namespace. Mount operations performed
inside the sandbox do not affect the host.

### PID namespace isolation

The wrapped process receives its own PID namespace. Processes inside the
sandbox cannot see host processes.

### IPC namespace isolation

IPC resources are isolated from the host.

### UTS namespace isolation

The process receives an isolated hostname namespace.

## Requirements

Linux kernel with support for user, network, PID, mount, and IPC namespaces,
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

```sh
./offline curl https://example.com   # fails: no network in the sandbox
./offline bash                       # an offline shell
```

More in [`docs/getting-started.md`](docs/getting-started.md).

## Documentation

Full docs live in [`docs/`](docs/) (mkdocs). Runnable examples live in
[`examples/`](examples/).

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
