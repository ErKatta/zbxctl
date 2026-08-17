package cmd

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
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

	"github.com/ErKatta/zbxctl/pkg/update"
)

func createCmdTestTarGz(t *testing.T, binaryName, content string) []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

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

func createCmdTestZip(t *testing.T, binaryName, content string) []byte {
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

func TestUpdateCommand(t *testing.T) {
	tmpDir := t.TempDir()
	targetBinary := filepath.Join(tmpDir, "zbxctl")
	if err := os.WriteFile(targetBinary, []byte("original-v0.1.0-binary"), 0755); err != nil {
		t.Fatalf("failed to write dummy target binary: %v", err)
	}

	newBinaryPayload := "new-v0.9.0-binary-content"
	var archiveBytes []byte
	var assetName string
	if runtime.GOOS == "windows" {
		assetName = fmt.Sprintf("zbxctl_0.9.0_%s_%s.zip", runtime.GOOS, runtime.GOARCH)
		archiveBytes = createCmdTestZip(t, "zbxctl.exe", newBinaryPayload)
	} else {
		assetName = fmt.Sprintf("zbxctl_0.9.0_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
		archiveBytes = createCmdTestTarGz(t, "zbxctl", newBinaryPayload)
	}

	hasher := sha256.New()
	hasher.Write(archiveBytes)
	archiveHash := hex.EncodeToString(hasher.Sum(nil))
	checksumsData := []byte(fmt.Sprintf("%s  %s\n", archiveHash, assetName))

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/releases/latest"):
			rel := update.Release{
				TagName:     "v0.9.0",
				Name:        "zbxctl v0.9.0 Release",
				PublishedAt: time.Now(),
				HTMLURL:     "https://github.com/ErKatta/zbxctl/releases/tag/v0.9.0",
				Assets: []update.Asset{
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

	os.Setenv("ZBXCTL_UPDATE_BASE_URL", server.URL)
	defer os.Unsetenv("ZBXCTL_UPDATE_BASE_URL")

	t.Run("zbxctl update --check", func(t *testing.T) {
		cmd := RootCmd
		ResetCommandFlags(cmd)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"update", "--check", "--target-binary", targetBinary})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("update --check failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "New version available: v0.9.0") {
			t.Errorf("expected 'New version available: v0.9.0' in output, got:\n%s", out)
		}
		if !strings.Contains(out, "Run 'zbxctl update' to install") {
			t.Errorf("expected run instructions in output, got:\n%s", out)
		}
	})

	t.Run("zbxctl update -o json", func(t *testing.T) {
		cmd := RootCmd
		ResetCommandFlags(cmd)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"-o", "json", "update", "--check", "--target-binary", targetBinary})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("update -o json failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, `"latest_version": "v0.9.0"`) || !strings.Contains(out, `"update_available": true`) {
			t.Errorf("expected json update metadata in output, got:\n%s", out)
		}
		if !strings.Contains(out, "zbxctl skill install --all") {
			t.Errorf("expected skill_reminder in json output, got:\n%s", out)
		}
	})

	t.Run("zbxctl update (apply update and remind skill install)", func(t *testing.T) {
		cmd := RootCmd
		ResetCommandFlags(cmd)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"update", "--target-binary", targetBinary})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("update execution failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "Successfully updated zbxctl to v0.9.0") {
			t.Errorf("expected success message in output, got:\n%s", out)
		}
		if !strings.Contains(out, "zbxctl skill install --all") {
			t.Errorf("expected skill install reminder in output, got:\n%s", out)
		}

		// Verify binary on disk was replaced
		content, err := os.ReadFile(targetBinary)
		if err != nil {
			t.Fatalf("failed to read updated target binary: %v", err)
		}
		if string(content) != newBinaryPayload {
			t.Errorf("expected updated content %q, got %q", newBinaryPayload, string(content))
		}
	})
}
