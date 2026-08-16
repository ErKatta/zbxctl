package prompt

import (
	"bytes"
	"strings"
	"testing"
)

func TestPromptPassword(t *testing.T) {
	inputPass := "supersecret123\n"
	StdinOverride = strings.NewReader(inputPass)
	outBuf := new(bytes.Buffer)
	StderrOverride = outBuf

	defer func() {
		StdinOverride = nil
		StderrOverride = nil
	}()

	pass, err := PromptPassword("Password: ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pass != "supersecret123" {
		t.Errorf("expected 'supersecret123', got %q", pass)
	}

	if outBuf.String() != "Password: " {
		t.Errorf("expected prompt 'Password: ', got %q", outBuf.String())
	}
}
