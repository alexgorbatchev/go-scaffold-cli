package scaffold

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestCleanPackageName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"godeps", "godeps"},
		{"my-lib", "mylib"},
		{"go-chromaprint", "gochromaprint"},
		{"123pkg", "pkg123pkg"},
		{"special@_chars!", "specialchars"},
		{"", "pkg"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := CleanPackageName(tt.input)
			if got != tt.want {
				t.Errorf("CleanPackageName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDynamicDetection(t *testing.T) {
	username := DetectUsername()
	if username == "" {
		t.Error("expected non-empty username")
	}

	owner := DetectGitHubOwner()
	if owner == "" {
		t.Error("expected non-empty owner")
	}

	author := DetectAuthor()
	if author == "" {
		t.Error("expected non-empty author")
	}
}

func TestDynamicDetection_Fallbacks(t *testing.T) {
	origLookup := lookupUserCurrent
	origRun := runCommand
	defer func() {
		lookupUserCurrent = origLookup
		runCommand = origRun
	}()

	// 1. user.Current fails, env USER is set
	lookupUserCurrent = func() (*user.User, error) {
		return nil, errors.New("lookup failed")
	}
	t.Setenv("USER", "testuser1")
	t.Setenv("USERNAME", "")
	if u := DetectUsername(); u != "testuser1" {
		t.Errorf("expected 'testuser1', got %q", u)
	}

	// 2. env USER empty, env USERNAME set
	t.Setenv("USER", "")
	t.Setenv("USERNAME", "testuser2")
	if u := DetectUsername(); u != "testuser2" {
		t.Errorf("expected 'testuser2', got %q", u)
	}

	// 3. env vars empty, whoami command succeeds
	t.Setenv("USER", "")
	t.Setenv("USERNAME", "")
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "whoami" {
			return []byte("whoamiuser\n"), nil
		}
		return nil, errors.New("command failed")
	}
	if u := DetectUsername(); u != "whoamiuser" {
		t.Errorf("expected 'whoamiuser', got %q", u)
	}

	// 4. all username methods fail -> fallback "user"
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, errors.New("command failed")
	}
	if u := DetectUsername(); u != "user" {
		t.Errorf("expected 'user', got %q", u)
	}

	// 5. DetectGitHubOwner fallbacks:
	// a. gh succeeds
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "gh" {
			return []byte("ghowner\n"), nil
		}
		return nil, errors.New("not found")
	}
	if o := DetectGitHubOwner(); o != "ghowner" {
		t.Errorf("expected 'ghowner', got %q", o)
	}

	// b. gh fails, git config github.user succeeds
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "git" && len(args) == 2 && args[1] == "github.user" {
			return []byte("gitghuser\n"), nil
		}
		return nil, errors.New("not found")
	}
	if o := DetectGitHubOwner(); o != "gitghuser" {
		t.Errorf("expected 'gitghuser', got %q", o)
	}

	// c. git config user.username succeeds
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "git" && len(args) == 2 && args[1] == "user.username" {
			return []byte("gituser\n"), nil
		}
		return nil, errors.New("not found")
	}
	if o := DetectGitHubOwner(); o != "gituser" {
		t.Errorf("expected 'gituser', got %q", o)
	}

	// d. all owner commands fail -> falls back to DetectUsername
	t.Setenv("USER", "fallbackuser")
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, errors.New("not found")
	}
	if o := DetectGitHubOwner(); o != "fallbackuser" {
		t.Errorf("expected 'fallbackuser', got %q", o)
	}

	// 6. DetectAuthor fallbacks:
	// a. git config user.name succeeds
	runCommand = func(name string, args ...string) ([]byte, error) {
		if name == "git" && len(args) == 2 && args[1] == "user.name" {
			return []byte("Git Author Name\n"), nil
		}
		return nil, errors.New("not found")
	}
	if a := DetectAuthor(); a != "Git Author Name" {
		t.Errorf("expected 'Git Author Name', got %q", a)
	}

	// b. git config fails, user.Current Name succeeds
	runCommand = func(name string, args ...string) ([]byte, error) {
		return nil, errors.New("not found")
	}
	lookupUserCurrent = func() (*user.User, error) {
		return &user.User{Name: "OS User Name", Username: "osuser"}, nil
	}
	if a := DetectAuthor(); a != "OS User Name" {
		t.Errorf("expected 'OS User Name', got %q", a)
	}

	// c. git config and user.Current Name empty -> falls back to username
	lookupUserCurrent = func() (*user.User, error) {
		return nil, errors.New("failed")
	}
	t.Setenv("USER", "authoruser")
	if a := DetectAuthor(); a != "authoruser" {
		t.Errorf("expected 'authoruser', got %q", a)
	}
}

