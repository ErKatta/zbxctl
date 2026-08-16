package output

import (
	"strings"
	"testing"
	"time"
)

func TestParseEpoch(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		wantEpoch int64
		wantOk    bool
	}{
		{"nil", nil, 0, false},
		{"empty string", "", 0, false},
		{"zero string", "0", 0, false},
		{"zero int", 0, 0, false},
		{"negative int", -10, 0, false},
		{"invalid string", "not-a-number", 0, false},
		{"valid string", "1700000000", 1700000000, true},
		{"valid int", 1700000000, 1700000000, true},
		{"valid int64", int64(1700000000), 1700000000, true},
		{"valid float64", float64(1700000000), 1700000000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEpoch, gotOk := ParseEpoch(tt.input)
			if gotOk != tt.wantOk {
				t.Errorf("ParseEpoch(%v) gotOk = %v, wantOk %v", tt.input, gotOk, tt.wantOk)
			}
			if gotEpoch != tt.wantEpoch {
				t.Errorf("ParseEpoch(%v) gotEpoch = %v, wantEpoch %v", tt.input, gotEpoch, tt.wantEpoch)
			}
		})
	}
}

func TestFormatClockRelative(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    interface{}
		contains string
	}{
		{"nil", nil, "-"},
		{"empty", "", "-"},
		{"zero", "0", "-"},
		{"just now", now.Unix() - 10, "10s"},
		{"12 minutes ago", now.Unix() - 12*60, "12m"},
		{"48 minutes ago", now.Unix() - 48*60, "48m"},
		{"2 hours ago", now.Unix() - 2*3600, "2h"},
		{"3 days ago", now.Unix() - 3*86400, "3d"},
		{"future within 1m", now.Unix() + 30, "just now"},
		{"future beyond 1m", now.Unix() + 3600, "in future"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := FormatClockRelative(tt.input)
			if !strings.Contains(res, tt.contains) {
				t.Errorf("FormatClockRelative(%v) = %q; want it to contain %q", tt.input, res, tt.contains)
			}
		})
	}
}

func TestFormatClockAgo(t *testing.T) {
	now := time.Now()
	res := FormatClockAgo(now.Unix() - 12*60)
	if !strings.Contains(res, "12m ago") {
		t.Errorf("FormatClockAgo expected '12m ago', got %q", res)
	}
}
