package zabbix

import (
	"fmt"
	"sort"
	"strings"
)

// SanitizeApplyParams normalizes and cleans manifest parameters so they strictly conform
// to Zabbix 7 JSON-RPC API mutation requirements (*.create, *.update).
func SanitizeApplyParams(resource string, rawParams map[string]interface{}) map[string]interface{} {
	if rawParams == nil {
		return make(map[string]interface{})
	}

	normRes := strings.ToLower(strings.TrimSpace(resource))
	params := make(map[string]interface{})
	for k, v := range rawParams {
		params[k] = v
	}

	switch normRes {
	case "host":
		// 1. Sanitize groups -> must be array of objects with groupid only
		if rawGroups, exists := params["groups"]; exists {
			params["groups"] = normalizeGroupIDs(rawGroups)
		}

		// 2. Map parentTemplates to templates -> array of objects with templateid only
		if pTmpls, exists := params["parentTemplates"]; exists {
			if _, hasTmpls := params["templates"]; !hasTmpls {
				params["templates"] = normalizeTemplateIDs(pTmpls)
			}
			delete(params, "parentTemplates")
		}
		if tmpls, exists := params["templates"]; exists {
			params["templates"] = normalizeTemplateIDs(tmpls)
		}

		// 3. Sanitize interfaces -> strip read-only telemetry fields
		if rawIfaces, exists := params["interfaces"]; exists {
			params["interfaces"] = sanitizeInterfaces(rawIfaces)
		}

		// 4. Strip read-only / unmodifiable host fields
		readOnlyHostKeys := []string{
			"flags", "active_available", "assigned_proxyid", "custom_interfaces",
			"maintenance_from", "maintenance_status", "maintenance_type", "maintenanceid",
			"monitored_by", "proxy_groupid", "proxyid", "templateid",
			"uuid", "vendor_name", "vendor_version", "discoveryRule", "hostDiscovery",
		}
		for _, k := range readOnlyHostKeys {
			if val, ok := params[k]; ok {
				strVal := strings.TrimSpace(fmt.Sprintf("%v", val))
				// Remove if it's a known read-only field or defaulted to 0/empty
				if k == "flags" || k == "active_available" || k == "custom_interfaces" ||
					k == "maintenance_from" || k == "maintenance_status" || k == "maintenance_type" ||
					k == "discoveryRule" || k == "hostDiscovery" ||
					strVal == "0" || strVal == "" || strVal == "-1" && (k == "proxyid" || k == "proxy_groupid") {
					delete(params, k)
				}
			}
		}

	case "template":
		// 1. Sanitize groups
		if rawGroups, exists := params["groups"]; exists {
			params["groups"] = normalizeGroupIDs(rawGroups)
		}

		// 2. Map parentTemplates to templates
		if pTmpls, exists := params["parentTemplates"]; exists {
			if _, hasTmpls := params["templates"]; !hasTmpls {
				params["templates"] = normalizeTemplateIDs(pTmpls)
			}
			delete(params, "parentTemplates")
		}
		if tmpls, exists := params["templates"]; exists {
			params["templates"] = normalizeTemplateIDs(tmpls)
		}

		// 3. Strip read-only template fields
		delete(params, "flags")
		delete(params, "hosts")
		delete(params, "discoveryRule")
		delete(params, "templateDiscovery")

	case "hostgroup":
		delete(params, "flags")
		if uuidVal, ok := params["uuid"]; ok && fmt.Sprintf("%v", uuidVal) == "" {
			delete(params, "uuid")
		}

	case "item":
		delete(params, "state")
		delete(params, "error")
		delete(params, "lastvalue")
		delete(params, "prevvalue")
		delete(params, "lastclock")
		delete(params, "lastns")
		delete(params, "flags")
		delete(params, "host")
		delete(params, "itemDiscovery")
		delete(params, "discoveryRule")
		if tmplID, ok := params["templateid"]; ok && fmt.Sprintf("%v", tmplID) == "0" {
			delete(params, "templateid")
		}

	case "trigger":
		delete(params, "value")
		delete(params, "state")
		delete(params, "error")
		delete(params, "lastchange")
		delete(params, "flags")
		delete(params, "hosts")
		delete(params, "functions")
		delete(params, "discoveryRule")
		delete(params, "triggerDiscovery")
	}

	return params
}