func TestDefaultOptions(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}

	optsCLI := DefaultOptions("", ProjectTypeCLI)
	if optsCLI.TargetDir != cwd {
		t.Errorf("DefaultOptions(\"\").TargetDir = %q, want %q", optsCLI.TargetDir, cwd)
	}
	if optsCLI.Type != ProjectTypeCLI {
		t.Errorf("DefaultOptions.Type = %q, want %q", optsCLI.Type, ProjectTypeCLI)
	}

	// Test with empty type (defaults to ProjectTypeCLI)
	optsEmptyType := DefaultOptions("rel/path", "")
	if optsEmptyType.Type != ProjectTypeCLI {
		t.Errorf("expected default Type ProjectTypeCLI, got %q", optsEmptyType.Type)
	}

	optsCLI.Normalize()
	expectedName := filepath.Base(cwd)
	if optsCLI.Name != expectedName {
		t.Errorf("Normalize().Name = %q, want %q", optsCLI.Name, expectedName)
	}
	if optsCLI.Author == "" {
		t.Error("expected non-empty Author in Normalize()")
	}
	if !strings.HasPrefix(optsCLI.Module, "github.com/") {
		t.Errorf("expected Module starting with 'github.com/', got %q", optsCLI.Module)
	}

	optsLib := DefaultOptions("/tmp/my-super-lib", ProjectTypeLib)
	optsLib.Normalize()
	if optsLib.Name != "my-super-lib" {
		t.Errorf("optsLib.Name = %q, want 'my-super-lib'", optsLib.Name)
	}
	if optsLib.PkgName != "mysuperlib" {
		t.Errorf("optsLib.PkgName = %q, want 'mysuperlib'", optsLib.PkgName)
	}
	if !strings.Contains(optsLib.Description, "Go library for my-super-lib") {
		t.Errorf("optsLib.Description = %q", optsLib.Description)
	}

	// Root path defaults
	optsRootLib := DefaultOptions("/", ProjectTypeLib)
	optsRootLib.Normalize()
	if optsRootLib.Name != "mylib" {
		t.Errorf("optsRootLib.Name = %q, want 'mylib'", optsRootLib.Name)
	}

	optsRootCLI := DefaultOptions("/", ProjectTypeCLI)
	optsRootCLI.Normalize()
	if optsRootCLI.Name != "my-cli" {
		t.Errorf("optsRootCLI.Name = %q, want 'my-cli'", optsRootCLI.Name)
	}

	// Test Normalize with empty fields
	emptyOpts := Options{}
	emptyOpts.Normalize()
	if emptyOpts.TargetDir == "" || emptyOpts.Name == "" || emptyOpts.Author == "" || emptyOpts.Year == 0 {
		t.Errorf("Normalize failed on empty opts: %+v", emptyOpts)
	}
}

func TestDeriveNames(t *testing.T) {
	owner := DetectGitHubOwner()

	tests := []struct {
		name        string
		pType       ProjectType
		targetDir   string
		inputName   string
		inputBinary string
		inputPkg    string
		inputModule string
		wantBinary  string
		wantPkg     string
		wantModule  string
	}{
		{
			name:        "cli suffix trimmed for binary",
			pType:       ProjectTypeCLI,
			targetDir:   "/tmp/sample-track-cli",
			inputName:   "sample-track-cli",
			inputBinary: "",
			inputModule: "",
			wantBinary:  "sample-track",
			wantModule:  "github.com/" + owner + "/sample-track-cli",
		},
		{
			name:        "plain lib name",
			pType:       ProjectTypeLib,
			targetDir:   "/tmp/godeps",
			inputName:   "godeps",
			inputModule: "",
			wantPkg:     "godeps",
			wantModule:  "github.com/" + owner + "/godeps",
		},
		{
			name:        "custom binary and module preserved",
			pType:       ProjectTypeCLI,
			targetDir:   "/tmp/custom",
			inputName:   "custom-project-cli",
			inputBinary: "custom-bin",
			inputModule: "github.com/example/custom",
			wantBinary:  "custom-bin",
			wantModule:  "github.com/example/custom",
		},
		{
			name:        "custom lib pkg and module preserved",
			pType:       ProjectTypeLib,
			targetDir:   "/tmp/custom-lib",
			inputName:   "custom-lib",
			inputPkg:    "custompkg",
			inputModule: "github.com/example/custom-lib",
			wantPkg:     "custompkg",
			wantModule:  "github.com/example/custom-lib",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultOptions(tt.targetDir, tt.pType)
			if tt.inputName != "" {
				opts.Name = tt.inputName
			}
			opts.Binary = tt.inputBinary
			opts.PkgName = tt.inputPkg
			opts.Module = tt.inputModule
			opts.Normalize()
			if tt.pType == ProjectTypeCLI && opts.Binary != tt.wantBinary {
				t.Errorf("Normalize().Binary = %q, want %q", opts.Binary, tt.wantBinary)
			}
			if tt.pType == ProjectTypeLib && opts.PkgName != tt.wantPkg {
				t.Errorf("Normalize().PkgName = %q, want %q", opts.PkgName, tt.wantPkg)
			}
			if opts.Module != tt.wantModule {
				t.Errorf("Normalize().Module = %q, want %q", opts.Module, tt.wantModule)
			}
		})
	}
}

