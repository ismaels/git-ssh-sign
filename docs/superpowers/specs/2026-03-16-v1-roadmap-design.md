# v1 Release Roadmap — Design Spec

**Date:** 2026-03-16
**Project:** git-ssh-sign
**Goal:** Ship a tested, polished v1.0.0 release via the existing GoReleaser + Homebrew pipeline.

---

## Current State

- Two commands: `setup` (interactive 4-step wizard) and `verify` (config checker)
- Two internal packages: `internal/gitconfig` and `internal/sshkeys`
- GoReleaser + GitHub Actions release pipeline already configured (tag-triggered)
- Zero tests
- No CI workflow for tests
- `main.version` referenced in ldflags but not declared or surfaced in the binary
- README has no contribution section

---

## Approach

Three ordered milestones. Each must be complete before the next starts. Tag `v1.0.0` only after all three are done and CI is green.

---

## Milestone 1 — Tests

### Strategy: mix of unit and integration tests

**`internal/gitconfig`**

- *Unit tests:* `SigningConfig` struct population and field mapping only
- *Integration tests:* `Get`, `Set`, `GitVersion`, `Read`, `Apply` — run against a real `git` in a `t.TempDir()`-based repo with `GIT_CONFIG_GLOBAL` pointed at a temp file to avoid touching the developer's real git config

**`internal/sshkeys`**

- *Unit tests:* `OnePasswordSignProgram`, `CommonKeyPaths` — pure string returns, no filesystem needed. `AllowedSignersPath` and `OnePasswordAgentSocketPath` are also unit-testable by overriding `HOME` via `t.Setenv("HOME", t.TempDir())` before calling them.
- *Integration tests:* `FindExistingKeys`, `ReadPublicKey` — use `t.TempDir()` with fixture `.pub` key files, override `HOME` via `t.Setenv`. `EnsureAllowedSigners` — override `HOME` via `t.Setenv` so `AllowedSignersPath()` resolves inside `t.TempDir()`; test both the append and idempotent (already-present) paths. `Has1Password` / `Has1PasswordSignBinary` — override `HOME` (for the socket path) and create/omit the expected file in the temp dir to test both true and false branches. Note: `Has1PasswordSignBinary` checks a hardcoded `/Applications/...` path that cannot be redirected via `HOME`; this function is tested on macOS CI only, or skipped on Linux with `t.Skip`.

**`cmd/`**

Not tested directly. The huh TUI forms are interactive and not suitable for automated testing. Coverage of all meaningful logic comes from the internal package tests. The cmd layer is thin wiring only.

---

## Milestone 2 — Version flag

- Declare `var version = "dev"` in `main.go` (the existing ldflags `-X main.version={{.Version}}` already targets this)
- Add an exported `func SetVersion(v string)` in `cmd/root.go` that sets `rootCmd.Version = v`
- Call `cmd.SetVersion(version)` from `main.go` before `cmd.Execute()` — Cobra then automatically handles `--version` / `-v` with no additional code
- No new files needed

---

## Milestone 3 — Docs + CI gate

### CI workflow

New file: `.github/workflows/ci.yml`

- Trigger: push and pull_request to `main`
- Steps: `actions/checkout`, `actions/setup-go` with `go-version-file: go.mod` (resolves the exact minimum version from `go.mod`, consistent with `release.yml`)
- Run `go vet ./...` — any output is a blocking failure
- Run `go test -race ./...` — any test failure is a blocking failure; `-race` is the standard quality bar for Go projects
- The existing `release.yml` is unchanged

**Definition of "CI is green":** the CI workflow job exits 0 (both `go vet` and `go test -race` pass with no failures) on the commit being tagged.

### README — How to Contribute

Add a "How to Contribute" section covering:

1. Clone and build (`go build ./...`, `go run . <command>`)
2. Run tests (`go test ./...`)
3. Submit a PR against `main` — CI must pass

Tone: small team, no issue templates, no CoC.

### Release

Once milestones 1–3 are complete and CI is green on `main`:

```bash
git tag v1.0.0
git push origin v1.0.0
```

GoReleaser builds darwin/linux amd64/arm64 binaries, publishes GitHub release, and updates the Homebrew tap automatically.

---

## Out of Scope for v1

- Testing the `cmd/` layer (requires TUI mocking)
- Windows builds
- Issue templates, code of conduct
- Changelog tooling beyond GoReleaser's built-in changelog
