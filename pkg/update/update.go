package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultGitHubRepo = "ErKatta/zbxctl"
	DefaultTimeout    = 30 * time.Second
	SkillReminderMsg  = "To update AI agent skills to match the new version, run:\n  zbxctl skill install --all"
)

// Release represents a GitHub release.
type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Body        string    `json:"body"`
	Assets      []Asset   `json:"assets"`
}

// Asset represents an attached asset to a GitHub release.
type Asset struct {
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
	ContentType        string `json:"content_type"`
}

// UpdateResult represents the output metadata of an update check or execution.
type UpdateResult struct {
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version"`
	UpdateAvailable bool   `json:"update_available"`
	Updated         bool   `json:"updated"`
	BinaryPath      string `json:"binary_path,omitempty"`
	ReleaseURL      string `json:"release_url,omitempty"`
	Message         string `json:"message"`
	SkillReminder   string `json:"skill_reminder,omitempty"`
}

// Options configures the update operation.
type Options struct {
	CurrentVersion   string
	TargetVersion    string // specific release tag (e.g. "v0.2.0") or empty for latest
	Repo             string // GitHub repository (e.g. "ErKatta/zbxctl")
	TargetExecutable string // Target executable path (defaults to current binary)
	CheckOnly        bool   // Only check if update is available without downloading/installing
	Force            bool   // Force update even if already on latest version
	BaseAPIURL       string // Base GitHub API URL (defaults to https://api.github.com)
	HTTPClient       *http.Client
	Writer           io.Writer
}

// Updater handles checking and applying updates.
type Updater struct {
	opts Options
}

// NewUpdater returns a new Updater configured with the given options.
func NewUpdater(opts Options) *Updater {
	if opts.Repo == "" {
		opts.Repo = DefaultGitHubRepo
	}
	if opts.BaseAPIURL == "" {
		opts.BaseAPIURL = "https://api.github.com"
	}
	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{
			Timeout: DefaultTimeout,
		}
	}
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}
	return &Updater{opts: opts}
}

// FetchRelease retrieves the release metadata from GitHub.
func (u *Updater) FetchRelease(ctx context.Context) (*Release, error) {
	var endpoint string
	if u.opts.TargetVersion != "" {
		tag := u.opts.TargetVersion
		if !strings.HasPrefix(tag, "v") && !strings.Contains(tag, ".") {
			tag = "v" + tag
		}
		endpoint = fmt.Sprintf("%s/repos/%s/releases/tags/%s", u.opts.BaseAPIURL, u.opts.Repo, tag)
	} else {
		endpoint = fmt.Sprintf("%s/repos/%s/releases/latest", u.opts.BaseAPIURL, u.opts.Repo)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create release request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	userAgent := "zbxctl"
	if u.opts.CurrentVersion != "" {
		userAgent = fmt.Sprintf("zbxctl/%s", u.opts.CurrentVersion)
	}
	req.Header.Set("User-Agent", userAgent)

	// Use GITHUB_TOKEN or GH_TOKEN if present to prevent rate limiting
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := u.opts.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("release not found on repository %s (status 404)", u.opts.Repo)
	}
	if resp.StatusCode == http.StatusForbidden {
		bodyBytes, _ := io.ReadAll(resp.Body)
		if strings.Contains(string(bodyBytes), "rate limit") {
			return nil, fmt.Errorf("GitHub API rate limit exceeded. Set GITHUB_TOKEN or try again later")
		}
		return nil, fmt.Errorf("GitHub API request forbidden (status 403): %s", string(bodyBytes))
	}
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch release (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("failed to parse release metadata: %w", err)
	}

	return &rel, nil
}

// Semver represents parsed semantic version components.
type Semver struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Raw        string
}

