package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SkillInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content,omitempty"`
}

var BuiltinSkills = map[string]SkillInfo{
	"zabbix-automation": {
		Name:        "zabbix-automation",
		Description: "Declarative Zabbix 7 resource management, GitOps manifests, apply/diff workflows, and spec synchronization.",
		Content: `---
name: zabbix-automation
description: "Declarative Zabbix 7 resource management, GitOps manifests, apply/diff workflows, and spec synchronization."
---

# Zabbix 7 Automation & GitOps Workflow Guide

Use this skill when managing Zabbix 7 monitoring infrastructure declaratively using ` + "`" + `zbxctl` + "`" + `.

## Core Principles
1. **Declarative Spec Manifests**: Define resources in YAML or JSON files containing ` + "`" + `kind` + "`" + ` and ` + "`" + `spec` + "`" + `.
2. **Drift Detection**: Always run ` + "`" + `zbxctl diff -f manifest.yaml` + "`" + ` before applying changes to inspect live server state differences.
3. **Safe Application**: Apply changes using ` + "`" + `zbxctl apply -f manifest.yaml` + "`" + `.

## Manifest Format Example

` + "```" + `yaml
kind: host
spec:
  host: web-prod-01
  name: "Web Production 01"
  interfaces:
    - type: 1 # Agent
      main: 1
      useip: 1
      ip: 192.168.1.50
      port: "10050"
  groups:
    - groupid: "2" # Linux servers
  tags:
    - tag: zbxctl
      value: "true"
    - tag: environment
      value: production
` + "```" + `

## Domain Disambiguation: Hosts vs. Inventory vs. Sizing
- **Configured Monitoring Hosts**: Use ` + "`" + `zbxctl get host` + "`" + ` or ` + "`" + `zbxctl get host --count` + "`" + ` to query all monitoring endpoints.
- **Instance Sizing / Sizing Overview**: Use ` + "`" + `zbxctl cluster-info` + "`" + ` to get total counts across all object types (hosts, items, triggers, problems).
- **Zabbix Asset / Hardware Inventory**: Use ` + "`" + `zbxctl get inventory` + "`" + ` (or ` + "`" + `zbxctl get inv` + "`" + `) for dedicated CMDB asset queries, ` + "`" + `zbxctl describe inventory <id>` + "`" + ` or ` + "`" + `zbxctl describe host <id>` + "`" + ` for detailed metadata, and ` + "`" + `zbxctl apply -f inv.yaml` + "`" + ` with ` + "`" + `kind: inventory` + "`" + ` for declarative asset synchronization.

## Recipes

` + "```" + `bash
# Check zbxctl version
zbxctl version

# Count total configured hosts directly (single integer output)
zbxctl get host --count

# Sizing overview of entire Zabbix instance
zbxctl cluster-info

# Query hardware and asset CMDB inventory records
zbxctl get inventory -o table
zbxctl get inventory -f hostid,name,vendor,model,macaddress_a -o table
zbxctl get inv --search="Dell" -o json

# Query resources with projection and alphabetical sorting
zbxctl get template --fields=templateid,name --sort=name -o table
zbxctl get host --search="prod" --sort=name -o table

# Export live resource as declarative manifest
zbxctl get host web-prod-01 -o yaml > host-manifest.yaml
zbxctl get template --export -o yaml

# Dry-run diff against live Zabbix server
zbxctl diff -f host-manifest.yaml

# Apply declarative spec (via file or stdin)
zbxctl apply -f host-manifest.yaml
zbxctl apply -f inv-manifest.yaml
cat inv-manifest.yaml | zbxctl apply -f -

# Stream declarative manifests directly from Python scripts via stdin:
# subprocess.run(["zbxctl", "apply", "-f", "-"], input=manifest_yaml, text=True)

# Verify created host (including its Zabbix Host Inventory block)
zbxctl describe host web-prod-01
zbxctl describe inventory web-prod-01
` + "```" + `
`,
	},
	"zabbix-troubleshooting": {
		Name:        "zabbix-troubleshooting",
		Description: "Incident investigation, problem analysis, trigger thresholds, and root-cause diagnosis for Zabbix 7.",
		Content: `---
name: zabbix-troubleshooting
description: "Incident investigation, problem analysis, trigger thresholds, and root-cause diagnosis for Zabbix 7."
---

# Zabbix 7 Incident Troubleshooting Guide

Use this skill when investigating active alerts, high-severity problems, or agent connectivity issues in Zabbix 7.

## Diagnostic Flow

1. **System Health & Sizing Check**: Run ` + "`" + `zbxctl doctor` + "`" + ` and ` + "`" + `zbxctl cluster-info` + "`" + ` to confirm API connectivity, authentication, and instance state.
2. **Problem Discovery**: Query active problem counts with ` + "`" + `zbxctl get problem --count` + "`" + ` and list details with ` + "`" + `zbxctl get problem --filter='{"severity": 4}' --sort=severity --sort-order=desc -o json` + "`" + `.
3. **Host Metadata Inspection**: Inspect target host configurations using ` + "`" + `zbxctl describe host <host_id_or_name>` + "`" + `.
4. **Metric Telemetry Analysis**: Retrieve numerical history for problem items using ` + "`" + `zbxctl get metric <item_id> --since=4h` + "`" + `.
5. **Resolution Polling**: Wait for problem resolution using ` + "`" + `zbxctl wait problem <event_id> --for=resolved --timeout=60s` + "`" + `.

## Command Examples

` + "```" + `bash
# Run connectivity diagnostic
zbxctl doctor

# Count active problems
zbxctl get problem --count

# List high severity active problems sorted by severity
zbxctl get problem --filter='{"severity": 4}' --sort=severity --sort-order=desc -o json

# Get items for problem host with search and field projection
zbxctl get item --host="Zabbix server" --search="cpu" --fields=itemid,name,key_,lastvalue

# Fetch telemetry over the last 4 hours
zbxctl get metric 23253 --since=4h -o table
` + "```" + `
`,
	},
	"zabbix-telemetry": {
		Name:        "zabbix-telemetry",
		Description: "High-frequency metric extraction, time window filtering (--since=4h), and LLM token optimization (-o toon).",
		Content: `---
name: zabbix-telemetry
description: "High-frequency metric extraction, time window filtering (--since=4h), and LLM token optimization (-o toon)."
---

# Zabbix 7 Telemetry & Metric Analysis Guide

Use this skill when extracting numerical time-series metric data for LLM context windows or pipeline reporting.

## Key Verbs & Aliases
- ` + "`" + `zbxctl get metric <item_id>` + "`" + `
- ` + "`" + `zbxctl get telemetry <item_id>` + "`" + `
- ` + "`" + `zbxctl get history <item_id>` + "`" + `

All three options access the Zabbix ` + "`" + `history.get` + "`" + ` engine.

## Duration Window Filtering (` + "`" + `--since` + "`" + `)
Specify relative duration windows:
- ` + "`" + `--since=30m` + "`" + ` (last 30 minutes)
- ` + "`" + `--since=4h` + "`" + ` (last 4 hours)
- ` + "`" + `--since=1d` + "`" + ` (last 24 hours)

## Token Optimization (` + "`" + `-o toon` + "`" + `)
When sending telemetry outputs to LLM agent prompt turns, use ` + "`" + `-o toon` + "`" + ` to compress JSON arrays by 40-60% without data loss.

## Recipes

` + "```" + `bash
# Fetch 4 hours of CPU telemetry in compressed TOON format
zbxctl get metric 23253 --since=4h --limit=20 -o toon

# Fetch ICMP response latency in JSON
zbxctl get telemetry 50317 --since=1h -o json
` + "```" + `
`,
	},
	"zabbix-safety": {
		Name:        "zabbix-safety",
		Description: "Client-side safety middleware levels (readonly, readwrite-mine, readwrite-all) and Exit Code 2 error envelopes.",
		Content: `---
name: zabbix-safety
description: "Client-side safety middleware levels (readonly, readwrite-mine, readwrite-all) and Exit Code 2 error envelopes."
---

# Zabbix 7 Safety Middleware & Guardrails Guide

Use this skill to configure and verify client-side safety execution guardrails in ` + "`" + `zbxctl` + "`" + `.

## Safety Levels
1. ` + "`" + `readonly` + "`" + `: Blocks all mutations (` + "`" + `create` + "`" + `, ` + "`" + `update` + "`" + `, ` + "`" + `delete` + "`" + `, ` + "`" + `exec` + "`" + `). Exits cleanly with **Exit Code 2**.
2. ` + "`" + `readwrite-mine` + "`" + `: Permits mutations only on resources tagged ` + "`" + `zbxctl=true` + "`" + ` or ` + "`" + `managed-by=zbxctl` + "`" + `.
3. ` + "`" + `readwrite-all` + "`" + `: Permits full API mutations, but blocks bulk deletions without ` + "`" + `--force` + "`" + `.
4. ` + "`" + `dangerously-unrestricted` + "`" + `: Disables client safety checks.

## Exit Code 2 Error Envelopes
Safety violations emit structured JSON on ` + "`" + `stderr` + "`" + ` with Exit Code 2:

` + "```" + `json
{
  "error": {
    "code": "SAFETY_LEVEL_VIOLATION",
    "method": "host.delete",
    "message": "Operation blocked by safety-level 'readonly' on context 'prod'."
  }
}
` + "```" + `
`,
	},
}

