package cmd

import (
	"fmt"

	"github.com/ErKatta/zbxctl/pkg/zabbix"
	"github.com/spf13/cobra"
)

var briefFlag bool

type CommandMeta struct {
	Name        string   `json:"name"`
	Aliases     []string `json:"aliases,omitempty"`
	Description string   `json:"description"`
}

type VerbTree struct {
	Verb        string        `json:"verb"`
	Description string        `json:"description"`
	Resources   []CommandMeta `json:"resources,omitempty"`
}

type CommandExport struct {
	Tool        string     `json:"tool"`
	Version     string     `json:"version"`
	Description string     `json:"description"`
	SafetyLevels []string  `json:"safety_levels"`
	Verbs       []VerbTree `json:"verbs"`
	RawEngine   string     `json:"raw_engine"`
}

var commandsCmd = &cobra.Command{
	Use:   "commands [--brief|--full]",
	Short: "Export a compact tree of all zbxctl commands & resources for LLM context loading",
	Long:  `commands exports a machine-readable summary of all verbs, resources, flags, safety levels, and Tier 2 raw API capabilities.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var resources []CommandMeta
		for _, info := range zabbix.ResourceMap {
			resources = append(resources, CommandMeta{
				Name:        info.Name,
				Aliases:     info.Aliases,
				Description: info.Description,
			})
		}

		verbs := []VerbTree{
			{
				Verb:        "get",
				Description: "Fetch resources or list items with optional filtering",
				Resources:   resources,
			},
			{
				Verb:        "describe",
				Description: "Inspect detailed metadata for a single resource by ID or name",
				Resources:   resources,
			},
			{
				Verb:        "apply",
				Description: "Declaratively create or update resources from JSON/YAML manifest files",
				Resources:   resources,
			},
			{
				Verb:        "edit",
				Description: "Edit a live resource directly in your preferred text editor",
				Resources:   resources,
			},
			{
				Verb:        "delete",
				Description: "Remove target resources by ID",
				Resources:   resources,
			},
			{
				Verb:        "query",
				Description: "Advanced search and filtering across Zabbix resources",
				Resources:   resources,
			},
			{
				Verb:        "exec",
				Description: "Execute a Zabbix script on a target host",
			},
			{
				Verb:        "wait",
				Description: "Wait for a specific condition or problem resolution",
			},
			{
				Verb:        "diff",
				Description: "Compare a local YAML manifest with the live Zabbix resource",
			},
			{
				Verb:        "doctor",
				Description: "Run connectivity and health diagnostics",
			},
			{
				Verb:        "cluster-info",
				Description: "Display Zabbix instance connection, version, and sizing statistics",
			},
			{
				Verb:        "version",
				Description: "Print the zbxctl client version and build metadata",
			},
			{
				Verb:        "update",
				Description: "Check for releases and self-update zbxctl binary in-place",
			},
			{
				Verb:        "skill",
				Description: "Manage packaged AI agent skills (install, list, show, export)",
			},
			{
				Verb:        "raw",
				Description: "Tier 2 raw JSON-RPC caller for 100% Zabbix 7 API coverage (e.g. proxygroup.get, ha.get)",
			},
		}

		export := CommandExport{
			Tool:        "zbxctl",
			Version:     "7.0.0-cli",
			Description: "AI-Agent Friendly Zabbix 7 CLI",
			SafetyLevels: []string{
				"readonly",
				"readwrite-mine",
				"readwrite-all",
				"dangerously-unrestricted",
			},
			Verbs:     verbs,
			RawEngine: "zbxctl raw <zabbix.method> [--params='<json>']",
		}

		if briefFlag {
			// Compact text tree representation
			fmt.Println("zbxctl Tiered Command Tree:")
			for _, v := range verbs {
				fmt.Printf("  zbxctl %s\n", v.Verb)
				if len(v.Resources) > 0 {
					for _, r := range v.Resources {
						fmt.Printf("    ├── %s (aliases: %v)\n", r.Name, r.Aliases)
					}
				}
			}
			fmt.Println("  zbxctl raw <zabbix.method> [--params='<json>']")
			return nil
		}

		return formatter.Print(export)
	},
}

func init() {
	commandsCmd.Flags().BoolVar(&briefFlag, "brief", false, "emit compact human-readable tree text")
	RootCmd.AddCommand(commandsCmd)
}
