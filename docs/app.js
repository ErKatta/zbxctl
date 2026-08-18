/* ==========================================================================
   zbxctl Interactive Logic, Mobile Navigation & Accessible UI
   ========================================================================== */
'use strict';

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
    name: "zbxctl get metric --since",
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
    name: "zbxctl edit host",
    category: "tier1",
    desc: "Fetch and edit a live resource directly in your text editor, applying changes upon save (kubectl edit style).",
    example: "zbxctl edit host web-prod-01"
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
    name: "zbxctl version",
    category: "agent",
    desc: "Print the zbxctl CLI version, git commit, build date, and platform architecture.",
    example: "zbxctl version"
  },
  {
    name: "zbxctl cluster-info",
    category: "agent",
    desc: "Display Zabbix instance connection, version, and sizing statistics (hosts, problems, items, triggers).",
    example: "zbxctl cluster-info"
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
    allowed: ["get", "describe", "query", "doctor", "cluster-info", "commands", "*.get"],
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
| HOSTID | HOST        | NAME                  | STATUS |
+--------+-------------+-----------------------+--------+
| 10084  | web-prod-01 | Production Web Node 1 | 0 (UP) |
+--------+-------------+-----------------------+--------+`,
  toon: `[{"hostid":"10084","host":"web-prod-01","name":"Production Web Node 01","status":"0"}]`,
  yaml: `- hostid: "10084"
  host: "web-prod-01"
  name: "Production Web Node 01"
  status: "0"`
};

const agentScenariosData = {
  incident: {
    id: "incident",
    title: "Incident Triage & Root-Cause Diagnosis",
    badge: "Skills: troubleshooting + telemetry",
    userPrompt: "Alert firing on prod-api-01: high response latency and CPU exhaustion in the last 2 hours. Investigate root cause and check recent item metrics.",
    thought1: "Analyzing active problem triggers on host 'prod-api-01' using token-optimized TOON output...",
    command1: "zbxctl get problem --host=prod-api-01 -o toon",
    output1: `[{"eventid":"20419","host":"prod-api-01","problem":"CPU utilization > 95%","severity":4,"age":"38m","ack":"No"},{"eventid":"20422","host":"prod-api-01","problem":"Worker pool thread starvation (>128 threads)","severity":3,"age":"32m","ack":"No"}]`,
    thought2: "Isolating worker thread telemetry and CPU queue metrics over the 2-hour window...",
    command2: "zbxctl get metric 48201 --since=2h --step=60s -o toon",
    output2: `{"itemid":"48201","key":"app.threads.active","samples":[{"t":"13:45","v":32},{"t":"14:10","v":128},{"t":"14:30","v":144},{"t":"15:00","v":140}]}`,
    agentResponse: "Root Cause Identified: Worker thread pool on prod-api-01 reached max saturation (144 active threads vs 128 max limit) starting at 14:10 UTC, cascading into 96.4% CPU usage and latency degradation.\n\nRecommended Fix: Scale worker pool limit to 256 or enable request shedding in /etc/app/config.yaml.",
    tokenSavings: "TOON output: 148 tokens vs Raw JSON: 1,640 tokens (91% LLM token reduction)"
  },
  gitops: {
    id: "gitops",
    title: "GitOps Declarative Spec Synchronization",
    badge: "Skill: zabbix-automation",
    userPrompt: "Review drift on our Redis cluster monitoring manifest redis-cluster.yaml and apply updates safely.",
    thought1: "Performing dry-run differential analysis against live Zabbix 7 server state...",
    command1: "zbxctl diff -f manifests/redis-cluster.yaml",
    output1: `--- live:redis-cluster-01
+++ manifest:redis-cluster.yaml
@@ -14,3 +14,5 @@
   macros:
-    {$REDIS_MAX_MEMORY}: "8GB"
+    {$REDIS_MAX_MEMORY}: "16GB"
   tags:
+    env: "production"
+    team: "data-infra"`,
    thought2: "Drift verified: 1 macro updated and 2 tags added. Applying changes under 'readwrite-mine' safety mode...",
    command2: "zbxctl apply -f manifests/redis-cluster.yaml --safety-level=readwrite-mine",
    output2: `[OK] manifests/redis-cluster.yaml applied successfully:
  - redis-cluster-01: 1 macro updated, 2 tags added (hostid: 10842)
  - redis-cluster-02: 1 macro updated, 2 tags added (hostid: 10843)
Summary: 2 resources synced, 0 errors.`,
    agentResponse: "Declarative synchronization complete: 2 Redis hosts updated with new 16GB memory threshold macros and production metadata tags. No unexpected resource drift detected.",
    tokenSavings: "Unified diff format: 112 tokens vs full JSON schema: 1,320 tokens (92% LLM token reduction)"
  },
  safety: {
    id: "safety",
    title: "Client-Side Safety Intercept (Guardrail Protection)",
    badge: "Skill: zabbix-safety",
    userPrompt: "Clean up and delete all unused staging host groups in our Zabbix instance.",
    thought1: "Executing host group deletion in current context 'prod' (safety-level: readonly)...",
    command1: "zbxctl delete hostgroup --all-unused",
    output1: `[EXIT CODE 2 - SAFETY_LEVEL_VIOLATION]
{
  "error": {
    "code": "SAFETY_LEVEL_VIOLATION",
    "method": "hostgroup.delete",
    "level": "readonly",
    "message": "Destructive bulk operation blocked client-side by zbxctl safety middleware.",
    "resolution": "To execute destructive mutations, switch context or use '--safety-level=readwrite-all --force'."
  }
}`,
    thought2: "The client-side safety guardrail intercepted this destructive action before any network packet was sent to the Zabbix API.",
    command2: "zbxctl config get-contexts",
    output2: `CURRENT   NAME       URL                              SAFETY_LEVEL   FORMAT
*         prod       https://zabbix.prod.company.io   readonly       json
          staging    https://zabbix.stg.company.io    readwrite-all  json`,
    agentResponse: "Safety Guardrail Engaged: zbxctl blocked the delete request locally because the active context 'prod' is in 'readonly' mode.\n\nTo proceed safely, switch to the staging context (zbxctl config use-context staging) or confirm with explicit administrative privileges.",
    tokenSavings: "Deterministic Exit Code 2 & Error Envelope: Prevents accidental destructive API calls"
  }
};

