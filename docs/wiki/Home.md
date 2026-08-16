# `zbxctl` Wiki & Technical Documentation

Welcome to the **`zbxctl`** official Wiki!

`zbxctl` is an open-source, developer-first command-line tool for **Zabbix 7.0+ LTS**. Inspired by **`dtctl`** (Dynatrace CLI) and **`kubectl`** (Kubernetes CLI), `zbxctl` moves from heavyweight Model Context Protocol (MCP) server daemons to direct shell execution—accelerating automation workflow development.

---

## Wiki Sitemap

- [**Architecture & Design Philosophy**](Architecture-and-Design.md): Tier 1 ergonomic verbs vs Tier 2 raw JSON-RPC caller.
- [**`zabbix-cli` Comparison & Analysis**](Zabbix-CLI-Comparison.md): Deep feature comparison & use-cases with the Python `zabbix-cli`.
- [**Safety Guardrails Matrix**](Safety-Guardrails-Matrix.md): Client-side execution safety levels and error handling.
- [**AI Agent Integration Guide**](AI-Agent-Integration-Guide.md): Token-optimized outputs (`-o toon`, `--brief`) for LLM pair programming.
- [**AI Agent Skill Management**](AI-Agent-Skill-Management.md): Multi-agent skill installation (`zbxctl skill install --all`) for Antigravity, Claude, and Cursor.
- [**CI/CD & Automation Recipes**](CI-CD-Automation-Recipes.md): Shell scripts, GitHub Actions, and automated Zabbix operations.
- [**Software Lifecycle Management (SLCM)**](Software-Lifecycle-Management.md): Branching strategy, multi-OS CI matrix, GoReleaser pipeline, and semantic versioning.

---

## Quick Commands Cheatsheet

```bash
# Initialize context configuration
zbxctl config init

# Register active context
zbxctl login https://zabbix.example.com/api_jsonrpc.php \
 --token="your_api_token" \
 --name="prod-us" \
 --safety-level="readonly"

# Diagnostic check
zbxctl doctor

# List high severity problems
zbxctl get problem --filter='{"severity": 4}' -o json

# Fetch metric telemetry for the last 4 hours
zbxctl get metric 23253 --since=4h -o table

# Declarative resource manifest apply
zbxctl apply -f host-manifest.yaml

# Universal raw API fallback (100% Zabbix 7 API coverage)
zbxctl raw host.get --params='{"output": ["hostid", "host"]}'
```
