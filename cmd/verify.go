package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ismaels/git-ssh-sign/internal/gitconfig"
	"github.com/ismaels/git-ssh-sign/internal/sshkeys"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Check your current SSH signing configuration",
	RunE:  runVerify,
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}

type checkResult struct {
	label string
	value string
	ok    bool
	hint  string
}

func runVerify(_ *cobra.Command, _ []string) error {
	fmt.Println(titleStyle.Render("\n🔍 git-ssh-sign verify\n"))

	cfg := gitconfig.Read()
	_, gitOk := gitconfig.GitVersion()

	checks := []checkResult{
		{
			label: "git installed",
			value: "git",
			ok:    gitOk,
			hint:  "Install git from https://git-scm.com",
		},
		{
			label: "gpg.format",
			value: cfg.GPGFormat,
			ok:    cfg.GPGFormat == "ssh",
			hint:  `Run: git config --global gpg.format ssh`,
		},
		{
			label: "user.signingkey",
			value: truncate(cfg.SigningKey, 60),
			ok:    cfg.SigningKey != "",
			hint:  "Run setup to configure a signing key",
		},
		{
			label: "commit.gpgsign",
			value: cfg.CommitGPGSign,
			ok:    cfg.CommitGPGSign == "true",
			hint:  `Run: git config --global commit.gpgsign true`,
		},
		{
			label: "tag.gpgsign",
			value: cfg.TagGPGSign,
			ok:    cfg.TagGPGSign == "true",
			hint:  `Run: git config --global tag.gpgsign true`,
		},
		{
			label: "gpg.ssh.allowedSignersFile",
			value: cfg.AllowedSignersFile,
			ok:    cfg.AllowedSignersFile != "" && fileExists(cfg.AllowedSignersFile),
			hint:  "Run setup to create the allowed signers file",
		},
	}

	// 1Password checks (informational)
	has1P := sshkeys.Has1Password()
	hasBin := sshkeys.Has1PasswordSignBinary()
	if has1P && hasBin {
		checks = append(checks, checkResult{
			label: "1Password SSH agent",
			value: "detected",
			ok:    true,
		})
		checks = append(checks, checkResult{
			label: "gpg.ssh.program",
			value: cfg.SSHProgram,
			ok:    cfg.SSHProgram == sshkeys.OnePasswordSignProgram(),
			hint:  fmt.Sprintf("Run: git config --global gpg.ssh.program %q", sshkeys.OnePasswordSignProgram()),
		})
	}

	allOk := true
	for _, c := range checks {
		icon := successStyle.Render("✓")
		valueStr := dimStyle.Render(c.value)
		if !c.ok {
			icon = errorStyle.Render("✗")
			valueStr = warnStyle.Render(c.value)
			allOk = false
		}
		fmt.Printf("  %s  %-35s %s\n", icon,
			lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Render(c.label),
			valueStr,
		)
		if !c.ok && c.hint != "" {
			fmt.Printf("       %s\n", dimStyle.Render("→ "+c.hint))
		}
	}

	fmt.Println()

	if allOk {
		fmt.Println(successStyle.Render("✓ Everything looks good!"))
		printGitHubTip(cfg)
	} else {
		fmt.Println(warnStyle.Render("Some settings are missing or incorrect."))
		fmt.Printf("Run %s to fix them.\n", codeStyle.Render("git-ssh-sign setup"))
	}

	return nil
}

func printGitHubTip(cfg gitconfig.SigningConfig) {
	out, err := exec.Command("ssh-add", "-L").Output()
	if err != nil {
		return
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	signingKeyShort := strings.Fields(cfg.SigningKey)

	for _, line := range lines {
		for _, f := range strings.Fields(line) {
			if len(signingKeyShort) > 1 && f == signingKeyShort[1] {
				return
			}
		}
	}

	fmt.Println()
	fmt.Println(dimStyle.Render("Tip: Make sure your public key is added to GitHub as a Signing Key:"))
	fmt.Println(dimStyle.Render("  Settings → SSH and GPG keys → New SSH key → Key type: Signing Key"))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
