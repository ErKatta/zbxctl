package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ErKatta/zbxctl/pkg/zabbix"
	"github.com/spf13/cobra"
)

var (
	waitForFlag   string
	waitTimeoutFlag time.Duration
)

var waitCmd = &cobra.Command{
	Use:   "wait <resource> <id> --for=<condition>",
	Short: "Wait for a condition or problem resolution on a Zabbix resource",
	Long:  `wait repeatedly polls Zabbix until the specified condition (e.g. resolved, status=0, value=OK) is met or timeout expires.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		resInfo, err := zabbix.ResolveResource(args[0])
		if err != nil {
			return err
		}

		id := args[1]
		if waitForFlag == "" {
			return fmt.Errorf("--for flag is required (e.g., --for=resolved, --for=status=0)")
		}

		deadline := time.Now().Add(waitTimeoutFlag)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		fmt.Printf("Waiting for %s %s to satisfy condition %q (timeout: %v)...\n", resInfo.Name, id, waitForFlag, waitTimeoutFlag)

		for {
			select {
			case <-cmd.Context().Done():
				return cmd.Context().Err()
			case now := <-ticker.C:
				if now.After(deadline) {
					return fmt.Errorf("timeout waiting for %s %s to reach condition %q", resInfo.Name, id, waitForFlag)
				}

				params := map[string]interface{}{
					"output": "extend",
				}
				if resInfo.PluralIDProperty != "" {
					params[resInfo.PluralIDProperty] = []string{id}
				} else {
					params[resInfo.IDProperty] = id
				}

				res, err := checkSafetyAndCall(cmd.Context(), resInfo.APIPrefix+".get", params)
				if err != nil {
					continue
				}

				var items []map[string]interface{}
				if err := json.Unmarshal(res.(json.RawMessage), &items); err == nil {
					if resInfo.Name == "problem" && waitForFlag == "resolved" {
						if len(items) == 0 {
							fmt.Printf("Condition met: problem %s is resolved.\n", id)
							return nil
						}
					} else if len(items) > 0 {
						item := items[0]
						if checkCondition(item, waitForFlag) {
							fmt.Printf("Condition met for %s %s.\n", resInfo.Name, id)
							return formatter.Print(item)
						}
					}
				}
			}
		}
	},
}

func checkCondition(item map[string]interface{}, cond string) bool {
	parts := strings.SplitN(cond, "=", 2)
	if len(parts) == 2 {
		k, expected := parts[0], parts[1]
		val, exists := item[k]
		if exists && fmt.Sprintf("%v", val) == expected {
			return true
		}
	}
	return false
}

func init() {
	waitCmd.Flags().StringVar(&waitForFlag, "for", "", "condition to wait for (e.g. resolved, status=0)")
	waitCmd.Flags().DurationVar(&waitTimeoutFlag, "timeout", 60*time.Second, "maximum wait timeout")
	_ = waitCmd.MarkFlagRequired("for")
	RootCmd.AddCommand(waitCmd)
}
