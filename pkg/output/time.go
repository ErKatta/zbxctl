package output

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseEpoch attempts to parse a generic value into a Unix epoch timestamp (int64).
// Returns the epoch and true if valid (epoch > 0), or (0, false) if invalid.
func ParseEpoch(val interface{}) (int64, bool) {
	if val == nil {
		return 0, false
	}

	switch v := val.(type) {
	case int:
		if v > 0 {
			return int64(v), true
		}
	case int32:
		if v > 0 {
			return int64(v), true
		}
	case int64:
		if v > 0 {
			return v, true
		}
	case float64:
		if v > 0 {
			return int64(v), true
		}
	case json.Number:
		if n, err := v.Int64(); err == nil && n > 0 {
			return n, true
		}
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" || trimmed == "0" {
			return 0, false
		}
		if n, err := strconv.ParseInt(trimmed, 10, 64); err == nil && n > 0 {
			return n, true
		}
	}

	return 0, false
}

// FormatDuration formats a time.Duration into a concise human-readable string (e.g. 12m, 2h, 3d, 45s).
func FormatDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}

	if d < time.Minute {
		secs := int(d.Seconds())
		if secs <= 0 {
			return "0s"
		}
		return fmt.Sprintf("%ds", secs)
	}

	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}

	if d < 24*time.Hour {
		hrs := int(d.Hours())
		mins := int(d.Minutes()) % 60
		if mins > 0 && hrs < 6 {
			return fmt.Sprintf("%dh%dm", hrs, mins)
		}
		return fmt.Sprintf("%dh", hrs)
	}

	days := int(d.Hours() / 24)
	if days < 30 {
		hrs := int(d.Hours()) % 24
		if hrs > 0 && days < 5 {
			return fmt.Sprintf("%dd%dh", days, hrs)
		}
		return fmt.Sprintf("%dd", days)
	}

	if days < 365 {
		return fmt.Sprintf("%dmo", days/30)
	}

	return fmt.Sprintf("%dy", days/365)
}

// FormatClockRelative formats an epoch timestamp into a human-friendly relative age (e.g. 12m, 2h, 3d, 45s).
func FormatClockRelative(val interface{}) string {
	epoch, ok := ParseEpoch(val)
	if !ok {
		if val == nil {
			return "-"
		}
		str := fmt.Sprintf("%v", val)
		if str == "" || str == "0" {
			return "-"
		}
		return str
	}

	t := time.Unix(epoch, 0)
	diff := time.Since(t)

	if diff < 0 {
		if diff > -time.Minute {
			return "just now"
		}
		return FormatDuration(-diff) + " in future"
	}

	return FormatDuration(diff)
}

// FormatClockAgo formats an epoch timestamp into a relative time with "ago" (e.g. 12m ago, 2h ago).
func FormatClockAgo(val interface{}) string {
	epoch, ok := ParseEpoch(val)
	if !ok {
		if val == nil {
			return "-"
		}
		str := fmt.Sprintf("%v", val)
		if str == "" || str == "0" {
			return "-"
		}
		return str
	}

	t := time.Unix(epoch, 0)
	diff := time.Since(t)

	if diff < 0 {
		if diff > -time.Minute {
			return "just now"
		}
		return FormatDuration(-diff) + " in future"
	}

	return FormatDuration(diff) + " ago"
}
