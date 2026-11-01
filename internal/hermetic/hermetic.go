// Package hermetic keeps a test process off the network.
package hermetic

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Run executes m with the GitHub CLI replaced by a failing stub and the Go
// module proxy and checksum database turned off, so generating a project reads
// only the local module cache. Run `go mod download` before the tests.
func Run(m *testing.M) int {
	dir, err := os.MkdirTemp("", "hermetic")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "gh"), []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	for name, value := range map[string]string{
		"PATH":    dir + string(os.PathListSeparator) + os.Getenv("PATH"),
		"GOPROXY": "off",
		"GOSUMDB": "off",
	} {
		if err := os.Setenv(name, value); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	return m.Run()
}
