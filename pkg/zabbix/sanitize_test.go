package zabbix

import (
	"testing"
)

func TestSanitizeApplyParamsHost(t *testing.T) {
	rawParams := map[string]interface{}{
		"hostid": "10693",
		"host":   "camera-garage.internal.net",
		"name":   "Garage Camera",
		"flags":  "0",
		"groups": []interface{}{
			map[string]interface{}{
				"flags":   "0",
				"groupid": "5",
				"name":    "Discovered hosts",
				"uuid":    "f2481361f99448eea617b7b1d4765566",
			},
		},
		"parentTemplates": []interface{}{
			map[string]interface{}{
				"templateid": "10564",
				"name":       "ICMP Ping",
				"flags":      "0",
			},
		},
		"interfaces": []interface{}{
			map[string]interface{}{
				"interfaceid":   "52",
				"hostid":        "10693",
				"main":          "1",
				"type":          "1",
				"useip":         "1",
				"ip":            "192.168.1.50",
				"dns":           "camera-garage.internal.net",
				"port":          "10050",
				"available":     "0",
				"error":         "",
				"errors_from":   "0",
				"disable_until": "0",
			},
		},
		"active_available":   "0",
		"assigned_proxyid":   "0",
		"custom_interfaces":  "0",
		"maintenance_from":   "0",
		"maintenance_status": "0",
		"maintenance_type":   "0",
		"maintenanceid":      "0",
		"monitored_by":       "0",
		"proxy_groupid":      "0",
		"proxyid":            "0",
		"templateid":         "0",
		"uuid":               "",
	}

	sanitized := SanitizeApplyParams("host", rawParams)

	// Check hostid and name preserved
	if sanitized["hostid"] != "10693" || sanitized["name"] != "Garage Camera" {
		t.Errorf("expected hostid and name to be preserved, got %v, %v", sanitized["hostid"], sanitized["name"])
	}

	// Check flags and other read-only fields removed
	unwantedKeys := []string{
		"flags", "active_available", "assigned_proxyid", "custom_interfaces",
		"maintenance_from", "maintenance_status", "maintenance_type", "maintenanceid",
		"monitored_by", "proxy_groupid", "proxyid", "templateid", "uuid", "parentTemplates",
	}
	for _, k := range unwantedKeys {
		if _, exists := sanitized[k]; exists {
			t.Errorf("did not expect key %q in sanitized host params", k)
		}
	}

	// Check groups sanitized to groupid only
	groups, ok := sanitized["groups"].([]map[string]interface{})
	if !ok || len(groups) != 1 {
		t.Fatalf("expected sanitized groups to be []map[string]interface{} of len 1, got %T", sanitized["groups"])
	}
	if groups[0]["groupid"] != "5" {
		t.Errorf("expected groupid '5', got %v", groups[0]["groupid"])
	}
	if _, hasFlags := groups[0]["flags"]; hasFlags {
		t.Errorf("did not expect 'flags' in sanitized groups")
	}
	if _, hasName := groups[0]["name"]; hasName {
		t.Errorf("did not expect 'name' in sanitized groups for apply")
	}

	// Check templates sanitized from parentTemplates
	templates, ok := sanitized["templates"].([]map[string]interface{})
	if !ok || len(templates) != 1 {
		t.Fatalf("expected templates to be []map[string]interface{} of len 1, got %T", sanitized["templates"])
	}
	if templates[0]["templateid"] != "10564" {
		t.Errorf("expected templateid '10564', got %v", templates[0]["templateid"])
	}
	if _, hasFlags := templates[0]["flags"]; hasFlags {
		t.Errorf("did not expect 'flags' in templates")
	}

	// Check interfaces sanitized (available, error, hostid removed)
	ifaces, ok := sanitized["interfaces"].([]map[string]interface{})
	if !ok || len(ifaces) != 1 {
		t.Fatalf("expected interfaces to be []map[string]interface{} of len 1, got %T", sanitized["interfaces"])
	}
	if ifaces[0]["interfaceid"] != "52" || ifaces[0]["ip"] != "192.168.1.50" {
		t.Errorf("expected interfaceid 52 and ip 192.168.1.50, got %v, %v", ifaces[0]["interfaceid"], ifaces[0]["ip"])
	}
	if _, hasAvail := ifaces[0]["available"]; hasAvail {
		t.Errorf("did not expect 'available' in sanitized interface")
	}
	if _, hasHostID := ifaces[0]["hostid"]; hasHostID {
		t.Errorf("did not expect 'hostid' in sanitized interface")
	}
}

func TestSanitizeExportSpecHost(t *testing.T) {
	rawSpec := map[string]interface{}{
		"hostid": "10693",
		"host":   "camera-garage.internal.net",
		"name":   "Garage Camera",
		"flags":  "0",
		"groups": []interface{}{
			map[string]interface{}{
				"flags":   "0",
				"groupid": "5",
				"name":    "Discovered hosts",
				"uuid":    "f2481361f99448eea617b7b1d4765566",
			},
		},
		"parentTemplates": []interface{}{
			map[string]interface{}{
				"templateid": "10564",
				"name":       "ICMP Ping",
				"flags":      "0",
			},
		},
		"active_available": "0",
	}

	exported := SanitizeExportSpec("host", rawSpec)

	if _, exists := exported["flags"]; exists {
		t.Errorf("did not expect 'flags' in exported spec")
	}
	if _, exists := exported["active_available"]; exists {
		t.Errorf("did not expect 'active_available' in exported spec")
	}
	if _, exists := exported["parentTemplates"]; exists {
		t.Errorf("did not expect 'parentTemplates' in exported spec (should be 'templates')")
	}

	templates, ok := exported["templates"].([]map[string]interface{})
	if !ok || len(templates) != 1 {
		t.Fatalf("expected templates in exported spec, got %v", exported["templates"])
	}
	if templates[0]["templateid"] != "10564" || templates[0]["name"] != "ICMP Ping" {
		t.Errorf("expected templateid 10564 and name ICMP Ping, got %v", templates[0])
	}
}

func TestSanitizeDiffSpecHost(t *testing.T) {
	localSpec := map[string]interface{}{
		"hostid": "10693",
		"host":   "camera-garage.internal.net",
		"name":   "Garage Camera Renamed",
		"flags":  "0",
		"groups": []interface{}{
			map[string]interface{}{
				"flags":   "0",
				"groupid": "5",
				"name":    "Discovered hosts",
			},
		},
		"parentTemplates": []interface{}{
			map[string]interface{}{
				"templateid": "10564",
				"name":       "ICMP Ping",
			},
		},
		"macros": []interface{}{},
		"tags":   []interface{}{},
	}

	remoteLive := map[string]interface{}{
		"hostid": "10693",
		"host":   "camera-garage.internal.net",
		"name":   "Garage Camera",
		"flags":  "0",
		"groups": []interface{}{
			map[string]interface{}{
				"groupid": "5",
				"name":    "Discovered hosts",
			},
		},
		"parentTemplates": []interface{}{
			map[string]interface{}{
				"templateid": "10564",
			},
		},
		"macros": []interface{}{},
		"tags":   []interface{}{},
	}

	normLocal := SanitizeDiffSpec("host", localSpec)
	normRemote := SanitizeDiffSpec("host", remoteLive)

	if normLocal["name"] == normRemote["name"] {
		t.Errorf("expected name to differ, but both were %v", normLocal["name"])
	}
	if normLocal["hostid"] != normRemote["hostid"] {
		t.Errorf("expected hostid to match")
	}
	if _, exists := normLocal["flags"]; exists {
		t.Errorf("expected flags to be removed from diff spec")
	}
}
