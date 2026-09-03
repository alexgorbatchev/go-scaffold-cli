package scaffold

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// InspectResult represents standards compliance checks on a Go CLI or library project.
type InspectResult struct {
	Type               ProjectType `json:"type"`
	ProjectDir         string      `json:"project_dir"`
	HasGoMod           bool        `json:"has_go_mod"`
	ModuleName         string      `json:"module_name,omitempty"`
	GoVersion          string      `json:"go_version,omitempty"`
	HasJustfile        bool        `json:"has_justfile"`
	HasGitignore       bool        `json:"has_gitignore"`
	HasGoreleaser      bool        `json:"has_goreleaser,omitempty"`
	HasCIWorkflow      bool        `json:"has_ci_workflow"`
	HasReleaseWorkflow bool        `json:"has_release_workflow,omitempty"`
	HasLicense         bool        `json:"has_license"`
	HasReadme          bool        `json:"has_readme"`
	HasAgentsDoc       bool        `json:"has_agents_doc"`
	HasAgentMode       bool        `json:"has_agent_mode,omitempty"`
	HasCobraHelpTree   bool        `json:"has_cobra_help_tree,omitempty"`
	HasGoSource        bool        `json:"has_go_source,omitempty"`
	HasGoTest          bool        `json:"has_go_test,omitempty"`
	Issues             []string    `json:"issues"`
	Score              float64     `json:"score"`
}

