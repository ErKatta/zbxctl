# Software Lifecycle Management (SLCM) Process

This document describes the end-to-end **Software Lifecycle Management (SLCM)** process for `zbxctl`. It defines the standards, workflows, quality guardrails, and automation pipelines governing code from initial design to production release.

---

## 1. Lifecycle Overview & Architecture

```
+-------------------+      +--------------------+      +--------------------+
| Local Development | ---> | Pull Request & CI  | ---> | Integration (dev)  |
| - gofmt / go vet  |      | - Linux / Mac / Win|      | - Nightly & Stage  |
| - go test -race   |      | - Race Detector    |      +--------------------+
+-------------------+      +--------------------+                 |
                                                                  v
+-------------------+      +--------------------+      +--------------------+
| GitHub Pages Docs | <--- | Release Tag (v*)   | <--- | Production (main)  |
| - Dynamic Version |      | - GoReleaser       |      | - Protected Branch |
| - Direct Binaries |      | - Cross-OS Archive |      +--------------------+
+-------------------+      +--------------------+
```

---

## 2. Branching Strategy & Git Workflow

`zbxctl` follows a structured Git workflow designed for stability and rapid iteration:

| Branch / Pattern | Purpose | Protection Rules |
| :--- | :--- | :--- |
| `main` | Production-ready branch. Releases are tagged exclusively from `main`. | Protected: Requires passing CI, branch up-to-date, and code review. |
| `dev` | Active integration branch for upcoming releases. | Protected: Direct pushes disallowed; PR merges only. |
| `feature/<name>` | New CLI verbs, flags, or major capabilities. | Branched off `dev` (or `main`). |
| `fix/<name>` | Bug fixes and schema corrections. | Branched off `dev` or `main`. |
| `docs/<name>` | Documentation, Wiki, and GitHub Pages updates. | Branched off `main`. |
| `chore/<name>` | Dependency updates, Go toolchain bumps, CI adjustments. | Branched off `dev` or `main`. |

---

## 3. Local Development Lifecycle

Before opening a pull request or requesting code review, all contributors must execute local verification:

### 3.1 Formatting & Static Analysis
```bash
# Format code to standard Go specifications
gofmt -s -w .

# Run static analysis
go vet ./...
```

### 3.2 Automated Test Execution
Run the full test suite with race condition detection:
```bash
# Execute unit and mock tests with race detection
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
```

### 3.3 Static Binary Build
Verify that the binary builds with zero external C dependencies (`CGO_ENABLED=0`):
```bash
CGO_ENABLED=0 go build -o zbxctl main.go
./zbxctl version
```

---

## 4. Continuous Integration (CI) Pipeline

Every pull request and push to `main` or `dev` automatically triggers the **CI Pipeline** ([`.github/workflows/ci.yml`](../../.github/workflows/ci.yml)).

### 4.1 Multi-Platform Test Matrix
The CI pipeline validates all code across a cross-platform operating system matrix:

```yaml
strategy:
  matrix:
    os: [ubuntu-latest, macos-latest, windows-latest]
    go-version: ['1.22', '1.23']
```

### 4.2 Automated Checks Executed in CI
1. **Module Integrity**: Verifies `go.mod` and `go.sum` consistency via `go vet` and `git diff --exit-code`.
2. **Race Detection**: Runs `go test -race` on POSIX runners (Ubuntu & macOS).
3. **Windows Native Testing**: Executes tests on Windows native runners without POSIX emulation.
4. **Compilation Verification**: Verifies compilation of the CLI entrypoint on all target runners.

---

## 5. Quality & Safety Guardrails Protocol

Because `zbxctl` executes operations against production monitoring infrastructure, every code modification must adhere to strict safety protocols:

1. **Client-Side Safety Middleware Integrity**:
   * Any mutating command (`create`, `update`, `delete`, `apply`, `exec`) must pass through `pkg/safety.CheckSafety()`.
   * Safety levels (`readonly`, `readwrite-mine`, `readwrite-all`, `dangerously-unrestricted`) must never be bypassed in client logic.
   * Safety violations must return standard **Exit Code 2** structured error envelopes.
2. **Credential Redaction**:
   * Tokens, passwords, and authorization headers must remain masked in terminal, JSON, YAML, and TOON outputs.
3. **Deterministic Output Schemas**:
   * Table output must follow the deterministic column schemas defined in `pkg/output/schema.go`.
   * Field validation (`ValidateFields`) must halt execution if an unknown column name is supplied to `--fields`.

---

## 6. Release & Versioning Lifecycle

Releases follow [Semantic Versioning 2.0.0](https://semver.org/) (`vMAJOR.MINOR.PATCH`):
* **MAJOR**: Incompatible CLI breaking changes or paradigm shifts.
* **MINOR**: New ergonomic verbs, resources, flags, or backward-compatible feature additions.
* **PATCH**: Bug fixes, schema corrections, and performance improvements.

### 6.1 Release Tagging Procedure
Releases are triggered by pushing a signed Git tag to `main`:

```bash
# 1. Ensure main is clean and up to date
git checkout main
git pull origin main

# 2. Create an annotated semantic tag
git tag -a v0.1.0 -m "Release v0.1.0: Add curated schemas, field validation, and relative age formatting"

# 3. Push the tag to trigger GitHub Actions
git push origin v0.1.0
```

### 6.2 Automated Release Pipeline ([`.github/workflows/release.yml`](../../.github/workflows/release.yml))
The release workflow activates on `push: tags: ['v*']` and executes **GoReleaser** ([`.goreleaser.yaml`](../../.goreleaser.yaml)):

1. **Cross-Compilation Matrix (`CGO_ENABLED=0`)**:
   * **Linux**: `amd64` (x86_64), `arm64` (aarch64)
   * **macOS**: `amd64` (Intel), `arm64` (Apple Silicon M1-M4)
   * **Windows**: `amd64` (x86_64)
2. **Packaging & Checksums**:
   * Generates `.tar.gz` archives for Linux and macOS.
   * Generates `.zip` archives for Windows.
   * Generates SHA256 checksums (`checksums.txt`).
3. **GitHub Release Publication**:
   * Automatically generates changelogs excluding internal maintenance commits.
   * Publishes release notes and attaches binary assets to the **GitHub Releases** page.

---

## 7. Documentation & GitHub Pages Deployment

The documentation website (`docs/`) is continuously deployed:

1. **Automated Deployment ([`.github/workflows/docs-gh-pages.yml`](../../.github/workflows/docs-gh-pages.yml))**:
   * Pushes to `main` automatically publish the `docs/` directory to **GitHub Pages**.
2. **Dynamic Client-Side Release Binding ([`docs/app.js`](../../docs/app.js))**:
   * The documentation site queries the GitHub API (`/releases/latest`) on page load.
   * Version badges, download URLs, and copyable install snippets for Linux, macOS, and Windows update dynamically without requiring manual documentation edits.

---

## 8. AI Agent Skill Maintenance Lifecycle

`zbxctl` packages built-in AI agent skills (`pkg/skill/builtin.go` & `SKILL.md`):

1. **Synchronization**:
   * Whenever new CLI verbs, flags, or schema definitions are added, update the built-in skill prompt definitions in `pkg/skill/builtin.go` and `SKILL.md`.
2. **Multi-Agent Verification**:
   * Verify skill distribution with `zbxctl skill install --all --agent=all` across Antigravity, Claude Code, and Cursor IDE configurations.
