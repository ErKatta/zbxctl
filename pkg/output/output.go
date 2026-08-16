package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

type Formatter struct {
	Format string
	Writer io.Writer
}

type PrintOptions struct {
	Resource  string
	Fields    []string
	AllFields bool
	Wide      bool
}

func NewFormatter(format string) *Formatter {
	return &Formatter{
		Format: strings.ToLower(format),
		Writer: os.Stdout,
	}
}

func IsTerminal() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

func (f *Formatter) Print(data interface{}) error {
	return f.PrintWithOptions(data, PrintOptions{})
}

func (f *Formatter) PrintResource(data interface{}, resource string, fields []string, allFields bool) error {
	return f.PrintWithOptions(data, PrintOptions{
		Resource:  resource,
		Fields:    fields,
		AllFields: allFields,
	})
}

func (f *Formatter) PrintWithOptions(data interface{}, opts PrintOptions) error {
	if len(opts.Fields) > 0 {
		if err := ValidateFields(data, opts.Resource, opts.Fields); err != nil {
			return err
		}
	}

	format := f.Format
	if opts.Wide || format == "wide" {
		format = "table"
		opts.AllFields = true
		opts.Wide = true
	} else if format == "" || format == "auto" {
		if IsTerminal() {
			format = "table"
		} else {
			format = "json"
		}
	}

	if len(opts.Fields) > 0 && format != "table" {
		data = filterDataByFields(data, opts.Fields)
	}

	switch format {
	case "json":
		return f.printJSON(data)
	case "yaml":
		return f.printYAML(data)
	case "toon":
		return f.printTOON(data)
	case "table":
		return f.printTableWithOptions(data, opts)
	default:
		return f.printJSON(data)
	}
}

func filterDataByFields(data interface{}, fields []string) interface{} {
	if len(fields) == 0 {
		return data
	}

	var rawData interface{} = data
	if raw, ok := data.(json.RawMessage); ok {
		if err := json.Unmarshal(raw, &rawData); err != nil {
			return data
		}
	}

	lowerFields := make(map[string]bool)
	for _, f := range fields {
		trimmed := strings.TrimSpace(f)
		if trimmed != "" {
			lowerFields[strings.ToLower(trimmed)] = true
		}
	}

	filterItem := func(m map[string]interface{}) map[string]interface{} {
		filtered := make(map[string]interface{})
		for k, v := range m {
			if lowerFields[strings.ToLower(k)] {
				filtered[k] = v
			}
		}
		return filtered
	}

	sliceValue := reflect.ValueOf(rawData)
	if sliceValue.Kind() == reflect.Slice {
		var filteredList []map[string]interface{}
		for i := 0; i < sliceValue.Len(); i++ {
			item := sliceValue.Index(i).Interface()
			if itemMap, ok := item.(map[string]interface{}); ok {
				filteredList = append(filteredList, filterItem(itemMap))
			}
		}
		return filteredList
	} else if m, ok := rawData.(map[string]interface{}); ok {
		return filterItem(m)
	}

	return data
}

func (f *Formatter) printJSON(data interface{}) error {
	var buf []byte
	var err error

	if raw, ok := data.(json.RawMessage); ok {
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, raw, "", "  "); err == nil {
			buf = pretty.Bytes()
		} else {
			buf = raw
		}
	} else {
		buf, err = json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("json marshal failed: %w", err)
		}
	}

	_, err = fmt.Fprintln(f.Writer, string(buf))
	return err
}

func (f *Formatter) printYAML(data interface{}) error {
	if raw, ok := data.(json.RawMessage); ok {
		var unmarshaled interface{}
		if err := json.Unmarshal(raw, &unmarshaled); err == nil {
			data = unmarshaled
		}
	}
	buf, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("yaml marshal failed: %w", err)
	}
	_, err = fmt.Fprintln(f.Writer, string(buf))
	return err
}