// SanitizeExportSpec cleans up an exported resource map before saving to a declarative manifest.
func SanitizeExportSpec(resource string, rawSpec map[string]interface{}) map[string]interface{} {
	if rawSpec == nil {
		return make(map[string]interface{})
	}

	normRes := strings.ToLower(strings.TrimSpace(resource))
	spec := make(map[string]interface{})
	for k, v := range rawSpec {
		spec[k] = v
	}

	switch normRes {
	case "host":
		// 1. Clean groups
		if rawGroups, exists := spec["groups"]; exists {
			spec["groups"] = cleanExportGroups(rawGroups)
		}

		// 2. Map parentTemplates to templates
		if pTmpls, exists := spec["parentTemplates"]; exists {
			spec["templates"] = cleanExportTemplates(pTmpls)
			delete(spec, "parentTemplates")
		} else if tmpls, exists := spec["templates"]; exists {
			spec["templates"] = cleanExportTemplates(tmpls)
		}

		// 3. Clean interfaces
		if rawIfaces, exists := spec["interfaces"]; exists {
			spec["interfaces"] = sanitizeInterfaces(rawIfaces)
		}

		// 4. Remove zero/empty noisy internal fields
		noisyKeys := []string{
			"flags", "active_available", "assigned_proxyid", "custom_interfaces",
			"maintenance_from", "maintenance_status", "maintenance_type", "maintenanceid",
			"monitored_by", "proxy_groupid", "proxyid", "templateid",
			"uuid", "vendor_name", "vendor_version", "discoveryRule", "hostDiscovery",
		}
		for _, k := range noisyKeys {
			if val, ok := spec[k]; ok {
				strVal := strings.TrimSpace(fmt.Sprintf("%v", val))
				if k == "flags" || k == "active_available" || k == "custom_interfaces" ||
					strVal == "0" || strVal == "" || strVal == "-1" && (k == "proxyid" || k == "proxy_groupid") {
					delete(spec, k)
				}
			}
		}

	case "template":
		if rawGroups, exists := spec["groups"]; exists {
			spec["groups"] = cleanExportGroups(rawGroups)
		}
		if pTmpls, exists := spec["parentTemplates"]; exists {
			spec["templates"] = cleanExportTemplates(pTmpls)
			delete(spec, "parentTemplates")
		} else if tmpls, exists := spec["templates"]; exists {
			spec["templates"] = cleanExportTemplates(tmpls)
		}
		delete(spec, "flags")
		delete(spec, "hosts")
	}

	return spec
}

// SanitizeDiffSpec normalizes both local and remote resource maps so that declarative
// diff comparison produces clean, meaningful diffs without false positives from read-only
// internal fields or different relation key naming.
func SanitizeDiffSpec(resource string, rawSpec map[string]interface{}) map[string]interface{} {
	if rawSpec == nil {
		return make(map[string]interface{})
	}

	normRes := strings.ToLower(strings.TrimSpace(resource))
	spec := make(map[string]interface{})
	for k, v := range rawSpec {
		spec[k] = v
	}

	switch normRes {
	case "host":
		// 1. Normalize groups -> array of map with groupid
		if rawGroups, exists := spec["groups"]; exists {
			spec["groups"] = normalizeGroupIDs(rawGroups)
		}

		// 2. Normalize templates -> array of map with templateid
		if pTmpls, exists := spec["parentTemplates"]; exists {
			if _, hasTmpls := spec["templates"]; !hasTmpls {
				spec["templates"] = normalizeTemplateIDs(pTmpls)
			}
			delete(spec, "parentTemplates")
		}
		if tmpls, exists := spec["templates"]; exists {
			spec["templates"] = normalizeTemplateIDs(tmpls)
		}

		// 3. Normalize interfaces
		if rawIfaces, exists := spec["interfaces"]; exists {
			spec["interfaces"] = sanitizeInterfaces(rawIfaces)
		}

		// 4. Normalize tags (sorted)
		if rawTags, exists := spec["tags"]; exists {
			spec["tags"] = normalizeTags(rawTags)
		}

		// 5. Normalize macros (sorted)
		if rawMacros, exists := spec["macros"]; exists {
			spec["macros"] = normalizeMacros(rawMacros)
		}

		// 6. Strip unmanaged/read-only noisy fields
		noisyKeys := []string{
			"flags", "active_available", "assigned_proxyid", "custom_interfaces",
			"maintenance_from", "maintenance_status", "maintenance_type", "maintenanceid",
			"monitored_by", "proxy_groupid", "proxyid", "templateid",
			"uuid", "vendor_name", "vendor_version", "discoveryRule", "hostDiscovery",
		}
		for _, k := range noisyKeys {
			if val, ok := spec[k]; ok {
				strVal := strings.TrimSpace(fmt.Sprintf("%v", val))
				if k == "flags" || k == "active_available" || k == "custom_interfaces" ||
					strVal == "0" || strVal == "" || strVal == "-1" && (k == "proxyid" || k == "proxy_groupid") {
					delete(spec, k)
				}
			}
		}

	case "template":
		if rawGroups, exists := spec["groups"]; exists {
			spec["groups"] = normalizeGroupIDs(rawGroups)
		}
		if pTmpls, exists := spec["parentTemplates"]; exists {
			if _, hasTmpls := spec["templates"]; !hasTmpls {
				spec["templates"] = normalizeTemplateIDs(pTmpls)
			}
			delete(spec, "parentTemplates")
		}
		if tmpls, exists := spec["templates"]; exists {
			spec["templates"] = normalizeTemplateIDs(tmpls)
		}
		if rawTags, exists := spec["tags"]; exists {
			spec["tags"] = normalizeTags(rawTags)
		}
		if rawMacros, exists := spec["macros"]; exists {
			spec["macros"] = normalizeMacros(rawMacros)
		}
		delete(spec, "flags")
		delete(spec, "hosts")
		delete(spec, "discoveryRule")
		delete(spec, "templateDiscovery")
	}

	return spec
}

