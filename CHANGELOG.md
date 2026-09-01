# Changelog

All notable changes to this project are documented here. The format is based
on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Bug Fixes

* pass the wrapped program's exit status through the outer stage instead of
  collapsing every failure to `1`, and stop writing `exit status N` to the
  program's own stderr. A program killed by a signal now reports `128+signal`
  rather than `255`.
* drop the `PID` namespace from the `--help` text: `CLONE_NEWPID` was removed
  in 1.1.1 and the help had been overstating what is isolated ever since.

## [1.1.1](https://github.com/fabiocicerchia/offline/compare/v1.1.0...v1.1.1) (2026-08-29)


### Bug Fixes

* drop the PID namespace so /proc describes the sandboxed process ([#48](https://github.com/fabiocicerchia/offline/issues/48)) ([fcb5fa6](https://github.com/fabiocicerchia/offline/commit/fcb5fa605c60b9f5b08df035eec5f37065587910))
* unblock quality and clear the Scorecard pinned-dependencies finding ([#54](https://github.com/fabiocicerchia/offline/issues/54)) ([04218f2](https://github.com/fabiocicerchia/offline/commit/04218f2c990c0262b39ff08c0bd0235dbeaefa21))

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

### Removed

- The PID namespace (`CLONE_NEWPID`) is no longer part of the sandbox. It
  renumbered the wrapped program without giving it a matching procfs, so a
  program looking itself up by `getpid()` read an unrelated host process —
  `/proc/1` is the host init, owned by an unmapped root, which fails with
  `EACCES`. Node and Bun runtimes do exactly that on startup and crashed.
  Mounting a private `/proc` is not an option either: Ubuntu's
  `kernel.apparmor_restrict_unprivileged_userns` confines anyone who creates
  an unprivileged user namespace to a profile that refuses every mount. The
  PID namespace never blocked network access, so nothing about the isolation
  guarantee changes.

## [0.1.0]

### Added

- Network-isolated execution via user/net/mount/PID/IPC/UTS namespaces.
- Privilege hardening: capability dropping and `PR_SET_NO_NEW_PRIVS`.
- Seccomp filter blocking `AF_INET`/`AF_INET6`/`AF_PACKET` sockets and the
  core network syscalls (`connect`, `bind`, `listen`, `accept`, ...).

[Unreleased]: https://github.com/fabiocicerchia/offline/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/fabiocicerchia/offline/releases/tag/v0.1.0
