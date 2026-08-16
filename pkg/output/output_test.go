package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

// errWriter is a helper writer that always fails on write
type errWriter struct{}

func (e *errWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write error")
}

// unmarshalableType is a type that fails json.Marshal
type unmarshalableType struct {
	Fn func()
}

type yamlErrStruct struct{}

func (y *yamlErrStruct) MarshalYAML() (interface{}, error) {
	return nil, errors.New("yaml marshal failed")
}




func TestNewFormatter(t *testing.T) {
	f := NewFormatter("JSON")
	if f == nil {
		t.Fatal("expected non-nil formatter")
	}
	if f.Format != "json" {
		t.Errorf("expected format 'json', got '%s'", f.Format)
	}
}

func TestIsTerminal(t *testing.T) {
	// Call IsTerminal to verify it doesn't panic and returns a boolean
	_ = IsTerminal()
}

func TestFormatterPrintAutoAndDefault(t *testing.T) {
	tests := []struct {
		name       string
		format     string
		data       interface{}
		expectJSON bool
	}{
		{
			name:       "empty format fallback",
			format:     "",
			data:       map[string]string{"foo": "bar"},
			expectJSON: true,
		},
		{
			name:       "auto format fallback",
			format:     "auto",
			data:       map[string]string{"foo": "bar"},
			expectJSON: true,
		},
		{
			name:       "unknown format fallback to json",
			format:     "unknown-format",
			data:       map[string]string{"foo": "bar"},
			expectJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			f := &Formatter{
				Format: tt.format,
				Writer: &buf,
			}
			err := f.Print(tt.data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			out := buf.String()
			if tt.expectJSON && !strings.Contains(out, `"foo": "bar"`) {
				t.Errorf("expected json output containing foo:bar, got: %s", out)
			}
		})
	}
}

func TestFormatterJSON(t *testing.T) {
	t.Run("map data", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "json", Writer: &buf}
		data := map[string]interface{}{
			"host":   "server-01",
			"status": "0",
		}
		if err := f.Print(data); err != nil {
			t.Fatalf("failed to print json: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, `"host": "server-01"`) {
			t.Errorf("expected json output to contain host string, got: %s", out)
		}
	})

	t.Run("valid json.RawMessage", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "json", Writer: &buf}
		raw := json.RawMessage(`{"host":"server-02","status":1}`)
		if err := f.Print(raw); err != nil {
			t.Fatalf("failed to print json.RawMessage: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "server-02") || !strings.Contains(out, "\n  \"host\":") {
			t.Errorf("expected indented json output for raw message, got: %s", out)
		}
	})

	t.Run("invalid json.RawMessage fallback", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "json", Writer: &buf}
		raw := json.RawMessage(`{not valid json}`)
		if err := f.Print(raw); err != nil {
			t.Fatalf("failed to print invalid json.RawMessage: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "{not valid json}") {
			t.Errorf("expected output to contain raw message, got: %s", out)
		}
	})

	t.Run("json marshal error", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "json", Writer: &buf}
		data := unmarshalableType{Fn: func() {}}
		err := f.Print(data)
		if err == nil {
			t.Fatal("expected json marshal error, got nil")
		}
		if !strings.Contains(err.Error(), "json marshal failed") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("writer error", func(t *testing.T) {
		f := &Formatter{Format: "json", Writer: &errWriter{}}
		err := f.Print(map[string]string{"key": "value"})
		if err == nil {
			t.Fatal("expected writer error, got nil")
		}
	})
}

func TestFormatterYAML(t *testing.T) {
	t.Run("map data", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "yaml", Writer: &buf}
		data := map[string]interface{}{
			"host":   "server-01",
			"status": "0",
		}
		if err := f.Print(data); err != nil {
			t.Fatalf("failed to print yaml: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "host: server-01") {
			t.Errorf("expected yaml output to contain 'host: server-01', got: %s", out)
		}
	})

	t.Run("valid json.RawMessage", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "yaml", Writer: &buf}
		raw := json.RawMessage(`{"host":"server-yaml","active":true}`)
		if err := f.Print(raw); err != nil {
			t.Fatalf("failed to print yaml with raw json: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "host: server-yaml") {
			t.Errorf("expected yaml to contain 'host: server-yaml', got: %s", out)
		}
	})

	t.Run("invalid json.RawMessage unmarshal fallback", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "yaml", Writer: &buf}
		raw := json.RawMessage(`invalid-raw`)
		// Should still marshal the raw slice to yaml
		if err := f.Print(raw); err != nil {
			t.Fatalf("failed to print invalid json.RawMessage in yaml: %v", err)
		}
	})

	t.Run("yaml marshal error", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "yaml", Writer: &buf}
		data := &yamlErrStruct{}
		err := f.Print(data)
		if err == nil {
			t.Fatal("expected yaml marshal error, got nil")
		}
		if !strings.Contains(err.Error(), "yaml marshal failed") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("writer error", func(t *testing.T) {
		f := &Formatter{Format: "yaml", Writer: &errWriter{}}
		err := f.Print(map[string]string{"key": "val"})
		if err == nil {
			t.Fatal("expected writer error, got nil")
		}
	})
}

