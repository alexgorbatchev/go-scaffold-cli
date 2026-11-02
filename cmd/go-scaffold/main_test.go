package main

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	cmd := newRootCommand()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.String(), err
}

func TestRootCommand_HelpAndVersion(t *testing.T) {
	t.Setenv("AGENT", "0")
	out, err := executeCommand("--help")
	if err != nil {
		t.Fatalf("help failed: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("expected help output")
	}

	// No args runs Help
	outNoArgs, errNoArgs := executeCommand()
	if errNoArgs != nil {
		t.Fatalf("no args failed: %v", errNoArgs)
	}
	if len(outNoArgs) == 0 {
		t.Fatal("expected help output on no args")
	}

	outVer, errVer := executeCommand("--version")
	if errVer != nil {
		t.Fatalf("version failed: %v", errVer)
	}
	if outVer != currentVersion()+"\n" {
		t.Errorf("expected '%s\\n', got %q", currentVersion(), outVer)
	}
}

func TestCLICommand_Help(t *testing.T) {
	t.Setenv("AGENT", "0")
	out, err := executeCommand("cli")
	if err != nil {
		t.Fatalf("cli command help failed: %v", err)
	}
	if !strings.Contains(out, "create") || !strings.Contains(out, "inspect") {
		t.Errorf("expected create and inspect in cli help, got:\n%s", out)
	}
}

func TestLibCommand_Help(t *testing.T) {
	t.Setenv("AGENT", "0")
	out, err := executeCommand("lib")
	if err != nil {
		t.Fatalf("lib command help failed: %v", err)
	}
	if !strings.Contains(out, "create") || !strings.Contains(out, "inspect") {
		t.Errorf("expected create and inspect in lib help, got:\n%s", out)
	}
}

func TestCLICreateCommand_HumanMode(t *testing.T) {
	t.Setenv("AGENT", "0")
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "subfolder")

	out, err := executeCommand("cli", "create", target,
		"--name", "custom-cli",
		"--binary", "custom-bin",
		"--module", "github.com/custom/custom-mod",
		"--description", "Custom CLI tool description",
		"--author", "Custom Author",
		"--go-version", "1.26.2",
		"--no-git",
		"--no-tidy",
	)
	if err != nil {
		t.Fatalf("cli create failed: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "[OK] Created CLI project custom-cli") {
		t.Errorf("expected '[OK] Created CLI project custom-cli' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "cmd/custom-bin/main.go") {
		t.Errorf("expected binary 'custom-bin', got:\n%s", out)
	}
	if !strings.Contains(out, "Next steps:") {
		t.Errorf("expected 'Next steps:' in output, got:\n%s", out)
	}

	// Verify go.mod was created with correct module and version
	data, err := os.ReadFile(filepath.Join(target, "go.mod"))
	if err != nil {
		t.Fatalf("expected go.mod to exist: %v", err)
	}
	if !strings.Contains(string(data), "module github.com/custom/custom-mod") {
		t.Errorf("expected module github.com/custom/custom-mod, got: %s", string(data))
	}
}

func TestCLICreateCommand_AgentMode(t *testing.T) {
	t.Setenv("AGENT", "1")
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "agent-created-cli")

	out, err := executeCommand("cli", "create", target, "--no-git", "--no-tidy")
	if err != nil {
		t.Fatalf("cli create failed: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "status: ok") {
		t.Errorf("expected 'status: ok' in agent output, got:\n%s", out)
	}
	if !strings.Contains(out, "project: agent-created-cli") {
		t.Errorf("expected 'project: agent-created-cli', got:\n%s", out)
	}
	if !strings.Contains(out, "binary: agent-created") {
		t.Errorf("expected 'binary: agent-created', got:\n%s", out)
	}
}

func TestCLICreateCommand_DryRun(t *testing.T) {
	t.Setenv("AGENT", "0")
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "dry-run-cli")

	out, err := executeCommand("cli", "create", target, "--dry-run")
	if err != nil {
		t.Fatalf("dry run failed: %v", err)
	}

	if !strings.Contains(out, "[DRY-RUN]") {
		t.Errorf("expected '[DRY-RUN]' in output, got:\n%s", out)
	}

	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf("expected target dir %q not to exist after dry run", target)
	}
}

func TestCLICreateCommand_ErrorOnNonEmpty(t *testing.T) {
	t.Setenv("AGENT", "0")
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "file.txt")
	_ = os.WriteFile(existingFile, []byte("content"), 0644)

	_, err := executeCommand("cli", "create", tmpDir, "--no-git", "--no-tidy")
	if err == nil {
		t.Fatal("expected error creating project in non-empty directory without --force")
	}

	// With --force, should succeed
	_, errForce := executeCommand("cli", "create", tmpDir, "--force", "--no-git", "--no-tidy")
	if errForce != nil {
		t.Fatalf("expected success with --force, got: %v", errForce)
	}
}

