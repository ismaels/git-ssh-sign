# git-ssh-sign

An interactive CLI wizard to configure SSH-based commit signing for Git — with optional 1Password integration.

SSH signing is simpler than GPG: it uses the same SSH keys you already have, no separate keychain required.

## Install

```bash
brew install ismaels/tap/git-ssh-sign
```

Or build from source:

```bash
go install github.com/ismaels/git-ssh-sign@latest
```

## Usage

### Set up SSH signing

```bash
git-ssh-sign setup
```

The wizard will:
1. Detect your git version and email
2. Check for 1Password SSH agent
3. Let you choose or paste a public key
4. Preview all git config changes before applying
5. Update `~/.ssh/allowed_signers`
6. Optionally run a test commit

### Verify existing config

```bash
git-ssh-sign verify
```

Checks all required settings and reports any that are missing or misconfigured.

## What it configures

| Git config key | Value |
|---|---|
| `gpg.format` | `ssh` |
| `user.signingkey` | your public key |
| `commit.gpgsign` | `true` |
| `tag.gpgsign` | `true` |
| `gpg.ssh.allowedSignersFile` | `~/.ssh/allowed_signers` |
| `gpg.ssh.program` | 1Password binary *(if detected)* |

## 1Password integration

If 1Password is installed with its SSH agent active, the wizard automatically configures `gpg.ssh.program` to route signing through the 1Password agent — giving you biometric approval per commit.

## After setup

Add your public key to GitHub as a **Signing Key**:
> Settings → SSH and GPG keys → New SSH key → Key type: **Signing Key**

Commits will show as **Verified** on GitHub.

## How to Contribute

```bash
git clone https://github.com/ismaels/git-ssh-sign.git
cd git-ssh-sign
go build ./...
go run . setup  # or: go run . verify
```

Run tests:

```bash
go test -race ./...
```

Submit a PR against `main`. CI must pass before merging.

## License

MIT
