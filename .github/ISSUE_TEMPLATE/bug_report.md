---
name: Bug Report
about: Create a report to help us fix a bug or unexpected error in zbxctl
title: '[BUG] '
labels: 'bug'
assignees: ''
---

**Describe the Bug**
A clear and concise description of what the bug is.

**Command Executed**
```bash
zbxctl get host --filter='...'
```

**Expected Behavior**
A clear description of what you expected to happen.

**Actual Behavior / Error Envelope**
Include stdout/stderr output (credentials are automatically redacted):
```json
{
 "error": "..."
}
```

**Environment Info:**
- `zbxctl` version (`zbxctl --version`):
- OS version:
- Zabbix server version (e.g. 7.0.2 LTS):
- Active safety level (`readonly`, `readwrite-mine`, etc.):

**Additional Context**
Add any other context about the problem here.