func TestLibCreateCommand_HumanMode(t *testing.T) {
	t.Setenv("AGENT", "0")
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "my-lib")

	out, err := executeCommand("lib", "create", target,
		"--name", "my-lib",
		"--pkg", "mylib",
		"--module", "github.com/alexgorbatchev/my-lib",
		"--description", "A powerful library",
		"--author", "Custom Lib Author",
		"--no-git",
		"--no-tidy",
	)
	if err != nil {
		t.Fatalf("lib create failed: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "[OK] Created library project my-lib") {
		t.Errorf("expected '[OK] Created library project my-lib' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "mylib.go") {
		t.Errorf("expected 'mylib.go' in output, got:\n%s", out)
	}

	data, err := os.ReadFile(filepath.Join(target, "go.mod"))
	if err != nil {
		t.Fatalf("expected go.mod to exist: %v", err)
	}
	if !strings.Contains(string(data), "module github.com/alexgorbatchev/my-lib") {
		t.Errorf("expected module line in go.mod, got: %s", string(data))
	}
}

func TestLibCreateCommand_AgentMode(t *testing.T) {
	t.Setenv("AGENT", "1")
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "agent-lib")

	out, err := executeCommand("lib", "create", target, "--no-git", "--no-tidy")
	if err != nil {
		t.Fatalf("lib create agent failed: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "status: ok") || !strings.Contains(out, "type: lib") {
		t.Errorf("expected agent lib output, got:\n%s", out)
	}
	if !strings.Contains(out, "package: agentlib") {
		t.Errorf("expected 'package: agentlib', got:\n%s", out)
	}
}

func TestLibCreateCommand_DryRun(t *testing.T) {
	t.Setenv("AGENT", "0")
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "dry-run-lib")

	out, err := executeCommand("lib", "create", target, "--dry-run")
	if err != nil {
		t.Fatalf("dry run failed: %v", err)
	}

	if !strings.Contains(out, "[DRY-RUN]") {
		t.Errorf("expected '[DRY-RUN]' in output, got:\n%s", out)
	}

	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Errorf("expected target dir %q not to exist after dry run", target)
	}
}

func TestLibCreateCommand_ErrorOnNonEmpty(t *testing.T) {
	t.Setenv("AGENT", "0")
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "file.txt")
	_ = os.WriteFile(existingFile, []byte("content"), 0644)

	_, err := executeCommand("lib", "create", tmpDir, "--no-git", "--no-tidy")
	if err == nil {
		t.Fatal("expected error creating lib in non-empty directory without --force")
	}

	// With --force, should succeed
	_, errForce := executeCommand("lib", "create", tmpDir, "--force", "--no-git", "--no-tidy")
	if errForce != nil {
		t.Fatalf("expected success with --force, got: %v", errForce)
	}
}

func TestCLICreateCommand_WithGitAndTidy(t *testing.T) {
	t.Setenv("AGENT", "0")
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "full-cli")

	out, err := executeCommand("cli", "create", target)
	if err != nil {
		t.Fatalf("cli create with git/tidy failed: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "[OK] Initialized git repository") {
		t.Errorf("expected '[OK] Initialized git repository', got:\n%s", out)
	}
	if !strings.Contains(out, "[OK] Ran go mod tidy") {
		t.Errorf("expected '[OK] Ran go mod tidy', got:\n%s", out)
	}
}

func TestLibCreateCommand_WithGitAndTidy(t *testing.T) {
	t.Setenv("AGENT", "0")
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "full-lib")

	out, err := executeCommand("lib", "create", target)
	if err != nil {
		t.Fatalf("lib create with git/tidy failed: %v\nOutput: %s", err, out)
	}

	if !strings.Contains(out, "[OK] Initialized git repository") {
		t.Errorf("expected '[OK] Initialized git repository', got:\n%s", out)
	}
	if !strings.Contains(out, "[OK] Ran go mod tidy") {
		t.Errorf("expected '[OK] Ran go mod tidy', got:\n%s", out)
	}
}