func TestFormatterTOON(t *testing.T) {
	t.Run("regular struct or map compaction", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "toon", Writer: &buf}
		data := []map[string]interface{}{
			{"itemid": "1001", "clock": 1700000000, "value": "42.5"},
			{"itemid": "1001", "clock": 1700000060, "value": "43.1"},
		}
		if err := f.Print(data); err != nil {
			t.Fatalf("failed to print toon: %v", err)
		}
		out := buf.String()
		// Compact JSON shouldn't have indent newlines between keys
		if strings.Contains(out, "\n  \"itemid\"") {
			t.Errorf("expected compact single line json in toon format, got: %s", out)
		}
		if !strings.Contains(out, `"itemid":"1001"`) {
			t.Errorf("expected compacted output with itemid:1001, got: %s", out)
		}
	})

	t.Run("valid json.RawMessage compaction", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "toon", Writer: &buf}
		raw := json.RawMessage(`{
			"host": "server-toon",
			"metrics": [ 1, 2, 3 ]
		}`)
		if err := f.Print(raw); err != nil {
			t.Fatalf("failed to print raw toon: %v", err)
		}
		out := strings.TrimSpace(buf.String())
		expected := `{"host":"server-toon","metrics":[1,2,3]}`
		if out != expected {
			t.Errorf("expected compact '%s', got '%s'", expected, out)
		}
	})

	t.Run("invalid json.RawMessage fallback", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "toon", Writer: &buf}
		raw := json.RawMessage(`not { valid json`)
		if err := f.Print(raw); err != nil {
			t.Fatalf("failed to print invalid raw toon: %v", err)
		}
		out := strings.TrimSpace(buf.String())
		if out != `not { valid json` {
			t.Errorf("expected fallback raw string, got '%s'", out)
		}
	})

	t.Run("toon marshal error", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "toon", Writer: &buf}
		data := unmarshalableType{Fn: func() {}}
		err := f.Print(data)
		if err == nil {
			t.Fatal("expected marshal error, got nil")
		}
	})

	t.Run("writer error", func(t *testing.T) {
		f := &Formatter{Format: "toon", Writer: &errWriter{}}
		err := f.Print(map[string]string{"foo": "bar"})
		if err == nil {
			t.Fatal("expected writer error, got nil")
		}
	})
}

