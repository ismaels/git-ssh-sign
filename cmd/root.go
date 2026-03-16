package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "git-ssh-sign",
	Short: "Set up and verify SSH commit signing for Git",
	Long: lipgloss.NewStyle().
		Foreground(lipgloss.Color("212")).
		Render(`
git-ssh-sign helps you configure SSH-based commit signing for Git.

SSH signing is simpler than GPG — it uses the same SSH keys you
already use for GitHub, with optional 1Password agent integration.
`),
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