func TestInspectCommand_CLIAndLib_HumanAndAgent(t *testing.T) {
	tmpDir := t.TempDir()
	targetCLI := filepath.Join(tmpDir, "app-cli")
	targetLib := filepath.Join(tmpDir, "app-lib")

	// 1. Create both
	_, err := executeCommand("cli", "create", targetCLI, "--no-git", "--no-tidy")
	if err != nil {
		t.Fatalf("create cli failed: %v", err)
	}
	_, err = executeCommand("lib", "create", targetLib, "--no-git", "--no-tidy")
	if err != nil {
		t.Fatalf("create lib failed: %v", err)
	}

	// 2. Inspect CLI in Human Mode
	t.Setenv("AGENT", "0")
	outCLI, err := executeCommand("cli", "inspect", targetCLI)
	if err != nil {
		t.Fatalf("inspect CLI failed: %v", err)
	}
	if !strings.Contains(outCLI, "Score: 10.0 / 10") {
		t.Errorf("expected score 10.0 for CLI inspect, got:\n%s", outCLI)
	}

	// 3. Inspect CLI in Agent Mode
	t.Setenv("AGENT", "1")
	outAgentCLI, err := executeCommand("cli", "inspect", targetCLI)
	if err != nil {
		t.Fatalf("inspect CLI in agent mode failed: %v", err)
	}
	if !strings.Contains(outAgentCLI, "status: ok") || !strings.Contains(outAgentCLI, "score: 10.0") {
		t.Errorf("expected agent score 10.0, got:\n%s", outAgentCLI)
	}

	// 4. Inspect Lib in Human Mode
	t.Setenv("AGENT", "0")
	outLib, err := executeCommand("lib", "inspect", targetLib)
	if err != nil {
		t.Fatalf("inspect Lib failed: %v", err)
	}
	if !strings.Contains(outLib, "Score: 10.0 / 10") {
		t.Errorf("expected score 10.0 for Lib inspect, got:\n%s", outLib)
	}

	// 5. Inspect Lib in Agent Mode
	t.Setenv("AGENT", "1")
	outAgentLib, err := executeCommand("lib", "inspect", targetLib)
	if err != nil {
		t.Fatalf("inspect Lib in agent mode failed: %v", err)
	}
	if !strings.Contains(outAgentLib, "status: ok") || !strings.Contains(outAgentLib, "score: 10.0") {
		t.Errorf("expected agent score 10.0, got:\n%s", outAgentLib)
	}

	// Inspect with default cwd
	origWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(origWd) }()
	_ = os.Chdir(targetCLI)
	t.Setenv("AGENT", "0")
	outCwdCLI, err := executeCommand("cli", "inspect")
	if err != nil {
		t.Fatalf("inspect cli cwd failed: %v", err)
	}
	if !strings.Contains(outCwdCLI, "Score: 10.0 / 10") {
		t.Errorf("expected score 10.0, got: %s", outCwdCLI)
	}

	_ = os.Chdir(targetLib)
	outCwdLib, err := executeCommand("lib", "inspect")
	if err != nil {
		t.Fatalf("inspect lib cwd failed: %v", err)
	}
	if !strings.Contains(outCwdLib, "Score: 10.0 / 10") {
		t.Errorf("expected score 10.0, got: %s", outCwdLib)
	}
}

func TestInspectCommand_EmptyWithIssues(t *testing.T) {
	tmpDir := t.TempDir()
	emptyDir := filepath.Join(tmpDir, "empty")
	_ = os.MkdirAll(emptyDir, 0755)

	// CLI issues in human mode
	t.Setenv("AGENT", "0")
	outCLI, err := executeCommand("cli", "inspect", emptyDir)
	if err != nil {
		t.Fatalf("cli inspect empty dir failed: %v", err)
	}
	if !strings.Contains(outCLI, "Issues found") || !strings.Contains(outCLI, "[WARN] Missing go.mod") {
		t.Errorf("expected warnings and issues list in CLI inspect, got:\n%s", outCLI)
	}

	// CLI issues in agent mode
	t.Setenv("AGENT", "1")
	outAgentCLI, err := executeCommand("cli", "inspect", emptyDir)
	if err != nil {
		t.Fatalf("cli inspect agent failed: %v", err)
	}
	if !strings.Contains(outAgentCLI, "issues_count: 12") {
		t.Errorf("expected 12 issues in agent CLI inspect, got: %s", outAgentCLI)
	}

	// Lib issues in human mode
	t.Setenv("AGENT", "0")
	outLib, err := executeCommand("lib", "inspect", emptyDir)
	if err != nil {
		t.Fatalf("lib inspect empty dir failed: %v", err)
	}
	if !strings.Contains(outLib, "Issues found") || !strings.Contains(outLib, "[WARN] Missing go.mod") {
		t.Errorf("expected warnings and issues list in Lib inspect, got:\n%s", outLib)
	}

	// Lib issues in agent mode
	t.Setenv("AGENT", "1")
	outAgentLib, err := executeCommand("lib", "inspect", emptyDir)
	if err != nil {
		t.Fatalf("lib inspect agent failed: %v", err)
	}
	if !strings.Contains(outAgentLib, "issues_count: 9") {
		t.Errorf("expected 9 issues in agent Lib inspect, got: %s", outAgentLib)
	}

	// Error on invalid directory
	_, errInvalidCLI := executeCommand("cli", "inspect", "/nonexistent/invalid/dir")
	if errInvalidCLI == nil {
		t.Error("expected error for nonexistent inspect path in cli")
	}
	_, errInvalidLib := executeCommand("lib", "inspect", "/nonexistent/invalid/dir")
	if errInvalidLib == nil {
		t.Error("expected error for nonexistent inspect path in lib")
	}
}

