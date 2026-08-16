package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ErKatta/zbxctl/pkg/config"
	"github.com/ErKatta/zbxctl/pkg/prompt"
	"github.com/ErKatta/zbxctl/pkg/zabbix"
)

func TestCLICommands(t *testing.T) {
	mockServer := zabbix.NewMockZabbixServer()
	defer mockServer.Close()

	tempDir, err := os.MkdirTemp("", "zbxctl-cmd-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfgPath := filepath.Join(tempDir, "config.yaml")

	testCfg := &config.Config{
		ActiveContext: "mock-context",
		Contexts: map[string]config.Context{
			"mock-context": {
				URL:          mockServer.URL,
				APIToken:     "test-token",
				SafetyLevel:  "readonly",
				OutputFormat: "json",
				Timeout:      5,
			},
			"rw-context": {
				URL:          mockServer.URL,
				APIToken:     "test-token",
				SafetyLevel:  "readwrite-all",
				OutputFormat: "json",
				Timeout:      5,
			},
		},
	}

	if err := config.SaveConfig(testCfg, cfgPath); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	cfgFile = cfgPath
	defer func() {
		cfgFile = ""
		cfg = nil
		activeCtx = nil
		actCtxName = ""
	}()

	// 1. Test zbxctl commands --brief
	t.Run("commands --brief", func(t *testing.T) {
		cmd := RootCmd
		ResetCommandFlags(cmd)
		cmd.SetArgs([]string{"--config", cfgPath, "commands", "--brief"})

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		if err := cmd.Execute(); err != nil {
			t.Fatalf("commands --brief failed: %v", err)
		}
	})

	// 1b. Test zbxctl version
	t.Run("version", func(t *testing.T) {
		cmd := RootCmd
		ResetCommandFlags(cmd)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "version"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("version failed: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "zbxctl version") {
			t.Errorf("expected 'zbxctl version' in output, got: %s", out)
		}
	})

	// 1c. Test zbxctl version --short
	t.Run("version --short", func(t *testing.T) {
		cmd := RootCmd
		ResetCommandFlags(cmd)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "version", "--short"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("version --short failed: %v", err)
		}
		out := strings.TrimSpace(buf.String())
		if out == "" {
			t.Errorf("expected non-empty short version string, got empty")
		}
	})

	// 1d. Test zbxctl version -o json
	t.Run("version -o json", func(t *testing.T) {
		cmd := RootCmd
		ResetCommandFlags(cmd)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "version", "-o", "json"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("version -o json failed: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, `"version":`) || !strings.Contains(out, `"go_version":`) {
			t.Errorf("expected JSON version structure, got: %s", out)
		}
	})

	// 2. Test zbxctl doctor
	t.Run("doctor", func(t *testing.T) {
		cmd := RootCmd
		cmd.SetArgs([]string{"--config", cfgPath, "doctor"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("doctor failed: %v", err)
		}
	})

	// 3. Test zbxctl cluster-info
	t.Run("cluster-info", func(t *testing.T) {
		cmd := RootCmd
		ResetCommandFlags(cmd)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "cluster-info", "-o", "json"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("cluster-info failed: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, `"total_hosts": 2`) {
			t.Errorf("expected total_hosts: 2 in cluster-info output, got:\n%s", out)
		}
	})

	// 3b. Test zbxctl cluster-info aliases
	t.Run("cluster-info aliases", func(t *testing.T) {
		for _, alias := range []string{"info", "overview", "sizing", "stats", "clusterinfo"} {
			cmd := RootCmd
			ResetCommandFlags(cmd)
			cmd.SetArgs([]string{"--config", cfgPath, alias})

			if err := cmd.Execute(); err != nil {
				t.Fatalf("cluster-info alias %q failed: %v", alias, err)
			}
		}
	})

	// 3c. Test zbxctl get inventory (table, json, yaml, toon, and field projection)
	t.Run("get inventory table schema and data", func(t *testing.T) {
		cmd := RootCmd
		ResetCommandFlags(cmd)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "-o", "table", "get", "inventory"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("get inventory table failed: %v", err)
		}

		out := buf.String()
		expectedCols := []string{"HOSTID", "HOST", "NAME", "MODE", "TYPE", "VENDOR", "MODEL", "MAC", "OS"}
		for _, col := range expectedCols {
			if !strings.Contains(out, col) {
				t.Errorf("expected column %q in get inventory output:\n%s", col, out)
			}
		}
		if !strings.Contains(out, "Dell") || !strings.Contains(out, "PowerEdge R740") {
			t.Errorf("expected inventory data (Dell PowerEdge R740) in output:\n%s", out)
		}
	})

	t.Run("get inventory aliases inv and host-inventory", func(t *testing.T) {
		for _, alias := range []string{"inv", "host-inventory"} {
			cmd := RootCmd
			ResetCommandFlags(cmd)
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetArgs([]string{"--config", cfgPath, "-o", "json", "get", alias})

			if err := cmd.Execute(); err != nil {
				t.Fatalf("get %s failed: %v", alias, err)
			}
			out := buf.String()
			if !strings.Contains(out, `"vendor": "Dell"`) {
				t.Errorf("expected vendor Dell in %s output:\n%s", alias, out)
			}
		}
	})

	t.Run("get inventory with custom fields projection", func(t *testing.T) {
		cmd := RootCmd
		ResetCommandFlags(cmd)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "-o", "table", "get", "inventory", "-f", "hostid,name,vendor,model"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("get inventory with custom fields failed: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "HOSTID") || !strings.Contains(out, "VENDOR") || !strings.Contains(out, "MODEL") {
			t.Errorf("expected HOSTID, VENDOR, MODEL columns, got:\n%s", out)
		}
	})

	t.Run("get inventory single host by id", func(t *testing.T) {
		cmd := RootCmd
		ResetCommandFlags(cmd)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "-o", "json", "get", "inventory", "10002"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("get inventory 10002 failed: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, `"vendor": "Dell"`) {
			t.Errorf("expected vendor Dell in get inventory 10002 output, got:\n%s", out)
		}
	})

	t.Run("describe host includes inventory block", func(t *testing.T) {
		cmd := RootCmd
		ResetCommandFlags(cmd)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "-o", "json", "describe", "host", "10002"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("describe host 10002 failed: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, `"inventory"`) || !strings.Contains(out, `"vendor": "Dell"`) {
			t.Errorf("expected inventory block in describe host output, got:\n%s", out)
		}
	})

	t.Run("describe inventory 10002", func(t *testing.T) {
		cmd := RootCmd
		ResetCommandFlags(cmd)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "-o", "json", "describe", "inventory", "10002"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("describe inventory 10002 failed: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, `"vendor": "Dell"`) {
			t.Errorf("expected vendor Dell in describe inventory output, got:\n%s", out)
		}
	})

	t.Run("query inventory with search", func(t *testing.T) {
		cmd := RootCmd
		ResetCommandFlags(cmd)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "-o", "json", "query", "inventory", "--search=Dell"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("query inventory failed: %v", err)
		}
	})

	t.Run("apply kind inventory declarative manifest", func(t *testing.T) {
		cmd := RootCmd
		ResetCommandFlags(cmd)

		invManifest := `kind: inventory
spec:
  hostid: "10002"
  inventory_mode: 0
  vendor: "Dell"
  model: "PowerEdge R740"
  macaddress_a: "00:1A:2B:3C:4D:5E"
`
		manifestFile := filepath.Join(tempDir, "inv_manifest.yaml")
		if err := os.WriteFile(manifestFile, []byte(invManifest), 0644); err != nil {
			t.Fatalf("failed to write inv_manifest: %v", err)
		}

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "--context=rw-context", "apply", "-f", manifestFile, "-o", "json"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("apply kind inventory failed: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, `"action": "updated"`) {
			t.Errorf("expected action updated in apply output, got:\n%s", out)
		}
	})

	t.Run("apply kind inventory via stdin", func(t *testing.T) {
		cmd := RootCmd
		ResetCommandFlags(cmd)

		invManifest := `kind: inventory
spec:
  host: "web-prod-01"
  mode: manual
  vendor: "Dell"
  model: "PowerEdge R740"
`
		oldStdin := os.Stdin
		r, w, _ := os.Pipe()
		os.Stdin = r
		w.WriteString(invManifest)
		w.Close()
		defer func() { os.Stdin = oldStdin }()

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "--context=rw-context", "apply", "-f", "-", "-o", "json"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("apply kind inventory via stdin failed: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, `"action": "updated"`) {
			t.Errorf("expected action updated in apply stdin output, got:\n%s", out)
		}
	})

	// 4. Test zbxctl get host
	t.Run("get host", func(t *testing.T) {
		cmd := RootCmd
		ResetCommandFlags(cmd)
		cmd.SetArgs([]string{"--config", cfgPath, "get", "host"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("get host failed: %v", err)
		}
	})

	// 4a. Test zbxctl get host --count
	t.Run("get host --count", func(t *testing.T) {
		cmd := RootCmd
		ResetCommandFlags(cmd)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "-o", "table", "get", "host", "--count"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("get host --count failed: %v", err)
		}
		out := strings.TrimSpace(buf.String())
		if out != "2" {
			t.Errorf("expected count output '2', got: %q", out)
		}
	})

	// 4a2. Test zbxctl get host --count -o json
	t.Run("get host --count -o json", func(t *testing.T) {
		cmd := RootCmd
		ResetCommandFlags(cmd)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "-o", "json", "get", "host", "--count"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("get host --count -o json failed: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, `"count": 2`) {
			t.Errorf("expected count JSON output with count: 2, got: %q", out)
		}
	})

	// 4b. Test zbxctl get problems (table schema, host enrichment, deterministic columns)
	t.Run("get problems table schema and host enrichment", func(t *testing.T) {
		ResetCommandFlags(RootCmd)
		cmd := RootCmd
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "-o", "table", "get", "problems"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("get problems failed: %v", err)
		}

		out := buf.String()
		expectedCols := []string{"EVENTID", "HOST", "PROBLEM", "SEVERITY", "STATUS", "AGE", "ACK"}
		for _, col := range expectedCols {
			if !strings.Contains(out, col) {
				t.Errorf("expected column %q in get problems output:\n%s", col, out)
			}
		}

		if !strings.Contains(out, "web-prod-01") {
			t.Errorf("expected enriched host 'web-prod-01' in get problems output:\n%s", out)
		}
		if !strings.Contains(out, "High CPU utilization") {
			t.Errorf("expected problem name in get problems output:\n%s", out)
		}
	})

	// 4c. Test zbxctl get problems with filter (verifying columns do not change)
	t.Run("get problems with filter preserves standard columns", func(t *testing.T) {
		ResetCommandFlags(RootCmd)
		cmd := RootCmd
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "-o", "table", "get", "problems", "--filter", `{"severity": 4}`})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("get problems with filter failed: %v", err)
		}

		out := buf.String()
		expectedCols := []string{"EVENTID", "HOST", "PROBLEM", "SEVERITY", "STATUS", "AGE", "ACK"}
		for _, col := range expectedCols {
			if !strings.Contains(out, col) {
				t.Errorf("expected column %q in get problems with filter output:\n%s", col, out)
			}
		}
	})

	// 4d. Test zbxctl get problems with custom --fields
	t.Run("get problems with custom fields", func(t *testing.T) {
		ResetCommandFlags(RootCmd)
		cmd := RootCmd
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "-o", "table", "get", "problems", "--fields", "eventid,name,clock"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("get problems with fields failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "EVENTID") || !strings.Contains(out, "NAME") || !strings.Contains(out, "CLOCK") {
			t.Errorf("expected EVENTID, NAME, CLOCK in output:\n%s", out)
		}
		if strings.Contains(out, "SEVERITY") {
			t.Errorf("did not expect SEVERITY in custom fields output:\n%s", out)
		}
	})

	// 4e. Test zbxctl get problems with non-existent field returns error
	t.Run("get problems with non-existent field returns error", func(t *testing.T) {
		ResetCommandFlags(RootCmd)
		cmd := RootCmd
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "-o", "table", "get", "problems", "--fields", "eventid,nonexistent_column"})

		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error for non-existent field, got nil")
		}
		if !strings.Contains(err.Error(), `field "nonexistent_column" does not exist`) {
			t.Errorf("expected field error message, got: %v", err)
		}
	})

	// 4f. Test zbxctl get template clean table schema (without DESCRIPTION bloat)
	t.Run("get template table schema is compact", func(t *testing.T) {
		ResetCommandFlags(RootCmd)
		cmd := RootCmd
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "-o", "table", "get", "template"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("get template failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "TEMPLATEID") || !strings.Contains(out, "NAME") {
			t.Errorf("expected TEMPLATEID and NAME in template table output:\n%s", out)
		}
		if strings.Contains(out, "DESCRIPTION") {
			t.Errorf("did not expect DESCRIPTION column in default template table view:\n%s", out)
		}
	})

	// 4g. Test zbxctl get template with --sort and --sort-order
	t.Run("get template with sort flag", func(t *testing.T) {
		ResetCommandFlags(RootCmd)
		cmd := RootCmd
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "get", "template", "--sort=name:desc"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("get template with sort failed: %v", err)
		}
	})

	// 4h. Test zbxctl get with search across default searchable fields
	t.Run("get template with general search", func(t *testing.T) {
		ResetCommandFlags(RootCmd)
		cmd := RootCmd
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "get", "template", "--search=Linux"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("get template with search failed: %v", err)
		}
	})

	// 4i. Test zbxctl get template with --search and --search-fields
	t.Run("get template with search and custom search-fields", func(t *testing.T) {
		ResetCommandFlags(RootCmd)
		cmd := RootCmd
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "get", "template", "--search=Linux", "--search-fields=name,host"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("get template with search-fields failed: %v", err)
		}
	})

	// 4j. Test zbxctl query command with search, sort, and fields
	t.Run("query host with search, sort, and fields", func(t *testing.T) {
		ResetCommandFlags(RootCmd)
		cmd := RootCmd
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--config", cfgPath, "query", "host", "--search=prod", "--sort=name", "-f=hostid,name"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("query host failed: %v", err)
		}
	})

	// 5. Test zbxctl raw proxygroup.get
	t.Run("raw proxygroup.get", func(t *testing.T) {
		cmd := RootCmd
		cmd.SetArgs([]string{"--config", cfgPath, "raw", "proxygroup.get"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("raw proxygroup.get failed: %v", err)
		}
	})

	// 6. Test zbxctl login
	t.Run("login via CLI", func(t *testing.T) {
		cmd := RootCmd
		cmd.SetArgs([]string{"--config", cfgPath, "login", mockServer.URL, "--token=test-token", "--name=new-login-ctx", "--safety-level=readwrite-mine"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("login failed: %v", err)
		}
	})

	// 6b. Test zbxctl login with user set but no password (should prompt for password)
	t.Run("login with username without password prompts for password", func(t *testing.T) {
		loginTokenFlag = ""
		loginUserFlag = ""
		loginPassFlag = ""
		loginNameFlag = "default"

		prompt.StdinOverride = strings.NewReader("prompted_pass\n")
		defer func() { prompt.StdinOverride = nil }()

		cmd := RootCmd
		cmd.SetArgs([]string{"--config", cfgPath, "login", mockServer.URL, "-u", "admin_user", "--name=prompt-ctx"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("login with username without password failed: %v", err)
		}

		loaded, err := config.LoadConfig(cfgPath)
		if err != nil {
			t.Fatalf("failed to load saved config: %v", err)
		}
		pCtx, ok := loaded.Contexts["prompt-ctx"]
		if !ok {
			t.Fatalf("context prompt-ctx not saved")
		}
		if pCtx.Username != "admin_user" || pCtx.Password != "prompted_pass" {
			t.Errorf("expected username admin_user and password prompted_pass, got user %q, pass %q", pCtx.Username, pCtx.Password)
		}
	})

	// 6c. Test zbxctl login with host:port without full API path
	t.Run("login with protocol host port normalizes URL", func(t *testing.T) {
		loginTokenFlag = ""
		loginUserFlag = ""
		loginPassFlag = ""
		loginNameFlag = "default"

		hostPortURL := strings.TrimSuffix(mockServer.URL, "/zabbix/api_jsonrpc.php")
		cmd := RootCmd
		cmd.SetArgs([]string{"--config", cfgPath, "login", hostPortURL, "--token=test-token", "--name=short-url-ctx"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("login with host:port failed: %v", err)
		}

		loaded, err := config.LoadConfig(cfgPath)
		if err != nil {
			t.Fatalf("failed to load saved config: %v", err)
		}
		sCtx, ok := loaded.Contexts["short-url-ctx"]
		if !ok {
			t.Fatalf("context short-url-ctx not saved")
		}
		if !strings.HasSuffix(sCtx.URL, "api_jsonrpc.php") {
			t.Errorf("expected URL ending with api_jsonrpc.php, got %q", sCtx.URL)
		}
	})

	// 6d. Test zbxctl login with --context flag creating and setting active context
	t.Run("login with --context sets active context", func(t *testing.T) {
		loginTokenFlag = ""
		loginUserFlag = ""
		loginPassFlag = ""
		loginNameFlag = "default"
		contextFlag = ""

		cmd := RootCmd
		cmd.SetArgs([]string{"--config", cfgPath, "login", mockServer.URL, "--token=test-token", "--context=local-zabbix"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("login with --context failed: %v", err)
		}

		loaded, err := config.LoadConfig(cfgPath)
		if err != nil {
			t.Fatalf("failed to load saved config: %v", err)
		}
		cCtx, ok := loaded.Contexts["local-zabbix"]
		if !ok {
			t.Fatalf("context local-zabbix not saved")
		}
		if loaded.ActiveContext != "local-zabbix" {
			t.Errorf("expected active context local-zabbix, got %q", loaded.ActiveContext)
		}
		if cCtx.APIToken != "test-token" {
			t.Errorf("expected APIToken test-token, got %q", cCtx.APIToken)
		}
	})

	// 7. Test zbxctl context list and use
	t.Run("context list and switch", func(t *testing.T) {
		cmd := RootCmd
		cmd.SetArgs([]string{"--config", cfgPath, "context", "list"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("context list failed: %v", err)
		}

		cmd.SetArgs([]string{"--config", cfgPath, "context", "use", "rw-context"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("context use failed: %v", err)
		}

		cmd.SetArgs([]string{"--config", cfgPath, "use-context", "mock-context"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("use-context failed: %v", err)
		}
	})

	// 8. Test zbxctl config current-config
	t.Run("config current-config", func(t *testing.T) {
		cmd := RootCmd
		cmd.SetArgs([]string{"--config", cfgPath, "config", "current-config"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("config current-config failed: %v", err)
		}
	})

	// 8. Test safety violation on delete host under readonly context
	t.Run("safety violation on delete under readonly", func(t *testing.T) {
		cmd := RootCmd
		cmd.SetArgs([]string{"--config", cfgPath, "--context=mock-context", "delete", "host", "10001"})

		err := cmd.Execute()
		if err == nil {
			t.Fatalf("expected safety violation error on delete under readonly, got nil")
		}
		if !strings.Contains(err.Error(), "blocked by safety-level 'readonly'") {
			t.Errorf("expected safety violation message, got %v", err)
		}
	})

	// 9. Test zbxctl skill commands
	t.Run("skill commands", func(t *testing.T) {
		cmd := RootCmd
		cmd.SetArgs([]string{"skill", "list"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("skill list failed: %v", err)
		}

		cmd.SetArgs([]string{"skill", "show", "zabbix-automation"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("skill show failed: %v", err)
		}

		cmd.SetArgs([]string{"skill", "install", "zabbix-automation", "--workspace"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("skill install failed: %v", err)
		}
	})
}
