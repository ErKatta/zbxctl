package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadAndSave(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "zbxctl-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")

	// Test load non-existing config -> returns default config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("expected no error loading non-existent config, got: %v", err)
	}
	if cfg.ActiveContext != "default" {
		t.Errorf("expected default active_context, got %q", cfg.ActiveContext)
	}

	// Modify config and save
	cfg.ActiveContext = "prod"
	cfg.Contexts["prod"] = Context{
		URL:         "https://zabbix.prod.company.com/api_jsonrpc.php",
		APIToken:    "test-token",
		SafetyLevel: "readwrite-mine",
	}

	if err := SaveConfig(cfg, configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Reload config and verify
	reloaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	if reloaded.ActiveContext != "prod" {
		t.Errorf("expected active context 'prod', got %q", reloaded.ActiveContext)
	}

	activeCtx, name, err := reloaded.GetActiveContext()
	if err != nil {
		t.Fatalf("failed to get active context: %v", err)
	}
	if name != "prod" {
		t.Errorf("expected context name 'prod', got %q", name)
	}
	if activeCtx.APIToken != "test-token" {
		t.Errorf("expected api token 'test-token', got %q", activeCtx.APIToken)
	}
	if activeCtx.SafetyLevel != "readwrite-mine" {
		t.Errorf("expected safety level 'readwrite-mine', got %q", activeCtx.SafetyLevel)
	}
}

func TestConfigRedacted(t *testing.T) {
	ctx := Context{
		URL:          "https://zabbix.example.com",
		APIToken:     "secret-token",
		Password:     "secret-pass",
		HTTPPassword: "proxy-pass",
		HTTPHeaders: map[string]string{
			"X-Auth": "secret-header-val",
		},
	}

	redacted := ctx.Redacted()

	if redacted.APIToken != "[REDACTED]" {
		t.Errorf("expected APIToken [REDACTED], got %q", redacted.APIToken)
	}
	if redacted.Password != "[REDACTED]" {
		t.Errorf("expected Password [REDACTED], got %q", redacted.Password)
	}
	if redacted.HTTPPassword != "[REDACTED]" {
		t.Errorf("expected HTTPPassword [REDACTED], got %q", redacted.HTTPPassword)
	}
	if redacted.HTTPHeaders["X-Auth"] != "[REDACTED]" {
		t.Errorf("expected HTTPHeader [REDACTED], got %q", redacted.HTTPHeaders["X-Auth"])
	}

	// Original struct should remain untouched
	if ctx.APIToken != "secret-token" || ctx.Password != "secret-pass" {
		t.Errorf("original Context modified during redaction")
	}
}

