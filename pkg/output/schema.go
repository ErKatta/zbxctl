package output

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
)

// ColumnDefinition defines a standard table column with header, field key lookups, and custom cell formatting.
type ColumnDefinition struct {
	Header    string
	Keys      []string
	Formatter func(val interface{}, row map[string]interface{}) string
}

func (col ColumnDefinition) ExtractValue(row map[string]interface{}, resource string) string {
	var rawVal interface{}
	var foundKey string

	for _, k := range col.Keys {
		if v, exists := row[k]; exists && v != nil {
			rawVal = v
			foundKey = k
			break
		}
		// Try case-insensitive matching
		for rk, rv := range row {
			if strings.EqualFold(rk, k) && rv != nil {
				rawVal = rv
				foundKey = rk
				break
			}
		}
		if rawVal != nil {
			break
		}
	}

	if col.Formatter != nil {
		return col.Formatter(rawVal, row)
	}

	return FormatDefaultCell(foundKey, rawVal, resource)
}

func formatSeverityCol(val interface{}, row map[string]interface{}) string {
	return FormatSeverity(val)
}

func formatProblemStatusCol(val interface{}, row map[string]interface{}) string {
	str := strings.TrimSpace(fmt.Sprintf("%v", val))
	if val == nil || str == "" || str == "<nil>" || str == "0" {
		if row != nil {
			if rEv, ok := row["r_eventid"].(string); ok && rEv != "0" && rEv != "" {
				return color.GreenString("RESOLVED")
			}
			if sup, ok := row["suppressed"].(string); ok && sup == "1" {
				return color.YellowString("SUPPRESSED")
			}
		}
		return color.RedString("PROBLEM")
	}

	if strings.EqualFold(str, "PROBLEM") || str == "1" {
		return color.RedString("PROBLEM")
	}
	if strings.EqualFold(str, "RESOLVED") {
		return color.GreenString("RESOLVED")
	}
	if strings.EqualFold(str, "SUPPRESSED") {
		return color.YellowString("SUPPRESSED")
	}
	return str
}

func formatAgeCol(val interface{}, row map[string]interface{}) string {
	return FormatClockRelative(val)
}

func formatAckCol(val interface{}, row map[string]interface{}) string {
	if val == nil {
		return color.HiBlackString("No")
	}
	str := strings.TrimSpace(fmt.Sprintf("%v", val))
	if str == "1" || strings.EqualFold(str, "yes") || strings.EqualFold(str, "true") {
		return color.GreenString("Yes")
	}
	return color.HiBlackString("No")
}

func formatHostStatusCol(val interface{}, row map[string]interface{}) string {
	if val == nil {
		return ""
	}
	str := strings.TrimSpace(fmt.Sprintf("%v", val))
	if str == "0" || strings.EqualFold(str, "monitored") {
		return color.GreenString("Monitored")
	}
	if str == "1" || strings.EqualFold(str, "unmonitored") {
		return color.HiBlackString("Unmonitored")
	}
	return str
}

func formatItemStatusCol(val interface{}, row map[string]interface{}) string {
	if val == nil {
		return ""
	}
	str := strings.TrimSpace(fmt.Sprintf("%v", val))
	if str == "0" || strings.EqualFold(str, "enabled") {
		return color.GreenString("Enabled")
	}
	if str == "1" || strings.EqualFold(str, "disabled") {
		return color.HiBlackString("Disabled")
	}
	return str
}

func formatTriggerValueCol(val interface{}, row map[string]interface{}) string {
	if val == nil {
		return ""
	}
	str := strings.TrimSpace(fmt.Sprintf("%v", val))
	if str == "0" || strings.EqualFold(str, "ok") {
		return color.GreenString("OK")
	}
	if str == "1" || strings.EqualFold(str, "problem") {
		return color.RedString("PROBLEM")
	}
	return str
}

func formatAvailabilityCol(val interface{}, row map[string]interface{}) string {
	if val == nil {
		return ""
	}
	str := strings.TrimSpace(fmt.Sprintf("%v", val))
	if str == "1" || strings.EqualFold(str, "available") {
		return color.GreenString("Available")
	}
	if str == "2" || strings.EqualFold(str, "unavailable") {
		return color.RedString("Unavailable")
	}
	if str == "0" || strings.EqualFold(str, "unknown") {
		return color.HiBlackString("Unknown")
	}
	return str
}

