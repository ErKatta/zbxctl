package cmd

import (
	"context"
	"encoding/json"
	"strings"
)

func enrichProblems(ctx context.Context, raw json.RawMessage) json.RawMessage {
	var items []map[string]interface{}
	if err := json.Unmarshal(raw, &items); err != nil || len(items) == 0 {
		return raw
	}

	var triggerIDs []string
	for _, it := range items {
		// Collect objectid (which is triggerid for trigger problems where object == "0" or omitted)
		obj, hasObj := it["object"].(string)
		if !hasObj || obj == "0" || obj == "" {
			if objID, ok := it["objectid"].(string); ok && objID != "" && objID != "0" {
				triggerIDs = append(triggerIDs, objID)
			}
		}
	}

	triggerHostMap := make(map[string]string)
	if len(triggerIDs) > 0 {
		uniqueIDs := uniqueStrings(triggerIDs)
		tRes, err := checkSafetyAndCall(ctx, "trigger.get", map[string]interface{}{
			"output":      []string{"triggerid"},
			"triggerids":  uniqueIDs,
			"selectHosts": []string{"hostid", "host", "name"},
		})
		if err == nil {
			if tRaw, ok := tRes.(json.RawMessage); ok {
				var triggers []map[string]interface{}
				if err := json.Unmarshal(tRaw, &triggers); err == nil {
					for _, tr := range triggers {
						tID, _ := tr["triggerid"].(string)
						if hosts, ok := tr["hosts"].([]interface{}); ok && len(hosts) > 0 {
							if hMap, ok := hosts[0].(map[string]interface{}); ok {
								hName, _ := hMap["name"].(string)
								if hName == "" {
									hName, _ = hMap["host"].(string)
								}
								if tID != "" && hName != "" {
									triggerHostMap[tID] = hName
								}
							}
						}
					}
				}
			}
		}
	}

	for i, it := range items {
		// 1. Resolve host
		if it["host"] == nil || it["host"] == "" {
			objID, _ := it["objectid"].(string)
			if hName, found := triggerHostMap[objID]; found {
				items[i]["host"] = hName
			} else if tags, ok := it["tags"].([]interface{}); ok {
				for _, tagObj := range tags {
					if tm, ok := tagObj.(map[string]interface{}); ok {
						tName, _ := tm["tag"].(string)
						if strings.EqualFold(tName, "host") || strings.EqualFold(tName, "hostname") {
							items[i]["host"] = tm["value"]
							break
						}
					}
				}
			}
		}

		// 2. Resolve status
		if it["status"] == nil || it["status"] == "" {
			if rEv, ok := it["r_eventid"].(string); ok && rEv != "0" && rEv != "" {
				items[i]["status"] = "RESOLVED"
			} else if sup, ok := it["suppressed"].(string); ok && sup == "1" {
				items[i]["status"] = "SUPPRESSED"
			} else {
				items[i]["status"] = "PROBLEM"
			}
		}

		// 3. Resolve ack
		if it["ack"] == nil || it["ack"] == "" {
			if ack, ok := it["acknowledged"].(string); ok && ack == "1" {
				items[i]["ack"] = "Yes"
			} else if ack, ok := it["acknowledged"].(float64); ok && ack == 1 {
				items[i]["ack"] = "Yes"
			} else {
				items[i]["ack"] = "No"
			}
		}
	}

	updated, err := json.Marshal(items)
	if err != nil {
		return raw
	}
	return json.RawMessage(updated)
}

func enrichEvents(ctx context.Context, raw json.RawMessage) json.RawMessage {
	var items []map[string]interface{}
	if err := json.Unmarshal(raw, &items); err != nil || len(items) == 0 {
		return raw
	}

	var triggerIDs []string
	for _, it := range items {
		if it["host"] == nil || it["host"] == "" {
			if hosts, ok := it["hosts"].([]interface{}); ok && len(hosts) > 0 {
				if hMap, ok := hosts[0].(map[string]interface{}); ok {
					hName, _ := hMap["name"].(string)
					if hName == "" {
						hName, _ = hMap["host"].(string)
					}
					it["host"] = hName
				}
			} else {
				if objID, ok := it["objectid"].(string); ok && objID != "" && objID != "0" {
					triggerIDs = append(triggerIDs, objID)
				}
			}
		}
	}

	triggerHostMap := make(map[string]string)
	if len(triggerIDs) > 0 {
		uniqueIDs := uniqueStrings(triggerIDs)
		tRes, err := checkSafetyAndCall(ctx, "trigger.get", map[string]interface{}{
			"output":      []string{"triggerid"},
			"triggerids":  uniqueIDs,
			"selectHosts": []string{"hostid", "host", "name"},
		})
		if err == nil {
			if tRaw, ok := tRes.(json.RawMessage); ok {
				var triggers []map[string]interface{}
				if err := json.Unmarshal(tRaw, &triggers); err == nil {
					for _, tr := range triggers {
						tID, _ := tr["triggerid"].(string)
						if hosts, ok := tr["hosts"].([]interface{}); ok && len(hosts) > 0 {
							if hMap, ok := hosts[0].(map[string]interface{}); ok {
								hName, _ := hMap["name"].(string)
								if hName == "" {
									hName, _ = hMap["host"].(string)
								}
								if tID != "" && hName != "" {
									triggerHostMap[tID] = hName
								}
							}
						}
					}
				}
			}
		}
	}

	for i, it := range items {
		if it["host"] == nil || it["host"] == "" {
			objID, _ := it["objectid"].(string)
			if hName, found := triggerHostMap[objID]; found {
				items[i]["host"] = hName
			}
		}

		if it["status"] == nil || it["status"] == "" {
			if val, ok := it["value"].(string); ok {
				if val == "1" {
					items[i]["status"] = "PROBLEM"
				} else {
					items[i]["status"] = "OK"
				}
			}
		}

		if it["ack"] == nil || it["ack"] == "" {
			if ack, ok := it["acknowledged"].(string); ok && ack == "1" {
				items[i]["ack"] = "Yes"
			} else if ack, ok := it["acknowledged"].(float64); ok && ack == 1 {
				items[i]["ack"] = "Yes"
			} else {
				items[i]["ack"] = "No"
			}
		}
	}

	updated, err := json.Marshal(items)
	if err != nil {
		return raw
	}
	return json.RawMessage(updated)
}

func uniqueStrings(input []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range input {
		if !seen[s] && s != "" {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
