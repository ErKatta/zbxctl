package cmd

import (
	"fmt"

	"github.com/ErKatta/zbxctl/pkg/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage zbxctl configuration and contexts",
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return formatter.Print(cfg.Redacted())
	},
}

var configGetContextsCmd = &cobra.Command{
	Use:     "get-contexts",
	Aliases: []string{"list"},
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

var configUseContextCmd = &cobra.Command{
	Use:   "use-context <context-name>",
	Short: "Switch the active context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetCtx := args[0]
		if _, ok := cfg.Contexts[targetCtx]; !ok {
			return fmt.Errorf("context %q not found in config", targetCtx)
		}
		cfg.ActiveContext = targetCtx
		path, _ := config.DefaultConfigPath()
		if err := config.SaveConfig(cfg, path); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("Switched active context to %q.\n", targetCtx)
		return nil
	},
}

var configSetSafetyCmd = &cobra.Command{
	Use:   "set-safety <readonly|readwrite-mine|readwrite-all|dangerously-unrestricted>",
	Short: "Set safety level for active or target context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		level := args[0]
		switch level {
		case "readonly", "readwrite-mine", "readwrite-all", "dangerously-unrestricted":
		default:
			return fmt.Errorf("invalid safety level %q. Must be one of: readonly, readwrite-mine, readwrite-all, dangerously-unrestricted", level)
		}

		targetCtx := cfg.ActiveContext
		if contextFlag != "" {
			targetCtx = contextFlag
		}

		ctxObj, ok := cfg.Contexts[targetCtx]
		if !ok {
			return fmt.Errorf("context %q not found", targetCtx)
		}

		ctxObj.SafetyLevel = level
		cfg.Contexts[targetCtx] = ctxObj

		path, _ := config.DefaultConfigPath()
		if err := config.SaveConfig(cfg, path); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("Updated safety level for context %q to %q.\n", targetCtx, level)
		return nil
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize default zbxctl configuration file (~/.zbxctl/config.yaml)",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := config.DefaultConfigPath()
		defCfg := config.DefaultConfig()
		if err := config.SaveConfig(defCfg, path); err != nil {
			return fmt.Errorf("failed to initialize config at %s: %w", path, err)
		}
		fmt.Printf("Initialized configuration file at %s\n", path)
		return nil
	},
}

var configCurrentConfigCmd = &cobra.Command{
	Use:     "current-config",
	Aliases: []string{"current", "current-context"},
	Short:   "Display active context configuration",
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

func init() {
	configCmd.AddCommand(configViewCmd)
	configCmd.AddCommand(configCurrentConfigCmd)
	configCmd.AddCommand(configGetContextsCmd)
	configCmd.AddCommand(configUseContextCmd)
	configCmd.AddCommand(configSetSafetyCmd)
	configCmd.AddCommand(configInitCmd)
	RootCmd.AddCommand(configCmd)
}

