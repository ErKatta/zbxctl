package cmd

import (
	"encoding/json"

	"github.com/ErKatta/zbxctl/pkg/zabbix"
	"github.com/spf13/cobra"
)

var describeCmd = &cobra.Command{
	Use:   "describe <resource> <id|name>",
	Short: "Inspect detailed extended configuration of a Zabbix resource",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		resInfo, err := zabbix.ResolveResource(args[0])
		if err != nil {
			return err
		}

		identifier := args[1]
		method := resInfo.APIPrefix + ".get"

		params := map[string]interface{}{
			"output": "extend",
		}
		if resInfo.Name == "host" {
			params["selectTags"] = "extend"
			params["selectInterfaces"] = "extend"
			params["selectGroups"] = "extend"
			params["selectParentTemplates"] = "extend"
			params["selectItems"] = "count"
			params["selectTriggers"] = "count"
		}

		if isNumeric(identifier) {
			if resInfo.PluralIDProperty != "" {
				params[resInfo.PluralIDProperty] = []string{identifier}
			} else {
				params[resInfo.IDProperty] = identifier
			}
		} else {
			params["filter"] = map[string]interface{}{
				resInfo.NameProperty: identifier,
			}
		}

		res, err := checkSafetyAndCall(cmd.Context(), method, params)
		if err != nil {
			return err
		}

		var items []map[string]interface{}
		if err := json.Unmarshal(res.(json.RawMessage), &items); err == nil && len(items) > 0 {
			return formatter.Print(items[0])
		}

		return formatter.Print(res)
	},
}

func init() {
	RootCmd.AddCommand(describeCmd)
}
