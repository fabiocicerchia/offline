# CLAUDE.md

Guidance for Claude Code (and other AI agents) working in this repo.

## Project

`offline` is a Linux sandbox wrapper, written in Go, that runs another
program with its network access fully isolated (user/net/mount/PID/IPC/UTS
namespaces + capability dropping + a seccomp filter blocking network
syscalls). Single-file entry point: `offline.go` (`package main`).

## Commands

```sh
# build: go build -o offline offline.go
# test:  go test ./...
# lint:  gofmt -l . && go vet ./...
# run:   ./offline <program> [args...]
```

Requires `libseccomp-dev`/`pkg-config` on the host (cgo binding); see README.

## Conventions

- Match existing style; don't reformat unrelated code. `gofmt` is
  non-negotiable.
- Conventional Commits for messages (see CONTRIBUTING.md).
- Update CHANGELOG.md (`## [Unreleased]`), docs/, and examples/ with behavior
  changes.
- Never commit secrets; CI runs gitleaks. Keep `.env` out of git.

## Guardrails

- This is a security sandbox: changes to namespace flags, capability
  dropping, or the seccomp filter directly affect the isolation guarantees —
  treat them as security-sensitive, not routine refactors.
- Don't add dependencies without a clear reason; prefer stdlib
  (`golang.org/x/sys/unix` and `github.com/seccomp/libseccomp-golang` are the
  only two, both load-bearing).
- Ask before large refactors or destructive operations.
