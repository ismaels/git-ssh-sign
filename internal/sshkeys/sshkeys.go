package sshkeys

import (
	"os"
	"path/filepath"
	"strings"
)

// CommonKeyPaths returns the default SSH public key paths to probe.
func CommonKeyPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, ".ssh", "id_ed25519.pub"),
		filepath.Join(home, ".ssh", "id_rsa.pub"),
		filepath.Join(home, ".ssh", "id_ecdsa.pub"),
	}
}

// FindExistingKeys returns public key paths that exist on disk.
func FindExistingKeys() []string {
	var found []string
	for _, p := range CommonKeyPaths() {
		if _, err := os.Stat(p); err == nil {
			found = append(found, p)
		}
	}
	return found
}

// ReadPublicKey reads the contents of a public key file.
func ReadPublicKey(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

// OnePasswordAgentSocketPath returns the 1Password SSH agent socket path for macOS.
func OnePasswordAgentSocketPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "Group Containers",
		"2BUA8C4S2C.com.1password", "t", "agent.sock")
}

// OnePasswordSignProgram returns the 1Password SSH signing program path.
func OnePasswordSignProgram() string {
	return "/Applications/1Password.app/Contents/MacOS/op-ssh-sign"
}

// Has1Password checks whether the 1Password SSH agent socket exists.
func Has1Password() bool {
	_, err := os.Stat(OnePasswordAgentSocketPath())
	return err == nil
}

// Has1PasswordSignBinary checks whether the op-ssh-sign binary is present.
func Has1PasswordSignBinary() bool {
	_, err := os.Stat(OnePasswordSignProgram())
	return err == nil
}

// AllowedSignersPath returns the default allowed signers file path.
func AllowedSignersPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ssh", "allowed_signers")
}

// EnsureAllowedSigners appends an entry to the allowed signers file if not already present.
func EnsureAllowedSigners(email, pubKey string) error {
	path := AllowedSignersPath()
	entry := email + " " + pubKey + "\n"

	existing, _ := os.ReadFile(path)
	if strings.Contains(string(existing), pubKey) {
		return nil // already present
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(entry)
	return err
}