func TestGenerate_CLIDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "demo-cli")

	opts := Options{
		Type:        ProjectTypeCLI,
		TargetDir:   target,
		Name:        "demo-cli",
		Binary:      "demo",
		Module:      "github.com/dj/demo-cli",
		Description: "A demo CLI utility",
		DryRun:      true,
		InitGit:     false,
		RunTidy:     false,
	}

	res, err := Generate(opts)
	if err != nil {
		t.Fatalf("Generate dry run failed: %v", err)
	}

	if len(res.Files) == 0 {
		t.Fatal("expected generated file list in dry run result")
	}

	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf("expected target dir %q not to exist in dry-run mode", target)
	}
}

func TestGenerate_CLIAgentsRequiresWrappedHelp(t *testing.T) {
	target := filepath.Join(t.TempDir(), "demo-cli")

	opts := Options{
		Type:        ProjectTypeCLI,
		TargetDir:   target,
		Name:        "demo-cli",
		Binary:      "demo",
		Module:      "github.com/dj/demo-cli",
		Description: "A demo CLI utility",
	}

	if _, err := Generate(opts); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	agents, err := os.ReadFile(filepath.Join(target, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read generated AGENTS.md: %v", err)
	}
	if !strings.Contains(string(agents), "wrapped to the terminal width and never truncated") {
		t.Error("generated AGENTS.md must require wrapped help")
	}
	for _, stale := range []string{"trimmed", "ellipsis", "cobra-help-tree/v2", "cobra-help-tree/v3", "go:embed", "cmd/demo/SKILL.md"} {
		if strings.Contains(string(agents), stale) {
			t.Errorf("generated AGENTS.md must not mention %q", stale)
		}
	}
	for _, want := range []string{"cobra-help-tree/v4", "skills/demo/SKILL.md", "Update the skill in the same change as any"} {
		if !strings.Contains(string(agents), want) {
			t.Errorf("generated AGENTS.md must mention %q", want)
		}
	}
}

// stubGitHubOwner makes owner detection offline.
func stubGitHubOwner(t *testing.T, owner string) {
	t.Helper()
	orig := runCommand
	t.Cleanup(func() { runCommand = orig })
	runCommand = func(string, ...string) ([]byte, error) {
		return []byte(owner + "\n"), nil
	}
}

func TestNormalizeRepository(t *testing.T) {
	stubGitHubOwner(t, "acme")
	tests := []struct {
		name   string
		module string
		want   string
	}{
		{"repository root", "github.com/dj/demo-cli", "dj/demo-cli"},
		{"major version suffix", "github.com/dj/demo-cli/v2", "dj/demo-cli"},
		{"repository subdirectory", "github.com/dj/monorepo/tools/demo", "dj/monorepo"},
		{"other host", "gitlab.com/dj/demo-cli", "acme/demo-cli"},
		{"missing repository segment", "github.com/dj", "acme/demo-cli"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{Type: ProjectTypeCLI, Name: "demo-cli", Module: tt.module, Author: "Test Author"}
			opts.Normalize()
			if opts.Repository != tt.want {
				t.Errorf("Repository = %q, want %q", opts.Repository, tt.want)
			}
		})
	}

	lib := Options{Type: ProjectTypeLib, Name: "demo", Module: "github.com/dj/demo", Author: "Test Author"}
	lib.Normalize()
	if lib.Repository != "" {
		t.Errorf("library Repository = %q, want empty", lib.Repository)
	}
}

func TestGenerate_CLIWarnsWhenModuleIsNotOnGitHub(t *testing.T) {
	stubGitHubOwner(t, "acme")
	tests := []struct {
		module string
		warn   bool
	}{
		{"github.com/dj/demo-cli", false},
		{"gitlab.com/dj/demo-cli", true},
	}
	for _, tt := range tests {
		t.Run(tt.module, func(t *testing.T) {
			res, err := Generate(Options{
				Type: ProjectTypeCLI, TargetDir: t.TempDir(), Name: "demo-cli", Binary: "demo",
				Module: tt.module, Author: "Test Author", DryRun: true,
			})
			if err != nil {
				t.Fatal(err)
			}
			warned := len(res.Warnings) == 1 && strings.Contains(res.Warnings[0], "cmd/demo/help.go")
			if warned != tt.warn {
				t.Errorf("warnings = %q, want warning %v", res.Warnings, tt.warn)
			}
		})
	}
}

