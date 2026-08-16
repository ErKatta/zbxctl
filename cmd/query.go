package cmd

import (
	"encoding/json"
	"strings"

	"github.com/ErKatta/zbxctl/pkg/zabbix"
	"github.com/spf13/cobra"
)

var (
	queryFilterFlag       string
	querySearchFlag       string
	querySearchFieldsFlag string
	queryFieldsFlag       string
	querySortFlag         string
	querySortOrderFlag    string
	queryLimitFlag        int
	queryAllFieldsFlag    bool
	queryWideFlag         bool
)

var queryCmd = &cobra.Command{
	Use:   "query <resource> [--filter='...'] [--search='...']",
	Short: "Advanced query and filter across Zabbix resources",
	Long:  `query searches and filters resources across standard Zabbix 7 endpoints with projection, pattern matching, and sorting.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resInfo, err := zabbix.ResolveResource(args[0])
		if err != nil {
			return err
		}

		var filterFlag, searchFlag, searchFieldsFlag, fieldsFlag, sortFlag, sortOrderFlag string
		var limitFlag int
		var allFieldsFlag, wideFlag bool

		if cmd.Flags().Changed("filter") {
			filterFlag, _ = cmd.Flags().GetString("filter")
		}
		if cmd.Flags().Changed("search") {
			searchFlag, _ = cmd.Flags().GetString("search")
		}
		if cmd.Flags().Changed("search-fields") {
			searchFieldsFlag, _ = cmd.Flags().GetString("search-fields")
		}
		if cmd.Flags().Changed("fields") {
			fieldsFlag, _ = cmd.Flags().GetString("fields")
		}
		if cmd.Flags().Changed("sort") {
			sortFlag, _ = cmd.Flags().GetString("sort")
		}
		if cmd.Flags().Changed("sort-order") {
			sortOrderFlag, _ = cmd.Flags().GetString("sort-order")
		}
		if cmd.Flags().Changed("limit") {
			limitFlag, _ = cmd.Flags().GetInt("limit")
		}
		if cmd.Flags().Changed("all-fields") {
			allFieldsFlag, _ = cmd.Flags().GetBool("all-fields")
		}
		if cmd.Flags().Changed("wide") {
			wideFlag, _ = cmd.Flags().GetBool("wide")
		}

		method := resInfo.APIPrefix + ".get"
		params := map[string]interface{}{
			"output": "extend",
		}

		if resInfo.Name == "inventory" {
			params["selectInventory"] = "extend"
			params["output"] = []string{"hostid", "host", "name", "inventory_mode"}
		}

		var fieldList []string
		if fieldsFlag != "" {
			for _, f := range strings.Split(fieldsFlag, ",") {
				if trimmed := strings.TrimSpace(f); trimmed != "" {
					fieldList = append(fieldList, trimmed)
				}
			}
			if len(fieldList) > 0 && resInfo.APIPrefix != "problem" && resInfo.APIPrefix != "event" && resInfo.Name != "inventory" {
				params["output"] = fieldList
			}
		}

		if limitFlag > 0 {
			params["limit"] = limitFlag
		}

		if filterFlag != "" {
			var filterMap map[string]interface{}
			if err := json.Unmarshal([]byte(filterFlag), &filterMap); err == nil {
				params["filter"] = filterMap
			} else {
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

		res, err := checkSafetyAndCall(cmd.Context(), method, params)
		if err != nil {
			return err
		}

		// Enrich problems / events / inventory relations
		if resInfo.APIPrefix == "problem" {
			if raw, ok := res.(json.RawMessage); ok {
				res = enrichProblems(cmd.Context(), raw)
			}
		} else if resInfo.APIPrefix == "event" {
			if raw, ok := res.(json.RawMessage); ok {
				res = enrichEvents(cmd.Context(), raw)
			}
		} else if resInfo.Name == "inventory" {
			if raw, ok := res.(json.RawMessage); ok {
				res = enrichInventory(raw)
			}
		}

		return formatter.PrintResource(res, resInfo.Name, fieldList, allFieldsFlag || wideFlag)
	},
}

func init() {
	queryCmd.Flags().StringVar(&queryFilterFlag, "filter", "", "filter criteria (JSON map or key=value)")
	queryCmd.Flags().StringVar(&querySearchFlag, "search", "", "search pattern (searches across primary text fields or key=value)")
	queryCmd.Flags().StringVar(&querySearchFieldsFlag, "search-fields", "", "comma-separated list of fields to search across when --search is used")
	queryCmd.Flags().StringVarP(&queryFieldsFlag, "fields", "f", "", "comma-separated list of output fields (e.g. eventid,name,clock)")
	queryCmd.Flags().StringVarP(&querySortFlag, "sort", "s", "", "sort by field name (e.g. name, host, clock, severity, or name:desc)")
	queryCmd.Flags().StringVar(&querySortOrderFlag, "sort-order", "", "sort order direction (asc or desc, default asc)")
	queryCmd.Flags().IntVar(&queryLimitFlag, "limit", 0, "limit output result size")
	queryCmd.Flags().BoolVar(&queryAllFieldsFlag, "all-fields", false, "display all available resource fields in table view")
	queryCmd.Flags().BoolVarP(&queryWideFlag, "wide", "w", false, "display wide output with all fields")
	RootCmd.AddCommand(queryCmd)
}
