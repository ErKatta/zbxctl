# Safety Guardrails Matrix

`zbxctl` includes compiled client-side safety middleware to protect Zabbix instances during rapid scripting or AI pair programming.

---

## Safety Levels

| Level | Allowed Operations | Blocked Operations | Typical Use Case |
| :--- | :--- | :--- | :--- |
| `readonly` | Read-only verbs (`get`, `describe`, `query`, `doctor`, `cluster-info`, `commands`, `*.get`) | Any mutation (`*.create`, `*.update`, `*.delete`, `exec`) | Default mode for AI agents & production inspection |
| `readwrite-mine` | Read-only + mutations on resources tagged `zbxctl=true` | Un-tagged or instance-wide infrastructure edits | Managed dev/test resources |
| `readwrite-all` | Full Zabbix API mutations | Bulk deletes without explicit `--force` | Admin operations & platform automation |
| `dangerously-unrestricted` | All API calls enabled without safety checks | None | Expert automated pipelines |

---

## Exit Code 2 Error Envelopes

When a command violates the context's `safety_level`, `zbxctl` halts execution locally **before** sending an HTTP request, and returns **Exit Code 2**:

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
