package zabbix

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ErKatta/zbxctl/pkg/config"
	"github.com/ErKatta/zbxctl/pkg/prompt"
)

func TestZabbixClientAPIToken(t *testing.T) {
	mockServer := NewMockZabbixServer()
	defer mockServer.Close()

	ctxCfg := &config.Context{
		URL:         mockServer.URL,
		APIToken:    "test-token",
		Timeout:     5,
		SafetyLevel: "readwrite-all",
	}

	client := NewClient(ctxCfg)

	ver, err := client.GetVersion(context.Background())
	if err != nil {
		t.Fatalf("failed to get version: %v", err)
	}
	if ver != "7.0.0" {
		t.Errorf("expected version 7.0.0, got %s", ver)
	}

	res, err := client.Call(context.Background(), "host.get", map[string]interface{}{"output": "extend"})
	if err != nil {
		t.Fatalf("failed to call host.get: %v", err)
	}

	var hosts []map[string]interface{}
	if err := json.Unmarshal(res, &hosts); err != nil {
		t.Fatalf("failed to parse host.get result: %v", err)
	}

	if len(hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d", len(hosts))
	}
}

func TestZabbixClientUserLogin(t *testing.T) {
	mockServer := NewMockZabbixServer()
	defer mockServer.Close()

	ctxCfg := &config.Context{
		URL:         mockServer.URL,
		Username:    "Admin",
		Password:    "zabbix",
		Timeout:     5,
		SafetyLevel: "readwrite-all",
	}

	client := NewClient(ctxCfg)

	res, err := client.Call(context.Background(), "host.get", map[string]interface{}{"output": "extend"})
	if err != nil {
		t.Fatalf("failed to call host.get with user login: %v", err)
	}

	var hosts []map[string]interface{}
	if err := json.Unmarshal(res, &hosts); err != nil {
		t.Fatalf("failed to parse host.get result: %v", err)
	}

	if len(hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d", len(hosts))
	}
}

func TestZabbixClientUserPromptPassword(t *testing.T) {
	mockServer := NewMockZabbixServer()
	defer mockServer.Close()

	prompt.StdinOverride = strings.NewReader("zabbix\n")
	defer func() { prompt.StdinOverride = nil }()

	ctxCfg := &config.Context{
		URL:         mockServer.URL,
		Username:    "Admin",
		Password:    "", // Password empty, should prompt
		Timeout:     5,
		SafetyLevel: "readwrite-all",
	}

	client := NewClient(ctxCfg)

	res, err := client.Call(context.Background(), "host.get", map[string]interface{}{"output": "extend"})
	if err != nil {
		t.Fatalf("failed to call host.get with prompted password: %v", err)
	}

	var hosts []map[string]interface{}
	if err := json.Unmarshal(res, &hosts); err != nil {
		t.Fatalf("failed to parse host.get result: %v", err)
	}

	if len(hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d", len(hosts))
	}
}

func TestZabbixClientHTTPBasicAuth(t *testing.T) {
	mockServer := NewMockZabbixServer()
	defer mockServer.Close()

	ctxCfg := &config.Context{
		URL:          mockServer.URL,
		HTTPUser:     "httpuser",
		HTTPPassword: "httppass",
		APIToken:     "test-token",
		Timeout:      5,
		SafetyLevel:  "readwrite-all",
	}

	client := NewClient(ctxCfg)

	res, err := client.Call(context.Background(), "host.get", map[string]interface{}{"output": "extend"})
	if err != nil {
		t.Fatalf("failed to call host.get with HTTP Basic Auth: %v", err)
	}

	var hosts []map[string]interface{}
	if err := json.Unmarshal(res, &hosts); err != nil {
		t.Fatalf("failed to parse host.get result: %v", err)
	}

	if len(hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d", len(hosts))
	}
}

func TestZabbixClientCustomHeaders(t *testing.T) {
	mockServer := NewMockZabbixServer()
	defer mockServer.Close()

	ctxCfg := &config.Context{
		URL:      mockServer.URL,
		APIToken: "test-token",
		HTTPHeaders: map[string]string{
			"X-Remote-User": "admin@company.com",
		},
		Timeout:     5,
		SafetyLevel: "readwrite-all",
	}

	client := NewClient(ctxCfg)

	res, err := client.Call(context.Background(), "host.get", map[string]interface{}{"output": "extend"})
	if err != nil {
		t.Fatalf("failed to call host.get with custom headers: %v", err)
	}

	var hosts []map[string]interface{}
	if err := json.Unmarshal(res, &hosts); err != nil {
		t.Fatalf("failed to parse host.get result: %v", err)
	}

	if len(hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d", len(hosts))
	}
}
