package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ErKatta/zbxctl/pkg/zabbix"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var applyFileFlag string

type ManifestItem struct {
	Kind     string                 `json:"kind,omitempty" yaml:"kind,omitempty"`
	Resource string                 `json:"resource,omitempty" yaml:"resource,omitempty"`
	Method   string                 `json:"method,omitempty" yaml:"method,omitempty"`
	Params   map[string]interface{} `json:"params,omitempty" yaml:"params,omitempty"`
	Spec     map[string]interface{} `json:"spec,omitempty" yaml:"spec,omitempty"`
}

var applyCmd = &cobra.Command{
	Use:   "apply -f <manifest.yaml|manifest.json>",
	Short: "Declaratively create or update Zabbix resources from manifest files",
	Long:  `apply reads a JSON or YAML resource manifest file and creates or updates Zabbix resources (hosts, templates, items, triggers).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if applyFileFlag == "" {
			return fmt.Errorf("manifest file path is required (--file / -f)")
		}

		var data []byte
		var err error
		if applyFileFlag == "-" {
			data, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read manifest from stdin: %w", err)
			}
		} else {
			data, err = os.ReadFile(applyFileFlag)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", applyFileFlag, err)
			}
		}

		var items []ManifestItem
		var single ManifestItem

		// Try unmarshaling as list or single object
		if err := yaml.Unmarshal(data, &items); err != nil || len(items) == 0 {
			if err2 := yaml.Unmarshal(data, &single); err2 == nil && (single.Kind != "" || single.Resource != "" || single.Method != "" || len(single.Params) > 0 || len(single.Spec) > 0) {
				items = []ManifestItem{single}
			} else {
				// Fallback: parse raw generic map
				var genericMap map[string]interface{}
				if err3 := yaml.Unmarshal(data, &genericMap); err3 == nil {
					items = []ManifestItem{{
						Kind:   fmt.Sprintf("%v", genericMap["kind"]),
						Params: genericMap,
					}}
				} else {
					return fmt.Errorf("failed to parse manifest file: %w", err)
				}
			}
		}

		type ApplyResult struct {
			Resource string      `json:"resource"`
			Action   string      `json:"action"`
			Response interface{} `json:"response"`
		}

		var results []ApplyResult
		for _, item := range items {
			resName := item.Kind
			if resName == "" {
				resName = item.Resource
			}
			if resName == "" && len(args) > 0 {
				resName = args[0]
			}

			params := item.Params
			if len(params) == 0 && len(item.Spec) > 0 {
				params = item.Spec
			}

			var method string
			if item.Method != "" {
				method = item.Method
			} else {
				resInfo, err := zabbix.ResolveResource(resName)
				if err != nil {
					return fmt.Errorf("failed to resolve resource for manifest item: %w", err)
				}

				if resInfo.Name == "inventory" {
					method = "host.update"
					if params["hostid"] == nil || fmt.Sprintf("%v", params["hostid"]) == "" {
						targetHost := fmt.Sprintf("%v", params["host"])
						if targetHost == "" || targetHost == "<nil>" {
							targetHost = fmt.Sprintf("%v", params["name"])
						}
						if targetHost != "" && targetHost != "<nil>" {
							hostRes, err := checkSafetyAndCall(cmd.Context(), "host.get", map[string]interface{}{
								"output": []string{"hostid"},
								"filter": map[string]interface{}{"host": targetHost},
							})
							var hosts []map[string]interface{}
							if err == nil {
								if rawMsg, ok := hostRes.(json.RawMessage); ok {
									_ = json.Unmarshal(rawMsg, &hosts)
								}
							}
							if len(hosts) == 0 {
								hostRes, err = checkSafetyAndCall(cmd.Context(), "host.get", map[string]interface{}{
									"output": []string{"hostid"},
									"filter": map[string]interface{}{"name": targetHost},
								})
								if err == nil {
									if rawMsg, ok := hostRes.(json.RawMessage); ok {
										_ = json.Unmarshal(rawMsg, &hosts)
									}
								}
							}
							if len(hosts) > 0 {
								if hID, ok := hosts[0]["hostid"].(string); ok {
									params["hostid"] = hID
								}
							}
						}
					}
				} else {
					// Check if ID is present in params -> update, else create
					idVal := params[resInfo.IDProperty]
					if idVal != nil && fmt.Sprintf("%v", idVal) != "" {
						method = resInfo.APIPrefix + ".update"
					} else {
						method = resInfo.APIPrefix + ".create"
					}
				}
			}

			// Clean and sanitize parameters according to Zabbix 7 mutation schema
			params = zabbix.SanitizeApplyParams(resName, params)

			res, err := checkSafetyAndCall(cmd.Context(), method, params)
			if err != nil {
				return fmt.Errorf("apply failed for method %s: %w", method, err)
			}

			action := "created"
			if strings.HasSuffix(method, ".update") {
				action = "updated"
			}

			results = append(results, ApplyResult{
				Resource: resName,
				Action:   action,
				Response: res,
			})
		}

		return formatter.Print(results)
	},
}

func init() {
	applyCmd.Flags().StringVarP(&applyFileFlag, "file", "f", "", "path to manifest file (.json or .yaml)")
	_ = applyCmd.MarkFlagRequired("file")
	RootCmd.AddCommand(applyCmd)
}