let activeCategory = 'all';

document.addEventListener('DOMContentLoaded', () => {
  const yearEl = document.getElementById('current-year');
  if (yearEl) yearEl.textContent = new Date().getFullYear();

  initNavigation();
  initBackToTop();
  initTablistKeyboardNavigation();
  renderCommands();
  selectSafety('readonly', document.getElementById('tab-readonly'));
  setFormat('json', document.getElementById('fmt-json'));
  selectScenario('incident', document.getElementById('tab-sim-incident'));
  fetchLatestRelease();
});

/* ==========================================================================
   Mobile Navigation Menu Management
   ========================================================================== */

function initNavigation() {
  const navToggle = document.getElementById('nav-toggle');
  const navMenu = document.getElementById('nav-menu');
  const navLinks = document.querySelectorAll('.nav-link, .nav-mobile-btn');

  if (!navToggle || !navMenu) return;

  function toggleMenu(open) {
    const shouldOpen = open !== undefined ? open : !navMenu.classList.contains('nav-open');
    navToggle.setAttribute('aria-expanded', shouldOpen ? 'true' : 'false');
    if (shouldOpen) {
      navMenu.classList.add('nav-open');
    } else {
      navMenu.classList.remove('nav-open');
    }
  }

  navToggle.addEventListener('click', (e) => {
    e.stopPropagation();
    toggleMenu();
  });

  // Close menu when clicking on any nav link
  navLinks.forEach(link => {
    link.addEventListener('click', () => {
      toggleMenu(false);
    });
  });

  // Close menu when clicking outside
  document.addEventListener('click', (e) => {
    if (!navMenu.contains(e.target) && !navToggle.contains(e.target)) {
      toggleMenu(false);
    }
  });

  // Close menu on Escape key press (WCAG 2.1.2)
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && navMenu.classList.contains('nav-open')) {
      toggleMenu(false);
      navToggle.focus();
    }
  });

  // Reset menu on resize to desktop width
  window.addEventListener('resize', () => {
    if (window.innerWidth > 1024 && navMenu.classList.contains('nav-open')) {
      toggleMenu(false);
    }
  });
}

/* ==========================================================================
   Screen Reader Live Announcements (WCAG 4.1.3)
   ========================================================================== */

function announceToScreenReader(message) {
  const announcer = document.getElementById('a11y-announcer');
  if (announcer) {
    announcer.textContent = '';
    setTimeout(() => {
      announcer.textContent = message;
    }, 50);
  }
}

/* ==========================================================================
   Interactive Command Playground
   ========================================================================== */