func GetAgentSkillDirs(agent string, global bool) ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	var dirs []string

	switch strings.ToLower(strings.TrimSpace(agent)) {
	case "claude":
		if global {
			dirs = append(dirs, filepath.Join(home, ".claude", "skills"))
		} else {
			dirs = append(dirs, filepath.Join(cwd, ".claude", "skills"))
		}
	case "cursor":
		if global {
			dirs = append(dirs, filepath.Join(home, ".cursor", "rules"))
		} else {
			dirs = append(dirs, filepath.Join(cwd, ".cursor", "rules"))
		}
	case "antigravity", "gemini":
		if global {
			dirs = append(dirs, filepath.Join(home, ".gemini", "config", "skills"))
		} else {
			dirs = append(dirs, filepath.Join(cwd, ".agents", "skills"))
		}
	default:
		if global {
			dirs = append(dirs,
				filepath.Join(home, ".gemini", "config", "skills"),
				filepath.Join(home, ".claude", "skills"),
				filepath.Join(home, ".config", "zbxctl", "skills"),
			)
		} else {
			dirs = append(dirs,
				filepath.Join(cwd, ".agents", "skills"),
				filepath.Join(cwd, ".claude", "skills"),
				filepath.Join(cwd, ".zbxctl", "skills"),
			)
		}
	}

	return dirs, nil
}

func GetGlobalSkillsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gemini", "config", "skills"), nil
}

func GetWorkspaceSkillsDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, ".agents", "skills"), nil
}

func InstallSkill(skillName string, targetDir string) (string, error) {
	skill, ok := BuiltinSkills[skillName]
	if !ok {
		return "", fmt.Errorf("skill %q not found. Use 'zbxctl skill list' to view available built-in skills", skillName)
	}

	skillFolder := filepath.Join(targetDir, skill.Name)
	if err := os.MkdirAll(skillFolder, 0755); err != nil {
		return "", fmt.Errorf("failed to create skill directory %s: %w", skillFolder, err)
	}

	skillFilePath := filepath.Join(skillFolder, "SKILL.md")
	if err := os.WriteFile(skillFilePath, []byte(skill.Content), 0644); err != nil {
		return "", fmt.Errorf("failed to write skill file %s: %w", skillFilePath, err)
	}

	return skillFilePath, nil
}
