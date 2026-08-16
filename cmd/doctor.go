package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Perform health checks and diagnose Zabbix API connectivity",
	Long:  `doctor verifies Zabbix API endpoint reachability, checks API version, verifies credentials, and inspects context safety configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()

		type CheckResult struct {
			Check   string `json:"check"`
			Status  string `json:"status"`
			Message string `json:"message"`
		}

		var results []CheckResult

		// Check 1: Context config
		results = append(results, CheckResult{
			Check:   "Context Configuration",
			Status:  "OK",
			Message: fmt.Sprintf("Active Context: %s (URL: %s, Safety: %s)", actCtxName, activeCtx.URL, activeCtx.SafetyLevel),
		})

		// Check 2: API Version Probe
		ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
		defer cancel()

		ver, err := zbxClient.GetVersion(ctx)
		if err != nil {
			results = append(results, CheckResult{
				Check:   "Zabbix API Connectivity & Version",
				Status:  "FAIL",
				Message: fmt.Sprintf("Failed to reach Zabbix API at %s: %v", activeCtx.URL, err),
			})
		} else {
			results = append(results, CheckResult{
				Check:   "Zabbix API Connectivity & Version",
				Status:  "OK",
				Message: fmt.Sprintf("Zabbix API Version: %s (Latency: %v)", ver, time.Since(start).Round(time.Millisecond)),
			})
		}

		// Check 3: Authentication probe (host.get output=count)
		if err == nil {
			_, authErr := zbxClient.Call(ctx, "host.get", map[string]interface{}{"countOutput": true})
			if authErr != nil {
				results = append(results, CheckResult{
					Check:   "Authentication & Scope",
					Status:  "FAIL",
					Message: fmt.Sprintf("Authentication test failed: %v", authErr),
				})
			} else {
				results = append(results, CheckResult{
					Check:   "Authentication & Scope",
					Status:  "OK",
					Message: "Successfully authenticated with Zabbix instance.",
				})
			}
		}

		return formatter.Print(results)
	},
}

func init() {
	RootCmd.AddCommand(doctorCmd)
}
