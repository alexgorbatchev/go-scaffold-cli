package agent

import (
	"testing"
)

func TestIsAgentMode(t *testing.T) {
	tests := []struct {
		name     string
		envVal   string
		setEnv   bool
		expected bool
	}{
		{"unset", "", false, false},
		{"empty", "", true, false},
		{"zero", "0", true, false},
		{"false", "false", true, false},
		{"no", "no", true, false},
		{"one", "1", true, true},
		{"true", "true", true, true},
		{"yes", "yes", true, true},
		{"TRUE uppercase", "TRUE", true, true},
		{"YES uppercase", "YES", true, true},
		{"with spaces", "  1  ", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv("AGENT", tt.envVal)
			} else {
				t.Setenv("AGENT", "")
			}
			actual := IsAgentMode()
			if actual != tt.expected {
				t.Errorf("IsAgentMode() = %v, want %v (env=%q)", actual, tt.expected, tt.envVal)
			}
		})
	}
}
