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
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"0.1.0", "v0.2.0", -1},
		{"0.2.0-dev", "v0.2.0", -1},
		{"v0.2.0", "0.2.0", 0},
		{"0.2.0", "v0.1.0", 1},
		{"1.0.0", "0.9.9", 1},
		{"0.3.0", "0.2.0", 1},
		{"0.2.0-alpha", "0.2.0-beta", -1},
		{"0.2.0", "0.2.0-rc1", 1},
		{"0.2.0-rc1", "0.2.0", -1},
		{"none", "0.1.0", -1},
		{"dev", "0.1.0", -1},
		{"unknown", "0.1.0", -1},
		{"v1.2.3", "v1.2.3", 0},
		{"1.2.4", "1.2.3", 1},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_vs_%s", tt.v1, tt.v2), func(t *testing.T) {
			res := CompareVersions(tt.v1, tt.v2)
			if res != tt.expected {
				t.Errorf("CompareVersions(%q, %q) = %d; expected %d", tt.v1, tt.v2, res, tt.expected)
			}
		})
	}
}

func TestFindMatchingAsset(t *testing.T) {
	assets := []Asset{
		{Name: "zbxctl_0.2.0_darwin_amd64.tar.gz", BrowserDownloadURL: "https://example.com/darwin_amd64.tar.gz"},
		{Name: "zbxctl_0.2.0_darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/darwin_arm64.tar.gz"},
		{Name: "zbxctl_0.2.0_linux_amd64.tar.gz", BrowserDownloadURL: "https://example.com/linux_amd64.tar.gz"},
		{Name: "zbxctl_0.2.0_linux_arm64.tar.gz", BrowserDownloadURL: "https://example.com/linux_arm64.tar.gz"},
		{Name: "zbxctl_0.2.0_windows_amd64.zip", BrowserDownloadURL: "https://example.com/windows_amd64.zip"},
		{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums.txt"},
	}

	t.Run("linux amd64", func(t *testing.T) {
		a, err := FindMatchingAsset(assets, "linux", "amd64")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a.Name != "zbxctl_0.2.0_linux_amd64.tar.gz" {
			t.Errorf("expected linux amd64 asset, got %s", a.Name)
		}
	})

	t.Run("linux arm64", func(t *testing.T) {
		a, err := FindMatchingAsset(assets, "linux", "arm64")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a.Name != "zbxctl_0.2.0_linux_arm64.tar.gz" {
			t.Errorf("expected linux arm64 asset, got %s", a.Name)
		}
	})

	t.Run("darwin arm64", func(t *testing.T) {
		a, err := FindMatchingAsset(assets, "darwin", "arm64")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a.Name != "zbxctl_0.2.0_darwin_arm64.tar.gz" {
			t.Errorf("expected darwin arm64 asset, got %s", a.Name)
		}
	})

	t.Run("windows amd64", func(t *testing.T) {
		a, err := FindMatchingAsset(assets, "windows", "amd64")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a.Name != "zbxctl_0.2.0_windows_amd64.zip" {
			t.Errorf("expected windows amd64 asset, got %s", a.Name)
		}
	})

	t.Run("unsupported platform", func(t *testing.T) {
		_, err := FindMatchingAsset(assets, "freebsd", "riscv64")
		if err == nil {
			t.Errorf("expected error for unsupported platform, got nil")
		}
	})

	t.Run("find checksum asset", func(t *testing.T) {
		chk := FindChecksumAsset(assets)
		if chk == nil || chk.Name != "checksums.txt" {
			t.Errorf("expected checksums.txt asset, got %v", chk)
		}
	})
}

func createTestTarGz(t *testing.T, binaryName, content string) []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Add a dummy README
	readmeContent := []byte("readme content")
	if err := tw.WriteHeader(&tar.Header{
		Name: "README.md",
		Mode: 0644,
		Size: int64(len(readmeContent)),
	}); err != nil {
		t.Fatalf("failed to write tar header: %v", err)
	}
	if _, err := tw.Write(readmeContent); err != nil {
		t.Fatalf("failed to write readme content: %v", err)
	}

	// Add binary
	binContent := []byte(content)
	if err := tw.WriteHeader(&tar.Header{
		Name: binaryName,
		Mode: 0755,
		Size: int64(len(binContent)),
	}); err != nil {
		t.Fatalf("failed to write tar header: %v", err)
	}
	if _, err := tw.Write(binContent); err != nil {
		t.Fatalf("failed to write binary content: %v", err)
	}

	if err := tw.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}

	return buf.Bytes()
}

func createTestZip(t *testing.T, binaryName, content string) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	binWriter, err := zw.Create(binaryName)
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	if _, err := binWriter.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write zip entry content: %v", err)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}

	return buf.Bytes()
}

