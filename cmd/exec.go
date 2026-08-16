package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var execHostIDFlag string

var execCmd = &cobra.Command{
	Use:   "exec <scriptid> --hostid=<hostid>",
	Short: "Execute a Zabbix script on a target host",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		scriptID := args[0]
		if execHostIDFlag == "" {
			return fmt.Errorf("hostid flag is required (--hostid)")
		}

		params := map[string]interface{}{
			"scriptid": scriptID,
			"hostid":   execHostIDFlag,
		}

		res, err := checkSafetyAndCall(cmd.Context(), "script.execute", params)
		if err != nil {
			return err
		}

		return formatter.Print(res)
	},
}

func init() {
	execCmd.Flags().StringVar(&execHostIDFlag, "hostid", "", "ID of target host")
	_ = execCmd.MarkFlagRequired("hostid")
	RootCmd.AddCommand(execCmd)
}
