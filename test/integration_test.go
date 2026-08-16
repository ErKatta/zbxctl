package test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ErKatta/zbxctl/cmd"
	"github.com/ErKatta/zbxctl/pkg/config"
	"github.com/ErKatta/zbxctl/pkg/zabbix"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestZabbixIntegrationTestcontainers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Request Zabbix Appliance container
	req := testcontainers.ContainerRequest{
		Image:        "zabbix/zabbix-appliance:latest",
		ExposedPorts: []string{"8080/tcp", "80/tcp"},
		WaitingFor: wait.ForHTTP("/api_jsonrpc.php").
			WithMethod(http.MethodPost).
			WithHeaders(map[string]string{"Content-Type": "application/json-rpc"}).
			WithBody(bytes.NewBufferString(`{"jsonrpc":"2.0","method":"apiinfo.version","params":[],"id":1}`)).
			WithResponseMatcher(func(body io.Reader) bool {
				b, _ := io.ReadAll(body)
				return strings.Contains(string(b), `"result"`)
			}).
			WithStartupTimeout(3 * time.Minute),
	}

	zabbixContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	// If Docker environment is unavailable (e.g. CI without docker socket), skip testcontainers test with informative message
	if err != nil {
		t.Skipf("Skipping Testcontainers integration test (Docker daemon not accessible or image pull failed: %v)", err)
		return
	}
	defer func() {
		_ = zabbixContainer.Terminate(context.Background())
	}()

	// Resolve dynamic mapped port and host
	host, err := zabbixContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}

	mappedPort, err := zabbixContainer.MappedPort(ctx, "80/tcp")
	if err != nil {
		mappedPort, err = zabbixContainer.MappedPort(ctx, "8080/tcp")
		if err != nil {
			t.Fatalf("failed to get mapped port: %v", err)
		}
	}

	// Probe endpoint path (/zabbix/api_jsonrpc.php vs /api_jsonrpc.php)
	apiURL := "http://" + host + ":" + mappedPort.Port() + "/zabbix/api_jsonrpc.php"
	clientTest := &http.Client{Timeout: 5 * time.Second}
	if resp, err := clientTest.Get("http://" + host + ":" + mappedPort.Port() + "/api_jsonrpc.php"); err == nil && resp.StatusCode != http.StatusNotFound {
		apiURL = "http://" + host + ":" + mappedPort.Port() + "/api_jsonrpc.php"
		resp.Body.Close()
	}
	t.Logf("Started Zabbix 7 Testcontainer at: %s", apiURL)

	// Build temporary zbxctl config targeting the live testcontainer
	tempDir, err := os.MkdirTemp("", "zbxctl-testcontainers-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfgPath := filepath.Join(tempDir, "config.yaml")

	testCfg := &config.Config{
		ActiveContext: "testcontainers-zabbix",
		Contexts: map[string]config.Context{
			"testcontainers-zabbix": {
				URL:          apiURL,
				Username:     "Admin",
				Password:     "zabbix",
				SafetyLevel:  "readwrite-all",
				OutputFormat: "json",
				Timeout:      30,
			},
		},
	}

	if err := config.SaveConfig(testCfg, cfgPath); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	// 1. Verify Zabbix Client API Version directly against container
	targetCtx := testCfg.Contexts["testcontainers-zabbix"]
	client := zabbix.NewClient(&targetCtx)
	ver, err := client.GetVersion(ctx)
	if err != nil {
		t.Fatalf("failed to get version from Zabbix testcontainer: %v", err)
	}
	t.Logf("Successfully connected to live Zabbix API (Version: %s)", ver)

	// 2. Test zbxctl CLI commands against Testcontainer instance
	t.Run("zbxctl login with Admin/zabbix credentials against Testcontainer", func(t *testing.T) {
		root := cmd.RootCmd
		root.SetArgs([]string{
			"--config", cfgPath,
			"login", apiURL,
			"--username", "Admin",
			"--password", "zabbix",
			"--name", "testcontainer-user-login",
			"--safety-level", "readwrite-all",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("login failed against Testcontainer: %v", err)
		}
	})

	t.Run("zbxctl doctor against Testcontainer", func(t *testing.T) {
		root := cmd.RootCmd
		root.SetArgs([]string{"--config", cfgPath, "doctor"})
		if err := root.Execute(); err != nil {
			t.Fatalf("doctor failed against Testcontainer: %v", err)
		}
	})

	t.Run("zbxctl cluster-info against Testcontainer", func(t *testing.T) {
		root := cmd.RootCmd
		cmd.ResetCommandFlags(root)
		root.SetArgs([]string{"--config", cfgPath, "cluster-info"})
		if err := root.Execute(); err != nil {
			t.Fatalf("cluster-info failed against Testcontainer: %v", err)
		}
	})

	t.Run("zbxctl inventory alias against Testcontainer", func(t *testing.T) {
		root := cmd.RootCmd
		cmd.ResetCommandFlags(root)
		root.SetArgs([]string{"--config", cfgPath, "inventory"})
		if err := root.Execute(); err != nil {
			t.Fatalf("inventory alias failed against Testcontainer: %v", err)
		}
	})

	t.Run("zbxctl get host against Testcontainer", func(t *testing.T) {
		root := cmd.RootCmd
		root.SetArgs([]string{"--config", cfgPath, "get", "host"})
		if err := root.Execute(); err != nil {
			t.Fatalf("get host failed against Testcontainer: %v", err)
		}
	})

	t.Run("zbxctl raw host.get against Testcontainer", func(t *testing.T) {
		root := cmd.RootCmd
		root.SetArgs([]string{"--config", cfgPath, "raw", "host.get", "--params", `{"output":["hostid","host"]}`})
		if err := root.Execute(); err != nil {
			t.Fatalf("raw host.get failed against Testcontainer: %v", err)
		}
	})
}