func TestExtractBinary(t *testing.T) {
	expectedContent := "#!/bin/sh\necho updated-zbxctl\n"

	t.Run("tar.gz extraction", func(t *testing.T) {
		archive := createTestTarGz(t, "zbxctl", expectedContent)
		extracted, err := ExtractBinary(archive, "zbxctl_0.2.0_linux_amd64.tar.gz")
		if err != nil {
			t.Fatalf("failed to extract tar.gz: %v", err)
		}
		if string(extracted) != expectedContent {
			t.Errorf("extracted content mismatch: got %q, expected %q", string(extracted), expectedContent)
		}
	})

	t.Run("zip extraction", func(t *testing.T) {
		archive := createTestZip(t, "zbxctl.exe", expectedContent)
		extracted, err := ExtractBinary(archive, "zbxctl_0.2.0_windows_amd64.zip")
		if err != nil {
			t.Fatalf("failed to extract zip: %v", err)
		}
		if string(extracted) != expectedContent {
			t.Errorf("extracted content mismatch: got %q, expected %q", string(extracted), expectedContent)
		}
	})
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("mock binary archive payload")
	hasher := sha256.New()
	hasher.Write(data)
	hashHex := hex.EncodeToString(hasher.Sum(nil))

	checksumsContent := fmt.Sprintf("%s  zbxctl_0.2.0_linux_amd64.tar.gz\n%s  zbxctl_0.2.0_darwin_arm64.tar.gz\n", hashHex, hashHex)

	t.Run("valid checksum", func(t *testing.T) {
		err := VerifyChecksum(data, "zbxctl_0.2.0_linux_amd64.tar.gz", []byte(checksumsContent))
		if err != nil {
			t.Errorf("expected checksum verification to pass, got: %v", err)
		}
	})

	t.Run("mismatch checksum", func(t *testing.T) {
		err := VerifyChecksum([]byte("corrupted data"), "zbxctl_0.2.0_linux_amd64.tar.gz", []byte(checksumsContent))
		if err == nil {
			t.Errorf("expected checksum mismatch error, got nil")
		}
	})

	t.Run("missing entry in checksums file", func(t *testing.T) {
		err := VerifyChecksum(data, "zbxctl_0.2.0_unknown.tar.gz", []byte(checksumsContent))
		if err == nil {
			t.Errorf("expected missing entry error, got nil")
		}
	})
}

func TestApplyUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	origBinPath := filepath.Join(tmpDir, "zbxctl")
	if err := os.WriteFile(origBinPath, []byte("old-binary-version-1"), 0755); err != nil {
		t.Fatalf("failed to write original binary: %v", err)
	}

	newBinaryData := []byte("new-binary-version-2")
	if err := ApplyUpdate(newBinaryData, origBinPath); err != nil {
		t.Fatalf("failed to apply update: %v", err)
	}

	readBack, err := os.ReadFile(origBinPath)
	if err != nil {
		t.Fatalf("failed to read updated binary: %v", err)
	}
	if string(readBack) != string(newBinaryData) {
		t.Errorf("binary content mismatch: got %q, expected %q", string(readBack), string(newBinaryData))
	}

	fi, err := os.Stat(origBinPath)
	if err != nil {
		t.Fatalf("failed to stat updated binary: %v", err)
	}
	if runtime.GOOS != "windows" && (fi.Mode()&0111 == 0) {
		t.Errorf("expected updated binary to have executable permissions, mode: %v", fi.Mode())
	}
}

