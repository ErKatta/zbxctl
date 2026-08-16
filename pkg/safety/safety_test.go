package safety

import (
	"testing"
)

func TestCheckSafety(t *testing.T) {
	tests := []struct {
		name        string
		level       SafetyLevel
		context     string
		method      string
		params      interface{}
		force       bool
		expectError bool
	}{
		{
			name:        "Readonly allows .get",
			level:       LevelReadOnly,
			context:     "prod",
			method:      "host.get",
			params:      nil,
			force:       false,
			expectError: false,
		},
		{
			name:        "Readonly blocks .create",
			level:       LevelReadOnly,
			context:     "prod",
			method:      "host.create",
			params:      map[string]interface{}{"host": "test"},
			force:       false,
			expectError: true,
		},
		{
			name:        "Readwrite-mine blocks untagged create",
			level:       LevelReadWriteMine,
			context:     "dev",
			method:      "host.create",
			params:      map[string]interface{}{"host": "test"},
			force:       false,
			expectError: true,
		},
		{
			name:        "Readwrite-mine allows tagged create",
			level:       LevelReadWriteMine,
			context:     "dev",
			method:      "host.create",
			params:      map[string]interface{}{"host": "test", "tags": []map[string]string{{"tag": "zbxctl", "value": "true"}}},
			force:       false,
			expectError: false,
		},
		{
			name:        "Readwrite-all blocks bulk delete without force",
			level:       LevelReadWriteAll,
			context:     "prod",
			method:      "host.delete",
			params:      []string{"10001", "10002"},
			force:       false,
			expectError: true,
		},
		{
			name:        "Readwrite-all allows bulk delete with force",
			level:       LevelReadWriteAll,
			context:     "prod",
			method:      "host.delete",
			params:      []string{"10001", "10002"},
			force:       true,
			expectError: false,
		},
		{
			name:        "Dangerously-unrestricted allows any operation",
			level:       LevelDangerouslyUnrestricted,
			context:     "lab",
			method:      "host.delete",
			params:      []string{"10001", "10002"},
			force:       false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckSafety(tt.level, tt.context, tt.method, tt.params, tt.force)
			if tt.expectError && err == nil {
				t.Errorf("expected error for method %s under level %s, got nil", tt.method, tt.level)
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error for method %s under level %s, got %v", tt.method, tt.level, err)
			}
		})
	}
}
