/* ==========================================================================
 zbxctl Interactive Logic & Data
 ========================================================================== */

const commandsData = [
 {
 name: "zbxctl config current-config",
 category: "config",
 desc: "Display active context configuration with automatic [REDACTED] credential masking.",
 example: "zbxctl config current-config -o json"
 },
 {
 name: "zbxctl get host",
 category: "tier1",
 desc: "Fetch Zabbix hosts with filtering and custom output formats.",
 example: "zbxctl get host web-prod-01 -o json"
 },
 {
 name: "zbxctl get problem",
 category: "tier1",
 desc: "List active problems filtered by severity or resource tags.",
 example: "zbxctl get problem --filter='{\"severity\": 4}'"
 },
 {
 name: "zbxctl describe host",
 category: "tier1",
 desc: "Inspect detailed extended configuration metadata for a single host.",
 example: "zbxctl describe host 10001"
 },
  {
  name: "zbxctl skill install --all",
  category: "tier1",
  desc: "Install built-in AI agent skills (automation, troubleshooting, telemetry, safety) into agent config.",
  example: "zbxctl skill install --all"
  },
 {
 name: "zbxctl login --context",
 category: "config",
 desc: "Authenticate against Zabbix API and automatically save & activate specified context name.",
 example: "zbxctl login http://zabbix.local:8080/ -u admin --context=prod-cluster"
 },
 {
 "name": "zbxctl get metric --since",
 category: "tier1",
 desc: "Fetch historical metric telemetry samples for a duration window (e.g. last 4h, 30m).",
 example: "zbxctl get metric 23253 --since=4h -o table"
 },
 {
 name: "zbxctl get item --host --fields",
 category: "tier1",
 desc: "Retrieve all items for a target host ID or name filtered by specific fields.",
 example: "zbxctl get item --host='Zabbix server' --fields=itemid,name,description"
 },
 {
 name: "zbxctl query item",
 category: "tier1",
 desc: "Advanced search and metric key filtering across items.",
 example: "zbxctl query item --search='key_=system.cpu' --limit=5"
 },
 {
 name: "zbxctl apply -f manifest.yaml",
 category: "tier1",
 desc: "Declaratively create or update Zabbix resources from YAML/JSON manifests.",
 example: "zbxctl apply -f host-manifest.yaml"
 },
 {
 name: "zbxctl diff -f manifest.yaml",
 category: "tier1",
 desc: "Compare local manifest definition against live Zabbix server state.",
 example: "zbxctl diff -f host.yaml --id=10001"
 },
 {
 name: "zbxctl delete host",
 category: "tier1",
 desc: "Delete host resources by ID (requires non-readonly safety level).",
 example: "zbxctl delete host 10001 10002 --force"
 },
 {
 name: "zbxctl exec",
 category: "tier1",
 desc: "Execute a registered Zabbix script directly on a target host.",
 example: "zbxctl exec 1 --hostid=10001"
 },
 {
 name: "zbxctl wait problem",
 category: "tier1",
 desc: "Poll and wait until a specific problem condition resolves.",
 example: "zbxctl wait problem 12345 --for=resolved --timeout=60s"
 },
 {
 name: "zbxctl doctor",
 category: "agent",
 desc: "Diagnose API connectivity, token validity, and context safety level.",
 example: "zbxctl doctor"
 },
 {
 name: "zbxctl inventory",
 category: "agent",
 desc: "Probe target Zabbix instance grounding statistics (hosts, problems, items).",
 example: "zbxctl inventory"
 },
 {
 name: "zbxctl commands --brief",
 category: "agent",
 desc: "Export compact text tree of all verbs, resources, flags for LLMs.",
 example: "zbxctl commands --brief"
 },
 {
 name: "zbxctl raw <zabbix.method>",
 category: "tier2",
 desc: "Invoke ANY Zabbix 7 JSON-RPC API endpoint with safety middleware filtering.",
 example: "zbxctl raw proxygroup.get --params='{\"output\":\"extend\"}'"
 },
 {
 name: "zbxctl config get-contexts",
 category: "config",
 desc: "List all configured contexts and highlight current active context.",
 example: "zbxctl config get-contexts"
 },
 {
 name: "zbxctl config use-context",
 category: "config",
 desc: "Switch the active context in ~/.zbxctl/config.yaml.",
 example: "zbxctl config use-context prod-us"
 }
];

