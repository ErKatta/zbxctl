---
name: zbxctl
description: Zabbix 7 CLI and Automation Engine for AI Agents and Systems Engineers
---

# `zbxctl` - AI-Agent Friendly Zabbix 7 CLI

`zbxctl` is a high-performance command-line tool for interacting with Zabbix 7.0+ LTS monitoring systems.
It provides a **Tiered Command Architecture**:
1. **Tier 1 Ergonomic Commands:** High-frequency verb-noun operations (`get`, `describe`, `apply`, `delete`, `query`, `exec`, `wait`, `diff`, `doctor`, `cluster-info`).
2. **Tier 2 Universal Raw Engine:** `zbxctl raw <zabbix.method> [--params='<json>']` providing 100% Zabbix 7 API coverage with full client-side safety middleware enforcement.

---

## 1. Supported Authentication Methods

`zbxctl` natively supports all authentication mechanisms available in Zabbix 7 & web infrastructure:

1. **Zabbix API Token (Recommended):**
 ```yaml
 active_context: prod
 contexts:
 prod:
 url: "https://zabbix.company.com/api_jsonrpc.php"
 api_token: "zabbix_api_token_here"
 safety_level: "readonly"
 ```

2. **Username / Password (`user.login` with internal Zabbix or LDAP/AD backend):**
 ```yaml
 contexts:
 internal:
 url: "https://zabbix.local/api_jsonrpc.php"
 username: "Admin"
 password: "secret_password"
 safety_level: "readwrite-mine"
 ```

3. **HTTP Basic Authentication (Reverse Proxy / Web Server Auth):**
 ```yaml
 contexts:
 proxy-auth:
 url: "https://zabbix.internal.net/api_jsonrpc.php"
 http_user: "proxy_user"
 http_password: "proxy_password"
 api_token: "zabbix_api_token"
 safety_level: "readwrite-all"
 ```

4. **SSO / Custom Proxy Headers (e.g. Authelia, Keycloak, OAuth2 Proxy, Remote User):**
 ```yaml
 contexts:
 sso-auth:
 url: "https://zabbix-sso.company.com/api_jsonrpc.php"
 api_token: "zabbix_api_token"
 http_headers:
 X-Remote-User: "admin@company.com"
 X-Forwarded-User: "admin"
 ```

5. **Client TLS Certificates (mTLS) & Custom CA Certificates:**
 ```yaml
 contexts:
 secure-mtls:
 url: "https://zabbix-mtls.company.com/api_jsonrpc.php"
 api_token: "zabbix_api_token"
 tls_cert_file: "/etc/ssl/client.crt"
 tls_key_file: "/etc/ssl/client.key"
 tls_ca_file: "/etc/ssl/custom_ca.crt"
 insecure_skip_verify: false
 ```

---

## 2. Safety Levels & Enforcement

Configuration is stored in `~/.zbxctl/config.yaml`.
Every command passes through safety middleware according to the active context's `safety_level`:

| Safety Level | Permitted Operations | Blocked Operations |
| :--- | :--- | :--- |
| `readonly` | Read-only actions (`*.get`, `describe`, `query`, `doctor`, `cluster-info`, `commands`) | Any mutating action (`*.create`, `*.update`, `*.delete`, `history.push`, `exec`) |
| `readwrite-mine` | Read-only + mutations on resources tagged `zbxctl=true` or `managed-by=zbxctl` | Un-tagged or instance-wide destructive edits |
| `readwrite-all` | Full administrative API mutations | Destructive bulk deletes without explicit `--force` |
| `dangerously-unrestricted` | All operations enabled without safety checks | None |

### Safety Violations & Exit Codes
- **Exit Code 2:** Safety level violation or missing required parameters. Emits a structured JSON error envelope on `stderr`:
```json
{
 "error": {
 "code": "SAFETY_LEVEL_VIOLATION",
 "method": "host.delete",
 "message": "Operation blocked by safety-level 'readonly' on context 'prod-us'.",
 "resolution": "Switch context or update safety-level in ~/.zbxctl/config.yaml."
 }
}
```
- **Exit Code 1:** Network or API processing error.
- **Exit Code 0:** Command succeeded cleanly.

---

## 3. Low-Token Agent Self-Discovery & Grounding Commands

When starting a session or grounding context, run:

```bash
# 1. Discover all verbs, resources, flags, and raw capabilities (low-token tree format)
zbxctl commands --brief

# 2. Check zbxctl CLI client version & build metadata
zbxctl version

# 3. Probe target Zabbix instance for sizing & stats (version, total hosts, active problems, items, triggers)
zbxctl cluster-info

# 4. Diagnose connectivity, credentials, and context safety
zbxctl doctor
```

---

## 4. Tier 1 Ergonomic Commands (High-Frequency Operations)

