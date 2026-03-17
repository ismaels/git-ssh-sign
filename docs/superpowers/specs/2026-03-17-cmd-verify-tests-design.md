---
title: cmd/verify testability refactor
date: 2026-03-17
status: approved
---

# cmd/verify testability refactor

## Goal

Add tests to the `cmd` package by extracting the check-building logic in `runVerify` into a pure, testable function.

## Problem

`runVerify` in `cmd/verify.go` mixes business logic (building `[]checkResult`) with I/O (printing to stdout). This makes it impossible to assert on the logic without capturing output.

## Background: relevant types

`checkResult` (already defined in `cmd/verify.go`):

```go
type checkResult struct {
    label string
    value string
    ok    bool
    hint  string
}
```

`gitconfig.SigningConfig` fields used by checks:
- `GPGFormat` — must equal `"ssh"`
- `SigningKey` — must be non-empty
- `CommitGPGSign` — must equal `"true"`
- `TagGPGSign` — must equal `"true"`
- `AllowedSignersFile` — must be non-empty and exist on disk
- `SSHProgram` — must equal `sshkeys.OnePasswordSignProgram()` (only when 1Password present)

`gitOk bool` — passed in from `gitconfig.GitVersion()`: whether git is installed and runnable. Gates the "git installed" check.

## Design

### Extract `buildChecks`

Move the check-building logic out of `runVerify` into a standalone function:

```go
func buildChecks(cfg gitconfig.SigningConfig, gitOk bool, has1P bool, hasBin bool) []checkResult
```

- Takes all inputs as parameters — no side effects, no I/O, no global state
- Returns `[]checkResult` — fully assertable in tests
- `runVerify` calls `buildChecks` and handles all printing (behavior unchanged)
- `fileExists` remains a private helper called inside `buildChecks`

The 1Password checks (labels `"1Password SSH agent"` and `"gpg.ssh.program"`) are only appended when both `has1P` and `hasBin` are true.

### Handling `fileExists` in tests

`buildChecks` calls `fileExists` internally for the `AllowedSignersFile` check. Tests pass a path that does not exist (e.g. `"/nonexistent/allowed_signers"`), so `fileExists` returns false and the check has `ok: false`. This is intentional and consistent across all environments. Tests that want `ok: true` for that check must use `t.TempDir()` to create a real file and pass its path in `cfg.AllowedSignersFile`.

`fileExists` is not made injectable — it is a one-line `os.Stat` wrapper and the path-based approach above is sufficient for test coverage.

### Test file: `cmd/verify_test.go`

**Tests for `buildChecks`:**

| Scenario | Setup | Assertion |
|---|---|---|
| Fully valid config | `cfg` with all fields set correctly (`GPGFormat: "ssh"`, `SigningKey: "ssh-ed25519 AAAA test"`, `CommitGPGSign: "true"`, `TagGPGSign: "true"`, `AllowedSignersFile`: path to a real temp file), `gitOk: true`, `has1P: false`, `hasBin: false` | All checks have `ok: true` |
| `gpg.format` wrong | Same valid config but `GPGFormat: "gpg"` | Check with `label: "gpg.format"` has `ok: false` and `hint` contains `"gpg.format ssh"` |
| Empty `user.signingkey` | `SigningKey: ""` | Check with `label: "user.signingkey"` has `ok: false` |
| `commit.gpgsign` not true | `CommitGPGSign: ""` | Check with `label: "commit.gpgsign"` has `ok: false` |
| 1Password present | `has1P: true`, `hasBin: true` | Result contains checks with labels `"1Password SSH agent"` and `"gpg.ssh.program"` |
| 1Password partial | `has1P: true`, `hasBin: false` | Result does NOT contain a check with label `"1Password SSH agent"` or `"gpg.ssh.program"` |

**Tests for `truncate`:**

| Scenario | Input | Expected |
|---|---|---|
| Shorter than n | `"hi"`, n=10 | `"hi"` |
| Exactly n | `"hello"`, n=5 | `"hello"` |
| Longer than n | `"hello world"`, n=5 | `"hello..."` |

`truncate` is already in `cmd/verify.go` (package `cmd`). Tests in `cmd/verify_test.go` (package `cmd`) access it directly — no export needed.

## Files changed

- `cmd/verify.go` — extract `buildChecks` from `runVerify`; no other changes
- `cmd/verify_test.go` — new file with tests for `buildChecks` and `truncate`

## Behavioral regression risk

`runVerify` is not directly tested (it does I/O). Regression risk from the extraction is low: `buildChecks` is carved out mechanically with the same logic, and `runVerify` is left as a thin loop over its results. The existing manual smoke test (`git-ssh-sign verify`) remains the behavioral check for the print layer.

## Out of scope

- `runSetup` — requires interactive TTY, not addressed here
- `fileExists` standalone tests — covered implicitly via `buildChecks` tests
- `printGitHubTip` — calls `ssh-add`, not testable without environment setup
