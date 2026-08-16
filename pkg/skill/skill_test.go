package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuiltinSkills(t *testing.T) {
	if len(BuiltinSkills) == 0 {
		t.Fatalf("expected built-in skills, got 0")
	}

	expectedSkills := []string{"zabbix-automation", "zabbix-troubleshooting", "zabbix-telemetry", "zabbix-safety"}
	for _, name := range expectedSkills {
		s, ok := BuiltinSkills[name]
		if !ok {
			t.Errorf("expected skill %q to be present", name)
		}
		if s.Content == "" {
			t.Errorf("expected skill %q to have non-empty content", name)
		}
	}
}

func TestInstallSkill(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "zbxctl-skill-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	path, err := InstallSkill("zabbix-automation", tempDir)
	if err != nil {
		t.Fatalf("failed to install skill: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("installed skill file does not exist at %s", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read installed skill file: %v", err)
	}

	if len(content) == 0 {
		t.Fatalf("installed skill file is empty")
	}

	expectedFile := filepath.Join(tempDir, "zabbix-automation", "SKILL.md")
	if path != expectedFile {
		t.Errorf("expected path %q, got %q", expectedFile, path)
	}
}
