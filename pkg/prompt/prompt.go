package prompt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

var (
	// StdinOverride can be set in tests to inject password inputs.
	StdinOverride io.Reader
	// StderrOverride can be set in tests to capture prompt output.
	StderrOverride io.Writer
)

// PromptPassword prompts the user for a password on stderr and reads the password.
// If running in an interactive terminal, characters typed are masked/hidden.
func PromptPassword(promptMsg string) (string, error) {
	r := StdinOverride
	if r == nil {
		r = os.Stdin
	}
	w := StderrOverride
	if w == nil {
		w = os.Stderr
	}

	if promptMsg != "" {
		fmt.Fprint(w, promptMsg)
	}

	if f, ok := r.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		bytePassword, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(w)
		if err != nil {
			return "", fmt.Errorf("failed to read password from terminal: %w", err)
		}
		return string(bytePassword), nil
	}

	reader := bufio.NewReader(r)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	return strings.TrimRight(line, "\r\n"), nil
}