func TestFormatterTable(t *testing.T) {
	t.Run("single map object (key-value table)", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "table", Writer: &buf}
		data := map[string]interface{}{
			"hostid": "10084",
			"name":   "Zabbix server",
			"nested": map[string]interface{}{"sub": "val"},
			"list":   []string{"a", "b"},
			"nilval": nil,
		}
		if err := f.Print(data); err != nil {
			t.Fatalf("failed to print single table: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "KEY") || !strings.Contains(out, "VALUE") {
			t.Errorf("expected table header KEY and VALUE, got: %s", out)
		}
		if !strings.Contains(out, "Zabbix server") {
			t.Errorf("expected table to contain 'Zabbix server', got: %s", out)
		}
		if !strings.Contains(out, `{"sub":"val"}`) {
			t.Errorf("expected nested map to be json formatted, got: %s", out)
		}
	})

	t.Run("non-map non-slice single object fallback to JSON", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "table", Writer: &buf}
		data := "plain string"
		if err := f.Print(data); err != nil {
			t.Fatalf("failed to print primitive table: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, `"plain string"`) {
			t.Errorf("expected json fallback for string, got: %s", out)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "table", Writer: &buf}
		data := []map[string]interface{}{}
		if err := f.Print(data); err != nil {
			t.Fatalf("failed to print empty slice table: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "No items found.") {
			t.Errorf("expected 'No items found.', got: %s", out)
		}
	})

	t.Run("slice of maps with priority headers and extra headers", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "table", Writer: &buf}
		data := []map[string]interface{}{
			{
				"hostid":      "10001",
				"host":        "db-master",
				"name":        "DB Master Node",
				"status":      "0",
				"severity":    "4",
				"description": "Primary database instance",
				"tags":        []map[string]string{{"tag": "env", "value": "prod"}},
				"extra_field": "custom-val",
				"nil_field":   nil,
			},
			{
				"hostid":   "10002",
				"host":     "db-replica",
				"name":     "DB Replica Node",
				"status":   "0",
				"severity": "1",
			},
		}
		if err := f.Print(data); err != nil {
			t.Fatalf("failed to print slice table: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "HOSTID") || !strings.Contains(out, "HOST") || !strings.Contains(out, "SEVERITY") {
			t.Errorf("expected priority headers in table, got: %s", out)
		}
		if !strings.Contains(out, "db-master") || !strings.Contains(out, "db-replica") {
			t.Errorf("expected row contents in table, got: %s", out)
		}
	})

	t.Run("slice of maps with non-map elements in between", func(t *testing.T) {
		sliceWithMixed := []interface{}{
			map[string]interface{}{"hostid": "101", "name": "srv1"},
			"not a map",
			map[string]interface{}{"hostid": "102", "name": "srv2"},
		}
		var buf bytes.Buffer
		f := &Formatter{Format: "table", Writer: &buf}
		if err := f.Print(sliceWithMixed); err != nil {
			t.Fatalf("failed to print mixed slice: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "srv1") || !strings.Contains(out, "srv2") {
			t.Errorf("expected table with valid map rows, got: %s", out)
		}
	})

	t.Run("slice of non-map objects (primitives) fallback to JSON", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "table", Writer: &buf}
		data := []string{"apple", "banana", "cherry"}
		if err := f.Print(data); err != nil {
			t.Fatalf("failed to print string slice table: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, `"apple"`) {
			t.Errorf("expected json fallback for string slice, got: %s", out)
		}
	})

	t.Run("slice with empty map header fallback to JSON", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "table", Writer: &buf}
		data := []map[string]interface{}{
			{},
		}
		if err := f.Print(data); err != nil {
			t.Fatalf("failed to print empty map slice: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "{}") {
			t.Errorf("expected json fallback for empty map slice, got: %s", out)
		}
	})

	t.Run("raw json slice converted to table", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "table", Writer: &buf}
		raw := json.RawMessage(`[{"hostid":"501","name":"web-app"}]`)
		if err := f.Print(raw); err != nil {
			t.Fatalf("failed to print raw json as table: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "HOSTID") || !strings.Contains(out, "web-app") {
			t.Errorf("expected table output from raw json slice, got: %s", out)
		}
	})

	t.Run("invalid raw json falls back to printJSON", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "table", Writer: &buf}
		raw := json.RawMessage(`[invalid json`)
		if err := f.Print(raw); err != nil {
			t.Fatalf("failed to handle invalid raw json: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, `[invalid json`) {
			t.Errorf("expected raw output fallback, got: %s", out)
		}
	})

	t.Run("extractTableHeadersAndRows empty reflect slice", func(t *testing.T) {
		emptySlice := []map[string]interface{}{}
		headers, rows, err := extractTableHeadersAndRows(reflect.ValueOf(emptySlice))
		if err != nil || headers != nil || rows != nil {
			t.Errorf("expected nil headers and rows and nil err for empty reflect slice, got %v, %v, %v", headers, rows, err)
		}
	})
}

func TestFormatSeverity(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{"5", "5 (Disaster)"},
		{"disaster", "5 (Disaster)"},
		{"4", "4 (High)"},
		{"high", "4 (High)"},
		{"3", "3 (Average)"},
		{"average", "3 (Average)"},
		{"2", "2 (Warning)"},
		{"warning", "2 (Warning)"},
		{"1", "1 (Info)"},
		{"information", "1 (Info)"},
		{"0", "0 (Not classified)"},
		{"unknown", "unknown (Not classified)"},
	}

	for _, tt := range tests {
		res := FormatSeverity(tt.input)
		if !strings.Contains(res, tt.expected) {
			t.Errorf("FormatSeverity(%v) = %q; want it to contain %q", tt.input, res, tt.expected)
		}
	}
}

