package safety

import (
	"encoding/json"
	"fmt"
	"strings"
)

type SafetyLevel string

const (
	LevelReadOnly                SafetyLevel = "readonly"
	LevelReadWriteMine           SafetyLevel = "readwrite-mine"
	LevelReadWriteAll            SafetyLevel = "readwrite-all"
	LevelDangerouslyUnrestricted SafetyLevel = "dangerously-unrestricted"
)

type SafetyError struct {
	Code       string `json:"code"`
	Method     string `json:"method"`
	Message    string `json:"message"`
	Resolution string `json:"resolution"`
}

func (e *SafetyError) Error() string {
	return e.Message
}

type SafetyErrorEnvelope struct {
	Error SafetyError `json:"error"`
}

func IsReadOnlyMethod(method string) bool {
	lower := strings.ToLower(method)
	if strings.HasSuffix(lower, ".get") {
		return true
	}
	switch lower {
	case "apiinfo.version", "user.checkauthentication", "user.login", "user.logout":
		return true
	}
	return false
}

func CheckSafety(level SafetyLevel, contextName, method string, params interface{}, force bool) error {
	switch level {
	case LevelDangerouslyUnrestricted:
		return nil

	case LevelReadOnly:
		if IsReadOnlyMethod(method) {
			return nil
		}
		return &SafetyError{
			Code:       "SAFETY_LEVEL_VIOLATION",
			Method:     method,
			Message:    fmt.Sprintf("Operation %q blocked by safety-level 'readonly' on context %q.", method, contextName),
			Resolution: "Switch context or update safety-level in ~/.zbxctl/config.yaml.",
		}

	case LevelReadWriteMine:
		if IsReadOnlyMethod(method) {
			return nil
		}
		// For readwrite-mine, mutations are allowed only if parameters indicate zbxctl tag/marker
		if isTaggedForZbxctl(params) {
			return nil
		}
		return &SafetyError{
			Code:       "SAFETY_LEVEL_VIOLATION",
			Method:     method,
			Message:    fmt.Sprintf("Operation %q blocked by safety-level 'readwrite-mine' on context %q (resource must be tagged or created with tag zbxctl=true).", method, contextName),
			Resolution: "Add tag zbxctl=true to the resource or upgrade safety-level to 'readwrite-all'.",
		}

	case LevelReadWriteAll:
		if IsReadOnlyMethod(method) {
			return nil
		}
		// Block bulk deletes unless force is explicit
		if strings.HasSuffix(strings.ToLower(method), ".delete") {
			if isBulkOperation(params) && !force {
				return &SafetyError{
					Code:       "SAFETY_LEVEL_VIOLATION",
					Method:     method,
					Message:    fmt.Sprintf("Bulk operation %q blocked by safety-level 'readwrite-all' on context %q without --force flag.", method, contextName),
					Resolution: "Re-run with --force flag to confirm bulk operation.",
				}
			}
		}
		return nil

	default:
		// Default to strict readonly if unknown
		if IsReadOnlyMethod(method) {
			return nil
		}
		return &SafetyError{
			Code:       "SAFETY_LEVEL_VIOLATION",
			Method:     method,
			Message:    fmt.Sprintf("Operation %q blocked by unknown safety-level %q on context %q.", method, level, contextName),
			Resolution: "Set a valid safety-level (readonly, readwrite-mine, readwrite-all, dangerously-unrestricted).",
		}
	}
}

func isTaggedForZbxctl(params interface{}) bool {
	if params == nil {
		return false
	}
	data, err := json.Marshal(params)
	if err != nil {
		return false
	}
	str := strings.ToLower(string(data))
	return strings.Contains(str, "zbxctl") || strings.Contains(str, "managed-by")
}

func isBulkOperation(params interface{}) bool {
	if params == nil {
		return true
	}
	// Check if array with > 1 items
	switch v := params.(type) {
	case []interface{}:
		return len(v) > 1
	case []string:
		return len(v) > 1
	case []int:
		return len(v) > 1
	case []int64:
		return len(v) > 1
	}

	data, err := json.Marshal(params)
	if err != nil {
		return false
	}
	var list []interface{}
	if err := json.Unmarshal(data, &list); err == nil {
		return len(list) > 1
	}
	return false
}
