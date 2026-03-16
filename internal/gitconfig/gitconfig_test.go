package gitconfig_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ismaels/git-ssh-sign/internal/gitconfig"
)

// setupGitConfig creates a temp dir with an empty git config file and sets
// GIT_CONFIG_GLOBAL to point at it, so tests never touch the developer's real
// git config. t.Setenv automatically restores the original value after the test.
//
// GIT_CONFIG_GLOBAL requires git 2.32+. This helper skips the test if the
// installed git is older, rather than silently writing to the real config.
func setupGitConfig(t *testing.T) string {
	t.Helper()

	// Check git version >= 2.32 before relying on GIT_CONFIG_GLOBAL.
	out, err := exec.Command("git", "--version").Output()
	if err != nil {
		t.Skip("git not found, skipping")
	}
	// "git version 2.39.0" → parse major.minor
	var major, minor int
	fmt.Sscanf(strings.TrimPrefix(strings.TrimSpace(string(out)), "git version "), "%d.%d", &major, &minor)
	if major < 2 || (major == 2 && minor < 32) {
		t.Skipf("git >= 2.32 required for GIT_CONFIG_GLOBAL isolation (have %d.%d)", major, minor)
	}

	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "gitconfig")
	if err := os.WriteFile(cfgFile, []byte{}, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", cfgFile)
	return cfgFile
}

func TestGitVersion(t *testing.T) {
	v, ok := gitconfig.GitVersion()
	if !ok {
		t.Fatal("GitVersion returned false — is git installed?")
	}
	if v == "" {
		t.Fatal("GitVersion returned empty string")
	}
}

func TestGetReturnsEmptyForMissingKey(t *testing.T) {
	setupGitConfig(t)
	got := gitconfig.Get("user.email")
	if got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestSetAndGet(t *testing.T) {
	setupGitConfig(t)
	if err := gitconfig.Set("user.email", "test@example.com"); err != nil {
		t.Fatal(err)
	}
	got := gitconfig.Get("user.email")
	if got != "test@example.com" {
		t.Fatalf("expected %q, got %q", "test@example.com", got)
	}
}

func TestRead(t *testing.T) {
	setupGitConfig(t)
	if err := gitconfig.Set("gpg.format", "ssh"); err != nil {
		t.Fatal(err)
	}
	cfg := gitconfig.Read()
	if cfg.GPGFormat != "ssh" {
		t.Fatalf("expected GPGFormat %q, got %q", "ssh", cfg.GPGFormat)
	}
	if cfg.SigningKey != "" {
		t.Fatalf("expected empty SigningKey, got %q", cfg.SigningKey)
	}
}

func TestApply(t *testing.T) {
	setupGitConfig(t)
	pairs := map[string]string{
		"gpg.format":     "ssh",
		"commit.gpgsign": "true",
	}
	if err := gitconfig.Apply(pairs); err != nil {
		t.Fatal(err)
	}
	if got := gitconfig.Get("gpg.format"); got != "ssh" {
		t.Fatalf("expected %q, got %q", "ssh", got)
	}
	if got := gitconfig.Get("commit.gpgsign"); got != "true" {
		t.Fatalf("expected %q, got %q", "true", got)
	}
}

func TestSigningConfigFieldMapping(t *testing.T) {
	setupGitConfig(t)
	pairs := map[string]string{
		"gpg.format":                 "ssh",
		"user.signingkey":            "ssh-ed25519 AAAA test",
		"commit.gpgsign":             "true",
		"tag.gpgsign":                "true",
		"gpg.ssh.allowedSignersFile": "/tmp/allowed_signers",
		"gpg.ssh.program":            "/usr/bin/op-ssh-sign",
	}
	if err := gitconfig.Apply(pairs); err != nil {
		t.Fatal(err)
	}
	cfg := gitconfig.Read()
	if cfg.GPGFormat != "ssh" {
		t.Errorf("GPGFormat: got %q", cfg.GPGFormat)
	}
	if cfg.SigningKey != "ssh-ed25519 AAAA test" {
		t.Errorf("SigningKey: got %q", cfg.SigningKey)
	}
	if cfg.CommitGPGSign != "true" {
		t.Errorf("CommitGPGSign: got %q", cfg.CommitGPGSign)
	}
	if cfg.TagGPGSign != "true" {
		t.Errorf("TagGPGSign: got %q", cfg.TagGPGSign)
	}
	if cfg.AllowedSignersFile != "/tmp/allowed_signers" {
		t.Errorf("AllowedSignersFile: got %q", cfg.AllowedSignersFile)
	}
	if cfg.SSHProgram != "/usr/bin/op-ssh-sign" {
		t.Errorf("SSHProgram: got %q", cfg.SSHProgram)
	}
}
