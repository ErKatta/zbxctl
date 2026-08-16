package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type ClusterInfoReport struct {
	Context        string `json:"context"`
	ZabbixVersion  string `json:"zabbix_version"`
	SafetyLevel    string `json:"safety_level"`
	TotalHosts     int    `json:"total_hosts"`
	TotalProblems  int    `json:"total_problems"`
	TotalItems     int    `json:"total_items"`
	TotalTriggers  int    `json:"total_triggers"`
	TotalTemplates int    `json:"total_templates"`
}

// InventoryReport is an alias for backwards compatibility
type InventoryReport = ClusterInfoReport

func parseCountResult(raw interface{}) int {
	if raw == nil {
		return 0
	}
	if rawMsg, ok := raw.(json.RawMessage); ok {
		var strVal string
		if err := json.Unmarshal(rawMsg, &strVal); err == nil {
			if n, err := strconv.Atoi(strings.TrimSpace(strVal)); err == nil {
				return n
			}
		}
		var intVal int
		if err := json.Unmarshal(rawMsg, &intVal); err == nil {
			return intVal
		}
		var floatVal float64
		if err := json.Unmarshal(rawMsg, &floatVal); err == nil {
			return int(floatVal)
		}
	}
	switch v := raw.(type) {
	case string:
		n, _ := strconv.Atoi(strings.TrimSpace(v))
		return n
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	return 0
}

var clusterInfoCmd = &cobra.Command{
	Use:     "cluster-info",
	Aliases: []string{"clusterinfo", "inventory", "info", "overview"},
	Short:   "Display Zabbix instance connection, version, and sizing statistics",
	Long:    `cluster-info performs a read-only probe of the connected Zabbix 7 instance, gathering host counts, active problem counts, item totals, and system context.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		ver, err := zbxClient.GetVersion(ctx)
		if err != nil {
			return fmt.Errorf("failed to fetch Zabbix version: %w", err)
		}

		report := ClusterInfoReport{
			Context:       actCtxName,
			ZabbixVersion: ver,
			SafetyLevel:   activeCtx.SafetyLevel,
		}

		// Count Hosts
		if res, err := checkSafetyAndCall(ctx, "host.get", map[string]interface{}{"countOutput": true}); err == nil {
			report.TotalHosts = parseCountResult(res)
		}

		// Count Problems
		if res, err := checkSafetyAndCall(ctx, "problem.get", map[string]interface{}{"countOutput": true}); err == nil {
			report.TotalProblems = parseCountResult(res)
		}

		// Count Items
		if res, err := checkSafetyAndCall(ctx, "item.get", map[string]interface{}{"countOutput": true}); err == nil {
			report.TotalItems = parseCountResult(res)
		}

		// Count Triggers
		if res, err := checkSafetyAndCall(ctx, "trigger.get", map[string]interface{}{"countOutput": true}); err == nil {
			report.TotalTriggers = parseCountResult(res)
		}

		// Count Templates
		if res, err := checkSafetyAndCall(ctx, "template.get", map[string]interface{}{"countOutput": true}); err == nil {
			report.TotalTemplates = parseCountResult(res)
		}

		return formatter.Print(report)
	},
}

func init() {
	RootCmd.AddCommand(clusterInfoCmd)
}
