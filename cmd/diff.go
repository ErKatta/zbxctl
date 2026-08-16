package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ErKatta/zbxctl/pkg/zabbix"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	diffFileFlag string
	diffIDFlag   string
)

var diffCmd = &cobra.Command{
	Use:   "diff -f <manifest.yaml> [--id=<resource_id>]",
	Short: "Show differences between local manifest spec and live Zabbix resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		if diffFileFlag == "" {
			return fmt.Errorf("--file (-f) is required")
		}

		var data []byte
		var err error
		if diffFileFlag == "-" {
			data, err = io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read manifest from stdin: %w", err)
			}
		} else {
			data, err = os.ReadFile(diffFileFlag)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", diffFileFlag, err)
			}
		}

		var item ManifestItem
		if err := yaml.Unmarshal(data, &item); err != nil {
			return fmt.Errorf("failed to parse manifest file: %w", err)
		}

		resName := item.Kind
		if resName == "" {
			resName = item.Resource
		}
		if resName == "" && len(args) > 0 {
			resName = args[0]
		}

		resInfo, err := zabbix.ResolveResource(resName)
		if err != nil {
			return fmt.Errorf("failed to resolve resource: %w", err)
		}

		targetID := diffIDFlag
		if targetID == "" {
			spec := item.Params
			if len(spec) == 0 {
				spec = item.Spec
			}
			if idVal, ok := spec[resInfo.IDProperty]; ok && idVal != nil {
				targetID = fmt.Sprintf("%v", idVal)
			}
		}
		if targetID == "" {
			return fmt.Errorf("--id flag is required (or specify %q in manifest spec)", resInfo.IDProperty)
		}

		params := map[string]interface{}{
			"output": "extend",
		}
		if resInfo.PluralIDProperty != "" {
			params[resInfo.PluralIDProperty] = []string{targetID}
		} else {
			params[resInfo.IDProperty] = targetID
		}

		switch resInfo.APIPrefix {
		case "host":
			params["selectTags"] = "extend"
			params["selectGroups"] = "extend"
			params["selectMacros"] = "extend"
			params["selectParentTemplates"] = "extend"
			params["selectInterfaces"] = "extend"
			params["selectInventory"] = "extend"
		case "template":
			params["selectTags"] = "extend"
			params["selectGroups"] = "extend"
			params["selectMacros"] = "extend"
			params["selectTemplates"] = "extend"
		case "hostgroup":
			params["selectTags"] = "extend"
		case "item":
			params["selectTags"] = "extend"
			params["selectPreprocessing"] = "extend"
		case "trigger":
			params["selectTags"] = "extend"
			params["selectDependencies"] = "extend"
		}

		res, err := checkSafetyAndCall(cmd.Context(), resInfo.APIPrefix+".get", params)
		if err != nil {
			return fmt.Errorf("failed to fetch live resource %s %s: %w", resName, targetID, err)
		}

		var liveList []map[string]interface{}
		_ = json.Unmarshal(res.(json.RawMessage), &liveList)
		if len(liveList) == 0 {
			return fmt.Errorf("live resource %s with ID %s not found", resName, targetID)
		}

		liveItem := liveList[0]
		rawSpec := item.Params
		if len(rawSpec) == 0 {
			rawSpec = item.Spec
		}

		localSpec := zabbix.SanitizeDiffSpec(resName, rawSpec)
		remoteSpec := zabbix.SanitizeDiffSpec(resName, liveItem)

		writer := cmd.OutOrStdout()
		fmt.Fprintf(writer, "Comparing local manifest (%s) vs Live Zabbix %s (ID: %s):\n", diffFileFlag, resName, targetID)
		diffCount := 0

		for k, expected := range localSpec {
			actual, exists := remoteSpec[k]

			if !exists {
				if isEmptyDiffVal(expected) {
					continue
				}
				fmt.Fprintln(writer, color.GreenString("+ %s: %s (missing on remote)", k, formatDiffValue(expected)))
				diffCount++
				continue
			}

			if isEmptyDiffVal(expected) && isEmptyDiffVal(actual) {
				continue
			}

			if !isEqualDiffValue(expected, actual) {
				fmt.Fprintln(writer, color.RedString("- %s: %s (remote)", k, formatDiffValue(actual)))
				fmt.Fprintln(writer, color.GreenString("+ %s: %s (local manifest)", k, formatDiffValue(expected)))
				diffCount++
			} else {
				if verboseFlag {
					fmt.Fprintf(writer, "  %s: %s (unchanged)\n", k, formatDiffValue(expected))
				}
			}
		}

		if diffCount == 0 {
			fmt.Fprintln(writer, "No differences found. Live resource matches local manifest spec perfectly.")
		}

		return nil
	},
}

func isEmptyDiffVal(v interface{}) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val) == ""
	case []interface{}:
		return len(val) == 0
	case []map[string]interface{}:
		return len(val) == 0
	case map[string]interface{}:
		return len(val) == 0
	}
	return false
}

func isEqualDiffValue(a, b interface{}) bool {
	jsonA, errA := json.Marshal(a)
	jsonB, errB := json.Marshal(b)
	if errA == nil && errB == nil {
		return string(jsonA) == string(jsonB)
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func formatDiffValue(v interface{}) string {
	switch v.(type) {
	case []interface{}, []map[string]interface{}, map[string]interface{}:
		data, err := json.Marshal(v)
		if err == nil {
			return string(data)
		}
	}
	return fmt.Sprintf("%v", v)
}

func init() {
	diffCmd.Flags().StringVarP(&diffFileFlag, "file", "f", "", "path to manifest file")
	diffCmd.Flags().StringVar(&diffIDFlag, "id", "", "ID of live resource to compare against (optional if present in manifest spec)")
	_ = diffCmd.MarkFlagRequired("file")
	RootCmd.AddCommand(diffCmd)
}