func TestPrintErrorEnvelope(t *testing.T) {
	var buf bytes.Buffer
	errData := map[string]interface{}{
		"error": map[string]string{
			"code":    "SAFETY_VIOLATION",
			"message": "Write operation blocked by safety middleware",
		},
	}

	PrintErrorEnvelope(&buf, errData)
	out := buf.String()

	if !strings.Contains(out, "SAFETY_VIOLATION") || !strings.Contains(out, "Write operation blocked") {
		t.Errorf("expected json error envelope output, got: %s", out)
	}
}

func TestPrintResourceSchemas(t *testing.T) {
	t.Run("problem standard schema with age, severity, status, and ack", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "table", Writer: &buf}

		data := []map[string]interface{}{
			{
				"eventid":      "10492",
				"host":         "web-prod-01",
				"name":         "High CPU utilization (> 95%)",
				"severity":     "4",
				"status":       "PROBLEM",
				"clock":        time.Now().Unix() - 12*60,
				"acknowledged": "0",
			},
			{
				"eventid":      "10501",
				"host":         "db-primary",
				"name":         "PostgreSQL replication lag > 15s",
				"severity":     "4",
				"status":       "PROBLEM",
				"clock":        time.Now().Unix() - 48*60,
				"acknowledged": "1",
			},
		}

		err := f.PrintResource(data, "problem", nil, false)
		if err != nil {
			t.Fatalf("failed to print problem resource table: %v", err)
		}
		out := buf.String()

		expectedHeaders := []string{"EVENTID", "HOST", "PROBLEM", "SEVERITY", "STATUS", "AGE", "ACK"}
		for _, h := range expectedHeaders {
			if !strings.Contains(out, h) {
				t.Errorf("expected table header %q in output:\n%s", h, out)
			}
		}

		if !strings.Contains(out, "10492") || !strings.Contains(out, "web-prod-01") || !strings.Contains(out, "12m") {
			t.Errorf("expected row 1 data in table output:\n%s", out)
		}
		if !strings.Contains(out, "10501") || !strings.Contains(out, "db-primary") || !strings.Contains(out, "48m") {
			t.Errorf("expected row 2 data in table output:\n%s", out)
		}
	})

	t.Run("problem custom fields selection", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "table", Writer: &buf}

		data := []map[string]interface{}{
			{
				"eventid":  "6123",
				"host":     "web-prod-01",
				"name":     "ICMP Ping unavailable",
				"severity": "4",
				"clock":    time.Now().Unix() - 5*60,
			},
		}

		err := f.PrintResource(data, "problem", []string{"eventid", "name", "clock"}, false)
		if err != nil {
			t.Fatalf("failed to print custom fields: %v", err)
		}
		out := buf.String()

		if !strings.Contains(out, "EVENTID") || !strings.Contains(out, "NAME") || !strings.Contains(out, "CLOCK") {
			t.Errorf("expected EVENTID, NAME, CLOCK headers in output:\n%s", out)
		}
		if strings.Contains(out, "SEVERITY") {
			t.Errorf("did not expect SEVERITY header when custom fields selected:\n%s", out)
		}
		if !strings.Contains(out, "5m") {
			t.Errorf("expected clock to be formatted as relative age '5m', got:\n%s", out)
		}
	})

	t.Run("problem all-fields / wide output", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "table", Writer: &buf}

		data := []map[string]interface{}{
			{
				"eventid":   "6123",
				"name":      "ICMP Ping",
				"ns":        "342714930",
				"r_ns":      "0",
				"custom_id": "xyz",
			},
		}

		err := f.PrintResource(data, "problem", nil, true)
		if err != nil {
			t.Fatalf("failed to print all-fields: %v", err)
		}
		out := buf.String()

		if !strings.Contains(out, "CUSTOM") || !strings.Contains(out, "NS") || !strings.Contains(out, "xyz") {
			t.Errorf("expected all raw fields in wide output:\n%s", out)
		}
	})

	t.Run("problem invalid field returns error", func(t *testing.T) {
		var buf bytes.Buffer
		f := &Formatter{Format: "table", Writer: &buf}

		data := []map[string]interface{}{
			{
				"eventid": "6123",
				"name":    "ICMP Ping",
			},
		}

		err := f.PrintResource(data, "problem", []string{"eventid", "nonexistent_column"}, false)
		if err == nil {
			t.Fatal("expected error for nonexistent column, got nil")
		}
		if !strings.Contains(err.Error(), `field "nonexistent_column" does not exist`) {
			t.Errorf("expected nonexistent column error message, got: %v", err)
		}
	})
}

