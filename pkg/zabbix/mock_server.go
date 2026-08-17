package zabbix

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

type MockServer struct {
	Server *httptest.Server
	URL    string
}

func NewMockZabbixServer() *MockServer {
	mux := http.NewServeMux()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json-rpc")

		// Check HTTP Basic Auth header if present
		user, pass, hasBasic := r.BasicAuth()
		if hasBasic && (user != "httpuser" || pass != "httppass") {
			http.Error(w, "Unauthorized HTTP Basic Auth", http.StatusUnauthorized)
			return
		}

		// Check custom header test if required
		if r.Header.Get("X-Test-Auth") == "invalid" {
			http.Error(w, "Forbidden custom header", http.StatusForbidden)
			return
		}

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp := Response{
			JSONRPC: "2.0",
			ID:      req.ID,
		}

		hasCount := false
		if pMap, ok := req.Params.(map[string]interface{}); ok {
			if cnt, exists := pMap["countOutput"]; exists && cnt == true {
				hasCount = true
			}
		}

		switch req.Method {
		case "apiinfo.version":
			resp.Result = json.RawMessage(`"7.0.0"`)

		case "user.login":
			resp.Result = json.RawMessage(`"mock-session-token-12345"`)

		case "host.get":
			if hasCount {
				resp.Result = json.RawMessage(`"2"`)
				break
			}
			host10001 := `{"hostid": "10001", "host": "Zabbix server", "name": "Zabbix server", "status": "0", "inventory_mode": "1", "inventory": {"type": "Server", "vendor": "Zabbix", "model": "Zabbix-Appliance", "os": "Linux", "macaddress_a": "52:54:00:12:34:56"}, "groups": [{"groupid": "1", "name": "Discovered hosts", "flags": "0"}], "parentTemplates": [{"templateid": "40001", "name": "Linux by Zabbix agent", "flags": "0"}], "interfaces": [{"interfaceid": "1", "main": "1", "type": "1", "useip": "1", "ip": "127.0.0.1", "dns": "", "port": "10050", "available": "0"}]}`
			host10002 := `{"hostid": "10002", "host": "web-prod-01", "name": "web-prod-01", "status": "0", "inventory_mode": "0", "inventory": {"type": "Web Server", "vendor": "Dell", "model": "PowerEdge R740", "os": "Ubuntu 24.04", "macaddress_a": "00:1A:2B:3C:4D:5E"}}`

			if pMap, ok := req.Params.(map[string]interface{}); ok {
				if hIDs, exists := pMap["hostids"].([]interface{}); exists && len(hIDs) == 1 {
					idStr := fmt.Sprintf("%v", hIDs[0])
					if idStr == "10001" {
						resp.Result = json.RawMessage("[" + host10001 + "]")
					} else if idStr == "10002" {
						resp.Result = json.RawMessage("[" + host10002 + "]")
					} else {
						resp.Result = json.RawMessage("[]")
					}
				} else if hIDStr, exists := pMap["hostids"].(string); exists {
					if hIDStr == "10001" {
						resp.Result = json.RawMessage("[" + host10001 + "]")
					} else if hIDStr == "10002" {
						resp.Result = json.RawMessage("[" + host10002 + "]")
					} else {
						resp.Result = json.RawMessage("[]")
					}
				} else if filterMap, exists := pMap["filter"].(map[string]interface{}); exists {
					matched := false
					if name, ok := filterMap["name"].(string); ok {
						if name == "Zabbix server" || name == "10001" {
							resp.Result = json.RawMessage("[" + host10001 + "]")
							matched = true
						} else if name == "web-prod-01" || name == "10002" {
							resp.Result = json.RawMessage("[" + host10002 + "]")
							matched = true
						}
					}
					if !matched {
						if hostName, ok := filterMap["host"].(string); ok {
							if hostName == "Zabbix server" {
								resp.Result = json.RawMessage("[" + host10001 + "]")
								matched = true
							} else if hostName == "web-prod-01" {
								resp.Result = json.RawMessage("[" + host10002 + "]")
								matched = true
							}
						}
					}
					if !matched {
						resp.Result = json.RawMessage("[]")
					}
				} else {
					resp.Result = json.RawMessage("[" + host10001 + ", " + host10002 + "]")
				}
			} else {
				resp.Result = json.RawMessage("[" + host10001 + ", " + host10002 + "]")
			}

		case "host.create":
			resp.Result = json.RawMessage(`{"hostids": ["10003"]}`)

		case "host.update":
			resp.Result = json.RawMessage(`{"hostids": ["10001"]}`)

		case "host.delete":
			resp.Result = json.RawMessage(`{"hostids": ["10001"]}`)

		case "problem.get":
			if hasCount {
				resp.Result = json.RawMessage(`"1"`)
				break
			}
			resp.Result = json.RawMessage(`[
				{"eventid": "1", "name": "High CPU utilization", "severity": "4", "objectid": "30001", "clock": "1700000000", "acknowledged": "0"}
			]`)

		case "item.get":
			if hasCount {
				resp.Result = json.RawMessage(`"1"`)
				break
			}
			resp.Result = json.RawMessage(`[
				{"itemid": "20001", "name": "CPU load", "key_": "system.cpu.load", "hostid": "10001"}
			]`)

		case "trigger.get":
			if hasCount {
				resp.Result = json.RawMessage(`"1"`)
				break
			}
			resp.Result = json.RawMessage(`[
				{"triggerid": "30001", "description": "CPU high", "priority": "4", "hosts": [{"hostid": "10002", "host": "web-prod-01", "name": "web-prod-01"}]}
			]`)

		case "template.get":
			if hasCount {
				resp.Result = json.RawMessage(`"1"`)
				break
			}
			resp.Result = json.RawMessage(`[
				{"templateid": "40001", "name": "Linux by Zabbix agent"}
			]`)

		case "hostgroup.get":
			if hasCount {
				resp.Result = json.RawMessage(`"2"`)
				break
			}
			resp.Result = json.RawMessage(`[
				{"groupid": "1", "name": "Discovered hosts"},
				{"groupid": "2", "name": "Linux servers"}
			]`)

		case "proxygroup.get":
			resp.Result = json.RawMessage(`[
				{"proxy_groupid": "1", "name": "Primary Proxy Group"}
			]`)

		case "history.push":
			resp.Result = json.RawMessage(`{"response": "success"}`)

		default:
			if strings.HasSuffix(req.Method, ".get") {
				resp.Result = json.RawMessage(`[]`)
			} else {
				resp.Result = json.RawMessage(`{"result": "ok"}`)
			}
		}

		_ = json.NewEncoder(w).Encode(resp)
	}

	mux.HandleFunc("/zabbix/api_jsonrpc.php", handler)
	mux.HandleFunc("/api_jsonrpc.php", handler)

	srv := httptest.NewServer(mux)
	return &MockServer{
		Server: srv,
		URL:    srv.URL + "/zabbix/api_jsonrpc.php",
	}
}

func (ms *MockServer) Close() {
	ms.Server.Close()
}
