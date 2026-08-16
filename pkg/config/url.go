package config

import (
	"net/url"
	"strings"
)

// NormalizeURL takes a URL or protocol/host/port string and ensures a valid Zabbix JSON-RPC API endpoint URL.
// Examples:
// - "http://localhost:8080" -> "http://localhost:8080/api_jsonrpc.php"
// - "https://zabbix.example.com" -> "https://zabbix.example.com/api_jsonrpc.php"
// - "localhost:8080" -> "http://localhost:8080/api_jsonrpc.php"
// - "http://localhost:8080/zabbix" -> "http://localhost:8080/zabbix/api_jsonrpc.php"
// - "http://localhost:8080/zabbix/api_jsonrpc.php" -> "http://localhost:8080/zabbix/api_jsonrpc.php"
func NormalizeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	// Add scheme if missing
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	path := parsed.Path
	if path == "" || path == "/" {
		parsed.Path = "/api_jsonrpc.php"
	} else if !strings.HasSuffix(path, "api_jsonrpc.php") {
		parsed.Path = strings.TrimSuffix(path, "/") + "/api_jsonrpc.php"
	}

	return parsed.String()
}
