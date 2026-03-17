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

## Design

### Extract `buildChecks`

Move the check-building logic out of `runVerify` into a standalone function:

```go
func buildChecks(cfg gitconfig.SigningConfig, gitOk bool, has1P bool, hasBin bool) []checkResult
```

- Takes all inputs as parameters — no side effects, no I/O, no global state
- Returns `[]checkResult` — fully assertable in tests
- `runVerify` calls `buildChecks` and handles all printing (unchanged behavior)
- `fileExists` remains a private helper, called inside `buildChecks`

### Test file: `cmd/verify_test.go`

Tests for `buildChecks`:

| Scenario | What's checked |
|---|---|
| Fully valid config | All checks have `ok: true` |
| `gpg.format` not `"ssh"` | That check has `ok: false` and correct hint |
| Empty `user.signingkey` | That check has `ok: false` |
| Empty `commit.gpgsign` | That check has `ok: false` |
| `has1P && hasBin` both true | 1Password checks are appended |
| `has1P` true, `hasBin` false | 1Password checks are NOT appended |

Tests for `truncate`:

| Scenario | Expected |
|---|---|
| String shorter than n | Returned unchanged |
| String exactly n chars | Returned unchanged |
| String longer than n | Truncated with `"..."` suffix |

### No mocking required

`buildChecks` is pure. `fileExists` returns false for any path used in tests (no real files on disk), which is acceptable — the tests control `cfg` inputs and assert on check presence/absence, not file resolution.

## Files changed

- `cmd/verify.go` — extract `buildChecks` from `runVerify`
- `cmd/verify_test.go` — new file with tests

## Out of scope

- `runSetup` — requires interactive TTY, not addressed here
- `fileExists` and `printGitHubTip` — no standalone tests