function renderCommands() {
  const container = document.getElementById('command-list');
  const searchInput = document.getElementById('command-search');
  const searchStatus = document.getElementById('search-status');
  if (!container || !searchInput) return;

  const searchVal = searchInput.value.trim().toLowerCase();

  const filtered = commandsData.filter(cmd => {
    const matchesCat = activeCategory === 'all' || cmd.category === activeCategory;
    const matchesSearch = cmd.name.toLowerCase().includes(searchVal) || cmd.desc.toLowerCase().includes(searchVal);
    return matchesCat && matchesSearch;
  });

  if (searchStatus) {
    if (searchVal || activeCategory !== 'all') {
      searchStatus.textContent = `Showing ${filtered.length} matching commands`;
    } else {
      searchStatus.textContent = '';
    }
  }

  if (filtered.length === 0) {
    container.innerHTML = `<div style="grid-column: 1/-1; text-align: center; color: var(--text-dim); padding: 2rem;">No matching commands found.</div>`;
    return;
  }

  container.innerHTML = filtered.map((cmd, idx) => `
    <div class="cmd-item">
      <div>
        <div class="cmd-item-header">
          <span class="cmd-name">${escapeHtml(cmd.name)}</span>
          <span class="cmd-tag">${escapeHtml(cmd.category.toUpperCase())}</span>
        </div>
        <p class="cmd-desc">${escapeHtml(cmd.desc)}</p>
      </div>
      <div class="cmd-example">
        <code id="cmd-sample-${idx}">${escapeHtml(cmd.example)}</code>
        <button type="button" class="copy-btn" aria-label="Copy ${escapeHtml(cmd.name)} command" onclick="copyElementText('cmd-sample-${idx}', this)">Copy</button>
      </div>
    </div>
  `).join('');
}

function filterCommands() {
  renderCommands();
}

function setCategory(cat, btn) {
  activeCategory = cat;
  document.querySelectorAll('.cat-btn').forEach(b => {
    b.classList.remove('active');
    b.setAttribute('aria-selected', 'false');
  });
  if (btn) {
    btn.classList.add('active');
    btn.setAttribute('aria-selected', 'true');
  }
  renderCommands();
}

/* ==========================================================================
   Safety Level Selector
   ========================================================================== */

function selectSafety(level, btn) {
  document.querySelectorAll('.safety-option').forEach(b => {
    b.classList.remove('active');
    b.setAttribute('aria-selected', 'false');
  });
  if (btn) {
    btn.classList.add('active');
    btn.setAttribute('aria-selected', 'true');
  }

  // Sync tabpanel aria-labelledby with the active tab (WAI-ARIA best practice)
  const panel = document.getElementById('safety-details');
  if (panel && btn) {
    panel.setAttribute('aria-labelledby', btn.id);
  }

  const data = safetyMatrixData[level];
  const container = document.getElementById('safety-details');
  if (!data || !container) return;

  const badgeClass = level === 'readonly' ? 'level-readonly' :
                     level === 'readwrite-mine' ? 'level-mine' :
                     level === 'readwrite-all' ? 'level-all' : 'level-danger';

  container.innerHTML = `
    <div class="safety-details-header">
      <div class="safety-level-title-row">
        <span class="safety-badge ${badgeClass}">${escapeHtml(data.title)}</span>
        <span class="safety-subtitle">${escapeHtml(data.subtitle)}</span>
      </div>
    </div>

    <div class="safety-columns">
      <div class="safety-box safety-box-allowed">
        <h4 class="safety-box-title"><span class="safety-icon-check" aria-hidden="true">✓</span> Permitted Operations</h4>
        <ul class="safety-ops-list">
          ${data.allowed.map(item => `<li><code>${escapeHtml(item)}</code></li>`).join('')}
        </ul>
      </div>
      <div class="safety-box safety-box-blocked">
        <h4 class="safety-box-title"><span class="safety-icon-cross" aria-hidden="true">✕</span> Blocked Operations</h4>
        <ul class="safety-ops-list">
          ${data.blocked.map(item => `<li><code>${escapeHtml(item)}</code></li>`).join('')}
        </ul>
      </div>
    </div>

    <div class="safety-error-section">
      <div class="safety-error-header">
        <span>EXIT CODE 2 &bull; SAFETY VIOLATION ENVELOPE</span>
        <button type="button" class="copy-btn mini-copy-btn" aria-label="Copy error envelope" onclick="copyText('${escapeHtml(data.sampleError).replace(/'/g, "\\'")}', this)">Copy</button>
      </div>
      <pre tabindex="0" class="safety-error-pre" aria-label="Violation error envelope"><code>${escapeHtml(data.sampleError)}</code></pre>
    </div>
  `;
}