func (f *Formatter) printTOON(data interface{}) error {
	var buf []byte
	if raw, ok := data.(json.RawMessage); ok {
		var compact bytes.Buffer
		if err := json.Compact(&compact, raw); err == nil {
			buf = compact.Bytes()
		} else {
			buf = raw
		}
	} else {
		var err error
		buf, err = json.Marshal(data)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(f.Writer, string(buf))
	return err
}

func (f *Formatter) printTable(data interface{}) error {
	return f.printTableWithOptions(data, PrintOptions{})
}

func (f *Formatter) printTableWithOptions(data interface{}, opts PrintOptions) error {
	var rawData interface{} = data
	if raw, ok := data.(json.RawMessage); ok {
		if err := json.Unmarshal(raw, &rawData); err != nil {
			return f.printJSON(data)
		}
	}

	sliceValue := reflect.ValueOf(rawData)
	if sliceValue.Kind() != reflect.Slice {
		return f.printSingleTable(rawData)
	}

	if sliceValue.Len() == 0 {
		fmt.Fprintln(f.Writer, "No items found.")
		return nil
	}

	headers, rows, err := extractTableHeadersAndRows(sliceValue, opts)
	if err != nil {
		return err
	}
	if len(headers) == 0 {
		return f.printJSON(data)
	}

	table := tablewriter.NewWriter(f.Writer)
	table.SetAutoWrapText(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	table.SetHeader(headers)
	for _, row := range rows {
		table.Append(row)
	}
	table.Render()
	return nil
}

func (f *Formatter) printSingleTable(data interface{}) error {
	var m map[string]interface{}
	if mapData, ok := data.(map[string]interface{}); ok {
		m = mapData
	} else if data != nil {
		val := reflect.ValueOf(data)
		if val.Kind() == reflect.Struct || (val.Kind() == reflect.Ptr && val.Elem().Kind() == reflect.Struct) {
			b, err := json.Marshal(data)
			if err == nil {
				_ = json.Unmarshal(b, &m)
			}
		}
	}
	if m == nil {
		return f.printJSON(data)
	}

	table := tablewriter.NewWriter(f.Writer)
	table.SetHeader([]string{"KEY", "VALUE"})
	table.SetAutoWrapText(true)

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := m[k]
		valStr := fmt.Sprintf("%v", v)
		if reflect.TypeOf(v) != nil && (reflect.TypeOf(v).Kind() == reflect.Map || reflect.TypeOf(v).Kind() == reflect.Slice) {
			b, _ := json.Marshal(v)
			valStr = string(b)
		}
		table.Append([]string{k, valStr})
	}
	table.Render()
	return nil
}

func extractTableHeadersAndRows(sliceVal reflect.Value, optionalOpts ...PrintOptions) ([]string, [][]string, error) {
	if sliceVal.Len() == 0 {
		return nil, nil, nil
	}

	var opts PrintOptions
	if len(optionalOpts) > 0 {
		opts = optionalOpts[0]
	}

	var rowMaps []map[string]interface{}
	for i := 0; i < sliceVal.Len(); i++ {
		item := sliceVal.Index(i).Interface()
		if itemMap, ok := item.(map[string]interface{}); ok {
			rowMaps = append(rowMaps, itemMap)
		}
	}
	if len(rowMaps) == 0 {
		return nil, nil, nil
	}

	normalizedRes := strings.ToLower(strings.TrimSpace(opts.Resource))

	// 1. User-specified fields (--fields)
	if len(opts.Fields) > 0 {
		if err := ValidateFieldsFromRowMaps(rowMaps, normalizedRes, opts.Fields); err != nil {
			return nil, nil, err
		}

		var headers []string
		var colDefs []ColumnDefinition
		for _, f := range opts.Fields {
			trimmed := strings.TrimSpace(f)
			if trimmed == "" {
				continue
			}
			header := strings.ToUpper(trimmed)
			// Check if this field corresponds to a defined schema column
			var matchedDef *ColumnDefinition
			if cols, found := ResourceSchemas[normalizedRes]; found {
				for _, c := range cols {
					if strings.EqualFold(c.Header, trimmed) {
						matchedDef = &c
						break
					}
					for _, k := range c.Keys {
						if strings.EqualFold(k, trimmed) {
							matchedDef = &c
							break
						}
					}
					if matchedDef != nil {
						break
					}
				}
			}

			if matchedDef != nil {
				headers = append(headers, header)
				colDefs = append(colDefs, ColumnDefinition{
					Header:    header,
					Keys:      append([]string{strings.ToLower(trimmed), trimmed}, matchedDef.Keys...),
					Formatter: matchedDef.Formatter,
				})
			} else {
				headers = append(headers, header)
				colDefs = append(colDefs, ColumnDefinition{
					Header: header,
					Keys:   []string{strings.ToLower(trimmed), trimmed},
				})
			}
		}
		if len(headers) > 0 {
			var rows [][]string
			for _, row := range rowMaps {
				var r []string
				for _, col := range colDefs {
					r = append(r, col.ExtractValue(row, opts.Resource))
				}
				rows = append(rows, r)
			}
			return headers, rows, nil
		}
	}

	// 2. Predefined Resource Schema (when not --all-fields / wide)
	if !opts.AllFields && !opts.Wide && normalizedRes != "" {
		if cols, found := ResourceSchemas[normalizedRes]; found {
			var activeCols []ColumnDefinition
			var headers []string
			for _, col := range cols {
				hasValue := false
				for _, row := range rowMaps {
					for _, k := range col.Keys {
						if v, exists := row[k]; exists && v != nil && fmt.Sprintf("%v", v) != "" {
							hasValue = true
							break
						}
					}
					if hasValue {
						break
					}
				}
				if hasValue || isEssentialColumn(normalizedRes, col.Header) {
					activeCols = append(activeCols, col)
					headers = append(headers, col.Header)
				}
			}

			if len(activeCols) > 0 {
				var rows [][]string
				for _, row := range rowMaps {
					var r []string
					for _, col := range activeCols {
						r = append(r, col.ExtractValue(row, normalizedRes))
					}
					rows = append(rows, r)
				}
				return headers, rows, nil
			}
		}
	}

	// 3. Fallback / All Fields: Deterministic gathering of all keys from all items
	allKeysMap := make(map[string]bool)
	for _, row := range rowMaps {
		for k := range row {
			allKeysMap[k] = true
		}
	}

	priorityCols := []string{
		"hostid", "host", "name", "problem", "eventid", "itemid", "triggerid",
		"templateid", "groupid", "userid", "status", "severity", "priority",
		"value", "age", "clock", "ack", "acknowledged", "description",
	}

	headersMap := make(map[string]bool)
	var orderedHeaders []string

	for _, p := range priorityCols {
		if allKeysMap[p] {
			orderedHeaders = append(orderedHeaders, strings.ToUpper(p))
			headersMap[p] = true
		}
	}

	var remainingKeys []string
	for k := range allKeysMap {
		if !headersMap[k] {
			remainingKeys = append(remainingKeys, k)
		}
	}
	sort.Strings(remainingKeys)

	for _, k := range remainingKeys {
		if !opts.AllFields && !opts.Wide && len(orderedHeaders) >= 10 {
			break
		}
		orderedHeaders = append(orderedHeaders, strings.ToUpper(k))
		headersMap[k] = true
	}

	var rows [][]string
	for _, row := range rowMaps {
		var r []string
		for _, h := range orderedHeaders {
			key := strings.ToLower(h)
			val := row[key]
			if val == nil {
				val = row[h]
			}
			r = append(r, FormatDefaultCell(key, val, opts.Resource))
		}
		rows = append(rows, r)
	}

	return orderedHeaders, rows, nil
}

func FormatSeverity(val interface{}) string {
	sevStr := strings.TrimSpace(fmt.Sprintf("%v", val))
	switch sevStr {
	case "5", "disaster":
		return color.RedString("5 (Disaster)")
	case "4", "high":
		return color.MagentaString("4 (High)")
	case "3", "average":
		return color.YellowString("3 (Average)")
	case "2", "warning":
		return color.YellowString("2 (Warning)")
	case "1", "information":
		return color.CyanString("1 (Info)")
	default:
		return color.WhiteString("%s (Not classified)", sevStr)
	}
}

func PrintErrorEnvelope(w io.Writer, errObj interface{}) {
	data, _ := json.MarshalIndent(errObj, "", "  ")
	fmt.Fprintln(w, string(data))
}