func TestGenerate_CLISkillLayout(t *testing.T) {
	res, err := Generate(Options{
		Type: ProjectTypeCLI, TargetDir: t.TempDir(), Name: "demo-cli", Binary: "custom-bin",
		Module: "github.com/dj/demo-cli", Author: "Test Author", DryRun: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	files := make(map[string]bool, len(res.Files))
	for _, f := range res.Files {
		files[filepath.ToSlash(f)] = true
	}
	if !files["skills/custom-bin/SKILL.md"] {
		t.Errorf("missing repository skill in %v", res.Files)
	}
	for _, f := range []string{"cmd/custom-bin/SKILL.md", "cmd/custom-bin/skill.go", "skills/app/SKILL.md", "internal/agent/agent.go"} {
		if files[f] {
			t.Errorf("generated %s", f)
		}
	}
}

func TestGenerate_LibDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "demo-lib")

	opts := Options{
		Type:        ProjectTypeLib,
		TargetDir:   target,
		Name:        "demo-lib",
		PkgName:     "demolib",
		Module:      "github.com/dj/demo-lib",
		Description: "A demo library utility",
		DryRun:      true,
		InitGit:     false,
		RunTidy:     false,
	}

	res, err := Generate(opts)
	if err != nil {
		t.Fatalf("Generate dry run failed: %v", err)
	}

	if len(res.Files) == 0 {
		t.Fatal("expected generated file list in dry run result")
	}

	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf("expected target dir %q not to exist in dry-run mode", target)
	}
}

func TestGenerate_ScaffoldedCLICompilesAndRuns(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "testapp-cli")

	opts := Options{
		Type:        ProjectTypeCLI,
		TargetDir:   target,
		Name:        "testapp-cli",
		Binary:      "testapp",
		Module:      "github.com/dj/testapp-cli",
		Description: "A fully functional test application",
		DryRun:      false,
		InitGit:     true,
		RunTidy:     true,
	}

	res, err := Generate(opts)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	for _, args := range [][]string{{"setup"}, {"setup"}, {"run", "status"}, {"check"}} {
		cmd := exec.Command("just", append([]string{"--justfile", "./justfile"}, args...)...)
		cmd.Dir = target
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("generated project lifecycle %v failed: %v\n%s", args, err, out)
		}
	}

	// Run go vet
	cmdVet := exec.Command("go", "vet", "./...")
	cmdVet.Dir = target
	if out, err := cmdVet.CombinedOutput(); err != nil {
		t.Fatalf("go vet failed: %v\nOutput:\n%s", err, string(out))
	}

	// Run go test
	cmdTest := exec.Command("go", "test", "-v", "-race", "-coverprofile=coverage.out", "./...")
	cmdTest.Dir = target
	if out, err := cmdTest.CombinedOutput(); err != nil {
		t.Fatalf("go test failed: %v\nOutput:\n%s", err, string(out))
	} else {
		t.Logf("generated CLI tests:\n%s", out)
	}

	// Run go build
	checkGeneratedCoverage(t, target)

	cmdBuild := exec.Command("go", "build", "-o", filepath.Join("bin", "testapp"), "./cmd/testapp")
	cmdBuild.Dir = target
	if out, err := cmdBuild.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\nOutput:\n%s", err, string(out))
	}

	// Run compiled binary --help
	binPath := filepath.Join(target, "bin", "testapp")
	cmdHelp := exec.Command(binPath, "--help")
	outHelp, err := cmdHelp.CombinedOutput()
	if err != nil {
		t.Fatalf("running bin --help failed: %v", err)
	}
	if !strings.Contains(string(outHelp), "status") {
		t.Errorf("expected --help output to contain status command, got:\n%s", string(outHelp))
	}

	// Run compiled binary --version
	cmdVersion := exec.Command(binPath, "--version")
	outVersion, err := cmdVersion.CombinedOutput()
	if err != nil {
		t.Fatalf("running bin --version failed: %v", err)
	}
	if strings.TrimSpace(string(outVersion)) != "dev" {
		t.Errorf("expected version dev for a source build, got: %q", string(outVersion))
	}

	// Run compiled binary in AGENT mode
	cmdAgent := exec.Command(binPath, "status")
	cmdAgent.Env = append(os.Environ(), "AGENT=1")
	outAgent, err := cmdAgent.CombinedOutput()
	if err != nil {
		t.Fatalf("running bin status in AGENT mode failed: %v", err)
	}
	if !strings.Contains(string(outAgent), "status: ready") {
		t.Errorf("expected AGENT output 'status: ready', got:\n%s", string(outAgent))
	}

	wantSkill := "Install the testapp skill using the Skills CLI:\n\n" +
		"  npx skills add dj/testapp-cli --skill testapp\n\n" +
		"Skill:\n\n" +
		"  https://github.com/dj/testapp-cli/tree/main/skills/testapp\n"
	for _, mode := range []string{"0", "1"} {
		cmdSkill := exec.Command(binPath, "skill")
		cmdSkill.Dir = t.TempDir() // An installed binary must work without its repository.
		cmdSkill.Env = append(os.Environ(), "AGENT="+mode)
		outSkill, err := cmdSkill.CombinedOutput()
		if err != nil || string(outSkill) != wantSkill {
			t.Fatalf("skill in AGENT=%s mode: %v\n%s", mode, err, outSkill)
		}
	}
	if _, err := os.Stat(filepath.Join(target, "skills", "testapp", "SKILL.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(target, "cmd", "testapp", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("the skill must not live beside the binary source: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "internal", "agent")); !os.IsNotExist(err) {
		t.Fatalf("agent detection belongs to cobra-help-tree: %v", err)
	}
	goMod, err := os.ReadFile(filepath.Join(target, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(goMod), "cobra-help-tree/v4 v4.0.0") {
		t.Errorf("generated go.mod does not use cobra-help-tree v4:\n%s", goMod)
	}

	// Errors print once, from run, with Cobra's own printing silenced.
	cmdBogus := exec.Command(binPath, "bogus")
	cmdBogus.Env = append(os.Environ(), "AGENT=0")
	var stdoutBogus, stderrBogus bytes.Buffer
	cmdBogus.Stdout = &stdoutBogus
	cmdBogus.Stderr = &stderrBogus
	if err := cmdBogus.Run(); err == nil {
		t.Fatal("expected unknown command to fail")
	}
	if !strings.HasPrefix(stderrBogus.String(), "[ERROR] unknown command") || strings.Count(stderrBogus.String(), "unknown command") != 1 {
		t.Errorf("errors must print once: stdout = %q, stderr = %q", stdoutBogus.String(), stderrBogus.String())
	}

	// Inspect generated CLI project
	inspectRes, err := Inspect(target, ProjectTypeCLI)
	if err != nil {
		t.Fatalf("Inspect CLI failed: %v", err)
	}
	if inspectRes.Score < 9.99 {
		t.Errorf("Inspect CLI score = %.2f, want 10.0", inspectRes.Score)
	}
	if len(inspectRes.Issues) > 0 {
		t.Errorf("Inspect CLI reported issues: %v", inspectRes.Issues)
	}

	// Generate again into already initialized git directory with force
	opts.Force = true
	res2, err2 := Generate(opts)
	if err2 != nil {
		t.Fatalf("Generate with existing .git failed: %v", err2)
	}
	_ = res2

	_ = res
}

