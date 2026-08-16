package config

import "testing"

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "http://localhost:8080",
			expected: "http://localhost:8080/api_jsonrpc.php",
		},
		{
			input:    "http://localhost:8080/",
			expected: "http://localhost:8080/api_jsonrpc.php",
		},
		{
			input:    "https://zabbix.example.com:8443",
			expected: "https://zabbix.example.com:8443/api_jsonrpc.php",
		},
		{
			input:    "localhost:8080",
			expected: "http://localhost:8080/api_jsonrpc.php",
		},
		{
			input:    "zabbix.internal",
			expected: "http://zabbix.internal/api_jsonrpc.php",
		},
		{
			input:    "http://localhost:8080/zabbix",
			expected: "http://localhost:8080/zabbix/api_jsonrpc.php",
		},
		{
			input:    "http://localhost:8080/zabbix/",
			expected: "http://localhost:8080/zabbix/api_jsonrpc.php",
		},
		{
			input:    "http://localhost:8080/zabbix/api_jsonrpc.php",
			expected: "http://localhost:8080/zabbix/api_jsonrpc.php",
		},
		{
			input:    "https://zabbix.company.com/api_jsonrpc.php",
			expected: "https://zabbix.company.com/api_jsonrpc.php",
		},
		{
			input:    "",
			expected: "",
		},
	}

	for _, tc := range tests {
		got := NormalizeURL(tc.input)
		if got != tc.expected {
			t.Errorf("NormalizeURL(%q) = %q; expected %q", tc.input, got, tc.expected)
		}
	}
}
