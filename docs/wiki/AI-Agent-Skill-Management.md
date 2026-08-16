# AI Agent Skill Management (`zbxctl skill`)

`zbxctl` includes a native skill management system designed to equip AI coding assistants (Antigravity, Claude Code, Cursor, Aider, Copilot) and human engineers with specialized domain knowledge, operational guardrails, and automation recipes for **Zabbix 7.0+ LTS**.

---

## 💡 What are `zbxctl` Skills?

A **Skill** is a packaged, structured markdown guide (`SKILL.md`) containing YAML frontmatter metadata, command recipes, workflow patterns, and safety rules. When installed, AI agents automatically discover and load these skills to perform high-precision Zabbix operations without manual prompt engineering.

---

## 📦 Built-In Skills Suite

| Skill Name | Focus Area | Description |
| :--- | :--- | :--- |
| **`zabbix-automation`** | GitOps & Declarative Spec | Declarative resource manifests, dry-run diffing (`diff`), spec application (`apply`), and template management. |
| **`zabbix-troubleshooting`** | Incident Response | Active problem analysis, host metadata inspection (`describe`), trigger threshold verification, and resolution polling (`wait`). |
| **`zabbix-telemetry`** | Metric Extraction | Time-series telemetry querying (`zbxctl get metric --since=4h`), duration window filtering, and LLM token optimization (`-o toon`). |
| **`zabbix-safety`** | Guardrails & Middleware | Client-side safety execution guardrail levels (`readonly`, `readwrite-mine`, `readwrite-all`) and Exit Code 2 error envelopes. |

---

## 🤖 Supported AI Coding Assistants

`zbxctl` supports target skill installation across all major AI agent frameworks:

| Agent Framework | Target Path (`--global`) | Target Path (`--workspace`) | CLI Flag |
| :--- | :--- | :--- | :--- |
| **Universal (All Agents)** | All global locations below | All workspace locations below | `--agent=all` *(Default)* |
| **Antigravity / Gemini CLI** | `~/.gemini/config/skills/` | `.agents/skills/` | `--agent=antigravity` |
| **Claude Code (Anthropic)** | `~/.claude/skills/` | `.claude/skills/` | `--agent=claude` |
| **Cursor IDE** | `~/.cursor/rules/` | `.cursor/rules/` | `--agent=cursor` |
| **zbxctl Native** | `~/.config/zbxctl/skills/` | `.zbxctl/skills/` | `--agent=zbxctl` |

---

## 🛠️ Command Reference

### 1. List Available Skills
Display all available built-in skills and their descriptions:

```bash
zbxctl skill list
```

### 2. Install Skills Across AI Agents
Install a single skill or all built-in skills to your target AI assistant:

```bash
# Install all skills across ALL AI agent frameworks (Universal)
zbxctl skill install --all --agent=all

# Install all skills globally for Claude Code
zbxctl skill install --all --agent=claude --global

# Install zabbix-automation skill for Cursor IDE in local project
zbxctl skill install zabbix-automation --agent=cursor --workspace
```

### 3. Show Skill Content
Display the raw markdown content of a built-in skill:

```bash
zbxctl skill show zabbix-automation
```

### 4. Export Skill to Workspace Repository
Export a skill into `.agents/skills/<skill-name>/SKILL.md` for team Git repository sharing:

```bash
zbxctl skill export zabbix-troubleshooting
```

---

## 🤝 Team GitOps Recipe

Commit installed skills into your team's Git repository so every engineer and AI agent sharing the repo automatically inherits Zabbix 7 operational knowledge:

```bash
# Export skills to workspace directory
zbxctl skill export zabbix-automation
zbxctl skill export zabbix-telemetry

# Commit to Git
git add .agents/skills/
git commit -m "chore: add zbxctl AI agent skills for Zabbix 7 automation"
git push origin main
```
