# Security Policy

## Security Commitments

1. **Automatic Credential Redaction:** API Tokens, passwords, and HTTP headers are sanitized to `[REDACTED]` before printing outputs to `stdout`, logs, or LLM prompt contexts.
2. **Client-Side Safety Middleware:** All mutations pass through local execution checks before hitting the Zabbix API.
3. **Zero Secrets Persistence in History:** Interactive login commands store tokens in local environment or encrypted config files rather than command history logs.

---

## Reporting a Vulnerability

If you discover a security vulnerability within `zbxctl`, please report it privately:

1. **Email:** Send details to `erkatta@gmail.com` or open a private security advisory on GitHub.
2. **Details to Include:**
 - Description of the vulnerability.
 - Steps to reproduce or proof-of-concept payload.
 - Affected `zbxctl` command or flags.
 - Impact assessment.

Please **do not** report security vulnerabilities via public GitHub issues. Thank you for keeping `zbxctl` and the open-source community safe!