func TestRun(t *testing.T) {
	for _, tt := range []struct {
		name string
		mode string
		args []string
		code int
		want string
	}{
		{"success", "0", []string{"--help"}, 0, ""},
		{"human error", "0", []string{"bogus"}, 1, "[ERROR] unknown command"},
		{"agent error", "1", []string{"bogus"}, 1, "ERR: unknown command"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("AGENT", tt.mode)
			stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
			cmd := newRootCommand()
			cmd.SetOut(stdout)
			cmd.SetErr(stderr)
			cmd.SetArgs(tt.args)
			if code := run(cmd); code != tt.code {
				t.Fatalf("exit code = %d, want %d", code, tt.code)
			}
			if !strings.HasPrefix(stderr.String(), tt.want) {
				t.Fatalf("stderr = %q, want prefix %q", stderr, tt.want)
			}
		})
	}
}

func TestCreateCommands_ReportWarnings(t *testing.T) {
	for _, tt := range []struct {
		name string
		mode string
		args []string
		want []string
	}{
		{"human foreign module", "0", []string{"cli", "create", "--module", "gitlab.com/dj/demo-cli", "--dry-run"},
			[]string{`[WARN] module "gitlab.com/dj/demo-cli" is not on github.com`}},
		{"agent foreign module", "1", []string{"cli", "create", "--module", "gitlab.com/dj/demo-cli", "--dry-run"},
			[]string{"warnings_count: 1\n", `warning: module "gitlab.com/dj/demo-cli" is not on github.com`}},
		{"human cli tidy failure", "0", []string{"cli", "create", "--module", "github.com/dj/demo-cli", "--go-version", "bogus", "--no-git"},
			[]string{"[WARN] go mod tidy failed"}},
		{"human lib tidy failure", "0", []string{"lib", "create", "--module", "github.com/dj/demo", "--go-version", "bogus", "--no-git"},
			[]string{"[WARN] go mod tidy failed"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("AGENT", tt.mode)
			out, err := executeCommand(append(tt.args, filepath.Join(t.TempDir(), "demo-cli"))...)
			if err != nil {
				t.Fatalf("create failed: %v\n%s", err, out)
			}
			for _, want := range tt.want {
				if !strings.Contains(out, want) {
					t.Errorf("output lacks %q:\n%s", want, out)
				}
			}
			if strings.Contains(out, "[OK] Ran go mod tidy") && strings.Contains(out, "go mod tidy failed") {
				t.Errorf("reported tidy as both failed and finished:\n%s", out)
			}
		})
	}
}

func TestDisplayedPathsFoldHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	target := filepath.Join(home, "projects", "demo-cli")
	createArgs := []string{"cli", "create", target, "--module", "github.com/dj/demo-cli", "--no-git", "--no-tidy"}
	tests := []struct {
		name string
		mode string
		args []string
		want string
	}{
		{"human create", "0", createArgs, "in ~/projects/demo-cli\n"},
		{"agent create", "1", createArgs, "target_dir: ~/projects/demo-cli\n"},
		{"human inspect", "0", []string{"cli", "inspect", target}, "CLI Project: ~/projects/demo-cli\n"},
		{"agent inspect", "1", []string{"lib", "inspect", target}, "target_dir: ~/projects/demo-cli\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("AGENT", tt.mode)
			args := tt.args
			if args[1] == "create" {
				args = append(slices.Clone(args), "--force")
			}
			out, err := executeCommand(args...)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out, tt.want) || strings.Contains(out, home) {
				t.Fatalf("output must fold %s and contain %q:\n%s", home, tt.want, out)
			}
		})
	}

	t.Run("error", func(t *testing.T) {
		t.Setenv("AGENT", "1")
		cmd := newRootCommand()
		stderr := new(bytes.Buffer)
		cmd.SetErr(stderr)
		cmd.SetArgs([]string{"cli", "inspect", filepath.Join(home, "missing")})
		if code := run(cmd); code != 1 {
			t.Fatalf("exit code = %d, want 1", code)
		}
		if !strings.Contains(stderr.String(), "~/missing") || strings.Contains(stderr.String(), home) {
			t.Fatalf("error must fold the home directory: %q", stderr.String())
		}
	})
}