func TestUpdaterEndToEnd(t *testing.T) {
	tmpDir := t.TempDir()
	targetBinPath := filepath.Join(tmpDir, "zbxctl")
	if err := os.WriteFile(targetBinPath, []byte("original-v0.1.0-binary"), 0755); err != nil {
		t.Fatalf("failed to create target binary: %v", err)
	}

	newBinaryPayload := "new-v0.3.0-binary-content"
	var archiveBytes []byte
	var assetName string
	if runtime.GOOS == "windows" {
		assetName = fmt.Sprintf("zbxctl_0.3.0_%s_%s.zip", runtime.GOOS, runtime.GOARCH)
		archiveBytes = createTestZip(t, "zbxctl.exe", newBinaryPayload)
	} else {
		assetName = fmt.Sprintf("zbxctl_0.3.0_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
		archiveBytes = createTestTarGz(t, "zbxctl", newBinaryPayload)
	}

	hasher := sha256.New()
	hasher.Write(archiveBytes)
	archiveHash := hex.EncodeToString(hasher.Sum(nil))
	checksumsData := []byte(fmt.Sprintf("%s  %s\n", archiveHash, assetName))

	// Mock server
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/releases/latest"):
			rel := Release{
				TagName:     "v0.3.0",
				Name:        "zbxctl v0.3.0 Release",
				PublishedAt: testingTime(),
				HTMLURL:     "https://github.com/ErKatta/zbxctl/releases/tag/v0.3.0",
				Assets: []Asset{
					{
						Name:               assetName,
						Size:               int64(len(archiveBytes)),
						BrowserDownloadURL: server.URL + "/download/" + assetName,
					},
					{
						Name:               "checksums.txt",
						Size:               int64(len(checksumsData)),
						BrowserDownloadURL: server.URL + "/download/checksums.txt",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(rel)
		case strings.HasSuffix(r.URL.Path, "/download/"+assetName):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(archiveBytes)
		case strings.HasSuffix(r.URL.Path, "/download/checksums.txt"):
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(checksumsData)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Run("check only when update is available", func(t *testing.T) {
		updater := NewUpdater(Options{
			CurrentVersion:   "0.2.0",
			Repo:             "ErKatta/zbxctl",
			TargetExecutable: targetBinPath,
			CheckOnly:        true,
			BaseAPIURL:       server.URL,
			HTTPClient:       server.Client(),
		})

		res, err := updater.Run(context.Background())
		if err != nil {
			t.Fatalf("updater.Run failed: %v", err)
		}

		if !res.UpdateAvailable {
			t.Errorf("expected update_available to be true")
		}
		if res.Updated {
			t.Errorf("expected updated to be false in check-only mode")
		}
		if !strings.Contains(res.Message, "A new version is available: v0.3.0") {
			t.Errorf("unexpected check message: %s", res.Message)
		}
	})

	t.Run("check when already up to date", func(t *testing.T) {
		updater := NewUpdater(Options{
			CurrentVersion:   "v0.3.0",
			Repo:             "ErKatta/zbxctl",
			TargetExecutable: targetBinPath,
			CheckOnly:        true,
			BaseAPIURL:       server.URL,
			HTTPClient:       server.Client(),
		})

		res, err := updater.Run(context.Background())
		if err != nil {
			t.Fatalf("updater.Run failed: %v", err)
		}

		if res.UpdateAvailable {
			t.Errorf("expected update_available to be false")
		}
		if !strings.Contains(res.Message, "zbxctl is up to date") {
			t.Errorf("unexpected check message: %s", res.Message)
		}
	})

	t.Run("apply update successfully and verify skill reminder", func(t *testing.T) {
		updater := NewUpdater(Options{
			CurrentVersion:   "0.2.0",
			Repo:             "ErKatta/zbxctl",
			TargetExecutable: targetBinPath,
			CheckOnly:        false,
			BaseAPIURL:       server.URL,
			HTTPClient:       server.Client(),
		})

		res, err := updater.Run(context.Background())
		if err != nil {
			t.Fatalf("updater.Run failed: %v", err)
		}

		if !res.Updated {
			t.Errorf("expected updated to be true")
		}
		if !strings.Contains(res.SkillReminder, "zbxctl skill install --all") {
			t.Errorf("expected skill reminder in update result, got: %s", res.SkillReminder)
		}

		// Verify file was updated on disk
		content, err := os.ReadFile(targetBinPath)
		if err != nil {
			t.Fatalf("failed to read updated file: %v", err)
		}
		if string(content) != newBinaryPayload {
			t.Errorf("expected updated content %q, got %q", newBinaryPayload, string(content))
		}
	})

	t.Run("already up to date without force", func(t *testing.T) {
		updater := NewUpdater(Options{
			CurrentVersion:   "0.3.0",
			Repo:             "ErKatta/zbxctl",
			TargetExecutable: targetBinPath,
			CheckOnly:        false,
			BaseAPIURL:       server.URL,
			HTTPClient:       server.Client(),
		})

		res, err := updater.Run(context.Background())
		if err != nil {
			t.Fatalf("updater.Run failed: %v", err)
		}

		if res.Updated {
			t.Errorf("expected updated to be false when already up to date")
		}
		if !strings.Contains(res.Message, "zbxctl is already up to date") {
			t.Errorf("unexpected message: %s", res.Message)
		}
	})

	t.Run("force update when on same version", func(t *testing.T) {
		updater := NewUpdater(Options{
			CurrentVersion:   "0.3.0",
			Repo:             "ErKatta/zbxctl",
			TargetExecutable: targetBinPath,
			CheckOnly:        false,
			Force:            true,
			BaseAPIURL:       server.URL,
			HTTPClient:       server.Client(),
		})

		res, err := updater.Run(context.Background())
		if err != nil {
			t.Fatalf("updater.Run failed: %v", err)
		}

		if !res.Updated {
			t.Errorf("expected updated to be true with force flag")
		}
	})
}

func testingTime() (t time.Time) {
	t, _ = time.Parse(time.RFC3339, "2026-08-17T12:00:00Z")
	return t
}
