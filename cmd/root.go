package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/ErKatta/zbxctl/pkg/config"
	"github.com/ErKatta/zbxctl/pkg/output"
	"github.com/ErKatta/zbxctl/pkg/safety"
	"github.com/ErKatta/zbxctl/pkg/zabbix"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	cfgFile      string
	contextFlag  string
	outputFlag   string
	safetyFlag   string
	forceFlag    bool
	verboseFlag  bool

	cfg        *config.Config
	activeCtx  *config.Context
	actCtxName string
	zbxClient  *zabbix.Client
	formatter  *output.Formatter
)

func getConfigPath(cmd *cobra.Command) (string, error) {
	if cmd != nil {
		if p, _ := cmd.Flags().GetString("config"); p != "" {
			return p, nil
		}
		if p, _ := cmd.Root().PersistentFlags().GetString("config"); p != "" {
			return p, nil
		}
	}
	if cfgFile != "" {
		return cfgFile, nil
	}
	return config.DefaultConfigPath()
}

var RootCmd = &cobra.Command{
	Use:   "zbxctl",
	Short: "zbxctl - AI-Agent Friendly Zabbix 7 CLI",
	Long: `zbxctl is a high-performance command-line tool tailored for both human engineers and AI coding agents.
It provides a Tier 1 ergonomic verb-noun interface for common operations and a Tier 2 raw JSON-RPC engine for 100% Zabbix 7 API coverage.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if contextFlag != "" {
			cfg.ActiveContext = contextFlag
		}

		activeCtx, actCtxName, err = cfg.GetActiveContext()
		if err != nil {
			// If config missing or active context invalid, create fallback default
			def := config.DefaultConfig()
			cfg = def
			activeCtx, actCtxName, _ = cfg.GetActiveContext()
		}

		if safetyFlag != "" {
			activeCtx.SafetyLevel = safetyFlag
		}
		if outputFlag != "" {
			activeCtx.OutputFormat = outputFlag
		}

		formatter = output.NewFormatter(activeCtx.OutputFormat)
		if cmd != nil && cmd.OutOrStdout() != nil {
			formatter.Writer = cmd.OutOrStdout()
		}
		zbxClient = zabbix.NewClient(activeCtx)

		return nil
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		handleError(err)
	}
}

func handleError(err error) {
	if err == nil {
		return
	}

	if sErr, ok := err.(*safety.SafetyError); ok {
		env := safety.SafetyErrorEnvelope{
			Error: *sErr,
		}
		if !output.IsTerminal() || activeCtx != nil && activeCtx.OutputFormat == "json" {
			output.PrintErrorEnvelope(os.Stderr, env)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %s\nResolution: %s\n", sErr.Message, sErr.Resolution)
		}
		os.Exit(2)
	}

	// For general errors in non-terminal or json format
	if !output.IsTerminal() || (activeCtx != nil && activeCtx.OutputFormat == "json") {
		env := map[string]interface{}{
			"error": map[string]string{
				"code":    "EXECUTION_ERROR",
				"message": err.Error(),
			},
		}
		output.PrintErrorEnvelope(os.Stderr, env)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	os.Exit(1)
}

func init() {
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.zbxctl/config.yaml)")
	RootCmd.PersistentFlags().StringVar(&contextFlag, "context", "", "override active context")
	RootCmd.PersistentFlags().StringVarP(&outputFlag, "output", "o", "", "output format (auto, table, json, toon, yaml)")
	RootCmd.PersistentFlags().StringVar(&safetyFlag, "safety-level", "", "override safety level (readonly, readwrite-mine, readwrite-all, dangerously-unrestricted)")
	RootCmd.PersistentFlags().BoolVar(&forceFlag, "force", false, "force bulk or high-risk mutations")
	RootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "enable verbose output")
}

func checkSafetyAndCall(ctx context.Context, method string, params interface{}) (interface{}, error) {
	sLevel := safety.SafetyLevel(activeCtx.SafetyLevel)
	if err := safety.CheckSafety(sLevel, actCtxName, method, params, forceFlag); err != nil {
		return nil, err
	}
	return zbxClient.Call(ctx, method, params)
}

func ResetCommandFlags(cmd *cobra.Command) {
	if cmd == nil {
		return
	}
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		f.Changed = false
		_ = f.Value.Set(f.DefValue)
	})
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		f.Changed = false
		_ = f.Value.Set(f.DefValue)
	})
	for _, child := range cmd.Commands() {
		ResetCommandFlags(child)
	}
}
