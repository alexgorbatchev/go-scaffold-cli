package scaffold

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCleanPreservesSessionStorage(t *testing.T) {
	for _, kind := range []ProjectType{ProjectTypeCLI, ProjectTypeLib} {
		t.Run(string(kind), func(t *testing.T) {
			dir := t.TempDir()
			_, err := Generate(Options{
				Type: kind, TargetDir: dir, Name: "sample", Binary: "sample",
				PkgName: "sample", Module: "example.com/sample",
			})
			if err != nil {
				t.Fatal(err)
			}
			checkCleanStorage(t, dir)
			cmd := exec.Command("git", "init", "--quiet")
			cmd.Dir = dir
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("git init: %v\n%s", err, out)
			}
			for _, path := range []string{".tmp/session/file", ".cache/build/file", ".state/run/file"} {
				cmd := exec.Command("git", "check-ignore", "--quiet", path)
				cmd.Dir = dir
				if err := cmd.Run(); err != nil {
					t.Errorf("generated project does not ignore %s: %v", path, err)
				}
			}
		})
	}
	t.Run("repository", func(t *testing.T) {
		dir := t.TempDir()
		data, err := os.ReadFile("../../justfile")
		if err != nil {
			t.Fatal(err)
		}
		writeInspectFile(t, dir, "justfile", string(data))
		checkCleanStorage(t, dir)
	})
}

func checkCleanStorage(t *testing.T, dir string) {
	t.Helper()
	outputs := []string{"bin/app", "dist/archive", "coverage.out"}
	retained := []string{".tmp/session/file", ".cache/build/file", ".state/run/file"}
	for _, path := range append(outputs, retained...) {
		writeInspectFile(t, dir, path, "keep")
	}
	cmd := exec.Command("just", "clean")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("just clean: %v\n%s", err, out)
	}
	for _, path := range outputs {
		if _, err := os.Stat(filepath.Join(dir, path)); !os.IsNotExist(err) {
			t.Errorf("clean left %s: %v", path, err)
		}
	}
	for _, path := range retained {
		if data, err := os.ReadFile(filepath.Join(dir, path)); err != nil || string(data) != "keep" {
			t.Errorf("clean damaged %s: %q, %v", path, data, err)
		}
	}
}
