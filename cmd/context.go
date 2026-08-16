package cmd

import (
	"fmt"

	"github.com/ErKatta/zbxctl/pkg/config"
	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "View and manage zbxctl contexts",
}

var contextListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"get", "get-contexts"},
	Short:   "List all configured contexts",
	RunE: func(cmd *cobra.Command, args []string) error {
		type ctxItem struct {
			Name        string `json:"name"`
			Active      bool   `json:"active"`
			URL         string `json:"url"`
			SafetyLevel string `json:"safety_level"`
		}

		var items []ctxItem
		for name, c := range cfg.Contexts {
			items = append(items, ctxItem{
				Name:        name,
				Active:      name == cfg.ActiveContext,
				URL:         c.URL,
				SafetyLevel: c.SafetyLevel,
			})
		}
		return formatter.Print(items)
	},
}

var contextUseCmd = &cobra.Command{
	Use:   "use <context-name>",
	Short: "Switch the active context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetCtx := args[0]
		if _, ok := cfg.Contexts[targetCtx]; !ok {
			return fmt.Errorf("context %q not found in config", targetCtx)
		}
		cfg.ActiveContext = targetCtx
		path, err := getConfigPath(cmd)
		if err != nil {
			return fmt.Errorf("failed to determine config path: %w", err)
		}
		if err := config.SaveConfig(cfg, path); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("Switched active context to %q.\n", targetCtx)
		return nil
	},
}

var contextCurrentCmd = &cobra.Command{
	Use:     "current",
	Aliases: []string{"show"},
	Short:   "Display active context details",
	RunE: func(cmd *cobra.Command, args []string) error {
		active, name, err := cfg.GetActiveContext()
		if err != nil {
			return err
		}
		redactedCtx := active.Redacted()
		type activeInfo struct {
			Name string          `json:"name"`
			Info *config.Context `json:"info"`
		}
		return formatter.Print(activeInfo{Name: name, Info: &redactedCtx})
	},
}

var useContextTopCmd = &cobra.Command{
	Use:    "use-context <context-name>",
	Short:  "Switch the active context (shortcut for zbxctl context use)",
	Hidden: false,
	Args:   cobra.ExactArgs(1),
	RunE:   contextUseCmd.RunE,
}

func init() {
	contextCmd.AddCommand(contextListCmd)
	contextCmd.AddCommand(contextUseCmd)
	contextCmd.AddCommand(contextCurrentCmd)
	RootCmd.AddCommand(contextCmd)
	RootCmd.AddCommand(useContextTopCmd)
}