func TestGenerate_ScaffoldedLibCompilesAndRuns(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "testlib")

	opts := Options{
		Type:        ProjectTypeLib,
		TargetDir:   target,
		Name:        "testlib",
		PkgName:     "testlib",
		Module:      "github.com/dj/testlib",
		Description: "A fully functional test library",
		DryRun:      false,
		InitGit:     true,
		RunTidy:     true,
	}

	res, err := Generate(opts)
	if err != nil {
		t.Fatalf("Generate lib failed: %v", err)
	}

	expectedFiles := []string{
		"go.mod",
		"justfile",
		".gitignore",
		filepath.Join(".github", "workflows", "ci.yml"),
		"LICENSE",
		"README.md",
		"AGENTS.md",
		"testlib.go",
		"testlib_test.go",
	}

	for _, f := range expectedFiles {
		fullPath := filepath.Join(target, f)
		if _, err := os.Stat(fullPath); err != nil {
			t.Errorf("missing expected library file %s: %v", f, err)
		}
	}

	// Run go vet
	cmdVet := exec.Command("go", "vet", "./...")
	cmdVet.Dir = target
	if out, err := cmdVet.CombinedOutput(); err != nil {
		t.Fatalf("go vet on lib failed: %v\nOutput:\n%s", err, string(out))
	}

	// Run go test with race detector and coverage
	cmdTest := exec.Command("go", "test", "-v", "-race", "-coverprofile=coverage.out", "./...")
	cmdTest.Dir = target
	if out, err := cmdTest.CombinedOutput(); err != nil {
		t.Fatalf("go test on lib failed: %v\nOutput:\n%s", err, string(out))
	}

	// Inspect generated Lib project
	checkGeneratedCoverage(t, target)

	inspectRes, err := Inspect(target, ProjectTypeLib)
	if err != nil {
		t.Fatalf("Inspect Lib failed: %v", err)
	}
	if inspectRes.Score < 9.99 {
		t.Errorf("Inspect Lib score = %.2f, want 10.0", inspectRes.Score)
	}
	if len(inspectRes.Issues) > 0 {
		t.Errorf("Inspect Lib reported issues: %v", inspectRes.Issues)
	}

	_ = res
}

func TestGenerate_TargetNonEmptyWithoutForce(t *testing.T) {
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "existing.txt")
	_ = os.WriteFile(existingFile, []byte("hello"), 0644)

	opts := Options{
		Type:      ProjectTypeCLI,
		TargetDir: tmpDir,
		Name:      "test-cli",
		Force:     false,
	}

	_, err := Generate(opts)
	if err == nil {
		t.Fatal("expected error when generating in non-empty directory without force")
	}
	if !strings.Contains(err.Error(), "not empty") {
		t.Errorf("unexpected error message: %v", err)
	}

	// With Force=true, should succeed
	opts.Force = true
	opts.InitGit = false
	opts.RunTidy = false
	_, err = Generate(opts)
	if err != nil {
		t.Fatalf("Generate with Force=true failed: %v", err)
	}
}