func formatItemTypeCol(val interface{}, row map[string]interface{}) string {
	if val == nil {
		return ""
	}
	str := strings.TrimSpace(fmt.Sprintf("%v", val))
	itemTypes := map[string]string{
		"0":  "Agent",
		"1":  "SNMPv1",
		"2":  "Trapper",
		"3":  "Simple",
		"4":  "SNMPv2c",
		"5":  "Internal",
		"6":  "SNMPv3",
		"7":  "Agent (act)",
		"9":  "Web",
		"10": "External",
		"11": "DB monitor",
		"12": "IPMI",
		"13": "SSH",
		"14": "TELNET",
		"15": "Calculated",
		"16": "JMX",
		"17": "SNMP trap",
		"18": "Dependent",
		"19": "HTTP",
		"20": "SNMP agent",
		"21": "Prototype",
	}
	if name, found := itemTypes[str]; found {
		return name
	}
	return str
}

func formatDateTimeCol(val interface{}, row map[string]interface{}) string {
	if epoch, ok := ParseEpoch(val); ok && epoch > 100000000 {
		return time.Unix(epoch, 0).Format("2006-01-02 15:04:05")
	}
	if val == nil {
		return "-"
	}
	return fmt.Sprintf("%v", val)
}

func formatSLOCol(val interface{}, row map[string]interface{}) string {
	if val == nil {
		return ""
	}
	str := fmt.Sprintf("%v", val)
	if !strings.HasSuffix(str, "%") {
		return str + "%"
	}
	return str
}

func formatPrivateCol(val interface{}, row map[string]interface{}) string {
	str := strings.TrimSpace(fmt.Sprintf("%v", val))
	if str == "1" || strings.EqualFold(str, "private") {
		return "Private"
	}
	return "Public"
}

// KnownResourceAttributes lists known standard field attributes for each resource type.
var KnownResourceAttributes = map[string][]string{
	"host":        {"hostid", "host", "name", "status", "description", "proxy_hostid", "ip", "dns", "port", "available", "error", "tls_connect", "tls_accept"},
	"template":    {"templateid", "name", "host", "description", "uuid", "vendor_name", "vendor_version"},
	"problem":     {"eventid", "source", "object", "objectid", "clock", "ns", "r_eventid", "r_clock", "r_ns", "correlationid", "userid", "name", "acknowledged", "severity", "cause_eventid", "opdata", "suppressed", "status", "host", "hostname", "age"},
	"item":        {"itemid", "type", "snmp_oid", "hostid", "name", "key_", "delay", "history", "trends", "status", "value_type", "trapper_hosts", "units", "description", "lastvalue", "prevvalue", "lastclock", "error", "host"},
	"trigger":     {"triggerid", "expression", "description", "url", "status", "value", "priority", "lastchange", "comments", "error", "templateid", "type", "state", "flags", "recovery_mode", "recovery_expression", "correlation_mode", "correlation_tag", "manual_close", "opdata", "event_name"},
	"hostgroup":   {"groupid", "name", "flags", "uuid"},
	"maintenance": {"maintenanceid", "name", "maintenance_type", "description", "active_since", "active_till", "tags_evaltype"},
	"event":       {"eventid", "source", "object", "objectid", "clock", "value", "acknowledged", "ns", "name", "severity", "r_eventid", "c_eventid", "correlationid", "userid", "suppressed", "opdata", "urls", "host", "hostname", "age", "status"},
	"user":        {"userid", "username", "name", "surname", "url", "autologin", "autologout", "lang", "refresh", "theme", "roleid", "userdirectoryid", "ts_provisioned"},
	"service":     {"serviceid", "name", "status", "algorithm", "description", "sortorder", "weight", "propagation_rule", "propagation_value", "readonly"},
	"sla":         {"slaid", "name", "period", "slo", "effective_date", "timezone", "status", "description", "service_tags"},
	"dashboard":   {"dashboardid", "name", "userid", "private", "display_period", "auto_start"},
	"history":     {"itemid", "clock", "value", "ns"},
}

