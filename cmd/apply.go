package cmd

import (
	"bytes"
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
	Use:   "apply -f <manifest.yaml|manifest.json|->",
	Short: "Declaratively create or update Zabbix resources from manifest files or stdin",
	Long: `apply reads a JSON or YAML resource manifest from a file or standard input ('-f -') and creates or updates Zabbix resources (hosts, templates, items, triggers, inventories).

Examples:
  # Apply a YAML manifest from file
  zbxctl apply -f host-manifest.yaml

  # Stream manifest from standard input (Unix pipe)
  cat host-manifest.yaml | zbxctl apply -f -

  # Stream manifest directly from Python subprocess
  # subprocess.run(["zbxctl", "apply", "-f", "-"], input=manifest_yaml, text=True)
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if applyFileFlag == "" {
			return fmt.Errorf("manifest file path is required (--file / -f)")
		}

		var data []byte
		var err error
		if applyFileFlag == "-" {
			data, err = io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return fmt.Errorf("failed to read manifest from stdin: %w", err)
			}
		} else {
			data, err = os.ReadFile(applyFileFlag)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", applyFileFlag, err)
			}
		}

		if len(strings.TrimSpace(string(data))) == 0 {
			return fmt.Errorf("manifest input is empty")
		}

		items, err := parseManifestItems(data)
		if err != nil {
			return fmt.Errorf("failed to parse manifest: %w", err)
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

func parseManifestItems(data []byte) ([]ManifestItem, error) {
	var items []ManifestItem
	dec := yaml.NewDecoder(bytes.NewReader(data))

	for {
		var docNode yaml.Node
		err := dec.Decode(&docNode)
		if err != nil {
			if err == io.EOF {
				break
			}
			// Fallback directly to single or array Unmarshal
			break
		}

		if docNode.Kind == 0 {
			continue
		}

		// 1. Try decoding document as array of ManifestItem
		var docItems []ManifestItem
		if err := docNode.Decode(&docItems); err == nil && len(docItems) > 0 {
			items = append(items, docItems...)
			continue
		}

		// 2. Try decoding document as single ManifestItem
		var single ManifestItem
		if err := docNode.Decode(&single); err == nil && (single.Kind != "" || single.Resource != "" || single.Method != "" || len(single.Params) > 0 || len(single.Spec) > 0) {
			items = append(items, single)
			continue
		}

		// 3. Fallback: decode document as generic map
		var genericMap map[string]interface{}
		if err := docNode.Decode(&genericMap); err == nil && len(genericMap) > 0 {
			items = append(items, ManifestItem{
				Kind:   fmt.Sprintf("%v", genericMap["kind"]),
				Params: genericMap,
			})
			continue
		}
	}

	if len(items) == 0 {
		// Fallback to top-level unmarshal
		var fallbackItems []ManifestItem
		var fallbackSingle ManifestItem
		if err := yaml.Unmarshal(data, &fallbackItems); err == nil && len(fallbackItems) > 0 {
			return fallbackItems, nil
		}
		if err := yaml.Unmarshal(data, &fallbackSingle); err == nil && (fallbackSingle.Kind != "" || fallbackSingle.Resource != "" || fallbackSingle.Method != "" || len(fallbackSingle.Params) > 0 || len(fallbackSingle.Spec) > 0) {
			return []ManifestItem{fallbackSingle}, nil
		}
		return nil, fmt.Errorf("no valid manifest objects found in input")
	}

	return items, nil
}

func init() {
	applyCmd.Flags().StringVarP(&applyFileFlag, "file", "f", "", "path to manifest file (.json or .yaml), or '-' to read from stdin")
	_ = applyCmd.MarkFlagRequired("file")
	RootCmd.AddCommand(applyCmd)
}
