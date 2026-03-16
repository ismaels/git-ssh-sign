package gitconfig

import (
	"os/exec"
	"strings"
)

// Get reads a git config value globally.
func Get(key string) string {
	out, err := exec.Command("git", "config", "--global", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// Set writes a git config value globally.
func Set(key, value string) error {
	return exec.Command("git", "config", "--global", key, value).Run()
}

// GitVersion returns the installed git version string.
func GitVersion() (string, bool) {
	out, err := exec.Command("git", "--version").Output()
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(out)), true
}

// SigningConfig holds all SSH signing related git config values.
type SigningConfig struct {
	GPGFormat          string
	SigningKey          string
	CommitGPGSign      string
	TagGPGSign         string
	AllowedSignersFile string
	SSHProgram         string
}

// Read reads the current SSH signing config from git globals.
func Read() SigningConfig {
	return SigningConfig{
		GPGFormat:          Get("gpg.format"),
		SigningKey:          Get("user.signingkey"),
		CommitGPGSign:      Get("commit.gpgsign"),
		TagGPGSign:         Get("tag.gpgsign"),
		AllowedSignersFile: Get("gpg.ssh.allowedSignersFile"),
		SSHProgram:         Get("gpg.ssh.program"),
	}
}

// Apply writes the provided config map to git globals.
func Apply(pairs map[string]string) error {
	for k, v := range pairs {
		if err := Set(k, v); err != nil {
			return err
		}
	}
	return nil
}
