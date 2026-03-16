# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build ./...
go run . <command>

# Install locally
go install .
```

## Testing

```bash
go test ./...
go test ./internal/gitconfig/...   # single package
go test -run TestFunctionName ./...  # single test
```

## Lint

```bash
go vet ./...
```

## Architecture

This is a Go CLI tool built with [Cobra](https://github.com/spf13/cobra) for command structure and [Charmbracelet Huh](https://github.com/charmbracelet/huh) for interactive TUI forms.

**Entry point:** `main.go` → `cmd.Execute()`

**`cmd/` — CLI layer**
- `root.go`: Registers the root `git-ssh-sign` command
- `setup.go`: Interactive 4-step wizard (`runSetup`); handles key selection, config preview, and application
- `verify.go`: Reports current SSH signing config status with pass/fail checks

**`internal/gitconfig/` — Git config abstraction**
- Thin wrapper around `git config --global` via `exec.Command`
- `SigningConfig` struct mirrors all relevant git SSH signing keys
- `Apply(map[string]string)` writes multiple keys atomically (sequential, no rollback)

**`internal/sshkeys/` — SSH key and 1Password detection**
- Probes `~/.ssh/id_ed25519.pub`, `id_rsa.pub`, `id_ecdsa.pub` for existing keys
- 1Password detection checks for the agent socket at `~/Library/Group Containers/2BUA8C4S2C.com.1password/t/agent.sock` and the sign binary at `/Applications/1Password.app/...`
- `EnsureAllowedSigners` appends to `~/.ssh/allowed_signers` only if the key isn't already present

**Module path:** `github.com/ismaels/git-ssh-sign` (go.mod) but imports in `cmd/` currently use `github.com/wecodepages/git-ssh-sign` — these must stay consistent if changed.
