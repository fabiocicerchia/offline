# Contributing

Thanks for taking the time to contribute!

## Getting started

1. Fork and clone the repo.
2. Install pre-commit hooks: `pre-commit install` (see `.pre-commit-config.yaml`).
3. Create a branch: `git checkout -b feat/short-description`.

## Making changes

- Keep changes focused; one logical change per PR.
- Update `docs/` and `examples/` when behavior changes.
- Ensure CI (`code-quality` + `security`) passes.
- Namespace, capability, and seccomp-filter changes affect the sandbox's
  security guarantees — call that out explicitly in the PR description.

Don't edit `CHANGELOG.md` by hand — it's generated from commit messages by
release-please (see [Releases](#releases)).

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/): `feat:`,
`fix:`, `docs:`, `chore:`, etc. This keeps history readable and drives the
version bump: `fix:` → patch, `feat:` → minor, `feat!:` or a
`BREAKING CHANGE:` footer → major.

## Releases

Releases are automated by [release-please](.github/workflows/release.yml);
you don't tag or edit the changelog manually.

1. Merge `feat:`/`fix:` PRs into `main` as normal — **no tag is created**.
2. release-please keeps an open **release PR** ("chore: release X.Y.Z"),
   recalculating the next version and changelog on every merge.
3. When you're ready to ship, **merge the release PR** — that (and only that)
   creates the `vX.Y.Z` tag and the GitHub Release.

So `main` is not released per-commit: changes accumulate into the release PR,
and merging it is the deliberate release step.

### Enabling release-please to open the release PR

By default GitHub blocks the built-in `GITHUB_TOKEN` from opening pull
requests, so release-please fails with _"GitHub Actions is not permitted to
create or approve pull requests"_. Pick one of:

- **Repo setting** — enable _Settings → Actions → General → Workflow
  permissions → "Allow GitHub Actions to create and approve pull requests"_.
  The default token then works, no secret needed.
- **Dedicated token** — add a `RELEASE_PLEASE_TOKEN` repository secret holding
  a PAT or GitHub App token with `contents: write` and `pull-requests: write`.
  The workflow uses it automatically when present.

## Pull requests

Fill out the PR template, link related issues, and request review. Be kind.
