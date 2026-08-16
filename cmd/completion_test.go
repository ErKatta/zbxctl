package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestCompletionCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedIn    string
		expectErr     bool
		errMsgContain string
	}{
		{
			name:       "bash completion",
			args:       []string{"completion", "bash"},
			expectedIn: "bash completion for zbxctl",
			expectErr:  false,
		},
		{
			name:       "zsh completion",
			args:       []string{"completion", "zsh"},
			expectedIn: "#compdef zbxctl",
			expectErr:  false,
		},
		{
			name:       "fish completion",
			args:       []string{"completion", "fish"},
			expectedIn: "complete -c zbxctl",
			expectErr:  false,
		},
		{
			name:       "powershell completion",
			args:       []string{"completion", "powershell"},
			expectedIn: "Register-ArgumentCompleter",
			expectErr:  false,
		},
		{
			name:          "invalid shell",
			args:          []string{"completion", "invalid_shell"},
			expectErr:     true,
			errMsgContain: `unsupported shell type "invalid_shell": must be one of [bash, zsh, fish, powershell]`,
		},
		{
			name:          "missing shell argument",
			args:          []string{"completion"},
			expectErr:     true,
			errMsgContain: "accepts 1 arg(s), received 0",
		},
		{
			name:          "too many arguments",
			args:          []string{"completion", "bash", "extra"},
			expectErr:     true,
			errMsgContain: "accepts 1 arg(s), received 2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			cmd := RootCmd
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tc.args)

			err := cmd.Execute()
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.errMsgContain)
				}
				if !strings.Contains(err.Error(), tc.errMsgContain) {
					t.Errorf("expected error containing %q, got %q", tc.errMsgContain, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				output := buf.String()
				if !strings.Contains(output, tc.expectedIn) {
					t.Errorf("expected output to contain %q, but got:\n%s", tc.expectedIn, output)
				}
			}
		})
	}
}
