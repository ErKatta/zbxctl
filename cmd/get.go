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
	getFilterFlag       string
	getSearchFlag       string
	getSearchFieldsFlag string
	getSortFlag         string
	getSortOrderFlag    string
	getLimitFlag        int
	getSinceFlag        string
	getHistoryTypeFlag  int
	getHostFlag         string
	getFieldsFlag       string
	getAllFieldsFlag    bool
	getWideFlag         bool
	getExportFlag       bool
	getCountFlag        bool
)

var getCmd = &cobra.Command{
	Use:   "get <resource> [id|name]",
	Short: "Get Zabbix resources (hosts, problems, items, triggers, templates, metrics, etc.)",
	Long: `get fetches resources from Zabbix.
Supported resources: host (h), problem (p), item (i), trigger (t), template (tmpl), hostgroup (hg), maintenance (maint), event (ev), user, service, sla, dashboard, history (metric, metrics).

When inspecting a single resource by ID/name or when exporting declarative manifests (--export), outputs in YAML/JSON are structured as declarative manifests (kind + spec) ready for 'zbxctl apply -f' and 'zbxctl diff -f'.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		resourceArg := args[0]
		resInfo, err := zabbix.ResolveResource(resourceArg)
		if err != nil {
			return err
		}

		var filterFlag, searchFlag, searchFieldsFlag, sortFlag, sortOrderFlag, hostFlag, fieldsFlag, sinceFlag string
		var limitFlag, historyTypeFlag int
		var allFieldsFlag, wideFlag, exportFlag, countFlag bool

		if cmd.Flags().Changed("filter") {
			filterFlag, _ = cmd.Flags().GetString("filter")
		}
		if cmd.Flags().Changed("search") {
			searchFlag, _ = cmd.Flags().GetString("search")
		}
		if cmd.Flags().Changed("search-fields") {
			searchFieldsFlag, _ = cmd.Flags().GetString("search-fields")
		}
		if cmd.Flags().Changed("sort") {
			sortFlag, _ = cmd.Flags().GetString("sort")
		}
		if cmd.Flags().Changed("sort-order") {
			sortOrderFlag, _ = cmd.Flags().GetString("sort-order")
		}
		if cmd.Flags().Changed("host") {
			hostFlag, _ = cmd.Flags().GetString("host")
		}
		if cmd.Flags().Changed("fields") {
			fieldsFlag, _ = cmd.Flags().GetString("fields")
		}
		if cmd.Flags().Changed("since") {
			sinceFlag, _ = cmd.Flags().GetString("since")
		}
		if cmd.Flags().Changed("limit") {
			limitFlag, _ = cmd.Flags().GetInt("limit")
		}
		if cmd.Flags().Changed("history-type") {
			historyTypeFlag, _ = cmd.Flags().GetInt("history-type")
		}
		if cmd.Flags().Changed("all-fields") {
			allFieldsFlag, _ = cmd.Flags().GetBool("all-fields")
		}
		if cmd.Flags().Changed("wide") {
			wideFlag, _ = cmd.Flags().GetBool("wide")
		}
		if cmd.Flags().Changed("export") {
			exportFlag, _ = cmd.Flags().GetBool("export")
		}
		if cmd.Flags().Changed("count") {
			countFlag, _ = cmd.Flags().GetBool("count")
		}

		method := resInfo.APIPrefix + ".get"
		params := map[string]interface{}{
			"output": "extend",
		}
		if countFlag {
			params["countOutput"] = true
		}

		isSingleTarget := len(args) == 2
		shouldExportManifest := !countFlag && (exportFlag || (isSingleTarget && (formatter.Format == "yaml" || formatter.Format == "json")))

		// When inspecting a single resource or exporting declarative manifests, include extended relations
		if shouldExportManifest {
			switch resInfo.APIPrefix {
			case "template":
				params["selectTags"] = "extend"
				params["selectGroups"] = "extend"
				params["selectMacros"] = "extend"
				params["selectTemplates"] = "extend"
			case "host":
				params["selectTags"] = "extend"
				params["selectGroups"] = "extend"
				params["selectMacros"] = "extend"
				params["selectParentTemplates"] = "extend"
				params["selectInterfaces"] = "extend"
			case "hostgroup":
				params["selectTags"] = "extend"
			case "item":
				params["selectTags"] = "extend"
				params["selectPreprocessing"] = "extend"
			case "trigger":
				params["selectTags"] = "extend"
				params["selectDependencies"] = "extend"
			}
		}

		if limitFlag > 0 {
			params["limit"] = limitFlag
		}

		if hostFlag != "" {
			if isNumeric(hostFlag) {
				params["hostids"] = []string{hostFlag}
			} else {
				// Resolve host ID by name or hostname
				hostRes, err := checkSafetyAndCall(cmd.Context(), "host.get", map[string]interface{}{
					"output": []string{"hostid"},
					"filter": map[string]interface{}{"name": hostFlag},
				})
				var hosts []map[string]interface{}
				if err == nil {
					_ = json.Unmarshal(hostRes.(json.RawMessage), &hosts)
				}
				if len(hosts) == 0 {
					// Retry filter by host key
					hostRes, err = checkSafetyAndCall(cmd.Context(), "host.get", map[string]interface{}{
						"output": []string{"hostid"},
						"filter": map[string]interface{}{"host": hostFlag},
					})
					if err == nil {
						_ = json.Unmarshal(hostRes.(json.RawMessage), &hosts)
					}
				}
				if len(hosts) > 0 {
					if hID, ok := hosts[0]["hostid"].(string); ok {
						params["hostids"] = []string{hID}
					}
				} else {
					return fmt.Errorf("host %q not found", hostFlag)
				}
			}
		}

		var fieldList []string
		if fieldsFlag != "" {
			for _, f := range strings.Split(fieldsFlag, ",") {
				if trimmed := strings.TrimSpace(f); trimmed != "" {
					fieldList = append(fieldList, trimmed)
				}
			}
			// Pass field list directly to Zabbix API if it's not a resource that requires relation lookup for enrichment
			if len(fieldList) > 0 && resInfo.APIPrefix != "problem" && resInfo.APIPrefix != "event" {
				params["output"] = fieldList
			}
		}

		if len(args) == 2 {
			identifier := args[1]
			// Check if identifier is numeric ID or string name
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
		}

		if filterFlag != "" {
			var filterMap map[string]interface{}
			if err := json.Unmarshal([]byte(filterFlag), &filterMap); err == nil {
				params["filter"] = filterMap
			} else {
				// Try key=value
				parts := strings.SplitN(filterFlag, "=", 2)
				if len(parts) == 2 {
					params["filter"] = map[string]interface{}{parts[0]: parts[1]}
				}
			}
		}

		if searchFlag != "" {
			params["searchWildcardsEnabled"] = true
			if strings.Contains(searchFlag, "=") {
				parts := strings.SplitN(searchFlag, "=", 2)
				params["search"] = map[string]interface{}{parts[0]: parts[1]}
			} else {
				var targetSearchFields []string
				if searchFieldsFlag != "" {
					for _, sf := range strings.Split(searchFieldsFlag, ",") {
						if trimmed := strings.TrimSpace(sf); trimmed != "" {
							targetSearchFields = append(targetSearchFields, trimmed)
						}
					}
				} else if len(resInfo.SearchFields) > 0 {
					targetSearchFields = resInfo.SearchFields
				} else if resInfo.NameProperty != "" {
					targetSearchFields = []string{resInfo.NameProperty}
				}

				if len(targetSearchFields) > 0 {
					searchMap := make(map[string]interface{})
					for _, f := range targetSearchFields {
						searchMap[f] = searchFlag
					}
					params["search"] = searchMap
					if len(targetSearchFields) > 1 {
						params["searchByAny"] = true
					}
				}
			}
		}

		if sortFlag != "" {
			sortField := sortFlag
			sortOrder := sortOrderFlag
			if strings.Contains(sortField, ":") {
				parts := strings.SplitN(sortField, ":", 2)
				sortField = parts[0]
				if sortOrder == "" {
					sortOrder = parts[1]
				}
			}
			params["sortfield"] = sortField
			if sortOrder != "" {
				params["sortorder"] = strings.ToUpper(sortOrder)
			} else {
				params["sortorder"] = "ASC"
			}
		} else if resInfo.Name == "history" {
			params["sortfield"] = "clock"
			if sortOrderFlag != "" {
				params["sortorder"] = strings.ToUpper(sortOrderFlag)
			} else {
				params["sortorder"] = "DESC"
			}
		} else if sortOrderFlag != "" {
			params["sortorder"] = strings.ToUpper(sortOrderFlag)
		}

		if resInfo.Name == "history" {
			params["history"] = historyTypeFlag
			if sinceFlag != "" {
				dur, err := parseSinceDuration(sinceFlag)
				if err != nil {
					return fmt.Errorf("invalid --since duration format %q (e.g., 4h, 30m, 1d): %w", sinceFlag, err)
				}
				params["time_from"] = time.Now().Add(-dur).Unix()
			}
		}

		res, err := checkSafetyAndCall(cmd.Context(), method, params)
		if err != nil {
			return err
		}

		// Enrich problems / events with host, status, and ack relations
		if resInfo.APIPrefix == "problem" {
			if raw, ok := res.(json.RawMessage); ok {
				res = enrichProblems(cmd.Context(), raw)
			}
		} else if resInfo.APIPrefix == "event" {
			if raw, ok := res.(json.RawMessage); ok {
				res = enrichEvents(cmd.Context(), raw)
			}
		}

		if countFlag {
			count := parseCountResult(res)
			if formatter.Format == "json" || formatter.Format == "toon" {
				return formatter.Print(map[string]int{"count": count})
			} else if formatter.Format == "yaml" {
				return formatter.Print(map[string]int{"count": count})
			}
			fmt.Fprintln(formatter.Writer, count)
			return nil
		}

		if shouldExportManifest {
			return outputAsExportManifest(res, resInfo.Name, isSingleTarget)
		}

		return formatter.PrintResource(res, resInfo.Name, fieldList, allFieldsFlag || wideFlag)
	},
}

func outputAsExportManifest(res interface{}, kind string, singleTarget bool) error {
	raw, ok := res.(json.RawMessage)
	if !ok {
		return formatter.Print(res)
	}

	var items []map[string]interface{}
	if err := json.Unmarshal(raw, &items); err != nil {
		var single map[string]interface{}
		if err2 := json.Unmarshal(raw, &single); err2 == nil {
			return formatter.Print(ManifestItem{
				Kind: kind,
				Spec: zabbix.SanitizeExportSpec(kind, single),
			})
		}
		return formatter.Print(res)
	}

	if len(items) == 0 {
		return formatter.Print(items)
	}

	if singleTarget || len(items) == 1 {
		return formatter.Print(ManifestItem{
			Kind: kind,
			Spec: zabbix.SanitizeExportSpec(kind, items[0]),
		})
	}

	var manifests []ManifestItem
	for _, it := range items {
		manifests = append(manifests, ManifestItem{
			Kind: kind,
			Spec: zabbix.SanitizeExportSpec(kind, it),
		})
	}
	return formatter.Print(manifests)
}

func parseSinceDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "d") {
		daysStr := strings.TrimSuffix(s, "d")
		var days int
		if _, err := fmt.Sscanf(daysStr, "%d", &days); err == nil {
			return time.Duration(days) * 24 * time.Hour, nil
		}
	}
	return time.ParseDuration(s)
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func init() {
	getCmd.Flags().StringVar(&getFilterFlag, "filter", "", "filter by JSON map or key=value")
	getCmd.Flags().StringVar(&getSearchFlag, "search", "", "search pattern (searches across primary text fields or key=value)")
	getCmd.Flags().StringVar(&getSearchFieldsFlag, "search-fields", "", "comma-separated list of fields to search across when --search is used")
	getCmd.Flags().StringVarP(&getSortFlag, "sort", "s", "", "sort by field name (e.g. name, host, clock, severity, or name:desc)")
	getCmd.Flags().StringVar(&getSortOrderFlag, "sort-order", "", "sort order direction (asc or desc, default asc)")
	getCmd.Flags().IntVar(&getLimitFlag, "limit", 0, "limit maximum number of returned items")
	getCmd.Flags().StringVar(&getSinceFlag, "since", "", "fetch history since duration (e.g. 4h, 30m, 1d)")
	getCmd.Flags().IntVar(&getHistoryTypeFlag, "history-type", 0, "Zabbix history value type (0=float, 1=str, 2=log, 3=uint, 4=text)")
	getCmd.Flags().StringVarP(&getHostFlag, "host", "H", "", "filter by target host ID or host name")
	getCmd.Flags().StringVarP(&getFieldsFlag, "fields", "f", "", "comma-separated list of output fields (e.g. eventid,name,clock)")
	getCmd.Flags().BoolVar(&getAllFieldsFlag, "all-fields", false, "display all available resource fields in table view")
	getCmd.Flags().BoolVarP(&getWideFlag, "wide", "w", false, "display wide output with all fields")
	getCmd.Flags().BoolVar(&getExportFlag, "export", false, "export as declarative GitOps manifests (kind + spec) with extended relations")
	getCmd.Flags().BoolVar(&getCountFlag, "count", false, "return only the total count of matched resources")
	RootCmd.AddCommand(getCmd)
}
