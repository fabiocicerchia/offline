# CLAUDE.md

Guidance for Claude Code (and other AI agents) working in this repo.

## Project

`offline` is a Linux sandbox wrapper, written in Go, that runs another
program with its network access fully isolated (user/net/mount/PID/IPC/UTS
namespaces + capability dropping + a seccomp filter blocking network
syscalls). Single-file entry point: `offline.go` (`package main`), with
`offline_test.go` beside it pinning the socket-family policy and the
exit-status passthrough.

## Commands

```sh
# build: go build -o offline offline.go
# test:  go test ./...
# lint:  gofmt -l . && go vet ./...
# run:   ./offline <program> [args...]
make help    # Show this help
make setup   # Install the pre-commit hook
make lint    # Run all pre-commit checks on the whole tree
make test    # Run the test suite
make build   # Build the offline binary
make install # Build and drop the binary in BINDIR (default ~/.local/bin)
```

Requires `libseccomp-dev`/`pkg-config` on the host (cgo binding); see README.

## Tooling

- `make setup` installs the pre-commit hook, and that is the whole of it.
  Don't add a `.githooks/` directory: `core.hooksPath` replaces `.git/hooks/`
  wholesale, so setting it silently stops every pre-commit hook from running.
- Hooks are pinned by commit SHA with the tag in a trailing comment. A tag can
  be moved, a SHA cannot.
- CI runs this same `.pre-commit-config.yaml` through `pre-commit/action`, so
  what passes locally is what gates the pull request.

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
