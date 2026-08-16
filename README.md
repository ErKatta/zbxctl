<p align="center">
 <img src="docs/logo.svg" alt="zbxctl Logo" width="140">
</p>

# `zbxctl` - Open Source Zabbix 7 CLI & Automation Engine

[![CI Pipeline](https://github.com/ErKatta/zbxctl/actions/workflows/ci.yml/badge.svg)](https://github.com/ErKatta/zbxctl/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/ErKatta/zbxctl?color=d1232a&label=release)](https://github.com/ErKatta/zbxctl/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/ErKatta/zbxctl.svg)](https://pkg.go.dev/github.com/ErKatta/zbxctl)
[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8.svg)](go.mod)
[![Zabbix Version](https://img.shields.io/badge/Zabbix-7.0%2B%20LTS-d1232a.svg)](https://www.zabbix.com)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

> `zbxctl` is an open-source CLI and automation engine for Zabbix 7.0+ LTS. Inspired by **`kubectl`** (Kubernetes CLI) and **`dtctl`** (Dynatrace CLI), its core goal is to transition developers and AI agents from heavyweight Model Context Protocol (MCP) server daemons to direct CLI execution—accelerating automation workflow development.

[**Website**](https://ErKatta.github.io/zbxctl/) | [**Wiki**](docs/wiki/Home.md) | [**Contributing**](CONTRIBUTING.md) | [**Security**](SECURITY.md) | [**Code of Conduct**](CODE_OF_CONDUCT.md)

---

## Why Shift from MCP to the CLI?

- **Lower Context Overhead:** MCP tool definitions and RPC envelopes bloat prompt turns. `zbxctl` emits token-optimized outputs (`-o toon`, `--brief`) directly over stdout, saving context tokens.
- **Universal Agent & Tooling Support:** Every AI coding assistant (Gemini, Claude, Cursor, Aider, custom scripts) and CI/CD runner natively supports shell execution out-of-the-box without running background MCP server daemons.
- **Accelerated Automation Workflows:** Fast, interactive shell feedback loops allow testing Zabbix queries, composing Unix pipes (`| jq`), and building scripts without writing custom Go or Python API wrappers.
- **Deterministic Client Guardrails:** Safety middleware (`readonly`, `readwrite-mine`) runs directly in compiled code before sending HTTP requests to Zabbix.

---

## Key Features

- **Tiered Command Architecture:** High-frequency ergonomic verbs (`get`, `describe`, `apply`, `delete`, `query`, `exec`, `wait`, `diff`, `doctor`, `inventory`) paired with a Tier 2 Universal Raw JSON-RPC engine (`zbxctl raw`).
- **Client-Side Safety Middleware:** Prevents accidental production outages with granular safety enforcement (`readonly`, `readwrite-mine`, `readwrite-all`, `dangerously-unrestricted`).
- **Credential Redaction:** Automatic redaction of sensitive API tokens, passwords, and HTTP headers in terminal/JSON output to keep credentials out of LLM contexts and logs.
- **LLM Native Context Loading:** Compact self-discovery tree (`zbxctl commands --brief`) and Token-Optimized Object Notation (`-o toon`) to optimize prompt token usage.
- **Universal Authentication:** Native support for API Tokens, Username/Password, HTTP Basic Auth, SSO custom headers (`X-Remote-User`), and mTLS client certificates.
- **Open Source (Apache 2.0):** Released under the permissive **Apache License 2.0**. Contributions, issue reports, and feedback are warmly welcomed!

---

## 1. Installation

`zbxctl` is distributed as a single static binary with zero external runtime dependencies. **You do not need to install the Go toolchain** to use it.

### Precompiled Binaries (Linux, macOS, Windows)

Precompiled binaries for Linux, macOS, and Windows are published with every release on the [GitHub Releases](https://github.com/ErKatta/zbxctl/releases/latest) page.

#### Linux (x86_64 / ARM64)
```bash
# Download and install the latest Linux binary
curl -sL https://github.com/ErKatta/zbxctl/releases/latest/download/zbxctl_$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/').tar.gz | tar -xz
sudo mv zbxctl /usr/local/bin/
```

#### macOS (Apple Silicon M-Series / Intel)
```bash
# Download and install the latest macOS binary
curl -sL https://github.com/ErKatta/zbxctl/releases/latest/download/zbxctl_darwin_$(uname -m | sed 's/x86_64/amd64/' | sed 's/arm64/arm64/').tar.gz | tar -xz
sudo mv zbxctl /usr/local/bin/
```

#### Windows (x86_64)
Download the latest [`zbxctl_windows_amd64.zip`](https://github.com/ErKatta/zbxctl/releases/latest) from GitHub Releases, extract the archive, and add `zbxctl.exe` to your `PATH`.
Or install via PowerShell:
```powershell
Invoke-WebRequest -Uri "https://github.com/ErKatta/zbxctl/releases/latest/download/zbxctl_windows_amd64.zip" -OutFile "zbxctl.zip"
Expand-Archive zbxctl.zip -DestinationPath "$HOME\bin"
```

---

### Alternative Installation Methods

#### Go Install (For Go Developers)
If you already have Go 1.22+ installed:
```bash
go install github.com/ErKatta/zbxctl@latest
```

#### Build From Source
```bash
git clone https://github.com/ErKatta/zbxctl.git
cd zbxctl
go build -o zbxctl main.go
sudo mv zbxctl /usr/local/bin/
```

---

## 2. Quickstart & Configuration

Initialize your configuration file (`~/.zbxctl/config.yaml`):

```bash
zbxctl config init
```

Log in and register a context:

```bash
zbxctl login https://zabbix.example.com/api_jsonrpc.php \
 --token="zabbix_api_token_here" \
 --name="prod-us" \
 --safety-level="readonly"
```

Verify connectivity and system diagnostics:

```bash
zbxctl doctor
```

Inspect current active configuration (with automatic credential redaction):

```bash
zbxctl config current-config
```

---

## 3. Supported Authentication Mechanisms

`zbxctl` supports all authentication models present in Zabbix 7:

```yaml
active_context: prod
contexts:
 prod:
 url: "https://zabbix.company.com/api_jsonrpc.php"
 api_token: "zabbix_api_token_here"
 safety_level: "readonly"
 internal-ldap:
 url: "https://zabbix.internal/api_jsonrpc.php"
 username: "Admin"
 password: "secret_password"
 safety_level: "readwrite-mine"
 proxy-basic:
 url: "https://zabbix-proxy.net/api_jsonrpc.php"
 http_user: "proxy_user"
 http_password: "proxy_password"
 api_token: "zabbix_api_token"
 safety_level: "readwrite-all"
 sso-headers:
 url: "https://zabbix-sso.company.com/api_jsonrpc.php"
 api_token: "zabbix_api_token"
 http_headers:
 X-Remote-User: "admin@company.com"
 secure-mtls:
 url: "https://zabbix-mtls.company.com/api_jsonrpc.php"
 api_token: "zabbix_api_token"
 tls_cert_file: "/etc/ssl/client.crt"
 tls_key_file: "/etc/ssl/client.key"
 tls_ca_file: "/etc/ssl/custom_ca.crt"
```

---

## 4. Safety Middleware & Enforcement

Every API call passes through client-side safety middleware enforced by the active context's `safety_level`:

| Safety Level | Permitted Operations | Blocked Operations |
| :--- | :--- | :--- |
| `readonly` | Read-only operations (`*.get`, `describe`, `query`, `doctor`, `inventory`) | Any mutation (`*.create`, `*.update`, `*.delete`, `exec`) |
| `readwrite-mine` | Read-only + mutations on resources tagged `zbxctl=true` or `managed-by=zbxctl` | Un-tagged or instance-wide destructive edits |
| `readwrite-all` | Full API mutations | Destructive bulk deletes without explicit `--force` |
| `dangerously-unrestricted` | All operations enabled without safety checks | None |

When a safety violation occurs, `zbxctl` emits a structured JSON error envelope on `stderr` with **Exit Code 2**:

```json
{
 "error": {
 "code": "SAFETY_LEVEL_VIOLATION",
 "method": "host.delete",
 "message": "Operation blocked by safety-level 'readonly' on context 'prod'.",
 "resolution": "Switch context or update safety-level in ~/.zbxctl/config.yaml."
 }
}
```

---

## 5. Command Reference

### Tier 1 Ergonomic Commands
```bash
# List hosts or filter problems
zbxctl get host
zbxctl get problem --filter='{"severity": 4}' -o json

# Fetch metric telemetry for a duration window (e.g. last 4h)
# (Note: 'metric', 'metrics', 'telemetry', and 'history' are zero-overhead aliases pointing to history.get)
zbxctl get metric 23253 --since=4h -o table
zbxctl get telemetry 23253 --since=4h -o table

# Get items belonging to a host with specific fields
zbxctl get item --host="Zabbix server" --fields=itemid,name,description

# Inspect detailed metadata for a resource
zbxctl describe host 10001

# Advanced query search
zbxctl query item --search="key_=system.cpu" --limit=5

# Declarative creation/update from YAML/JSON manifest
zbxctl apply -f manifest.yaml

# Compare local manifest with live resource
zbxctl diff -f manifest.yaml --id=10001

# Delete resources (requires non-readonly safety level)
zbxctl delete host 10003
zbxctl delete host 10001 10002 --force

# Execute Zabbix script on host
zbxctl exec 1 --hostid=10001

# Wait for problem resolution or metric condition
zbxctl wait problem 12345 --for=resolved --timeout=60s
```

### Tier 2 Universal Raw Engine (100% API Coverage)
```bash
zbxctl raw proxygroup.get --params='{"output": "extend"}' -o json
zbxctl raw history.push --params='[{"itemid": 10001, "value": "42.0"}]' -o json
zbxctl raw ha.get --params='{}' -o json
```

### Configuration & Context Commands
```bash
# Display active context configuration (with [REDACTED] credential masking)
zbxctl config current-config

# List all configured contexts
zbxctl config get-contexts

# Switch active context
zbxctl config use-context prod-us

# Update context safety level
zbxctl config set-safety readwrite-mine
```

### AI Agent Skill Commands
```bash
# List available built-in skills
zbxctl skill list

# Install all built-in skills to agent configuration (~/.gemini/config/skills)
zbxctl skill install --all

# Show content of a specific skill
zbxctl skill show zabbix-automation

# Export skill to workspace repository (.agents/skills)
zbxctl skill export zabbix-troubleshooting
```

---

## 6. Output Formats

Pass `-o` or `--output`:
- `json`: Pretty-printed JSON object/array output.
- `table`: Auto-formatted terminal table with ANSI color coding for problem severities.
- `toon`: Token-Optimized Object Notation for LLM prompts.
- `yaml`: Standard YAML representation.
- `auto`: Automatically selects `table` when stdout is TTY, and `json` when non-TTY.

---

## 7. License & Commercial Usage

`zbxctl` is licensed under the **Apache License 2.0**. This allows free commercial usage, modification, distribution, and private use without copyleft restrictions, making it 100% compatible with Zabbix ecosystem tooling and commercial enterprise deployments.

---

## 8. Acknowledgements & Citations

`zbxctl` draws inspiration from the design patterns and contributions of existing open-source tools:

- **[dtctl](https://github.com/dynatrace-oss/dtctl)** (*Dynatrace OSS*): For introducing ergonomic CLI verbs and structured observability workflows. If referencing `dtctl`, please see their [CITATION.cff](https://github.com/dynatrace-oss/dtctl/blob/main/CITATION.cff).
- **[kubectl](https://kubernetes.io/)** (*Kubernetes Authors / CNCF*): For declarative resource management (`apply`/`diff`), standard verbs, and multi-context configuration patterns.
- **[zabbix-cli](https://github.com/unioslo/zabbix-cli)** (*University of Oslo / unioslo*): For providing a dedicated terminal administration tool for the Zabbix community since 2014.

### Citing `zbxctl`
If you use `zbxctl` in your work, please cite it using the following citation file [`CITATION.cff`](CITATION.cff).