// ParseSemver parses a version string into Semver components.
func ParseSemver(v string) Semver {
	raw := strings.TrimSpace(v)
	s := strings.TrimPrefix(raw, "v")
	s = strings.TrimPrefix(s, "V")

	if s == "" || s == "dev" || s == "none" || s == "unknown" {
		return Semver{Major: 0, Minor: 0, Patch: 0, Prerelease: "dev", Raw: raw}
	}

	prerelease := ""
	if idx := strings.IndexAny(s, "-+"); idx != -1 {
		prerelease = s[idx+1:]
		s = s[:idx]
	}

	parts := strings.Split(s, ".")
	var major, minor, patch int
	if len(parts) > 0 {
		major, _ = strconv.Atoi(parts[0])
	}
	if len(parts) > 1 {
		minor, _ = strconv.Atoi(parts[1])
	}
	if len(parts) > 2 {
		patch, _ = strconv.Atoi(parts[2])
	}

	return Semver{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: prerelease,
		Raw:        raw,
	}
}

// CompareVersions compares two version strings.
// Returns:
// -1 if v1 < v2
//
//	0 if v1 == v2
//	1 if v1 > v2
func CompareVersions(v1, v2 string) int {
	s1 := ParseSemver(v1)
	s2 := ParseSemver(v2)

	if s1.Major != s2.Major {
		if s1.Major < s2.Major {
			return -1
		}
		return 1
	}
	if s1.Minor != s2.Minor {
		if s1.Minor < s2.Minor {
			return -1
		}
		return 1
	}
	if s1.Patch != s2.Patch {
		if s1.Patch < s2.Patch {
			return -1
		}
		return 1
	}

	// If main version numbers match, check prerelease:
	// Standard semver rule: a version with prerelease has lower precedence than normal release
	if s1.Prerelease == "" && s2.Prerelease != "" {
		return 1
	}
	if s1.Prerelease != "" && s2.Prerelease == "" {
		return -1
	}
	if s1.Prerelease != s2.Prerelease {
		if s1.Prerelease < s2.Prerelease {
			return -1
		}
		return 1
	}

	return 0
}

// FindMatchingAsset finds the binary archive asset matching the target OS and architecture.
func FindMatchingAsset(assets []Asset, goos, goarch string) (*Asset, error) {
	osNorm := strings.ToLower(goos)
	archNorm := strings.ToLower(goarch)

	var expectedExt string
	if osNorm == "windows" {
		expectedExt = ".zip"
	} else {
		expectedExt = ".tar.gz"
	}

	var candidates []Asset
	for _, a := range assets {
		name := strings.ToLower(a.Name)
		if strings.Contains(name, "checksums") || strings.HasSuffix(name, ".txt") || strings.HasSuffix(name, ".sig") {
			continue
		}

		// Match OS
		osMatch := false
		if osNorm == "darwin" && (strings.Contains(name, "darwin") || strings.Contains(name, "macos") || strings.Contains(name, "osx")) {
			osMatch = true
		} else if osNorm == "linux" && strings.Contains(name, "linux") {
			osMatch = true
		} else if osNorm == "windows" && strings.Contains(name, "windows") {
			osMatch = true
		}

		if !osMatch {
			continue
		}

		// Match Arch
		archMatch := false
		if (archNorm == "amd64" || archNorm == "x86_64") && (strings.Contains(name, "amd64") || strings.Contains(name, "x86_64")) {
			archMatch = true
		} else if (archNorm == "arm64" || archNorm == "aarch64") && (strings.Contains(name, "arm64") || strings.Contains(name, "aarch64")) {
			archMatch = true
		} else if (archNorm == "386" || archNorm == "i386") && (strings.Contains(name, "386") || strings.Contains(name, "i386")) {
			archMatch = true
		}

		if !archMatch {
			continue
		}

		if strings.HasSuffix(name, expectedExt) || (osNorm != "windows" && strings.HasSuffix(name, ".tgz")) {
			candidates = append(candidates, a)
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no matching release asset found for platform %s/%s", goos, goarch)
	}

	return &candidates[0], nil
}

// FindChecksumAsset returns the checksums file asset if available in the release.
func FindChecksumAsset(assets []Asset) *Asset {
	for _, a := range assets {
		name := strings.ToLower(a.Name)
		if strings.Contains(name, "checksums") || strings.HasSuffix(name, "sha256sums.txt") {
			return &a
		}
	}
	return nil
}

// DownloadAsset downloads the bytes of the given asset.
func (u *Updater) DownloadAsset(ctx context.Context, asset *Asset) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.BrowserDownloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}
	req.Header.Set("User-Agent", "zbxctl-updater")

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := u.opts.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read downloaded asset: %w", err)
	}

	return data, nil
}

