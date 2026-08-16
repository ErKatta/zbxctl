package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion scripts for zbxctl commands.

To load completions:

Bash:
  $ source <(zbxctl completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ zbxctl completion bash > /etc/bash_completion.d/zbxctl
  # macOS:
  $ zbxctl completion bash > $(brew --prefix)/etc/bash_completion.d/zbxctl

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ zbxctl completion zsh > "${fpath[1]}/_zbxctl"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ zbxctl completion fish | source

  # To load completions for each session, execute once:
  $ zbxctl completion fish > ~/.config/fish/completions/zbxctl.fish

PowerShell:
  PS> zbxctl completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> zbxctl completion powershell > zbxctl.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCompletion(cmd.OutOrStdout(), cmd.Root(), args[0])
	},
}

func runCompletion(out io.Writer, root *cobra.Command, shell string) error {
	switch shell {
	case "bash":
		return root.GenBashCompletion(out)
	case "zsh":
		return root.GenZshCompletion(out)
	case "fish":
		return root.GenFishCompletion(out, true)
	case "powershell":
		return root.GenPowerShellCompletionWithDesc(out)
	default:
		return fmt.Errorf("unsupported shell type %q: must be one of [bash, zsh, fish, powershell]", shell)
	}
}

func init() {
	RootCmd.AddCommand(completionCmd)
}
