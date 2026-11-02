package main

import (
	"runtime/debug"
	"testing"
)

func TestResolveVersion(t *testing.T) {
	tests := []struct {
		name     string
		injected string
		info     *debug.BuildInfo
		want     string
	}{
		{"release build", "1.2.3", &debug.BuildInfo{Main: debug.Module{Version: "v9.9.9"}}, "1.2.3"},
		{"go install at a tag", "", &debug.BuildInfo{Main: debug.Module{Version: "v1.2.3"}}, "1.2.3"},
		{"source build", "", &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, "dev"},
		{"no module version", "", &debug.BuildInfo{}, "dev"},
		{"no build info", "", nil, "dev"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveVersion(tt.injected, tt.info); got != tt.want {
				t.Fatalf("resolveVersion = %q, want %q", got, tt.want)
			}
		})
	}
}
