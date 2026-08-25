# Changelog

All notable changes to this project are documented here. The format is based
on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0](https://github.com/fabiocicerchia/offline/compare/v1.0.2...v1.1.0) (2026-08-25)


### Features

* **docs:** build the docs site in Actions and drop Read the Docs ([#46](https://github.com/fabiocicerchia/offline/issues/46)) ([3f10500](https://github.com/fabiocicerchia/offline/commit/3f1050053a9699c379191c1cf4952b883c26ab4b))

## [1.0.2](https://github.com/fabiocicerchia/offline/compare/v1.0.1...v1.0.2) (2026-08-13)


### Bug Fixes

* security and code-quality findings ([#31](https://github.com/fabiocicerchia/offline/issues/31)) ([557bba0](https://github.com/fabiocicerchia/offline/commit/557bba0215ccc2d480f55f706ffeef36652d39e1))

## [1.0.1](https://github.com/fabiocicerchia/offline/compare/v1.0.0...v1.0.1) (2026-08-06)


### Bug Fixes

* **deps:** pin the Go toolchain to 1.26.5 for the stdlib CVE fixes ([b5ea232](https://github.com/fabiocicerchia/offline/commit/b5ea23221892c67c6739b0213ace91094a756967))
* **pre-commit:** stop check-yaml failing on Helm templates and multi-doc manifests ([dad11c0](https://github.com/fabiocicerchia/offline/commit/dad11c0e44c63cae41ca6f302d2d7e235e1381d7))
* **security:** skip the SARIF upload on private repos ([bf5f667](https://github.com/fabiocicerchia/offline/commit/bf5f66731e6b94db16a5b2c001b561352de84990))

## 1.0.0 (2026-08-01)


### Features

* add --keep-loopback and --log-external flags ([169af4c](https://github.com/fabiocicerchia/offline/commit/169af4cbdd256caddfb10344004a7005efc76a43))

## [Unreleased]

### Added

- `--keep-loopback` flag: brings `lo` up inside the network namespace so
  127.0.0.1/::1 traffic works, while external addresses stay unreachable
  (no other interface, no routes).
- `--log-external` flag: logs each blocked network syscall (name + pid) to
  stderr before denying it, via a seccomp `ActNotify` filter instead of
  a plain errno.

## [0.1.0]

### Added

- Network-isolated execution via user/net/mount/PID/IPC/UTS namespaces.
- Privilege hardening: capability dropping and `PR_SET_NO_NEW_PRIVS`.
- Seccomp filter blocking `AF_INET`/`AF_INET6`/`AF_PACKET` sockets and the
  core network syscalls (`connect`, `bind`, `listen`, `accept`, ...).

[Unreleased]: https://github.com/fabiocicerchia/offline/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/fabiocicerchia/offline/releases/tag/v0.1.0