const safetyMatrixData = {
 readonly: {
 title: "readonly",
 subtitle: "Default & Safest Mode for Autonomous LLM Agents",
 allowed: ["get", "describe", "query", "doctor", "inventory", "commands", "*.get"],
 blocked: ["*.create", "*.update", "*.delete", "history.push", "exec"],
 sampleError: `{
 "error": {
 "code": "SAFETY_LEVEL_VIOLATION",
 "method": "host.delete",
 "message": "Operation blocked by safety-level 'readonly' on context 'prod'.",
 "resolution": "Switch context or update safety-level in ~/.zbxctl/config.yaml."
 }
}`
 },
 "readwrite-mine": {
 title: "readwrite-mine",
 subtitle: "Scoped Mutations for Managed Infrastructure",
 allowed: ["All read operations", "Mutations on resources tagged zbxctl=true or managed-by=zbxctl"],
 blocked: ["Un-tagged infrastructure", "Instance-wide destructive edits"],
 sampleError: `{
 "error": {
 "code": "SAFETY_LEVEL_VIOLATION",
 "method": "host.update",
 "message": "Target resource is not tagged 'zbxctl=true' under safety-level 'readwrite-mine'.",
 "resolution": "Add 'zbxctl=true' tag to resource or upgrade safety level."
 }
}`
 },
 "readwrite-all": {
 title: "readwrite-all",
 subtitle: "Full Administrative Mutations (Guardrails Preserved)",
 allowed: ["All Zabbix API operations", "Single resource mutations", "Bulk actions with --force"],
 blocked: ["Bulk deletes without explicit --force flag"],
 sampleError: `{
 "error": {
 "code": "SAFETY_LEVEL_VIOLATION",
 "method": "host.delete",
 "message": "Bulk deletion of multiple hosts requires explicit '--force' flag.",
 "resolution": "Pass '--force' flag to confirm destructive bulk action."
 }
}`
 },
 "dangerously-unrestricted": {
 title: "dangerously-unrestricted",
 subtitle: "Zero Guardrails - Expert Systems Only",
 allowed: ["All 100% Zabbix API calls without safety checks"],
 blocked: ["None"],
 sampleError: `// No safety checks applied. All commands execute directly.`
 }
};

const formatPreviews = {
 json: `[
 {
 "hostid": "10084",
 "host": "web-prod-01",
 "name": "Production Web Node 01",
 "status": "0",
 "safety_level": "readonly"
 }
]`,
 table: `+--------+-------------+-----------------------+--------+
| HOSTID | HOST | NAME | STATUS |
+--------+-------------+-----------------------+--------+
| 10084 | web-prod-01 | Production Web Node 1 | 0 (UP) |
+--------+-------------+-----------------------+--------+`,
 toon: `[{"hostid":"10084","host":"web-prod-01","name":"Production Web Node 01","status":"0"}]`,
 yaml: `- hostid: "10084"
 host: "web-prod-01"
 name: "Production Web Node 01"
 status: "0"`
};

let activeCategory = 'all';

document.addEventListener('DOMContentLoaded', () => {
  const yearEl = document.getElementById('current-year');
  if (yearEl) yearEl.textContent = new Date().getFullYear();

  renderCommands();
  selectSafety('readonly', document.querySelector('.safety-option[data-level="readonly"]'));
  setFormat('json', document.querySelector('.format-tab'));
  fetchLatestRelease();
});

function renderCommands() {
 const container = document.getElementById('command-list');
 const searchVal = document.getElementById('command-search').value.toLowerCase();

 const filtered = commandsData.filter(cmd => {
 const matchesCat = activeCategory === 'all' || cmd.category === activeCategory;
 const matchesSearch = cmd.name.toLowerCase().includes(searchVal) || cmd.desc.toLowerCase().includes(searchVal);
 return matchesCat && matchesSearch;
 });

 if (filtered.length === 0) {
 container.innerHTML = `<div style="grid-column: 1/-1; text-align: center; color: var(--text-dim); padding: 2rem;">No matching commands found.</div>`;
 return;
 }

 container.innerHTML = filtered.map(cmd => `
 <div class="cmd-item">
 <div>
 <div class="cmd-item-header">
 <span class="cmd-name">${cmd.name}</span>
 <span class="cmd-tag">${cmd.category.toUpperCase()}</span>
 </div>
 <p class="cmd-desc">${cmd.desc}</p>
 </div>
 <div class="cmd-example">
 <code>${cmd.example}</code>
 <button class="copy-btn" onclick="copyText('${cmd.example}', this)">Copy</button>
 </div>
 </div>
 `).join('');
}

function filterCommands() {
 renderCommands();
}

function setCategory(cat, btn) {
 activeCategory = cat;
 document.querySelectorAll('.cat-btn').forEach(b => b.classList.remove('active'));
 btn.classList.add('active');
 renderCommands();
}