func TestGenerate_GitAndTidyErrorsRecorded(t *testing.T) {
	opts := Options{
		Type:      ProjectTypeLib,
		TargetDir: filepath.Join(t.TempDir(), "bad-project"),
		Name:      "bad-project",
		PkgName:   "badproject",
		Module:    "github.com/test/bad-project",
		InitGit:   true,
		RunTidy:   true,
	}
	t.Setenv("PATH", t.TempDir()) // Neither git nor go can start.

	res, err := Generate(opts)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if res.GitError == nil || res.TidyError == nil {
		t.Fatalf("GitError = %v, TidyError = %v, want both set", res.GitError, res.TidyError)
	}
	if res.GitInitialized || res.TidyCompleted {
		t.Fatalf("GitInitialized = %t, TidyCompleted = %t, want false", res.GitInitialized, res.TidyCompleted)
	}
	want := []string{res.GitError.Error(), res.TidyError.Error()}
	if !slices.Equal(res.Warnings, want) {
		t.Fatalf("Warnings = %q, want %q", res.Warnings, want)
	}
}

func TestInspect_EdgeCases(t *testing.T) {
	// 1. Nonexistent directory
	_, err := Inspect("/nonexistent/dir/12345", ProjectTypeCLI)
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}

	// 2. Target is a file instead of directory
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "somefile.txt")
	_ = os.WriteFile(filePath, []byte("data"), 0644)
	_, err = Inspect(filePath, ProjectTypeCLI)
	if err == nil {
		t.Error("expected error for file path")
	}

	// 3. Empty directory inspection for CLI
	emptyDir := filepath.Join(tmpDir, "empty-cli")
	_ = os.MkdirAll(emptyDir, 0755)
	resCLI, err := Inspect(emptyDir, ProjectTypeCLI)
	if err != nil {
		t.Fatalf("Inspect empty dir failed: %v", err)
	}
	if resCLI.Score != 0.0 {
		t.Errorf("expected score 0.0 for empty dir, got %v", resCLI.Score)
	}
	if len(resCLI.Issues) != 12 {
		t.Errorf("expected 12 issues for empty CLI dir, got %d: %v", len(resCLI.Issues), resCLI.Issues)
	}

	// 4. Empty directory inspection for Lib
	emptyLibDir := filepath.Join(tmpDir, "empty-lib")
	_ = os.MkdirAll(emptyLibDir, 0755)
	resLib, err := Inspect(emptyLibDir, ProjectTypeLib)
	if err != nil {
		t.Fatalf("Inspect empty lib dir failed: %v", err)
	}
	if resLib.Score != 0.0 {
		t.Errorf("expected score 0.0 for empty lib dir, got %v", resLib.Score)
	}
	if len(resLib.Issues) != 9 {
		t.Errorf("expected 9 issues for empty Lib dir, got %d: %v", len(resLib.Issues), resLib.Issues)
	}

	// 5. Inspect with default empty project type (defaults to CLI) and alternative yaml names
	customDir := filepath.Join(tmpDir, "custom")
	_ = os.MkdirAll(filepath.Join(customDir, ".github", "workflows"), 0755)
	_ = os.MkdirAll(filepath.Join(customDir, "internal", "custom"), 0755)
	_ = os.WriteFile(filepath.Join(customDir, "go.mod"), []byte("module github.com/test/custom\n\ngo 1.26.2\n\nrequire github.com/alexgorbatchev/cobra-help-tree/v2 v2.0.1\n"), 0644)
	_ = os.WriteFile(filepath.Join(customDir, "Justfile"), []byte("default:\n"), 0644)
	_ = os.WriteFile(filepath.Join(customDir, ".goreleaser.yaml"), []byte("version: 2\n"), 0644)
	_ = os.WriteFile(filepath.Join(customDir, ".github", "workflows", "ci.yaml"), []byte("name: CI\n"), 0644)
	_ = os.WriteFile(filepath.Join(customDir, ".github", "workflows", "release.yaml"), []byte("name: Release\n"), 0644)
	_ = os.WriteFile(filepath.Join(customDir, "internal", "custom", "agent.go"), []byte(`package custom
import "os"
func Check() bool { return os.Getenv("AGENT") == "1" }
`), 0644)

	resCustom, err := Inspect(customDir, "")
	if err != nil {
		t.Fatalf("Inspect custom dir with empty pType failed: %v", err)
	}
	if !resCustom.HasJustfile || !resCustom.HasGoreleaser || !resCustom.HasCIWorkflow || !resCustom.HasReleaseWorkflow || !resCustom.HasAgentMode {
		t.Errorf("Inspect failed to detect yaml alternatives: %+v", resCustom)
	}
}