// VerifyChecksum verifies the SHA256 checksum of data against the checksums file.
func VerifyChecksum(data []byte, assetName string, checksumsData []byte) error {
	hasher := sha256.New()
	hasher.Write(data)
	actualHash := hex.EncodeToString(hasher.Sum(nil))

	lines := strings.Split(string(checksumsData), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			expectedHash := strings.ToLower(fields[0])
			fileName := filepath.Base(strings.TrimPrefix(fields[1], "*"))
			if strings.EqualFold(fileName, assetName) {
				if actualHash != expectedHash {
					return fmt.Errorf("checksum mismatch for %s: expected %s, calculated %s", assetName, expectedHash, actualHash)
				}
				return nil
			}
		}
	}

	return fmt.Errorf("checksum entry for %s not found in checksums file", assetName)
}

// ExtractBinary extracts the zbxctl binary from a tar.gz or zip archive.
func ExtractBinary(archiveData []byte, assetName string) ([]byte, error) {
	nameLower := strings.ToLower(assetName)
	if strings.HasSuffix(nameLower, ".zip") {
		return extractBinaryFromZip(archiveData)
	}
	return extractBinaryFromTarGz(archiveData)
}

func extractBinaryFromTarGz(archiveData []byte) ([]byte, error) {
	gzr, err := gzip.NewReader(bytes.NewReader(archiveData))
	if err != nil {
		return nil, fmt.Errorf("failed to decompress gzip archive: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed reading tar archive: %w", err)
		}

		baseName := filepath.Base(hdr.Name)
		if baseName == "zbxctl" || baseName == "zbxctl.exe" {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("failed extracting binary from archive: %w", err)
			}
			return data, nil
		}
	}

	return nil, fmt.Errorf("binary 'zbxctl' not found inside release archive")
}

func extractBinaryFromZip(archiveData []byte) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(archiveData), int64(len(archiveData)))
	if err != nil {
		return nil, fmt.Errorf("failed to open zip archive: %w", err)
	}

	for _, f := range zr.File {
		baseName := filepath.Base(f.Name)
		if baseName == "zbxctl" || baseName == "zbxctl.exe" {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("failed opening binary inside zip: %w", err)
			}
			data, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("failed reading binary inside zip: %w", err)
			}
			return data, nil
		}
	}

	return nil, fmt.Errorf("binary 'zbxctl' not found inside zip archive")
}

// GetCurrentExecutablePath determines the absolute, symlink-resolved path of the running executable.
func GetCurrentExecutablePath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to locate current executable: %w", err)
	}
	realPath, err := filepath.EvalSymlinks(execPath)
	if err == nil && realPath != "" {
		return realPath, nil
	}
	return filepath.Abs(execPath)
}

