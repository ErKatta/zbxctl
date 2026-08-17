package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/ErKatta/zbxctl/pkg/output"
	"github.com/ErKatta/zbxctl/pkg/safety"
	"github.com/ErKatta/zbxctl/pkg/zabbix"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	editFileFlag               string
	editEditorFlag             string
	editOutputFlag             string
	editWindowsLineEndingsFlag bool
)

var editCmd = &cobra.Command{
	Use:   "edit (RESOURCE/NAME | RESOURCE NAME | -f FILENAME)",
	Short: "Edit a Zabbix resource directly in your preferred text editor",
	Long: `edit fetches a live Zabbix resource, converts it into a declarative manifest, and opens it in your default text editor.
When you save and exit the editor, zbxctl validates your changes and applies them directly to the Zabbix server.

Supported invocations:
  zbxctl edit host web-prod-01
  zbxctl edit host/web-prod-01
  zbxctl edit template 40001
  zbxctl edit trigger 30001
  zbxctl edit -f host-manifest.yaml

Editor Resolution Order:
  1. --editor flag
  2. ZBX_EDITOR environment variable
  3. EDITOR environment variable
  4. VISUAL environment variable
  5. System default ('nano', 'vim', 'vi' on Unix / 'notepad.exe' on Windows)
`,
	Args: cobra.RangeArgs(0, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var resName string
		var identifier string
		var isFileMode bool

		if editFileFlag != "" {
			isFileMode = true
		} else if len(args) == 1 {
			arg := args[0]
			if strings.Contains(arg, "/") {
				parts := strings.SplitN(arg, "/", 2)
				resName = parts[0]
				identifier = parts[1]
			} else {
				return fmt.Errorf("when passing a single positional argument, use 'RESOURCE/NAME' (e.g. 'host/web-prod-01') or specify both resource and name/id")
			}
		} else if len(args) == 2 {
			resName = args[0]
			identifier = args[1]
		} else {
			return fmt.Errorf("requires either a resource and name/id (e.g. 'zbxctl edit host web-prod-01' or 'zbxctl edit host/web-prod-01') or -f flag")
		}

		var initialContent []byte
		var targetID string
		var displayName string
		var resInfo zabbix.ResourceInfo
		var err error

		if isFileMode {
			initialContent, err = os.ReadFile(editFileFlag)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", editFileFlag, err)
			}
			displayName = editFileFlag
		} else {
			resInfo, err = zabbix.ResolveResource(resName)
			if err != nil {
				return fmt.Errorf("failed to resolve resource %q: %w", resName, err)
			}

			params := map[string]interface{}{
				"output": "extend",
			}

			switch resInfo.APIPrefix {
			case "host":
				params["selectTags"] = "extend"
				params["selectGroups"] = "extend"
				params["selectMacros"] = "extend"
				params["selectParentTemplates"] = "extend"
				params["selectInterfaces"] = "extend"
				params["selectInventory"] = "extend"
			case "template":
				params["selectTags"] = "extend"
				params["selectGroups"] = "extend"
				params["selectMacros"] = "extend"
				params["selectTemplates"] = "extend"
			case "hostgroup":
				params["selectTags"] = "extend"
			case "item":
				params["selectTags"] = "extend"
				params["selectPreprocessing"] = "extend"
			case "trigger":
				params["selectTags"] = "extend"
				params["selectDependencies"] = "extend"
			}

			if resInfo.Name == "inventory" {
				params["selectInventory"] = "extend"
				params["output"] = []string{"hostid", "host", "name", "inventory_mode"}
			}

			if isNumeric(identifier) {
				if resInfo.PluralIDProperty != "" {
					params[resInfo.PluralIDProperty] = []string{identifier}
				} else {
					params[resInfo.IDProperty] = identifier
				}
			} else {
				params["filter"] = map[string]interface{}{
					resInfo.NameProperty: identifier,
				}
			}

			res, err := checkSafetyAndCall(cmd.Context(), resInfo.APIPrefix+".get", params)
			if err != nil {
				return fmt.Errorf("failed to fetch live resource %s %s: %w", resInfo.Name, identifier, err)
			}

			var liveList []map[string]interface{}
			_ = json.Unmarshal(res.(json.RawMessage), &liveList)

			// If empty and searched by name on host, retry with host field
			if len(liveList) == 0 && resInfo.APIPrefix == "host" && !isNumeric(identifier) {
				params["filter"] = map[string]interface{}{
					"host": identifier,
				}
				resRetry, errRetry := checkSafetyAndCall(cmd.Context(), "host.get", params)
				if errRetry == nil {
					_ = json.Unmarshal(resRetry.(json.RawMessage), &liveList)
				}
			}

			if len(liveList) == 0 {
				return fmt.Errorf("live %s %q not found", resInfo.Name, identifier)
			}

			liveItem := liveList[0]
			if idVal, ok := liveItem[resInfo.IDProperty]; ok && idVal != nil {
				targetID = fmt.Sprintf("%v", idVal)
			}
			displayName = identifier
			if nameVal, ok := liveItem[resInfo.NameProperty]; ok && nameVal != nil && fmt.Sprintf("%v", nameVal) != "" {
				displayName = fmt.Sprintf("%v", nameVal)
			}

			cleanSpec := zabbix.SanitizeExportSpec(resInfo.Name, liveItem)
			if targetID != "" {
				cleanSpec[resInfo.IDProperty] = targetID
			}

			manifest := ManifestItem{
				Kind: resInfo.Name,
				Spec: cleanSpec,
			}

			var manifestBytes []byte
			outFmt := editOutputFlag
			if outFmt == "" {
				outFmt = activeCtx.OutputFormat
			}
			if outFmt == "json" {
				manifestBytes, err = json.MarshalIndent(manifest, "", "  ")
			} else {
				manifestBytes, err = yaml.Marshal(manifest)
			}
			if err != nil {
				return fmt.Errorf("failed to marshal resource manifest: %w", err)
			}

			var buf bytes.Buffer
			buf.WriteString("# Please edit the object below. Lines beginning with a '#' will be ignored,\n")
			buf.WriteString("# and an empty file will abort the edit. If an error occurs while saving this file will be\n")
			buf.WriteString("# reopened with the relevant failures.\n#\n")
			buf.WriteString(fmt.Sprintf("# Resource: %s/%s", resInfo.Name, displayName))
			if targetID != "" {
				buf.WriteString(fmt.Sprintf(" (ID: %s)", targetID))
			}
			buf.WriteString("\n")
			buf.WriteString(fmt.Sprintf("# Context:  %s (safety-level: %s)\n", actCtxName, activeCtx.SafetyLevel))
			buf.WriteString("#\n")
			buf.Write(manifestBytes)

			initialContent = buf.Bytes()
		}

		// Editor resolution
		editorStr, err := resolveEditor(editEditorFlag)
		if err != nil {
			return err
		}
		editorArgs, err := parseEditorCommand(editorStr)
		if err != nil {
			return fmt.Errorf("invalid editor command %q: %w", editorStr, err)
		}

		ext := "yaml"
		if editOutputFlag == "json" {
			ext = "json"
		}
		tmpFile, err := os.CreateTemp("", fmt.Sprintf("zbxctl-edit-*.%s", ext))
		if err != nil {
			return fmt.Errorf("failed to create temporary edit file: %w", err)
		}
		tmpFilePath := tmpFile.Name()
		defer func() {
			_ = os.Remove(tmpFilePath)
		}()

		// Write initial buffer
		contentToWrite := initialContent
		if editWindowsLineEndingsFlag || (runtime.GOOS == "windows" && !strings.Contains(string(contentToWrite), "\r\n")) {
			contentToWrite = bytes.ReplaceAll(contentToWrite, []byte("\n"), []byte("\r\n"))
		}
		if _, err := tmpFile.Write(contentToWrite); err != nil {
			_ = tmpFile.Close()
			return fmt.Errorf("failed to write to temporary file: %w", err)
		}
		_ = tmpFile.Close()

		origSanitized := stripYamlComments(initialContent)

		// Interactive edit & retry loop
		for {
			runArgs := append([]string{}, editorArgs[1:]...)
			runArgs = append(runArgs, tmpFilePath)
			editorProc := exec.Command(editorArgs[0], runArgs...)
			editorProc.Stdin = cmd.InOrStdin()
			editorProc.Stdout = cmd.OutOrStdout()
			editorProc.Stderr = cmd.ErrOrStderr()

			if err := editorProc.Run(); err != nil {
				return fmt.Errorf("there was a problem with the editor %q: %w", editorStr, err)
			}

			editedBytes, err := os.ReadFile(tmpFilePath)
			if err != nil {
				return fmt.Errorf("failed to read back edited file: %w", err)
			}

			editedSanitized := stripYamlComments(editedBytes)

			// Check if unchanged or empty
			if len(bytes.TrimSpace(editedSanitized)) == 0 || bytes.Equal(bytes.TrimSpace(origSanitized), bytes.TrimSpace(editedSanitized)) {
				fmt.Fprintln(formatter.Writer, "Edit cancelled, no changes made.")
				return nil
			}

			// Parse manifest items
			items, parseErr := parseManifestItems(editedSanitized)
			var applyErr error
			if parseErr != nil {
				applyErr = fmt.Errorf("failed to parse manifest: %w", parseErr)
			} else if len(items) == 0 {
				applyErr = fmt.Errorf("no valid manifest object found")
			} else {
				// Apply the items
				for _, item := range items {
					itemResName := item.Kind
					if itemResName == "" {
						itemResName = item.Resource
					}
					if itemResName == "" {
						itemResName = resInfo.Name
					}

					itemResInfo, rErr := zabbix.ResolveResource(itemResName)
					if rErr != nil {
						applyErr = fmt.Errorf("failed to resolve resource %q: %w", itemResName, rErr)
						break
					}

					params := item.Params
					if len(params) == 0 && len(item.Spec) > 0 {
						params = item.Spec
					}

					// Preserve target ID if missing from spec
					if targetID != "" && (params[itemResInfo.IDProperty] == nil || fmt.Sprintf("%v", params[itemResInfo.IDProperty]) == "") {
						params[itemResInfo.IDProperty] = targetID
					}

					var method string
					if item.Method != "" {
						method = item.Method
					} else if itemResInfo.Name == "inventory" {
						method = "host.update"
					} else {
						method = itemResInfo.APIPrefix + ".update"
					}

					params = zabbix.SanitizeApplyParams(itemResName, params)

					_, callErr := checkSafetyAndCall(cmd.Context(), method, params)
					if callErr != nil {
						applyErr = callErr
						break
					}
				}
			}

			if applyErr == nil {
				// Update manifest file if in -f mode
				if isFileMode {
					_ = os.WriteFile(editFileFlag, editedBytes, 0644)
				}
				if resInfo.Name != "" {
					fmt.Fprintf(formatter.Writer, "%s %q edited\n", resInfo.Name, displayName)
				} else {
					fmt.Fprintf(formatter.Writer, "resource %q edited\n", displayName)
				}
				return nil
			}

			// If it's a safety violation error, do not prompt retry, fail immediately with safety error envelope
			if _, isSafetyErr := applyErr.(*safety.SafetyError); isSafetyErr {
				return applyErr
			}

			// Report error to stderr
			fmt.Fprintf(cmd.ErrOrStderr(), "error: %v\n", applyErr)

			// If non-interactive or cannot prompt, preserve edit and return error
			if !output.IsTerminal() || cmd.InOrStdin() == nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Your edits have been saved to %q\n", tmpFilePath)
				// Re-create temp file without auto-deletion so user doesn't lose work
				preservePath := fmt.Sprintf("%s.saved", tmpFilePath)
				_ = os.WriteFile(preservePath, editedBytes, 0600)
				fmt.Fprintf(cmd.ErrOrStderr(), "Preserved edits file: %q\n", preservePath)
				return applyErr
			}

			// Prompt user to re-edit or abort (kubectl standard behavior)
			fmt.Fprint(cmd.ErrOrStderr(), "Press [Enter] to re-edit, or enter 'x' to abort: ")
			var response string
			_, _ = fmt.Fscanln(cmd.InOrStdin(), &response)
			trimmedResp := strings.ToLower(strings.TrimSpace(response))
			if trimmedResp == "x" || trimmedResp == "q" || trimmedResp == "abort" || trimmedResp == "n" {
				preservePath := fmt.Sprintf("%s.saved", tmpFilePath)
				_ = os.WriteFile(preservePath, editedBytes, 0600)
				fmt.Fprintf(cmd.ErrOrStderr(), "Edit aborted. Your edits have been saved to %q\n", preservePath)
				return fmt.Errorf("edit aborted by user")
			}

			// Prepend error comments to temp file before re-opening editor
			var errBuf bytes.Buffer
			errBuf.WriteString("# An error occurred while updating the resource:\n")
			for _, line := range strings.Split(applyErr.Error(), "\n") {
				errBuf.WriteString(fmt.Sprintf("# %s\n", line))
			}
			errBuf.WriteString("#\n")
			errBuf.Write(editedSanitized)

			_ = os.WriteFile(tmpFilePath, errBuf.Bytes(), 0600)
		}
	},
}

