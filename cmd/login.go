package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ErKatta/zbxctl/pkg/config"
	"github.com/ErKatta/zbxctl/pkg/prompt"
	"github.com/ErKatta/zbxctl/pkg/zabbix"
	"github.com/spf13/cobra"
)

var (
	loginTokenFlag        string
	loginUserFlag         string
	loginPassFlag         string
	loginNameFlag         string
	loginSafetyFlag       string
	loginHTTPUserFlag     string
	loginHTTPPassFlag     string
	loginInsecureSkipFlag bool
)

var loginCmd = &cobra.Command{
	Use:   "login <zabbix_url> [--token=TOKEN | --username=USER [--password=PASS]]",
	Short: "Authenticate against a Zabbix 7 instance and save active context to ~/.zbxctl/config.yaml",
	Long: `login authenticates against the specified Zabbix API URL using an API token or username/password.
Upon successful authentication test, it creates or updates a context in ~/.zbxctl/config.yaml and sets it as active.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rawURL := args[0]
		url := config.NormalizeURL(rawURL)

		if loginUserFlag != "" && loginPassFlag == "" && loginTokenFlag == "" {
			pass, err := prompt.PromptPassword(fmt.Sprintf("Enter password for Zabbix user %q: ", loginUserFlag))
			if err != nil {
				return err
			}
			loginPassFlag = pass
		}

		if loginHTTPUserFlag != "" && loginHTTPPassFlag == "" {
			pass, err := prompt.PromptPassword(fmt.Sprintf("Enter HTTP Basic Auth password for %q: ", loginHTTPUserFlag))
			if err != nil {
				return err
			}
			loginHTTPPassFlag = pass
		}

		ctxName := "default"
		if cmd.Flags().Changed("name") && loginNameFlag != "" && loginNameFlag != "default" {
			ctxName = loginNameFlag
		} else if contextFlag != "" {
			ctxName = contextFlag
		} else if loginNameFlag != "" {
			ctxName = loginNameFlag
		}

		if loginSafetyFlag == "" {
			loginSafetyFlag = "readonly"
		}

		newCtx := config.Context{
			URL:                url,
			APIToken:           loginTokenFlag,
			Username:           loginUserFlag,
			Password:           loginPassFlag,
			HTTPUser:           loginHTTPUserFlag,
			HTTPPassword:       loginHTTPPassFlag,
			SafetyLevel:        loginSafetyFlag,
			OutputFormat:       "auto",
			Timeout:            30,
			InsecureSkipVerify: loginInsecureSkipFlag,
		}

		// Test connection before saving
		client := zabbix.NewClient(&newCtx)
		timeoutCtx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
		defer cancel()

		ver, err := client.GetVersion(timeoutCtx)
		if err != nil && !strings.Contains(rawURL, "api_jsonrpc.php") {
			if altURL := probeAlternativeURL(rawURL, url); altURL != "" {
				newCtx.URL = altURL
				client = zabbix.NewClient(&newCtx)
				ver, err = client.GetVersion(timeoutCtx)
			}
		}
		if err != nil {
			return fmt.Errorf("login failed: unable to connect to Zabbix API at %s: %w", newCtx.URL, err)
		}

		// Test auth scope
		if _, err := client.Call(timeoutCtx, "host.get", map[string]interface{}{"countOutput": true}); err != nil {
			return fmt.Errorf("login failed: authentication failed against Zabbix API: %w", err)
		}

		// Save context
		path, err := getConfigPath(cmd)
		if err != nil {
			return fmt.Errorf("failed to determine config path: %w", err)
		}

		diskCfg, loadErr := config.LoadConfig(path)
		if loadErr != nil || diskCfg == nil {
			diskCfg = config.DefaultConfig()
		}
		if diskCfg.Contexts == nil {
			diskCfg.Contexts = make(map[string]config.Context)
		}
		diskCfg.Contexts[ctxName] = newCtx
		diskCfg.ActiveContext = ctxName

		if err := config.SaveConfig(diskCfg, path); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		cfg = diskCfg
		actCtxName = ctxName
		activeCtx = &newCtx

		fmt.Printf("Successfully logged into Zabbix API at %s (Version: %s)!\nActive context updated to %q in %s\n", newCtx.URL, ver, ctxName, path)
		return nil
	},
}

func probeAlternativeURL(rawURL, triedURL string) string {
	if strings.Contains(rawURL, "api_jsonrpc.php") {
		return ""
	}
	if strings.HasSuffix(triedURL, "/api_jsonrpc.php") && !strings.HasSuffix(triedURL, "/zabbix/api_jsonrpc.php") {
		return strings.TrimSuffix(triedURL, "/api_jsonrpc.php") + "/zabbix/api_jsonrpc.php"
	}
	if strings.HasSuffix(triedURL, "/zabbix/api_jsonrpc.php") {
		return strings.TrimSuffix(triedURL, "/zabbix/api_jsonrpc.php") + "/api_jsonrpc.php"
	}
	return ""
}

func init() {
	loginCmd.Flags().StringVar(&loginTokenFlag, "token", "", "Zabbix API Bearer token")
	loginCmd.Flags().StringVarP(&loginUserFlag, "username", "u", "", "Zabbix username")
	loginCmd.Flags().StringVarP(&loginPassFlag, "password", "p", "", "Zabbix password")
	loginCmd.Flags().StringVar(&loginNameFlag, "name", "default", "context name to save")
	loginCmd.Flags().StringVar(&loginSafetyFlag, "safety-level", "readonly", "safety level (readonly, readwrite-mine, readwrite-all, dangerously-unrestricted)")
	loginCmd.Flags().StringVar(&loginHTTPUserFlag, "http-user", "", "HTTP Basic Auth username")
	loginCmd.Flags().StringVar(&loginHTTPPassFlag, "http-password", "", "HTTP Basic Auth password")
	loginCmd.Flags().BoolVar(&loginInsecureSkipFlag, "insecure", false, "skip TLS certificate verification")
	RootCmd.AddCommand(loginCmd)
}