// ResourceSchemas maps each resource type to its standard, curated table columns.
var ResourceSchemas = map[string][]ColumnDefinition{
	"problem": {
		{Header: "EVENTID", Keys: []string{"eventid"}},
		{Header: "HOST", Keys: []string{"host", "hostname", "host_name"}},
		{Header: "PROBLEM", Keys: []string{"name", "problem", "description"}},
		{Header: "SEVERITY", Keys: []string{"severity", "priority"}, Formatter: formatSeverityCol},
		{Header: "STATUS", Keys: []string{"status", "value"}, Formatter: formatProblemStatusCol},
		{Header: "AGE", Keys: []string{"age", "clock"}, Formatter: formatAgeCol},
		{Header: "ACK", Keys: []string{"ack", "acknowledged"}, Formatter: formatAckCol},
	},
	"host": {
		{Header: "HOSTID", Keys: []string{"hostid"}},
		{Header: "HOST", Keys: []string{"host"}},
		{Header: "NAME", Keys: []string{"name"}},
		{Header: "STATUS", Keys: []string{"status"}, Formatter: formatHostStatusCol},
		{Header: "AVAILABILITY", Keys: []string{"available", "availability"}, Formatter: formatAvailabilityCol},
	},
	"item": {
		{Header: "ITEMID", Keys: []string{"itemid"}},
		{Header: "HOST", Keys: []string{"host", "hostid"}},
		{Header: "NAME", Keys: []string{"name"}},
		{Header: "KEY", Keys: []string{"key_", "key"}},
		{Header: "TYPE", Keys: []string{"type"}, Formatter: formatItemTypeCol},
		{Header: "STATUS", Keys: []string{"status"}, Formatter: formatItemStatusCol},
		{Header: "DELAY", Keys: []string{"delay"}},
		{Header: "LASTVALUE", Keys: []string{"lastvalue", "prevvalue", "value"}},
	},
	"trigger": {
		{Header: "TRIGGERID", Keys: []string{"triggerid"}},
		{Header: "TRIGGER", Keys: []string{"description", "name"}},
		{Header: "SEVERITY", Keys: []string{"priority", "severity"}, Formatter: formatSeverityCol},
		{Header: "STATUS", Keys: []string{"status"}, Formatter: formatItemStatusCol},
		{Header: "VALUE", Keys: []string{"value"}, Formatter: formatTriggerValueCol},
		{Header: "EXPRESSION", Keys: []string{"expression"}},
	},
	"template": {
		{Header: "TEMPLATEID", Keys: []string{"templateid"}},
		{Header: "NAME", Keys: []string{"name"}},
		{Header: "HOST", Keys: []string{"host"}},
	},
	"hostgroup": {
		{Header: "GROUPID", Keys: []string{"groupid"}},
		{Header: "NAME", Keys: []string{"name"}},
		{Header: "FLAGS", Keys: []string{"flags"}},
	},
	"maintenance": {
		{Header: "MAINTENANCEID", Keys: []string{"maintenanceid"}},
		{Header: "NAME", Keys: []string{"name"}},
		{Header: "TYPE", Keys: []string{"maintenance_type"}},
		{Header: "ACTIVE_SINCE", Keys: []string{"active_since"}, Formatter: formatDateTimeCol},
		{Header: "ACTIVE_TILL", Keys: []string{"active_till"}, Formatter: formatDateTimeCol},
	},
	"event": {
		{Header: "EVENTID", Keys: []string{"eventid"}},
		{Header: "HOST", Keys: []string{"host", "hostname"}},
		{Header: "NAME", Keys: []string{"name"}},
		{Header: "SEVERITY", Keys: []string{"severity", "priority"}, Formatter: formatSeverityCol},
		{Header: "STATUS", Keys: []string{"value", "status"}, Formatter: formatTriggerValueCol},
		{Header: "AGE", Keys: []string{"age", "clock"}, Formatter: formatAgeCol},
		{Header: "ACK", Keys: []string{"acknowledged", "ack"}, Formatter: formatAckCol},
	},
	"user": {
		{Header: "USERID", Keys: []string{"userid"}},
		{Header: "USERNAME", Keys: []string{"username", "alias"}},
		{Header: "NAME", Keys: []string{"name"}},
		{Header: "SURNAME", Keys: []string{"surname"}},
		{Header: "ROLEID", Keys: []string{"roleid"}},
		{Header: "AUTOLOGOUT", Keys: []string{"autologout"}},
	},
	"service": {
		{Header: "SERVICEID", Keys: []string{"serviceid"}},
		{Header: "NAME", Keys: []string{"name"}},
		{Header: "STATUS", Keys: []string{"status"}, Formatter: formatTriggerValueCol},
		{Header: "ALGORITHM", Keys: []string{"algorithm"}},
	},
	"sla": {
		{Header: "SLAID", Keys: []string{"slaid"}},
		{Header: "NAME", Keys: []string{"name"}},
		{Header: "SLO", Keys: []string{"slo"}, Formatter: formatSLOCol},
		{Header: "PERIOD", Keys: []string{"period"}},
		{Header: "TIMEZONE", Keys: []string{"timezone"}},
		{Header: "STATUS", Keys: []string{"status"}, Formatter: formatItemStatusCol},
	},
	"dashboard": {
		{Header: "DASHBOARDID", Keys: []string{"dashboardid"}},
		{Header: "NAME", Keys: []string{"name"}},
		{Header: "USERID", Keys: []string{"userid"}},
		{Header: "PRIVATE", Keys: []string{"private"}, Formatter: formatPrivateCol},
	},
	"history": {
		{Header: "ITEMID", Keys: []string{"itemid"}},
		{Header: "CLOCK", Keys: []string{"clock"}, Formatter: formatAgeCol},
		{Header: "VALUE", Keys: []string{"value"}},
		{Header: "NS", Keys: []string{"ns"}},
	},
}

