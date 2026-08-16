package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

type InventoryReport struct {
	Context       string `json:"context"`
	ZabbixVersion string `json:"zabbix_version"`
	SafetyLevel   string `json:"safety_level"`
	TotalHosts    int    `json:"total_hosts"`
	TotalProblems int    `json:"total_problems"`
	TotalItems    int    `json:"total_items"`
	TotalTriggers int    `json:"total_triggers"`
	TotalTemplates int   `json:"total_templates"`
}

var inventoryCmd = &cobra.Command{
	Use:   "inventory",
	Short: "Probe Zabbix instance for low-token system grounding & statistics",
	Long:  `inventory performs a read-only probe of the connected Zabbix 7 instance, gathering host counts, active problem counts, item totals, and system context.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		ver, err := zbxClient.GetVersion(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch Zabbix version: %w", err)
		}

		report := InventoryReport{
			Context:       actCtxName,
			ZabbixVersion: ver,
			SafetyLevel:   activeCtx.SafetyLevel,
		}

		// Count Hosts
		if res, err := checkSafetyAndCall(ctx, "host.get", map[string]interface{}{"countOutput": true}); err == nil {
			var count int
			_ = json.Unmarshal(res.(json.RawMessage), &count)
			report.TotalHosts = count
		}

		// Count Problems
		if res, err := checkSafetyAndCall(ctx, "problem.get", map[string]interface{}{"countOutput": true}); err == nil {
			var count int
			_ = json.Unmarshal(res.(json.RawMessage), &count)
			report.TotalProblems = count
		}

		// Count Items
		if res, err := checkSafetyAndCall(ctx, "item.get", map[string]interface{}{"countOutput": true}); err == nil {
			var count int
			_ = json.Unmarshal(res.(json.RawMessage), &count)
			report.TotalItems = count
		}

		// Count Triggers
		if res, err := checkSafetyAndCall(ctx, "trigger.get", map[string]interface{}{"countOutput": true}); err == nil {
			var count int
			_ = json.Unmarshal(res.(json.RawMessage), &count)
			report.TotalTriggers = count
		}

		// Count Templates
		if res, err := checkSafetyAndCall(ctx, "template.get", map[string]interface{}{"countOutput": true}); err == nil {
			var count int
			_ = json.Unmarshal(res.(json.RawMessage), &count)
			report.TotalTemplates = count
		}

		return formatter.Print(report)
	},
}

func init() {
	RootCmd.AddCommand(inventoryCmd)
}
