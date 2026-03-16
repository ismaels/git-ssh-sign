package sshkeys_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ismaels/git-ssh-sign/internal/sshkeys"
)

// --- Unit tests for pure functions ---

func TestOnePasswordSignProgram(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("OnePasswordSignProgram returns a macOS path, skipping on non-darwin")
	}
	got := sshkeys.OnePasswordSignProgram()
	want := "/Applications/1Password.app/Contents/MacOS/op-ssh-sign"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestCommonKeyPaths(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	paths := sshkeys.CommonKeyPaths()
	if len(paths) != 3 {
		t.Fatalf("expected 3 paths, got %d", len(paths))
	}
	for _, p := range paths {
		if !strings.HasPrefix(p, fakeHome) {
			t.Errorf("HOME isolation failed: path %q does not start with fakeHome %q", p, fakeHome)
		}
		if !strings.HasSuffix(p, ".pub") {
			t.Errorf("path %q does not end with .pub", p)
		}
	}
}

func TestAllowedSignersPath(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	got := sshkeys.AllowedSignersPath()
	want := filepath.Join(fakeHome, ".ssh", "allowed_signers")
	if got != want {
		t.Fatalf("expected %q, got %q (HOME isolation may have failed)", want, got)
	}
}

func TestOnePasswordAgentSocketPath(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	got := sshkeys.OnePasswordAgentSocketPath()
	if !strings.HasPrefix(got, fakeHome) {
		t.Fatalf("HOME isolation failed: expected path under %q, got %q", fakeHome, got)
	}
	if !strings.HasSuffix(got, "agent.sock") {
		t.Fatalf("expected path to end with agent.sock, got %q", got)
	}
}

// --- Integration tests: FindExistingKeys and ReadPublicKey ---

func TestFindExistingKeys_nonePresent(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	got := sshkeys.FindExistingKeys()
	if len(got) != 0 {
		t.Fatalf("expected no keys, got %v", got)
	}
}

func TestFindExistingKeys_somePresent(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	sshDir := filepath.Join(fakeHome, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}
	keyPath := filepath.Join(sshDir, "id_ed25519.pub")
	if err := os.WriteFile(keyPath, []byte("ssh-ed25519 AAAA test\n"), 0600); err != nil {
		t.Fatal(err)
	}

	got := sshkeys.FindExistingKeys()
	if len(got) != 1 {
		t.Fatalf("expected 1 key, got %v", got)
	}
	if got[0] != keyPath {
		t.Fatalf("expected %q, got %q", keyPath, got[0])
	}
}

func TestReadPublicKey(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "id_ed25519.pub")
	content := "ssh-ed25519 AAAA test"
	if err := os.WriteFile(keyPath, []byte(content+"\n"), 0600); err != nil {
		t.Fatal(err)
	}

	got, err := sshkeys.ReadPublicKey(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	if got != content {
		t.Fatalf("expected %q, got %q", content, got)
	}
}

func TestReadPublicKey_missingFile(t *testing.T) {
	_, err := sshkeys.ReadPublicKey("/nonexistent/path/key.pub")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// --- Integration tests: EnsureAllowedSigners ---

func TestEnsureAllowedSigners_appendsNewEntry(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	// EnsureAllowedSigners writes to AllowedSignersPath() which is under HOME.
	// Pre-create the .ssh dir because the function only creates the file, not the parent.
	sshDir := filepath.Join(fakeHome, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}

	email := "user@example.com"
	pubKey := "ssh-ed25519 AAAA test"

	if err := sshkeys.EnsureAllowedSigners(email, pubKey); err != nil {
		t.Fatal(err)
	}

	allowedPath := sshkeys.AllowedSignersPath()
	content, err := os.ReadFile(allowedPath)
	if err != nil {
		t.Fatal(err)
	}
	want := email + " " + pubKey + "\n"
	if string(content) != want {
		t.Fatalf("expected %q, got %q", want, string(content))
	}
}

func TestEnsureAllowedSigners_idempotent(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	sshDir := filepath.Join(fakeHome, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}

	email := "user@example.com"
	pubKey := "ssh-ed25519 AAAA test"

	if err := sshkeys.EnsureAllowedSigners(email, pubKey); err != nil {
		t.Fatal(err)
	}
	if err := sshkeys.EnsureAllowedSigners(email, pubKey); err != nil {
		t.Fatal(err)
	}

	allowedPath := sshkeys.AllowedSignersPath()
	content, err := os.ReadFile(allowedPath)
	if err != nil {
		t.Fatal(err)
	}
	want := email + " " + pubKey + "\n"
	if string(content) != want {
		t.Fatalf("expected entry exactly once, got %q", string(content))
	}
}

func TestEnsureAllowedSigners_existingOtherKey(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	sshDir := filepath.Join(fakeHome, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}

	email := "user@example.com"
	otherKey := "ssh-rsa BBBB other"
	newKey := "ssh-ed25519 AAAA new"

	allowedPath := sshkeys.AllowedSignersPath()
	existing := email + " " + otherKey + "\n"
	if err := os.WriteFile(allowedPath, []byte(existing), 0600); err != nil {
		t.Fatal(err)
	}

	if err := sshkeys.EnsureAllowedSigners(email, newKey); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(allowedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), otherKey) {
		t.Errorf("original key missing from file: %q", string(content))
	}
	if !strings.Contains(string(content), newKey) {
		t.Errorf("new key not appended to file: %q", string(content))
	}
}

// --- Integration tests: Has1Password ---

func TestHas1Password_false(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)
	// Socket path resolves under HOME — temp dir has no socket
	if sshkeys.Has1Password() {
		t.Fatal("expected Has1Password to return false with empty temp HOME")
	}
}

func TestHas1Password_true(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	socketPath := sshkeys.OnePasswordAgentSocketPath()
	if err := os.MkdirAll(filepath.Dir(socketPath), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(socketPath, []byte{}, 0600); err != nil {
		t.Fatal(err)
	}

	if !sshkeys.Has1Password() {
		t.Fatal("expected Has1Password to return true when socket file exists")
	}
}

func TestHas1PasswordSignBinary(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip("Has1PasswordSignBinary checks a hardcoded macOS path, skipping on Linux")
	}
	// Verify it returns a bool without panicking; actual value depends on environment.
	_ = sshkeys.Has1PasswordSignBinary()
}
