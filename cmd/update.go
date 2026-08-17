package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/ErKatta/zbxctl/pkg/update"
	"github.com/spf13/cobra"
)

var (
	updateCheckFlag  bool
	updateForceFlag  bool
	updateRepoFlag   string
	updateTagFlag    string
	updateTargetFlag string
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for updates and self-update zbxctl binary in-place",
	Long: `update queries GitHub releases for the latest zbxctl version, verifies SHA256 checksums,
and updates the executable in-place at its current installation path.
After updating, it reminds the user to rerun 'zbxctl skill install --all' to update AI agent skills.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		currentVer := Version
		if currentVer == "" {
			currentVer = "0.0.0-dev"
		}

		targetExec := updateTargetFlag
		if targetExec == "" {
			var err error
			targetExec, err = update.GetCurrentExecutablePath()
			if err != nil {
				return err
			}
		}

		outputExplicit, _ := cmd.Root().PersistentFlags().GetString("output")
		if !cmd.Root().PersistentFlags().Changed("output") {
			outputExplicit = ""
		}
		isMachineOutput := outputExplicit == "json" || outputExplicit == "yaml" || outputExplicit == "toon"

		if !isMachineOutput {
			if updateCheckFlag {
				fmt.Fprintf(formatter.Writer, "Checking for latest zbxctl release (current: %s)...\n", currentVer)
			} else {
				fmt.Fprintf(formatter.Writer, "Checking for latest zbxctl release and self-updating (current: %s)...\n", currentVer)
			}
		}

		baseURL := os.Getenv("ZBXCTL_UPDATE_BASE_URL")

		updater := update.NewUpdater(update.Options{
			CurrentVersion:   currentVer,
			TargetVersion:    updateTagFlag,
			Repo:             updateRepoFlag,
			TargetExecutable: targetExec,
			CheckOnly:        updateCheckFlag,
			Force:            updateForceFlag,
			BaseAPIURL:       baseURL,
			Writer:           formatter.Writer,
		})

		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		result, err := updater.Run(ctx)
		if err != nil {
			return err
		}

		if isMachineOutput {
			return formatter.Print(result)
		}

		// Human-friendly output
		if updateCheckFlag {
			if result.UpdateAvailable {
				fmt.Fprintf(formatter.Writer, "\nNew version available: %s (current: %s)\n", result.LatestVersion, result.CurrentVersion)
				fmt.Fprintf(formatter.Writer, "Run 'zbxctl update' to install the latest release.\n")
			} else {
				fmt.Fprintf(formatter.Writer, "zbxctl is up to date (%s).\n", result.CurrentVersion)
			}
			return nil
		}

		if result.Updated {
			fmt.Fprintf(formatter.Writer, "\nSuccessfully updated zbxctl to %s!\nBinary location: %s\n", result.LatestVersion, result.BinaryPath)
			fmt.Fprintf(formatter.Writer, "\n%s\n", update.SkillReminderMsg)
		} else {
			fmt.Fprintf(formatter.Writer, "%s\n", result.Message)
		}

		return nil
	},
}

func init() {
	updateCmd.Flags().BoolVarP(&updateCheckFlag, "check", "c", false, "check for updates without downloading or installing")
	updateCmd.Flags().BoolVarP(&updateForceFlag, "force", "f", false, "force update/reinstall even if already at latest version")
	updateCmd.Flags().StringVar(&updateRepoFlag, "repo", update.DefaultGitHubRepo, "GitHub repository (owner/repo)")
	updateCmd.Flags().StringVar(&updateTagFlag, "version", "", "target specific version/tag to install (e.g. v0.2.0)")
	updateCmd.Flags().StringVar(&updateTargetFlag, "target-binary", "", "override target binary path (advanced/testing)")
	_ = updateCmd.Flags().MarkHidden("target-binary")

	RootCmd.AddCommand(updateCmd)
}
