package scaffold

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// cobraHelpTreeModule is the release line generated CLIs depend on.
const cobraHelpTreeModule = "github.com/alexgorbatchev/cobra-help-tree/v4"

var embeddedSkillRegex = regexp.MustCompile(`(?m)^//go:embed\s.*SKILL\.md`)

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
	HasSkill           bool        `json:"has_skill,omitempty"`
	HasGoSource        bool        `json:"has_go_source,omitempty"`
	HasGoTest          bool        `json:"has_go_test,omitempty"`
	Issues             []string    `json:"issues"`
	Score              float64     `json:"score"`
}

// findGoSource returns the first non-test Go file under cmd/ or internal/ whose
// source satisfies match, or "" when none does.
func findGoSource(projectDir string, match func(source string) bool) string {
	found := ""
	for _, root := range []string{"cmd", "internal"} {
		// Unreadable paths count as source that does not match, so the walk has no error to report.
		_ = filepath.WalkDir(filepath.Join(projectDir, root), func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
				return nil
			}
			data, err := os.ReadFile(p)
			if err != nil {
				return nil
			}
			if match(string(data)) {
				found = p
				return fs.SkipAll
			}
			return nil
		})
		if found != "" {
			break
		}
	}
	return found
}

// usesAgentMode reports whether non-test Go source under cmd/ or internal/ reads
// AGENT, directly or through cobrahelptree.IsAgentMode.
func usesAgentMode(projectDir string) bool {
	return findGoSource(projectDir, func(source string) bool {
		return strings.Contains(source, `os.Getenv("AGENT")`) || strings.Contains(source, "cobrahelptree.IsAgentMode()")
	}) != ""
}

// requiresCurrentHelpTree reports whether go.mod names the module exactly, so an
// older major version does not count.
func requiresCurrentHelpTree(goMod string) bool {
	for line := range strings.Lines(goMod) {
		if slices.Contains(strings.Fields(line), cobraHelpTreeModule) {
			return true
		}
	}
	return false
}

// checkSkill returns what keeps each binary in cmd/ from having its skill at
// skills/<binary>/SKILL.md in the repository root, named for the binary and kept
// out of the executable.
func checkSkill(projectDir, module string) []string {
	var issues []string
	if p := findGoSource(projectDir, embeddedSkillRegex.MatchString); p != "" {
		rel, err := filepath.Rel(projectDir, p)
		if err != nil {
			rel = p
		}
		issues = append(issues, fmt.Sprintf("%s embeds SKILL.md: keep the skill in the repository and print its install command", filepath.ToSlash(rel)))
	}

	entries, err := os.ReadDir(filepath.Join(projectDir, "cmd"))
	if err != nil {
		return append(issues, "missing cmd/ directory, so no binary has a skill")
	}
	root := filepath.Join(projectDir, filepath.FromSlash(repositoryRoot(moduleSubdirectory(module))))
	binaries := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		binaries++
		data, err := os.ReadFile(filepath.Join(root, "skills", e.Name(), "SKILL.md"))
		if err != nil {
			issues = append(issues, fmt.Sprintf("missing skills/%s/SKILL.md at the repository root", e.Name()))
			continue
		}
		name := regexp.MustCompile(`(?m)^name:\s*"?` + regexp.QuoteMeta(e.Name()) + `"?\s*$`)
		if !name.Match(data) {
			issues = append(issues, fmt.Sprintf("skills/%s/SKILL.md must set name to %s", e.Name(), e.Name()))
		}
	}
	if binaries == 0 {
		issues = append(issues, "missing cmd/<binary> directory, so no binary has a skill")
	}
	return issues
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
			switch {
			case requiresCurrentHelpTree(string(data)):
				res.HasCobraHelpTree = true
			case strings.Contains(string(data), "cobra-help-tree"):
				res.Issues = append(res.Issues, "go.mod must require "+cobraHelpTreeModule)
			default:
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

		if usesAgentMode(absDir) {
			res.HasAgentMode = true
		} else {
			res.Issues = append(res.Issues, "missing AGENT=1 dual-mode helper")
		}

		skillIssues := checkSkill(absDir, res.ModuleName)
		res.HasSkill = len(skillIssues) == 0
		res.Issues = append(res.Issues, skillIssues...)

		totalChecks := 12.0
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
		if res.HasSkill {
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