func TestInspect_AgentModeDetection(t *testing.T) {
	const (
		library = "package main\nvar agent = cobrahelptree.IsAgentMode()\n"
		direct  = "package agent\nvar on = os.Getenv(\"AGENT\") == \"1\"\n"
		other   = "package main\n"
	)
	tests := []struct {
		name string
		path string
		body string
		want bool
	}{
		{"library helper in cmd", "cmd/app/main.go", library, true},
		{"library helper in internal", "internal/cli/mode.go", library, true},
		{"direct read in internal", "internal/agent/agent.go", direct, true},
		{"direct read in cmd", "cmd/app/mode.go", direct, true},
		{"test file only", "cmd/app/main_test.go", library, false},
		{"not Go source", "cmd/app/notes.txt", library, false},
		{"no agent mode", "cmd/app/main.go", other, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, filepath.FromSlash(tt.path))
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(tt.body), 0644); err != nil {
				t.Fatal(err)
			}
			res, err := Inspect(dir, ProjectTypeCLI)
			if err != nil {
				t.Fatal(err)
			}
			if res.HasAgentMode != tt.want {
				t.Errorf("HasAgentMode = %v, want %v", res.HasAgentMode, tt.want)
			}
		})
	}
}

func TestNormalizeRepositoryRoot(t *testing.T) {
	tests := []struct {
		module string
		want   string
	}{
		{"github.com/dj/demo-cli", "."},
		{"github.com/dj/demo-cli/v2", "."},
		{"github.com/dj/demo-cli/v10", "."},
		{"github.com/dj/monorepo/demo", ".."},
		{"github.com/dj/monorepo/tools/demo", "../.."},
		{"github.com/dj/monorepo/tools/demo/v3", "../.."},
		{"github.com/dj/monorepo/tools/v1", "../.."},
		{"gitlab.com/dj/monorepo/tools/demo", "."},
	}
	for _, tt := range tests {
		t.Run(tt.module, func(t *testing.T) {
			opts := Options{Type: ProjectTypeCLI, Name: "demo", Module: tt.module, Author: "Test Author"}
			opts.Normalize()
			if opts.RepositoryRoot != tt.want {
				t.Errorf("RepositoryRoot = %q, want %q", opts.RepositoryRoot, tt.want)
			}
		})
	}
}

func subdirectoryOptions(target string) Options {
	return Options{
		Type: ProjectTypeCLI, TargetDir: target, Name: "demo", Binary: "demo",
		Module: "github.com/dj/monorepo/tools/demo", Author: "Test Author",
		GoVersion: "1.26.2", InitGit: true, RunTidy: true,
	}
}