### Resource Short Aliases
- `host` (`h`, `hosts`), `problem` (`p`, `problems`), `item` (`i`, `items`), `trigger` (`t`, `triggers`), `template` (`tmpl`, `templates`), `hostgroup` (`hg`, `hostgroups`), `maintenance` (`maint`), `event` (`ev`), `user`, `service`, `sla`, `dashboard`.

### Search, Sort, and Projection Flags
- `--count`: Return only the total integer count of matched resources (e.g. `zbxctl get host --count`, `zbxctl get problem --count`, `zbxctl get item --host="web-01" --count`).
- `--fields, -f`: Select specific output fields (e.g. `--fields=templateid,name` or `-f hostid,name,status`).
- `--sort, -s`: Sort by field name (e.g. `--sort=name`, `--sort=clock`, `--sort=severity`, or `--sort=name:desc`).
- `--sort-order`: Sort direction (`asc` or `desc`, default `asc`).
- `--search`: Search term across primary resource text fields (partial match with `*` wildcards supported) or `key=value`.
- `--search-fields`: Target specific fields for `--search` (e.g. `--search-fields=name,description`).
- `--filter`: Exact key-value match or JSON map filter (e.g. `--filter='{"severity": 4}'` or `--filter=status=0`).
- `--wide, -w`: Display wide table with all fields.
- `--export`: Export resources as declarative GitOps manifests (`kind` + `spec`) ready for `apply`/`diff`.

### Domain Disambiguation: Hosts vs. Inventory vs. Sizing
- **Configured Monitoring Hosts**: Use `zbxctl get host` or `zbxctl get host --count` to query all monitoring endpoints.
- **Instance Sizing / Sizing Overview**: Use `zbxctl cluster-info` to get total counts across all object types.
- **Zabbix Asset / Hardware Inventory**: Use `zbxctl describe host <id>` to inspect the hardware/OS inventory block, or `zbxctl get host --filter='{"inventory_mode": [0,1]}'` to filter inventory-enabled hosts.

### Recipes & Examples
```bash
# Count total configured hosts directly (single integer output)
zbxctl get host --count

# Count active problems
zbxctl get problem --count

# Sizing overview of entire instance
zbxctl cluster-info

# List templates with projection and alphabetical sorting
zbxctl get template --fields=templateid,name --sort=name -o table

# Search hosts across host and name fields, sorted alphabetically
zbxctl get host --search="prod" --sort=name -o table

# Get items for a host with field projection and search
zbxctl get item --host="web-prod-01" --search="cpu" -f itemid,name,key_,lastvalue

# Query high-severity problems sorted by severity descending
zbxctl get problem --filter='{"severity": 4}' --sort=severity --sort-order=desc -o json

# Inspect detailed extended configuration for a host (including Host Inventory)
zbxctl describe host 10001

# Declarative creation/update from manifest
zbxctl apply -f manifest.yaml

# Compare local manifest against live Zabbix resource
zbxctl diff -f manifest.yaml --id=10001

# Export declarative manifests for all templates or a single host
zbxctl get template --export -o yaml
zbxctl get host 10001 -o yaml

# Delete resources (requires appropriate safety level and --force if bulk)
zbxctl delete host 10003
zbxctl delete host 10001 10002 --force

# Execute Zabbix script on host
zbxctl exec 1 --hostid=10001

# Wait for problem resolution or metric condition
zbxctl wait problem 12345 --for=resolved --timeout=60s
```

---

## 5. Tier 2 Universal Raw API Engine (100% Zabbix 7 API Coverage)

> **Rule for AI Agents:** Always prefer `zbxctl get` or `zbxctl query` for querying, filtering, sorting, and projecting standard Zabbix resources. Use `zbxctl raw` **only** for unmapped API methods (e.g. `connector.*`, `ha.*`, `proxygroup.*`, `valuemap.*`, `task.create`) or custom non-standard JSON payloads.

Invoke *any* Zabbix 7 JSON-RPC API endpoint with safety middleware filtering applied:

```bash
# Get proxy groups
zbxctl raw proxygroup.get --params='{"output": "extend"}' -o json

# Push metric history
zbxctl raw history.push --params='[{"itemid": 10001, "value": "42.0"}]' -o json

# High availability nodes probe
zbxctl raw ha.get --params='{}' -o json

# Get connector configurations
zbxctl raw connector.get --params='{"output": "extend"}' -o json
```

---

## 6. Machine Output Formats

Pass `-o` or `--output`:
- `json`: Clean JSON object/array output (or declarative manifests when inspecting single resources / `--export`).
- `table`: Auto-formatted compact terminal table with ANSI color coding for problem severities.
- `toon`: Token-Optimized Object Notation for LLM prompts.
- `yaml`: Standard YAML representation.
- `auto`: Automatically selects `table` when stdout is TTY, and `json` when non-TTY.
