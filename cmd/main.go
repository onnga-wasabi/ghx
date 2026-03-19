package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/onnga-wasabi/ghx/internal/api"
	"github.com/onnga-wasabi/ghx/internal/auth"
	"github.com/onnga-wasabi/ghx/internal/repo"
	"github.com/onnga-wasabi/ghx/internal/tui"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "ghx",
		Short:   "A TUI dashboard for GitHub",
		Version: version + " (" + commit + ")",
		RunE:    run,
	}

	rootCmd.Flags().StringP("repo", "R", "", "owner/repo to use (defaults to current git repo)")

	completionCmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script",
		Long: `Generate shell completion script for ghx.

  # zsh
  ghx completion zsh > "${fpath[1]}/_ghx"

  # bash
  ghx completion bash > /etc/bash_completion.d/ghx

  # fish
  ghx completion fish > ~/.config/fish/completions/ghx.fish`,
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}
	rootCmd.AddCommand(completionCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	token, err := auth.GetToken()
	if err != nil {
		return fmt.Errorf("authentication failed: %w\nRun 'gh auth login' to authenticate", err)
	}

	owner, repoName, err := resolveRepo(cmd)
	if err != nil {
		return err
	}

	client := api.NewClient(token)
	app := tui.NewApp(client, owner, repoName)
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err = p.Run()
	return err
}

func resolveRepo(cmd *cobra.Command) (string, string, error) {
	repoFlag, _ := cmd.Flags().GetString("repo")
	if repoFlag != "" {
		parts := splitRepo(repoFlag)
		if parts == nil {
			return "", "", fmt.Errorf("invalid repo format %q, expected owner/repo", repoFlag)
		}
		return parts[0], parts[1], nil
	}
	return repo.Detect()
}

func splitRepo(s string) []string {
	for i, c := range s {
		if c == '/' {
			if i > 0 && i < len(s)-1 {
				return []string{s[:i], s[i+1:]}
			}
			return nil
		}
	}
	return nil
}