func TestGenerate_SubdirectoryModuleKeepsSkillAtRepositoryRoot(t *testing.T) {
	repo := t.TempDir()
	target := filepath.Join(repo, "tools", "demo")

	res, err := Generate(subdirectoryOptions(target))
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if res.TidyError != nil {
		t.Fatalf("go mod tidy failed: %v", res.TidyError)
	}

	if _, err := os.Stat(filepath.Join(repo, "skills", "demo", "SKILL.md")); err != nil {
		t.Errorf("skill is not at the repository root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "skills")); !os.IsNotExist(err) {
		t.Errorf("skill directory created inside the module: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, ".git")); !os.IsNotExist(err) || res.GitInitialized {
		t.Errorf("initialized a nested repository: %v", err)
	}
	warned := false
	for _, w := range res.Warnings {
		warned = warned || strings.Contains(w, ".github/workflows")
	}
	if !warned {
		t.Errorf("warnings = %q, want the workflow location", res.Warnings)
	}
	readme, err := os.ReadFile(filepath.Join(target, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(readme), "(../../skills/demo/SKILL.md)") {
		t.Errorf("README skill link does not reach the repository root:\n%s", readme)
	}

	// The generated test reads the skill through RepositoryRoot, so it fails if
	// the path, the printed URL, and the file disagree.
	cmd := exec.Command("go", "test", "-run", "TestSkillDocumentsCommandInterface", "./...")
	cmd.Dir = target
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generated skill test failed: %v\n%s", err, out)
	}
}

func TestGenerate_SubdirectoryModuleRejections(t *testing.T) {
	t.Run("target outside the module path", func(t *testing.T) {
		repo := t.TempDir()
		opts := subdirectoryOptions(filepath.Join(repo, "elsewhere"))
		_, err := Generate(opts)
		if err == nil || !strings.Contains(err.Error(), "must end in that path") {
			t.Fatalf("err = %v", err)
		}
		if entries, _ := os.ReadDir(repo); len(entries) != 0 {
			t.Fatalf("rejected target wrote %v", entries)
		}
	})

	t.Run("existing skill", func(t *testing.T) {
		repo := t.TempDir()
		skill := filepath.Join(repo, "skills", "demo", "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(skill), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(skill, []byte("mine"), 0644); err != nil {
			t.Fatal(err)
		}
		opts := subdirectoryOptions(filepath.Join(repo, "tools", "demo"))
		opts.RunTidy = false
		if _, err := Generate(opts); err == nil || !strings.Contains(err.Error(), "already exists") {
			t.Fatalf("err = %v", err)
		}
		if data, _ := os.ReadFile(skill); string(data) != "mine" {
			t.Fatalf("skill overwritten: %q", data)
		}
		opts.Force = true
		if _, err := Generate(opts); err != nil {
			t.Fatalf("Generate with force failed: %v", err)
		}
		if data, _ := os.ReadFile(skill); string(data) == "mine" {
			t.Fatal("force did not overwrite the skill")
		}
	})
}

func writeInspectFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestInspect_CurrentStandards(t *testing.T) {
	const currentHelpTree = "module github.com/test/demo\n\nrequire github.com/alexgorbatchev/cobra-help-tree/v4 v4.0.0\n"
	const skill = "---\nname: \"demo\"\n---\n"
	tests := []struct {
		name      string
		goMod     string
		files     map[string]string
		wantTree  bool
		wantSkill bool
		wantIssue string
	}{
		{
			name:      "skill at the repository root",
			goMod:     currentHelpTree,
			files:     map[string]string{"cmd/demo/main.go": "package main\n", "skills/demo/SKILL.md": skill},
			wantTree:  true,
			wantSkill: true,
		},
		{
			name:      "older help tree major version",
			goMod:     "module github.com/test/demo\n\nrequire github.com/alexgorbatchev/cobra-help-tree/v2 v2.0.1\n",
			files:     map[string]string{"cmd/demo/main.go": "package main\n", "skills/demo/SKILL.md": skill},
			wantSkill: true,
			wantIssue: "go.mod must require " + cobraHelpTreeModule,
		},
		{
			name:      "newer help tree major version with the same prefix",
			goMod:     "module github.com/test/demo\n\nrequire github.com/alexgorbatchev/cobra-help-tree/v40 v40.0.0\n",
			files:     map[string]string{"cmd/demo/main.go": "package main\n", "skills/demo/SKILL.md": skill},
			wantSkill: true,
			wantIssue: "go.mod must require " + cobraHelpTreeModule,
		},
		{
			name:      "skill under cmd",
			goMod:     currentHelpTree,
			files:     map[string]string{"cmd/demo/main.go": "package main\n", "cmd/demo/SKILL.md": skill},
			wantTree:  true,
			wantIssue: "missing skills/demo/SKILL.md at the repository root",
		},
		{
			name:  "embedded skill",
			goMod: currentHelpTree,
			files: map[string]string{
				"cmd/demo/main.go":     "package main\n\n//go:embed SKILL.md\nvar skill string\n",
				"skills/demo/SKILL.md": skill,
			},
			wantTree:  true,
			wantIssue: "cmd/demo/main.go embeds SKILL.md: keep the skill in the repository and print its install command",
		},
		{
			name:      "skill named for another binary",
			goMod:     currentHelpTree,
			files:     map[string]string{"cmd/demo/main.go": "package main\n", "skills/demo/SKILL.md": "---\nname: other\n---\n"},
			wantTree:  true,
			wantIssue: "skills/demo/SKILL.md must set name to demo",
		},
		{
			name:      "no binary",
			goMod:     currentHelpTree,
			files:     map[string]string{"cmd/.keep": ""},
			wantTree:  true,
			wantIssue: "missing cmd/<binary> directory, so no binary has a skill",
		},
		{
			name:      "no cmd directory",
			goMod:     currentHelpTree,
			wantTree:  true,
			wantIssue: "missing cmd/ directory, so no binary has a skill",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeInspectFile(t, dir, "go.mod", tt.goMod)
			for name, content := range tt.files {
				writeInspectFile(t, dir, name, content)
			}
			res, err := Inspect(dir, ProjectTypeCLI)
			if err != nil {
				t.Fatal(err)
			}
			if res.HasCobraHelpTree != tt.wantTree || res.HasSkill != tt.wantSkill {
				t.Fatalf("HasCobraHelpTree = %t, HasSkill = %t, issues %q", res.HasCobraHelpTree, res.HasSkill, res.Issues)
			}
			if tt.wantIssue != "" && !slices.Contains(res.Issues, tt.wantIssue) {
				t.Fatalf("issues %q do not include %q", res.Issues, tt.wantIssue)
			}
		})
	}
}

func TestInspect_SubdirectoryModuleReadsRepositoryRootSkill(t *testing.T) {
	root := t.TempDir()
	module := filepath.Join(root, "tools", "demo")
	writeInspectFile(t, module, "go.mod", "module github.com/test/repo/tools/demo\n")
	writeInspectFile(t, module, "cmd/demo/main.go", "package main\n")
	writeInspectFile(t, root, "skills/demo/SKILL.md", "---\nname: demo\n---\n")

	res, err := Inspect(module, ProjectTypeCLI)
	if err != nil {
		t.Fatal(err)
	}
	if !res.HasSkill {
		t.Fatalf("HasSkill = false, issues %q", res.Issues)
	}
}