func normalizeTags(tags interface{}) []map[string]interface{} {
	var result []map[string]interface{}
	switch tList := tags.(type) {
	case []interface{}:
		for _, it := range tList {
			if tMap, ok := it.(map[string]interface{}); ok {
				tag := fmt.Sprintf("%v", tMap["tag"])
				val := fmt.Sprintf("%v", tMap["value"])
				result = append(result, map[string]interface{}{"tag": tag, "value": val})
			}
		}
	case []map[string]interface{}:
		for _, tMap := range tList {
			tag := fmt.Sprintf("%v", tMap["tag"])
			val := fmt.Sprintf("%v", tMap["value"])
			result = append(result, map[string]interface{}{"tag": tag, "value": val})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return fmt.Sprintf("%v:%v", result[i]["tag"], result[i]["value"]) < fmt.Sprintf("%v:%v", result[j]["tag"], result[j]["value"])
	})
	return result
}

func normalizeMacros(macros interface{}) []map[string]interface{} {
	var result []map[string]interface{}
	switch mList := macros.(type) {
	case []interface{}:
		for _, it := range mList {
			if mMap, ok := it.(map[string]interface{}); ok {
				cleanMap := make(map[string]interface{})
				if m, ok := mMap["macro"]; ok {
					cleanMap["macro"] = fmt.Sprintf("%v", m)
				}
				if v, ok := mMap["value"]; ok {
					cleanMap["value"] = fmt.Sprintf("%v", v)
				}
				if len(cleanMap) > 0 {
					result = append(result, cleanMap)
				}
			}
		}
	case []map[string]interface{}:
		for _, mMap := range mList {
			cleanMap := make(map[string]interface{})
			if m, ok := mMap["macro"]; ok {
				cleanMap["macro"] = fmt.Sprintf("%v", m)
			}
			if v, ok := mMap["value"]; ok {
				cleanMap["value"] = fmt.Sprintf("%v", v)
			}
			if len(cleanMap) > 0 {
				result = append(result, cleanMap)
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return fmt.Sprintf("%v", result[i]["macro"]) < fmt.Sprintf("%v", result[j]["macro"])
	})
	return result
}

func normalizeGroupIDs(groups interface{}) []map[string]interface{} {
	var result []map[string]interface{}

	switch gList := groups.(type) {
	case []interface{}:
		for _, it := range gList {
			if gMap, ok := it.(map[string]interface{}); ok {
				if id, found := gMap["groupid"]; found && id != nil && fmt.Sprintf("%v", id) != "" {
					result = append(result, map[string]interface{}{
						"groupid": fmt.Sprintf("%v", id),
					})
				}
			} else if it != nil && fmt.Sprintf("%v", it) != "" {
				result = append(result, map[string]interface{}{
					"groupid": fmt.Sprintf("%v", it),
				})
			}
		}
	case []map[string]interface{}:
		for _, gMap := range gList {
			if id, found := gMap["groupid"]; found && id != nil && fmt.Sprintf("%v", id) != "" {
				result = append(result, map[string]interface{}{
					"groupid": fmt.Sprintf("%v", id),
				})
			}
		}
	case map[string]interface{}:
		if id, found := gList["groupid"]; found && id != nil && fmt.Sprintf("%v", id) != "" {
			result = append(result, map[string]interface{}{
				"groupid": fmt.Sprintf("%v", id),
			})
		}
	}

	return result
}

func cleanExportGroups(groups interface{}) []map[string]interface{} {
	var result []map[string]interface{}
	switch gList := groups.(type) {
	case []interface{}:
		for _, it := range gList {
			if gMap, ok := it.(map[string]interface{}); ok {
				cleanMap := make(map[string]interface{})
				if id, found := gMap["groupid"]; found && id != nil {
					cleanMap["groupid"] = fmt.Sprintf("%v", id)
				}
				if name, found := gMap["name"]; found && name != nil && fmt.Sprintf("%v", name) != "" {
					cleanMap["name"] = name
				}
				if len(cleanMap) > 0 {
					result = append(result, cleanMap)
				}
			}
		}
	case []map[string]interface{}:
		for _, gMap := range gList {
			cleanMap := make(map[string]interface{})
			if id, found := gMap["groupid"]; found && id != nil {
				cleanMap["groupid"] = fmt.Sprintf("%v", id)
			}
			if name, found := gMap["name"]; found && name != nil && fmt.Sprintf("%v", name) != "" {
				cleanMap["name"] = name
			}
			if len(cleanMap) > 0 {
				result = append(result, cleanMap)
			}
		}
	}
	return result
}

func normalizeTemplateIDs(templates interface{}) []map[string]interface{} {
	var result []map[string]interface{}

	switch tList := templates.(type) {
	case []interface{}:
		for _, it := range tList {
			if tMap, ok := it.(map[string]interface{}); ok {
				if id, found := tMap["templateid"]; found && id != nil && fmt.Sprintf("%v", id) != "" {
					result = append(result, map[string]interface{}{
						"templateid": fmt.Sprintf("%v", id),
					})
				}
			} else if it != nil && fmt.Sprintf("%v", it) != "" {
				result = append(result, map[string]interface{}{
					"templateid": fmt.Sprintf("%v", it),
				})
			}
		}
	case []map[string]interface{}:
		for _, tMap := range tList {
			if id, found := tMap["templateid"]; found && id != nil && fmt.Sprintf("%v", id) != "" {
				result = append(result, map[string]interface{}{
					"templateid": fmt.Sprintf("%v", id),
				})
			}
		}
	}

	return result
}

func cleanExportTemplates(templates interface{}) []map[string]interface{} {
	var result []map[string]interface{}
	switch tList := templates.(type) {
	case []interface{}:
		for _, it := range tList {
			if tMap, ok := it.(map[string]interface{}); ok {
				cleanMap := make(map[string]interface{})
				if id, found := tMap["templateid"]; found && id != nil {
					cleanMap["templateid"] = fmt.Sprintf("%v", id)
				}
				if name, found := tMap["name"]; found && name != nil && fmt.Sprintf("%v", name) != "" {
					cleanMap["name"] = name
				}
				if len(cleanMap) > 0 {
					result = append(result, cleanMap)
				}
			}
		}
	case []map[string]interface{}:
		for _, tMap := range tList {
			cleanMap := make(map[string]interface{})
			if id, found := tMap["templateid"]; found && id != nil {
				cleanMap["templateid"] = fmt.Sprintf("%v", id)
			}
			if name, found := tMap["name"]; found && name != nil && fmt.Sprintf("%v", name) != "" {
				cleanMap["name"] = name
			}
			if len(cleanMap) > 0 {
				result = append(result, cleanMap)
			}
		}
	}
	return result
}

func sanitizeInterfaces(ifaces interface{}) []map[string]interface{} {
	var result []map[string]interface{}

	processInterface := func(iMap map[string]interface{}) map[string]interface{} {
		clean := make(map[string]interface{})
		allowedKeys := []string{
			"interfaceid", "main", "type", "useip", "ip", "dns", "port", "details",
		}
		for _, k := range allowedKeys {
			if v, exists := iMap[k]; exists && v != nil {
				// Don't include empty interfaceid on create
				if k == "interfaceid" && fmt.Sprintf("%v", v) == "" {
					continue
				}
				clean[k] = v
			}
		}
		return clean
	}

	switch iList := ifaces.(type) {
	case []interface{}:
		for _, it := range iList {
			if iMap, ok := it.(map[string]interface{}); ok {
				result = append(result, processInterface(iMap))
			}
		}
	case []map[string]interface{}:
		for _, iMap := range iList {
			result = append(result, processInterface(iMap))
		}
	case map[string]interface{}:
		result = append(result, processInterface(iList))
	}

	return result
}
