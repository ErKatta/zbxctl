package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var rawParamsFlag string

var rawCmd = &cobra.Command{
	Use:   "raw <zabbix.method> [--params='<json>']",
	Short: "Invoke any Zabbix 7 JSON-RPC API endpoint directly",
	Long: `raw allows invoking any Zabbix 7 API method (e.g., proxygroup.get, mfa.get, ha.get, history.push, connector.create).
All calls are validated through client-side safety middleware.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		method := args[0]
		var params interface{}

		if rawParamsFlag != "" {
			if err := json.Unmarshal([]byte(rawParamsFlag), &params); err != nil {
				return fmt.Errorf("invalid json in --params: %w", err)
			}
		}

		res, err := checkSafetyAndCall(cmd.Context(), method, params)
		if err != nil {
			return err
		}

		return formatter.Print(res)
	},
}

func init() {
	rawCmd.Flags().StringVarP(&rawParamsFlag, "params", "p", "{}", "JSON parameters string for the API call")
	RootCmd.AddCommand(rawCmd)
}
