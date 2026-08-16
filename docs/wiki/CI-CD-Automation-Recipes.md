# CI/CD Automation Recipes

`zbxctl` seamlessly embeds into standard CI/CD runners, deployment scripts, and automated testing pipelines.

---

## GitHub Actions Example

```yaml
name: Apply Zabbix Monitoring Manifests

on:
 push:
 branches: [ main ]

jobs:
 zabbix-apply:
 runs-on: ubuntu-latest
 steps:
 - name: Checkout Code
 uses: actions/checkout@v4

 - name: Install zbxctl
 run: go install github.com/ErKatta/zbxctl@latest

 - name: Login to Zabbix
 run: |
 zbxctl login ${{ secrets.ZABBIX_URL }} \
 --token="${{ secrets.ZABBIX_API_TOKEN }}" \
 --name="ci-prod" \
 --safety-level="readwrite-all"

 - name: Run Dry-Run Diff
 run: zbxctl diff -f monitoring/host-manifest.yaml

 - name: Apply Monitoring Configuration
 run: zbxctl apply -f monitoring/host-manifest.yaml
```

---

## Bash Script Recipe: Auto-Acknowledge Resolved Problems

```bash
#!/usr/bin/env bash
set -euo pipefail

# Fetch high severity problems in JSON
PROBLEMS=$(zbxctl get problem --filter='{"severity": 4}' -o json)

# Process problem IDs with jq
echo "$PROBLEMS" | jq -r '.[] | .eventid' | while read -r EVENTID; do
 echo "Inspecting problem $EVENTID..."
 # Wait up to 30 seconds for resolution
 zbxctl wait problem "$EVENTID" --for=resolved --timeout=30s || true
done
```

---

## Telemetry Recipe: Export Host Items & 4-Hour Metrics

```bash
#!/usr/bin/env bash
set -euo pipefail

# Retrieve items for Zabbix server filtered by ID, name, and description
HOST_ITEMS=$(zbxctl get item --host="Zabbix server" --fields=itemid,name,description -o json)
echo "Retrieved items for Zabbix server:"
echo "$HOST_ITEMS" | jq .

# Fetch historical metric telemetry for the last 4 hours
zbxctl get metric 23253 --since=4h -o json > cpu_telemetry_4h.json
```