function selectSafety(level, element) {
 document.querySelectorAll('.safety-option').forEach(el => el.classList.remove('active'));
 if (element) element.classList.add('active');

 const data = safetyMatrixData[level];
 const container = document.getElementById('safety-details');

 container.innerHTML = `
 <div style="margin-bottom: 1.5rem;">
 <h3 style="font-family: var(--font-heading); font-size: 1.4rem; color: #0F172A; margin-bottom: 0.25rem;">
 Safety Level: <span style="color: var(--zabbix-blue);">${data.title}</span>
 </h3>
 <p style="color: var(--text-muted); font-size: 0.95rem;">${data.subtitle}</p>
 </div>

 <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1.5rem; margin-bottom: 1.5rem;">
 <div style="background: #ECFDF5; border: 1px solid #A7F3D0; padding: 1rem; border-radius: 8px;">
 <h4 style="color: #065F46; font-size: 0.9rem; margin-bottom: 0.5rem;">Permitted Operations</h4>
 <ul style="padding-left: 1.2rem; color: #1F2937; font-size: 0.88rem;">
 ${data.allowed.map(item => `<li><code>${item}</code></li>`).join('')}
 </ul>
 </div>
 <div style="background: #FEF2F2; border: 1px solid #FECACA; padding: 1rem; border-radius: 8px;">
 <h4 style="color: #991B1B; font-size: 0.9rem; margin-bottom: 0.5rem;">Blocked Operations</h4>
 <ul style="padding-left: 1.2rem; color: #1F2937; font-size: 0.88rem;">
 ${data.blocked.map(item => `<li><code>${item}</code></li>`).join('')}
 </ul>
 </div>
 </div>

 <div>
 <h4 style="font-size: 0.85rem; color: var(--text-dim); margin-bottom: 0.5rem;">Sample Exit Code 2 Violation Envelope</h4>
 <pre style="background: #1E222D; border: 1px solid #374151; padding: 1rem; border-radius: 8px; font-family: var(--font-mono); font-size: 0.82rem; color: #F87171;"><code>${data.sampleError}</code></pre>
 </div>
 `;
}

function setFormat(fmt, btn) {
 document.querySelectorAll('.format-tab').forEach(b => b.classList.remove('active'));
 if (btn) btn.classList.add('active');

 const preview = document.getElementById('format-preview');
 preview.textContent = formatPreviews[fmt] || formatPreviews.json;
}

function copyText(text, btn) {
  navigator.clipboard.writeText(text).then(() => {
    const originalText = btn.textContent;
    btn.textContent = 'Copied!';
    btn.style.background = 'var(--zabbix-blue)';
    btn.style.color = '#FFF';
    setTimeout(() => {
      btn.textContent = originalText;
      btn.style.background = '';
      btn.style.color = '';
    }, 1800);
  });
}

function copyElementText(elementId, btn) {
  const el = document.getElementById(elementId);
  if (el) {
    copyText(el.textContent, btn);
  }
}

async function fetchLatestRelease() {
  try {
    const res = await fetch('https://api.github.com/repos/ErKatta/zbxctl/releases/latest');
    if (!res.ok) return;
    const release = await res.json();
    const tag = release.tag_name || 'latest';

    // Update navbar version badge
    const badge = document.getElementById('version-badge');
    if (badge) badge.textContent = tag;

    // Process assets
    if (Array.isArray(release.assets)) {
      release.assets.forEach(asset => {
        const name = (asset.name || '').toLowerCase();
        const url = asset.browser_download_url;
        if (!url) return;

        if (name.includes('linux') && (name.includes('amd64') || name.includes('x86_64'))) {
          const el = document.getElementById('dl-linux-amd64');
          if (el) el.href = url;
          const cmd = document.getElementById('cmd-linux');
          if (cmd) cmd.textContent = `curl -sL ${url} | tar -xz && sudo mv zbxctl /usr/local/bin/`;
        } else if (name.includes('linux') && (name.includes('arm64') || name.includes('aarch64'))) {
          const el = document.getElementById('dl-linux-arm64');
          if (el) el.href = url;
        } else if (name.includes('darwin') && (name.includes('arm64') || name.includes('aarch64'))) {
          const el = document.getElementById('dl-darwin-arm64');
          if (el) el.href = url;
          const cmd = document.getElementById('cmd-darwin');
          if (cmd) cmd.textContent = `curl -sL ${url} | tar -xz && sudo mv zbxctl /usr/local/bin/`;
        } else if (name.includes('darwin') && (name.includes('amd64') || name.includes('x86_64'))) {
          const el = document.getElementById('dl-darwin-amd64');
          if (el) el.href = url;
        } else if (name.includes('windows') && (name.includes('amd64') || name.includes('x86_64'))) {
          const el = document.getElementById('dl-windows-amd64');
          if (el) el.href = url;
          const cmd = document.getElementById('cmd-windows');
          if (cmd) cmd.textContent = `Invoke-WebRequest -Uri "${url}" -OutFile "zbxctl.zip"; Expand-Archive zbxctl.zip -DestinationPath "$HOME\\bin"`;
        }
      });
    }
  } catch (err) {
    console.debug('GitHub release fetch skipped or failed:', err);
  }
}
