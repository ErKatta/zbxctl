# AI Agent Integration Guide

`zbxctl` is designed to streamline how AI coding assistants (Gemini, Claude, Cursor, Aider, custom scripts) interact with Zabbix monitoring infrastructure.

---

## Why CLI over MCP for AI Agents?

1. **Low Context Overhead:** Traditional MCP tools inject large JSON schemas and RPC envelopes into prompt context. `zbxctl` provides low-token summaries (`--brief`) and Token-Optimized Object Notation (`-o toon`), saving thousands of context tokens.
2. **Universal Shell Tool Call:** AI agents already support executing terminal commands. Using `zbxctl` eliminates the need to configure and run background MCP server daemons for every Zabbix instance.
3. **Deterministic Safety:** Safety levels (`readonly`, `readwrite-mine`) are enforced in compiled binary code, preventing accidental mutations even if an LLM hallucinates destructive flags.

---

## Token-Optimization Flags

### 1. Compact Command Tree (`zbxctl commands --brief`)
Export a low-token summary of available verbs and flags:

```bash
zbxctl commands --brief
```

### 2. Token-Optimized Object Notation (`-o toon`)
Output compressed JSON structures without whitespace padding:

```bash
zbxctl get host -o toon
```

---

## Multi-Agent Skill Installation (`zbxctl skill`)

Equip your AI assistant (Antigravity, Claude Code, Cursor IDE) with Zabbix 7 domain knowledge and workflow recipes:

```bash
# Install built-in skills across ALL AI agent frameworks
zbxctl skill install --all --agent=all

# Install skills specifically for Claude Code
zbxctl skill install --all --agent=claude

# Export skills to local workspace repository for team Git sharing
zbxctl skill export zabbix-automation
```