// ValidateFields verifies that each field in fields exists either in rowMaps, ResourceSchemas[resource], or KnownResourceAttributes[resource].
func ValidateFields(data interface{}, resource string, fields []string) error {
	if len(fields) == 0 {
		return nil
	}

	var rawData interface{} = data
	if raw, ok := data.(json.RawMessage); ok {
		if err := json.Unmarshal(raw, &rawData); err != nil {
			return nil
		}
	}

	sliceValue := reflect.ValueOf(rawData)
	var rowMaps []map[string]interface{}
	if sliceValue.Kind() == reflect.Slice {
		for i := 0; i < sliceValue.Len(); i++ {
			item := sliceValue.Index(i).Interface()
			if itemMap, ok := item.(map[string]interface{}); ok {
				rowMaps = append(rowMaps, itemMap)
			}
		}
	} else if m, ok := rawData.(map[string]interface{}); ok {
		rowMaps = append(rowMaps, m)
	}

	return ValidateFieldsFromRowMaps(rowMaps, resource, fields)
}

// ValidateFieldsFromRowMaps validates that each field in fields exists either in rowMaps, ResourceSchemas[resource], or KnownResourceAttributes[resource].
func ValidateFieldsFromRowMaps(rowMaps []map[string]interface{}, resource string, fields []string) error {
	if len(fields) == 0 {
		return nil
	}

	normalizedRes := strings.ToLower(strings.TrimSpace(resource))
	availableMap := make(map[string]bool)

	for _, row := range rowMaps {
		for k := range row {
			availableMap[strings.ToLower(k)] = true
		}
	}

	if schemaCols, found := ResourceSchemas[normalizedRes]; found {
		for _, col := range schemaCols {
			availableMap[strings.ToLower(col.Header)] = true
			for _, k := range col.Keys {
				availableMap[strings.ToLower(k)] = true
			}
		}
	}

	if known, found := KnownResourceAttributes[normalizedRes]; found {
		for _, k := range known {
			availableMap[strings.ToLower(k)] = true
		}
	}

	for _, f := range fields {
		trimmed := strings.TrimSpace(f)
		if trimmed == "" {
			continue
		}
		if !availableMap[strings.ToLower(trimmed)] {
			var availableList []string
			for k := range availableMap {
				availableList = append(availableList, k)
			}
			sort.Strings(availableList)
			if normalizedRes != "" {
				return fmt.Errorf("field %q does not exist for resource %q (available fields: %s)", trimmed, resource, strings.Join(availableList, ", "))
			}
			return fmt.Errorf("field %q does not exist (available fields: %s)", trimmed, strings.Join(availableList, ", "))
		}
	}

	return nil
}

// FormatDefaultCell applies contextual formatting based on field name and value.
func FormatDefaultCell(key string, val interface{}, resource string) string {
	if val == nil {
		return ""
	}
	lowerKey := strings.ToLower(key)

	if lowerKey == "severity" || lowerKey == "priority" {
		return FormatSeverity(val)
	}
	if lowerKey == "clock" || lowerKey == "age" {
		return FormatClockRelative(val)
	}
	if lowerKey == "status" {
		switch resource {
		case "host":
			return formatHostStatusCol(val, nil)
		case "item", "trigger", "sla":
			return formatItemStatusCol(val, nil)
		case "event":
			return formatTriggerValueCol(val, nil)
		case "problem":
			return formatProblemStatusCol(val, nil)
		}
	}
	if lowerKey == "ack" || lowerKey == "acknowledged" {
		return formatAckCol(val, nil)
	}

	if reflect.TypeOf(val) != nil && (reflect.TypeOf(val).Kind() == reflect.Map || reflect.TypeOf(val).Kind() == reflect.Slice) {
		b, _ := json.Marshal(val)
		return string(b)
	}

	return fmt.Sprintf("%v", val)
}

func isEssentialColumn(resource, header string) bool {
	switch resource {
	case "problem":
		return header == "EVENTID" || header == "HOST" || header == "PROBLEM" || header == "SEVERITY" || header == "STATUS" || header == "AGE" || header == "ACK"
	case "host":
		return header == "HOSTID" || header == "HOST" || header == "NAME" || header == "STATUS"
	case "item":
		return header == "ITEMID" || header == "NAME" || header == "KEY" || header == "STATUS"
	case "trigger":
		return header == "TRIGGERID" || header == "TRIGGER" || header == "SEVERITY" || header == "STATUS" || header == "VALUE"
	case "template":
		return header == "TEMPLATEID" || header == "NAME" || header == "HOST"
	case "hostgroup":
		return header == "GROUPID" || header == "NAME"
	case "event":
		return header == "EVENTID" || header == "NAME" || header == "SEVERITY" || header == "STATUS" || header == "AGE"
	default:
		return false
	}
}