/* ==========================================================================
   Output Format Tabs
   ========================================================================== */

function setFormat(fmt, btn) {
  document.querySelectorAll('.format-tab').forEach(b => {
    b.classList.remove('active');
    b.setAttribute('aria-selected', 'false');
  });
  if (btn) {
    btn.classList.add('active');
    btn.setAttribute('aria-selected', 'true');
  }

  // Sync tabpanel aria-labelledby with the active tab (WAI-ARIA best practice)
  const formatPanel = document.getElementById('format-preview-panel');
  if (formatPanel && btn) {
    formatPanel.setAttribute('aria-labelledby', btn.id);
  }

  const preview = document.getElementById('format-preview');
  if (preview) {
    preview.innerHTML = `<code>${escapeHtml(formatPreviews[fmt] || formatPreviews.json)}</code>`;
  }
}

/* ==========================================================================
   AI Agent Simulation Scenario Switcher
   ========================================================================== */

function selectScenario(scenarioKey, btn) {
  document.querySelectorAll('.sim-tab').forEach(b => {
    b.classList.remove('active');
    b.setAttribute('aria-selected', 'false');
  });
  if (btn) {
    btn.classList.add('active');
    btn.setAttribute('aria-selected', 'true');
  }

  const simPanel = document.getElementById('agent-sim-panel');
  if (simPanel && btn) {
    simPanel.setAttribute('aria-labelledby', btn.id);
  }

  const data = agentScenariosData[scenarioKey] || agentScenariosData.incident;
  const container = document.getElementById('agent-chat-body');
  const tokenBadge = document.getElementById('sim-token-savings');

  if (tokenBadge) {
    tokenBadge.textContent = data.tokenSavings;
  }

  if (container) {
    container.innerHTML = `
      <!-- User Turn -->
      <div class="chat-row user-turn">
        <div class="chat-author-bar">
          <span class="chat-avatar-badge avatar-user" aria-hidden="true">DEV</span>
          <span class="chat-author-name">Developer / SRE</span>
        </div>
        <div class="chat-bubble-user">${escapeHtml(data.userPrompt)}</div>
      </div>

      <!-- Agent Turn -->
      <div class="chat-row agent-turn">
        <div class="chat-author-bar">
          <span class="chat-avatar-badge avatar-agent" aria-hidden="true">AI</span>
          <span class="chat-author-name">AI Coding Assistant</span>
          <span class="skill-badge">${escapeHtml(data.badge)}</span>
        </div>

        <div class="chat-thought-box">
          <strong>Thought:</strong> ${escapeHtml(data.thought1)}
        </div>

        <!-- Tool Call 1 -->
        <div class="chat-cli-block" role="region" aria-label="Shell tool execution 1">
          <div class="chat-cli-header">
            <span class="tool-call-label"><span class="tool-call-dot"></span> TOOL CALL &bull; SHELL / BASH</span>
            <button type="button" class="copy-btn mini-copy-btn" aria-label="Copy tool command 1" onclick="copyText('${escapeHtml(data.command1)}', this)">Copy</button>
          </div>
          <div class="chat-cli-body">
            <div class="chat-cli-cmd"><span class="prompt">$</span> ${escapeHtml(data.command1)}</div>
            <pre class="chat-cli-output"><code>${escapeHtml(data.output1)}</code></pre>
          </div>
        </div>

        ${data.thought2 ? `
        <div class="chat-thought-box">
          <strong>Thought:</strong> ${escapeHtml(data.thought2)}
        </div>
        ` : ''}

        ${data.command2 ? `
        <!-- Tool Call 2 -->
        <div class="chat-cli-block" role="region" aria-label="Shell tool execution 2">
          <div class="chat-cli-header">
            <span class="tool-call-label"><span class="tool-call-dot"></span> TOOL CALL &bull; SHELL / BASH</span>
            <button type="button" class="copy-btn mini-copy-btn" aria-label="Copy tool command 2" onclick="copyText('${escapeHtml(data.command2)}', this)">Copy</button>
          </div>
          <div class="chat-cli-body">
            <div class="chat-cli-cmd"><span class="prompt">$</span> ${escapeHtml(data.command2)}</div>
            <pre class="chat-cli-output"><code>${escapeHtml(data.output2)}</code></pre>
          </div>
        </div>
        ` : ''}

        <div class="chat-bubble-agent">
          <div class="agent-response-text">${escapeHtml(data.agentResponse).replace(/\n\n/g, '<br><br>')}</div>
        </div>
      </div>
    `;
  }

  announceToScreenReader(`Switched to scenario: ${data.title}`);
}