// Inspect checks an existing directory against Go CLI or library project standards.
func Inspect(targetDir string, pType ProjectType) (*InspectResult, error) {
	absDir, err := filepath.Abs(targetDir)
	if err != nil {
		absDir = targetDir
	}

	info, err := os.Stat(absDir)
	if err != nil {
		return nil, fmt.Errorf("stat target dir %q: %w", absDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("target path %q is not a directory", absDir)
	}

	if pType == "" {
		pType = ProjectTypeCLI
	}

	res := &InspectResult{
		Type:       pType,
		ProjectDir: absDir,
		Issues:     make([]string, 0),
	}

	// 1. go.mod
	goModPath := filepath.Join(absDir, "go.mod")
	if data, err := os.ReadFile(goModPath); err == nil {
		res.HasGoMod = true
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "module ") {
				res.ModuleName = strings.TrimSpace(strings.TrimPrefix(line, "module"))
			} else if strings.HasPrefix(line, "go ") {
				res.GoVersion = strings.TrimSpace(strings.TrimPrefix(line, "go"))
			}
		}
		if pType == ProjectTypeCLI {
			if strings.Contains(string(data), "cobra-help-tree") {
				res.HasCobraHelpTree = true
			} else {
				res.Issues = append(res.Issues, "missing cobra-help-tree dependency in go.mod")
			}
		}
	} else {
		res.Issues = append(res.Issues, "missing go.mod")
		if pType == ProjectTypeCLI {
			res.Issues = append(res.Issues, "missing cobra-help-tree dependency in go.mod")
		}
	}

	// 2. justfile
	justfilePath := filepath.Join(absDir, "justfile")
	if _, err := os.Stat(justfilePath); err == nil {
		res.HasJustfile = true
	} else {
		if _, err := os.Stat(filepath.Join(absDir, "Justfile")); err == nil {
			res.HasJustfile = true
		} else {
			res.Issues = append(res.Issues, "missing justfile")
		}
	}

	// 3. .gitignore
	if _, err := os.Stat(filepath.Join(absDir, ".gitignore")); err == nil {
		res.HasGitignore = true
	} else {
		res.Issues = append(res.Issues, "missing .gitignore")
	}

	// 4. CI Workflow
	if _, err := os.Stat(filepath.Join(absDir, ".github", "workflows", "ci.yml")); err == nil {
		res.HasCIWorkflow = true
	} else if _, err := os.Stat(filepath.Join(absDir, ".github", "workflows", "ci.yaml")); err == nil {
		res.HasCIWorkflow = true
	} else {
		res.Issues = append(res.Issues, "missing .github/workflows/ci.yml")
	}

	// 5. LICENSE
	if _, err := os.Stat(filepath.Join(absDir, "LICENSE")); err == nil {
		res.HasLicense = true
	} else {
		res.Issues = append(res.Issues, "missing LICENSE")
	}

	// 6. README.md
	if _, err := os.Stat(filepath.Join(absDir, "README.md")); err == nil {
		res.HasReadme = true
	} else {
		res.Issues = append(res.Issues, "missing README.md")
	}

	// 7. AGENTS.md
	if _, err := os.Stat(filepath.Join(absDir, "AGENTS.md")); err == nil {
		res.HasAgentsDoc = true
	} else {
		res.Issues = append(res.Issues, "missing AGENTS.md")
	}

	if pType == ProjectTypeCLI {
		// CLI-specific checks: .goreleaser, release workflow, agent mode
		if _, err := os.Stat(filepath.Join(absDir, ".goreleaser.yml")); err == nil {
			res.HasGoreleaser = true
		} else if _, err := os.Stat(filepath.Join(absDir, ".goreleaser.yaml")); err == nil {
			res.HasGoreleaser = true
		} else {
			res.Issues = append(res.Issues, "missing .goreleaser.yml")
		}

		if _, err := os.Stat(filepath.Join(absDir, ".github", "workflows", "release.yml")); err == nil {
			res.HasReleaseWorkflow = true
		} else if _, err := os.Stat(filepath.Join(absDir, ".github", "workflows", "release.yaml")); err == nil {
			res.HasReleaseWorkflow = true
		} else {
			res.Issues = append(res.Issues, "missing .github/workflows/release.yml")
		}

		if _, err := os.Stat(filepath.Join(absDir, "internal", "agent", "agent.go")); err == nil {
			res.HasAgentMode = true
		} else {
			_ = filepath.Walk(filepath.Join(absDir, "internal"), func(p string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() && strings.HasSuffix(p, ".go") {
					if d, err := os.ReadFile(p); err == nil && strings.Contains(string(d), `os.Getenv("AGENT")`) {
						res.HasAgentMode = true
					}
				}
				return nil
			})
			if !res.HasAgentMode {
				res.Issues = append(res.Issues, "missing AGENT=1 dual-mode helper")
			}
		}

		totalChecks := 11.0
		passed := 0.0
		if res.HasGoMod {
			passed++
		}
		if res.HasJustfile {
			passed++
		}
		if res.HasGitignore {
			passed++
		}
		if res.HasGoreleaser {
			passed++
		}
		if res.HasCIWorkflow {
			passed++
		}
		if res.HasReleaseWorkflow {
			passed++
		}
		if res.HasLicense {
			passed++
		}
		if res.HasReadme {
			passed++
		}
		if res.HasAgentsDoc {
			passed++
		}
		if res.HasAgentMode {
			passed++
		}
		if res.HasCobraHelpTree {
			passed++
		}

		res.Score = (passed / totalChecks) * 10.0
	} else {
		// Library-specific checks: root .go sources and tests
		entries, err := os.ReadDir(absDir)
		if err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") {
					if strings.HasSuffix(e.Name(), "_test.go") {
						res.HasGoTest = true
					} else {
						res.HasGoSource = true
					}
				}
			}
		}
		if !res.HasGoSource {
			res.Issues = append(res.Issues, "missing root package Go source file (*.go)")
		}
		if !res.HasGoTest {
			res.Issues = append(res.Issues, "missing root package Go test file (*_test.go)")
		}

		totalChecks := 9.0
		passed := 0.0
		if res.HasGoMod {
			passed++
		}
		if res.HasJustfile {
			passed++
		}
		if res.HasGitignore {
			passed++
		}
		if res.HasCIWorkflow {
			passed++
		}
		if res.HasLicense {
			passed++
		}
		if res.HasReadme {
			passed++
		}
		if res.HasAgentsDoc {
			passed++
		}
		if res.HasGoSource {
			passed++
		}
		if res.HasGoTest {
			passed++
		}

		res.Score = (passed / totalChecks) * 10.0
	}

	return res, nil
}
