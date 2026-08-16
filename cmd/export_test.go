package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ErKatta/zbxctl/pkg/config"
	"github.com/ErKatta/zbxctl/pkg/zabbix"
	"gopkg.in/yaml.v3"
)

func TestGetDeclarativeWorkflow(t *testing.T) {
	mockServer := zabbix.NewMockZabbixServer()
	defer mockServer.Close()

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &config.Config{
		ActiveContext: "mock-context",
		Contexts: map[string]config.Context{
			"mock-context": {
				URL:          mockServer.URL,
				APIToken:     "mock-token-xyz",
				SafetyLevel:  "readwrite-all",
				OutputFormat: "yaml",
			},
		},
	}
	if err := config.SaveConfig(cfg, cfgPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	t.Run("get template -o yaml outputs declarative manifest directly", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := RootCmd
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--config", cfgPath, "-o", "yaml", "get", "template", "40001"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("get template -o yaml failed: %v", err)
		}

		outStr := buf.String()
		if !strings.Contains(outStr, "kind: template") {
			t.Errorf("expected 'kind: template' in manifest output, got:\n%s", outStr)
		}
		if !strings.Contains(outStr, "spec:") {
			t.Errorf("expected 'spec:' in manifest output, got:\n%s", outStr)
		}
		if !strings.Contains(outStr, "Linux by Zabbix agent") {
			t.Errorf("expected template name in manifest output, got:\n%s", outStr)
		}

		// Verify that this YAML is a valid ManifestItem
		var item ManifestItem
		if err := yaml.Unmarshal(buf.Bytes(), &item); err != nil {
			t.Fatalf("failed to unmarshal exported manifest: %v", err)
		}
		if item.Kind != "template" {
			t.Errorf("expected unmarshaled kind 'template', got %q", item.Kind)
		}
		if item.Spec["templateid"] != "40001" {
			t.Errorf("expected templateid '40001', got %v", item.Spec["templateid"])
		}

		// Test diff using exported manifest with auto-detected ID
		manifestFile := filepath.Join(tmpDir, "template-exported.yaml")
		if err := os.WriteFile(manifestFile, buf.Bytes(), 0644); err != nil {
			t.Fatalf("failed to write manifest file: %v", err)
		}

		var diffBuf bytes.Buffer
		diffCmdRun := RootCmd
		diffCmdRun.SetOut(&diffBuf)
		diffCmdRun.SetArgs([]string{"--config", cfgPath, "diff", "-f", manifestFile})
		if err := diffCmdRun.Execute(); err != nil {
			t.Fatalf("diff with auto-detected ID failed: %v", err)
		}

		// Test apply of the exported manifest
		var applyBuf bytes.Buffer
		applyCmdRun := RootCmd
		applyCmdRun.SetOut(&applyBuf)
		applyCmdRun.SetArgs([]string{"--config", cfgPath, "apply", "-f", manifestFile})
		if err := applyCmdRun.Execute(); err != nil {
			t.Fatalf("apply of exported manifest failed: %v", err)
		}
	})

	t.Run("get host -o json outputs declarative manifest directly", func(t *testing.T) {
		ResetCommandFlags(RootCmd)
		var buf bytes.Buffer
		cmd := RootCmd
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--config", cfgPath, "-o", "json", "get", "host", "10001"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("get host -o json failed: %v", err)
		}

		outStr := buf.String()
		if !strings.Contains(outStr, `"kind": "host"`) {
			t.Errorf("expected '\"kind\": \"host\"' in json output, got:\n%s", outStr)
		}
		if !strings.Contains(outStr, `"spec"`) {
			t.Errorf("expected '\"spec\"' in json output, got:\n%s", outStr)
		}
	})

	t.Run("get host list -o json outputs clean array without manifest wrapper", func(t *testing.T) {
		ResetCommandFlags(RootCmd)
		var buf bytes.Buffer
		cmd := RootCmd
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--config", cfgPath, "-o", "json", "get", "host"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("get host list -o json failed: %v", err)
		}

		outStr := buf.String()
		if strings.Contains(outStr, `"kind": "host"`) {
			t.Errorf("did not expect 'kind: host' in list output, got:\n%s", outStr)
		}
		if !strings.Contains(outStr, `"hostid": "10001"`) {
			t.Errorf("expected hostid 10001 in json output, got:\n%s", outStr)
		}
	})

	t.Run("get host list --export -o json outputs declarative manifests", func(t *testing.T) {
		ResetCommandFlags(RootCmd)
		var buf bytes.Buffer
		cmd := RootCmd
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--config", cfgPath, "-o", "json", "get", "host", "--export"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("get host list --export -o json failed: %v", err)
		}

		outStr := buf.String()
		if !strings.Contains(outStr, `"kind": "host"`) {
			t.Errorf("expected 'kind: host' in exported list output, got:\n%s", outStr)
		}
	})

	t.Run("get template --fields=templateid,name -o json outputs projected fields", func(t *testing.T) {
		ResetCommandFlags(RootCmd)
		var buf bytes.Buffer
		cmd := RootCmd
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--config", cfgPath, "-o", "json", "get", "template", "--fields=templateid,name"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("get template --fields failed: %v", err)
		}

		outStr := buf.String()
		if !strings.Contains(outStr, `"templateid": "40001"`) || !strings.Contains(outStr, `"name": "Linux by Zabbix agent"`) {
			t.Errorf("expected templateid and name in output, got:\n%s", outStr)
		}
	})

	t.Run("apply host manifest with complex groups, parentTemplates, and flags", func(t *testing.T) {
		ResetCommandFlags(RootCmd)
		hostYaml := `
kind: host
spec:
    hostid: "10693"
    host: camera-garage.internal.net
    name: Garage Camera Renamed
    flags: "0"
    active_available: "0"
    assigned_proxyid: "0"
    custom_interfaces: "0"
    description: ""
    groups:
        - flags: "0"
          groupid: "5"
          name: Discovered hosts
          uuid: f2481361f99448eea617b7b1d4765566
    parentTemplates:
        - flags: "0"
          host: ICMP Ping
          templateid: "10564"
    interfaces:
        - available: "0"
          details: []
          disable_until: "0"
          dns: camera-garage.internal.net
          error: ""
          errors_from: "0"
          hostid: "10693"
          interfaceid: "52"
          ip: 192.168.1.50
          main: "1"
          port: "10050"
          type: "1"
          useip: "1"
`
		testManifestPath := filepath.Join(tmpDir, "complex-host.yaml")
		if err := os.WriteFile(testManifestPath, []byte(hostYaml), 0644); err != nil {
			t.Fatalf("failed to write test manifest: %v", err)
		}

		var buf bytes.Buffer
		cmd := RootCmd
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--config", cfgPath, "apply", "-f", testManifestPath})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("apply of complex host manifest failed: %v", err)
		}
	})

	t.Run("diff on host manifest ignores noisy fields and formats differences cleanly", func(t *testing.T) {
		ResetCommandFlags(RootCmd)
		hostYaml := `
kind: host
spec:
    hostid: "10001"
    host: Zabbix server
    name: Zabbix server Renamed
    flags: "0"
    active_available: "0"
    groups:
        - flags: "0"
          groupid: "1"
          name: Discovered hosts
    parentTemplates:
        - flags: "0"
          templateid: "40001"
          name: Linux by Zabbix agent
    interfaces:
        - available: "0"
          interfaceid: "1"
          ip: 127.0.0.1
          dns: ""
          main: "1"
          port: "10050"
          type: "1"
          useip: "1"
`
		testManifestPath := filepath.Join(tmpDir, "diff-host.yaml")
		if err := os.WriteFile(testManifestPath, []byte(hostYaml), 0644); err != nil {
			t.Fatalf("failed to write test manifest: %v", err)
		}

		var buf bytes.Buffer
		cmd := RootCmd
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"--config", cfgPath, "diff", "-f", testManifestPath})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("diff of complex host manifest failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "name: Zabbix server") || !strings.Contains(out, "name: Zabbix server Renamed") {
			t.Errorf("expected diff on name, got:\n%s", out)
		}
		if strings.Contains(out, "flags") || strings.Contains(out, "active_available") || strings.Contains(out, "missing on remote") {
			t.Errorf("did not expect noisy flags or false-positive 'missing on remote' in diff:\n%s", out)
		}
	})
}
