# Contributing to `zbxctl`

Thank you for your interest in contributing to `zbxctl`! We welcome contributions from developers, DevOps engineers, and AI enthusiasts of all experience levels.

`zbxctl` is an open-source command-line tool for Zabbix 7.0+ LTS inspired by `kubectl` and `dtctl`. Our mission is to accelerate automation workflow development for engineers and AI agents while maintaining deterministic client-side safety guardrails.

---

## Getting Started

### Prerequisites

- **Go 1.22+**: Make sure Go is installed (`go version`).
- **Zabbix 7.0+ LTS** instance (or standard Docker Zabbix container for testing).
- **Git**: For version control.

### Setting Up Your Development Environment

1. **Fork and Clone the Repository:**
 ```bash
 git clone https://github.com/YOUR-USERNAME/zbxctl.git
 cd zbxctl
 ```

2. **Verify Building from Source:**
 ```bash
 go build -o zbxctl main.go
 ./zbxctl --version
 ```

3. **Run Unit & Integration Tests:**
 ```bash
 go test -v ./...
 ```

---

## Codebase Architecture

`zbxctl` relies on a clean, decoupled Go package layout:

- **`main.go`**: Entry point delegating command parsing to `cmd/`.
- **`cmd/`**: Cobra command registry containing Tier 1 ergonomic verbs (`get`, `describe`, `apply`, `delete`, `query`, `exec`, `wait`, `doctor`, `cluster-info`) and Tier 2 raw JSON-RPC caller (`raw`).
- **`pkg/zabbix/`**: Zabbix JSON-RPC HTTP client driver supporting API Tokens, Basic Auth, User/Password, and mTLS.
- **`pkg/safety/`**: Client-side safety middleware enforcing execution boundaries (`readonly`, `readwrite-mine`, `readwrite-all`, `dangerously-unrestricted`).
- **`pkg/formatter/`**: Multi-format output renderers (JSON, Terminal Table, TOON, YAML) and automatic credential redaction middleware.

---

## Testing Guidelines

Before opening a pull request, ensure all tests pass cleanly:

```bash
# Run standard tests with race detection
go test -race -v ./...

# Run code format verification
gofmt -s -w .

# Run static analysis
go vet ./...
```

---

## Software Lifecycle Management (SLCM)

`zbxctl` enforces a structured software development and release lifecycle:

### 1. Branching Strategy
* **`main`**: Production-ready branch. Releases are tagged exclusively from `main`. Direct pushes are blocked.
* **`dev`**: Active integration branch for upcoming features and sprint milestones.
* **`feature/<name>`** / **`fix/<name>`**: Topic branches branched from `dev` or `main`.

### 2. Continuous Integration (CI) Matrix
Every pull request is automatically tested across:
* **Operating Systems**: Linux (`ubuntu-latest`), macOS (`macos-latest`), Windows (`windows-latest`).
* **Go Versions**: Go `1.22`, Go `1.23`.
* **Checks**: `go vet`, `gofmt` verification, and POSIX race detection (`go test -race`).

### 3. Release & Distribution Lifecycle
* Releases follow **Semantic Versioning 2.0.0** (`vMAJOR.MINOR.PATCH`).
* When a release tag (`v*`) is pushed to `main`, GitHub Actions triggers **GoReleaser** to cross-compile standalone static binaries (`CGO_ENABLED=0`) for Linux (`amd64`/`arm64`), macOS (`amd64`/`arm64`), and Windows (`amd64`).
* Checksums (`checksums.txt`) and `.tar.gz`/`.zip` archives are automatically attached to the GitHub Release.
* Detailed guide: [Software Lifecycle Management (SLCM) Process](docs/wiki/Software-Lifecycle-Management.md).

---

## Commit & PR Guidelines

We follow clear git commit conventions:

- Use imperative present tense in commit messages (e.g. `add host wait command timeout flag` instead of `added host wait command`).
- Keep commit titles under 60 characters.
- Reference relevant issue numbers when applicable (`Fixes #42`).

### Pull Request Process

1. Create a feature branch off `dev` (or `main`):
   ```bash
   git checkout -b feature/my-new-verb
   ```
2. Write clean code and accompanying unit tests.
3. Ensure safety middleware checks are preserved for any mutating actions.
4. Verify tests pass locally:
   ```bash
   go test -race -v ./...
   gofmt -s -w .
   go vet ./...
   ```
5. Push your branch and open a PR.
6. Fill out the provided Pull Request template completely.

---

## License

By contributing to `zbxctl`, you agree that your contributions will be licensed under the project's [Apache License 2.0](LICENSE).