// ApplyUpdate writes the binary atomically in place of the target executable path.
func ApplyUpdate(binaryData []byte, targetPath string) error {
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("invalid executable path: %w", err)
	}

	// Resolve symlink if possible
	if realPath, err := filepath.EvalSymlinks(absPath); err == nil && realPath != "" {
		absPath = realPath
	}

	dir := filepath.Dir(absPath)

	// Create temp file in the same directory to guarantee atomic rename across filesystems
	tempFile, err := os.CreateTemp(dir, ".zbxctl-update-*.tmp")
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied writing to %s: please rerun with appropriate privileges (e.g. sudo zbxctl update)", dir)
		}
		return fmt.Errorf("failed to create temporary file in %s: %w", dir, err)
	}
	tempPath := tempFile.Name()

	// Ensure cleanup in case of error
	cleanedUp := false
	defer func() {
		if !cleanedUp {
			_ = os.Remove(tempPath)
		}
	}()

	if _, err := tempFile.Write(binaryData); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to write new binary to temporary file: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary binary file: %w", err)
	}

	// Set executable permissions
	if err := os.Chmod(tempPath, 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}

	// Windows atomic replacement handling
	if runtime.GOOS == "windows" {
		backupPath := absPath + ".old"
		_ = os.Remove(backupPath)
		if err := os.Rename(absPath, backupPath); err == nil {
			if err := os.Rename(tempPath, absPath); err != nil {
				// Rollback
				_ = os.Rename(backupPath, absPath)
				return fmt.Errorf("failed to replace binary on Windows: %w", err)
			}
			cleanedUp = true
			_ = os.Remove(backupPath) // Best effort
			return nil
		}
	}

	// Unix atomic rename
	if err := os.Rename(tempPath, absPath); err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied replacing %s: please rerun with appropriate privileges (e.g. sudo zbxctl update)", absPath)
		}
		return fmt.Errorf("failed to replace binary at %s: %w", absPath, err)
	}

	cleanedUp = true
	return nil
}

// Run executes the complete update check and installation flow.
func (u *Updater) Run(ctx context.Context) (*UpdateResult, error) {
	targetExec := u.opts.TargetExecutable
	if targetExec == "" {
		var err error
		targetExec, err = GetCurrentExecutablePath()
		if err != nil {
			return nil, err
		}
	}

	currentVer := u.opts.CurrentVersion
	if currentVer == "" {
		currentVer = "0.0.0-dev"
	}

	rel, err := u.FetchRelease(ctx)
	if err != nil {
		return nil, err
	}

	latestVer := rel.TagName
	cmp := CompareVersions(currentVer, latestVer)
	updateAvailable := cmp < 0

	res := &UpdateResult{
		CurrentVersion:  currentVer,
		LatestVersion:   latestVer,
		UpdateAvailable: updateAvailable,
		Updated:         false,
		BinaryPath:      targetExec,
		ReleaseURL:      rel.HTMLURL,
		SkillReminder:   SkillReminderMsg,
	}

	if u.opts.CheckOnly {
		if updateAvailable {
			res.Message = fmt.Sprintf("A new version is available: %s (current: %s). Run 'zbxctl update' to install.", latestVer, currentVer)
		} else if cmp == 0 {
			res.Message = fmt.Sprintf("zbxctl is up to date (%s).", currentVer)
		} else {
			res.Message = fmt.Sprintf("zbxctl is up to date (current: %s, latest release: %s).", currentVer, latestVer)
		}
		return res, nil
	}

	if !updateAvailable && !u.opts.Force {
		if cmp == 0 {
			res.Message = fmt.Sprintf("zbxctl is already up to date (%s).", currentVer)
		} else {
			res.Message = fmt.Sprintf("zbxctl is already up to date (current: %s, latest release: %s). Use --force to reinstall.", currentVer, latestVer)
		}
		return res, nil
	}

	// Find matching binary asset
	asset, err := FindMatchingAsset(rel.Assets, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return nil, err
	}

	// Download archive
	archiveData, err := u.DownloadAsset(ctx, asset)
	if err != nil {
		return nil, fmt.Errorf("failed to download release asset %s: %w", asset.Name, err)
	}

	// Verify checksum if available
	if checksumAsset := FindChecksumAsset(rel.Assets); checksumAsset != nil {
		checksumData, err := u.DownloadAsset(ctx, checksumAsset)
		if err == nil {
			if err := VerifyChecksum(archiveData, asset.Name, checksumData); err != nil {
				return nil, fmt.Errorf("security check failed: %w", err)
			}
		}
	}

	// Extract binary
	binaryData, err := ExtractBinary(archiveData, asset.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to extract binary: %w", err)
	}

	// Apply update in-place
	if err := ApplyUpdate(binaryData, targetExec); err != nil {
		return nil, err
	}

	res.Updated = true
	res.Message = fmt.Sprintf("Successfully updated zbxctl to %s at %s!", latestVer, targetExec)
	return res, nil
}
