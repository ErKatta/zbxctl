package cmd

import (
	"github.com/ErKatta/zbxctl/pkg/zabbix"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <resource> <id...>",
	Short: "Delete Zabbix resources by ID",
	Long:  `delete removes resources from Zabbix (hosts, templates, items, triggers, etc.).`,
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		resInfo, err := zabbix.ResolveResource(args[0])
		if err != nil {
			return err
		}

		ids := args[1:]
		method := resInfo.APIPrefix + ".delete"

		res, err := checkSafetyAndCall(cmd.Context(), method, ids)
		if err != nil {
			return err
		}

		return formatter.Print(res)
	},
}

func init() {
	RootCmd.AddCommand(deleteCmd)
}
