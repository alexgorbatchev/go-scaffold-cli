package scaffold

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func coverageCommand(t *testing.T, target string) string {
	t.Helper()
	workflow, err := os.ReadFile(filepath.Join(target, ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	_, block, ok := strings.Cut(string(workflow), "      - name: Check coverage\n        run: |\n")
	if !ok {
		t.Fatal("generated CI has no coverage gate")
	}
	var lines []string
	for line := range strings.SplitSeq(block, "\n") {
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "          ") {
			break
		}
		lines = append(lines, strings.TrimPrefix(line, "          "))
	}
	return strings.Join(lines, "\n")
}

func checkGeneratedCoverage(t *testing.T, target string) {
	t.Helper()
	cmd := exec.Command("bash", "-e", "-o", "pipefail", "-c", coverageCommand(t, target))
	cmd.Dir = target
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generated coverage gate failed: %v\n%s", err, out)
	} else {
		t.Logf("generated coverage gate:\n%s", out)
	}
}

func TestGenerate_CICoverageGate(t *testing.T) {
	for _, kind := range []ProjectType{ProjectTypeCLI, ProjectTypeLib, "repository"} {
		t.Run(string(kind), func(t *testing.T) {
			target := filepath.Join(t.TempDir(), "project")
			if kind == "repository" {
				target = filepath.Join("..", "..")
			} else {
				_, err := Generate(Options{
					Type: kind, TargetDir: target, Name: "project", Binary: "project",
					Module: "example.com/project", Author: "Test Author",
				})
				if err != nil {
					t.Fatal(err)
				}
			}
			command := coverageCommand(t, target)
			for _, tt := range []struct {
				name    string
				profile string
				wantErr bool
			}{
				{"below", "mode: atomic\na/a.go:1.1,2.1 89 7\nb/b.go:1.1,2.1 11 0\n", true},
				{"boundary", "mode: atomic\na/a.go:1.1,2.1 90 7\nb/b.go:1.1,2.1 10 0\n", false},
				{"above", "mode: atomic\na/a.go:1.1,2.1 91 7\nb/b.go:1.1,2.1 9 0\n", false},
				{"rounding", "mode: atomic\na/a.go:1.1,2.1 8999 1\nb/b.go:1.1,2.1 1001 0\n", true},
				{"empty", "mode: atomic\n", true},
				{"missing", "", true},
			} {
				t.Run(tt.name, func(t *testing.T) {
					dir := t.TempDir()
					if tt.profile != "" {
						if err := os.WriteFile(filepath.Join(dir, "coverage.out"), []byte(tt.profile), 0644); err != nil {
							t.Fatal(err)
						}
					}
					cmd := exec.Command("bash", "-e", "-o", "pipefail", "-c", command)
					cmd.Dir = dir
					out, err := cmd.CombinedOutput()
					if (err != nil) != tt.wantErr {
						t.Fatalf("coverage gate error = %v, wantErr %v\n%s", err, tt.wantErr, out)
					}
				})
			}
		})
	}
}
