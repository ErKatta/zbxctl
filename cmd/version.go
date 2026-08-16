package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version   = "0.2.0-dev"
	Commit    = "none"
	BuildDate = "unknown"

	versionShortFlag bool
)

type VersionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Compiler  string `json:"compiler"`
	Platform  string `json:"platform"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the zbxctl client version and build information",
	Long:  `version displays the client version, git commit, build date, Go compiler version, and target OS/architecture platform.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		info := VersionInfo{
			Version:   Version,
			Commit:    Commit,
			BuildDate: BuildDate,
			GoVersion: runtime.Version(),
			Compiler:  runtime.Compiler,
			Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		}

		if versionShortFlag {
			fmt.Fprintln(formatter.Writer, info.Version)
			return nil
		}

		outputExplicit, _ := cmd.Root().PersistentFlags().GetString("output")
		if !cmd.Root().PersistentFlags().Changed("output") {
			outputExplicit = ""
		}

		if outputExplicit == "json" || outputExplicit == "toon" {
			return formatter.Print(info)
		} else if outputExplicit == "yaml" {
			return formatter.Print(info)
		} else if outputExplicit == "table" {
			return formatter.Print(info)
		}

		// Standard human-friendly text format
		if Commit != "none" && BuildDate != "unknown" {
			fmt.Fprintf(formatter.Writer, "zbxctl version %s (commit: %s, built: %s, %s, %s)\n",
				info.Version, info.Commit, info.BuildDate, info.Platform, info.GoVersion)
		} else {
			fmt.Fprintf(formatter.Writer, "zbxctl version %s (%s, %s)\n",
				info.Version, info.Platform, info.GoVersion)
		}
		return nil
	},
}

func init() {
	RootCmd.Version = Version
	versionCmd.Flags().BoolVarP(&versionShortFlag, "short", "s", false, "print only the version string")
	RootCmd.AddCommand(versionCmd)
}