/* ==========================================================================
   Keyboard Tablist Navigation (WCAG 2.1.1)
   ========================================================================== */

function initTablistKeyboardNavigation() {
  const tablists = document.querySelectorAll('[role="tablist"]');
  tablists.forEach(tablist => {
    const tabs = Array.from(tablist.querySelectorAll('[role="tab"]'));
    tablist.addEventListener('keydown', (e) => {
      const activeElement = document.activeElement;
      const currentIndex = tabs.indexOf(activeElement);
      if (currentIndex === -1) return;

      let nextIndex = -1;
      if (e.key === 'ArrowRight' || e.key === 'ArrowDown') {
        e.preventDefault();
        nextIndex = (currentIndex + 1) % tabs.length;
      } else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') {
        e.preventDefault();
        nextIndex = (currentIndex - 1 + tabs.length) % tabs.length;
      } else if (e.key === 'Home') {
        e.preventDefault();
        nextIndex = 0;
      } else if (e.key === 'End') {
        e.preventDefault();
        nextIndex = tabs.length - 1;
      }

      if (nextIndex !== -1) {
        tabs[nextIndex].focus();
        tabs[nextIndex].click();
      }
    });
  });
}

/* ==========================================================================
   Copy to Clipboard Functionality with Feedback
   ========================================================================== */

function copyText(text, btn) {
  if (!navigator.clipboard) {
    fallbackCopyText(text, btn);
    return;
  }
  navigator.clipboard.writeText(text).then(() => {
    showCopyFeedback(btn);
  }).catch(() => {
    fallbackCopyText(text, btn);
  });
}

function copyElementText(elementId, btn) {
  const el = document.getElementById(elementId);
  if (el) {
    copyText(el.textContent.trim(), btn);
  }
}

function fallbackCopyText(text, btn) {
  const textArea = document.createElement('textarea');
  textArea.value = text;
  textArea.style.position = 'fixed';
  textArea.style.opacity = '0';
  document.body.appendChild(textArea);
  textArea.focus();
  textArea.select();
  try {
    document.execCommand('copy');
    showCopyFeedback(btn);
  } catch (err) {
    console.error('Fallback copy failed:', err);
  }
  document.body.removeChild(textArea);
}

function showCopyFeedback(btn) {
  if (!btn) return;
  const originalText = btn.textContent;
  btn.textContent = 'Copied!';
  btn.style.background = 'var(--zabbix-blue)';
  btn.style.color = '#FFFFFF';
  announceToScreenReader('Copied to clipboard');

  setTimeout(() => {
    btn.textContent = originalText;
    btn.style.background = '';
    btn.style.color = '';
  }, 1800);
}

function escapeHtml(str) {
  if (typeof str !== 'string') return '';
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

/* ==========================================================================
   GitHub Release Fetching
   ========================================================================== */

async function fetchLatestRelease() {
  if (typeof fetch !== 'function') return;
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

/* ==========================================================================
   Floating Action Button (Back to Top)
   ========================================================================== */

function scrollToTop() {
  const prefersReduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  window.scrollTo({
    top: 0,
    behavior: prefersReduced ? 'auto' : 'smooth'
  });

  const mainContent = document.getElementById('main-content');
  if (mainContent) {
    mainContent.focus({ preventScroll: true });
  }
}

function initBackToTop() {
  const backToTopBtn = document.getElementById('btn-back-to-top');
  if (!backToTopBtn) return;

  const updateVisibility = () => {
    const scrollTop = window.pageYOffset || document.documentElement.scrollTop || document.body.scrollTop || window.scrollY || 0;
    if (scrollTop > 200) {
      backToTopBtn.classList.add('visible');
    } else {
      backToTopBtn.classList.remove('visible');
    }
  };

  let ticking = false;
  const onScroll = () => {
    if (!ticking) {
      window.requestAnimationFrame(() => {
        updateVisibility();
        ticking = false;
      });
      ticking = true;
    }
  };

  window.addEventListener('scroll', onScroll, { passive: true });
  window.addEventListener('touchmove', onScroll, { passive: true });
  document.addEventListener('scroll', onScroll, { passive: true });

  // Initial check
  updateVisibility();
}