func stripYamlComments(data []byte) []byte {
	var out []string
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		out = append(out, line)
	}
	return []byte(strings.Join(out, "\n"))
}

func parseEditorCommand(editorStr string) ([]string, error) {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := byte(0)

	for i := 0; i < len(editorStr); i++ {
		c := editorStr[i]
		if inQuote {
			if c == quoteChar {
				inQuote = false
			} else {
				current.WriteByte(c)
			}
		} else {
			if c == '\'' || c == '"' {
				inQuote = true
				quoteChar = c
			} else if c == ' ' || c == '\t' {
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
			} else {
				current.WriteByte(c)
			}
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	if len(args) == 0 {
		return nil, fmt.Errorf("empty editor command")
	}
	return args, nil
}

func resolveEditor(flagEditor string) (string, error) {
	if flagEditor != "" {
		return flagEditor, nil
	}
	if env := os.Getenv("ZBX_EDITOR"); env != "" {
		return env, nil
	}
	if env := os.Getenv("EDITOR"); env != "" {
		return env, nil
	}
	if env := os.Getenv("VISUAL"); env != "" {
		return env, nil
	}
	if runtime.GOOS == "windows" {
		if p, err := exec.LookPath("notepad.exe"); err == nil {
			return p, nil
		}
		return "notepad.exe", nil
	}
	for _, candidate := range []string{"nano", "vim", "vi"} {
		if p, err := exec.LookPath(candidate); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("unable to find an editor. Please set EDITOR or ZBX_EDITOR environment variable or use --editor")
}

func init() {
	editCmd.Flags().StringVarP(&editFileFlag, "file", "f", "", "path to manifest file (.json or .yaml) to edit and apply")
	editCmd.Flags().StringVar(&editEditorFlag, "editor", "", "editor binary/command to use (overrides $EDITOR and $ZBX_EDITOR)")
	editCmd.Flags().StringVarP(&editOutputFlag, "output", "o", "", "output format of temporary edit file (yaml, json)")
	editCmd.Flags().BoolVar(&editWindowsLineEndingsFlag, "windows-line-endings", false, "use Windows line endings (CRLF) in temporary file")
	RootCmd.AddCommand(editCmd)
}
