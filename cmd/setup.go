package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/ismaels/git-ssh-sign/internal/gitconfig"
	"github.com/ismaels/git-ssh-sign/internal/sshkeys"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive wizard to configure SSH commit signing",
	RunE:  runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

// styles
var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	codeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Background(lipgloss.Color("236")).Padding(0, 1)
)

func runSetup(cmd *cobra.Command, args []string) error {
	fmt.Println(titleStyle.Render("\n⚙  git-ssh-sign setup wizard\n"))

	// ── Step 1: Preflight ───────────────────────────────────────────────────
	fmt.Println(dimStyle.Render("Step 1/4 · Checking environment"))

	gitVersion, ok := gitconfig.GitVersion()
	if !ok {
		return fmt.Errorf("%s", errorStyle.Render("✗ git not found. Please install git first."))
	}
	fmt.Println(successStyle.Render("✓ " + gitVersion))

	userEmail := gitconfig.Get("user.email")
	if userEmail == "" {
		return fmt.Errorf("%s", errorStyle.Render("✗ git user.email is not set. Run: git config --global user.email you@example.com"))
	}
	fmt.Println(successStyle.Render("✓ git email: " + userEmail))

	has1P := sshkeys.Has1Password() && sshkeys.Has1PasswordSignBinary()
	if has1P {
		fmt.Println(successStyle.Render("✓ 1Password SSH agent detected"))
	} else {
		fmt.Println(warnStyle.Render("  1Password SSH agent not detected (will use system SSH)"))
	}

	// ── Step 2: Key source ──────────────────────────────────────────────────
	fmt.Println(dimStyle.Render("\nStep 2/4 · Choose signing key"))

	existingKeys := sshkeys.FindExistingKeys()

	var keySourceOptions []huh.Option[string]
	for _, kp := range existingKeys {
		pub, err := sshkeys.ReadPublicKey(kp)
		if err == nil {
			keySourceOptions = append(keySourceOptions,
				huh.NewOption(kp, pub))
		}
	}
	keySourceOptions = append(keySourceOptions,
		huh.NewOption("Paste a public key manually", "__paste__"))

	var selectedKey string

	keyForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Which SSH public key should Git use for signing?").
				Options(keySourceOptions...).
				Value(&selectedKey),
		),
	)

	if err := keyForm.Run(); err != nil {
		return err
	}

	if selectedKey == "__paste__" {
		var pasted string
		pasteForm := huh.NewForm(
			huh.NewGroup(
				huh.NewText().
					Title("Paste your SSH public key").
					Description("e.g. from 1Password or ssh-keygen output").
					Value(&pasted),
			),
		)
		if err := pasteForm.Run(); err != nil {
			return err
		}
		selectedKey = pasted
	}

	// ── Step 3: Confirm config ──────────────────────────────────────────────
	fmt.Println(dimStyle.Render("\nStep 3/4 · Review changes"))

	allowedSignersPath := sshkeys.AllowedSignersPath()
	configMap := map[string]string{
		"gpg.format":                 "ssh",
		"user.signingkey":            selectedKey,
		"commit.gpgsign":             "true",
		"tag.gpgsign":                "true",
		"gpg.ssh.allowedSignersFile": allowedSignersPath,
	}

	if has1P {
		configMap["gpg.ssh.program"] = sshkeys.OnePasswordSignProgram()
	}

	fmt.Println("\nThe following will be written to your global git config:")
	for k, v := range configMap {
		display := v
		if len(display) > 60 {
			display = display[:57] + "..."
		}
		fmt.Printf("  %s = %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Render(k),
			dimStyle.Render(display),
		)
	}
	fmt.Printf("  %s → append entry for %s\n",
		lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Render("~/.ssh/allowed_signers"),
		dimStyle.Render(userEmail),
	)

	var confirmed bool
	confirmForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Apply these settings?").
				Value(&confirmed),
		),
	)
	if err := confirmForm.Run(); err != nil {
		return err
	}

	if !confirmed {
		fmt.Println(warnStyle.Render("\nAborted. No changes made."))
		return nil
	}

	// ── Step 4: Apply ───────────────────────────────────────────────────────
	fmt.Println(dimStyle.Render("\nStep 4/4 · Applying configuration"))

	if err := gitconfig.Apply(configMap); err != nil {
		return fmt.Errorf("failed to write git config: %w", err)
	}
	fmt.Println(successStyle.Render("✓ Git config updated"))

	if err := sshkeys.EnsureAllowedSigners(userEmail, selectedKey); err != nil {
		return fmt.Errorf("failed to update allowed_signers: %w", err)
	}
	fmt.Println(successStyle.Render("✓ allowed_signers updated"))

	// ── Done ────────────────────────────────────────────────────────────────
	fmt.Println(titleStyle.Render("\n✓ SSH signing is configured!\n"))
	fmt.Println("Test it with:")
	fmt.Println(codeStyle.Render("  git commit --allow-empty -m \"test ssh signing\""))
	fmt.Println(codeStyle.Render("  git log --show-signature -1"))

	if !has1P {
		fmt.Println(warnStyle.Render("\nTip: Add the following to your ~/.ssh/config for agent support:"))
		fmt.Println(dimStyle.Render("Host *\n  IdentityAgent ~/.ssh/agent.sock"))
	}

	// Offer a test commit
	fmt.Println()
	var runTest bool
	testForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Run a test empty commit now to verify?").
				Value(&runTest),
		),
	)
	if err := testForm.Run(); err != nil {
		return err
	}

	if runTest {
		out, err := exec.Command("git", "commit", "--allow-empty", "-m", "chore: test SSH signing").CombinedOutput()
		if err != nil {
			fmt.Println(errorStyle.Render("✗ Test commit failed:"))
			fmt.Fprintln(os.Stderr, string(out))
		} else {
			fmt.Println(successStyle.Render("✓ Test commit succeeded!"))
		}
	}

	return nil
}
