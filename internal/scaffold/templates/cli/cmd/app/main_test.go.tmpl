package main

import (
	"bytes"
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

func TestRun(t *testing.T) {
	for _, tt := range []struct {
		name string
		mode string
		args []string
		code int
		want string
	}{
		{"success", "0", []string{"--version"}, 0, version + "\n"},
		{"human error", "0", []string{"unknown-command"}, 1, "[ERROR] unknown command"},
		{"agent error", "1", []string{"unknown-command"}, 1, "ERR: unknown command"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("AGENT", tt.mode)
			stdout, stderr := new(bytes.Buffer), new(bytes.Buffer)
			cmd := newRootCommand()
			cmd.SetArgs(tt.args)
			cmd.SetOut(stdout)
			cmd.SetErr(stderr)
			if code := run(cmd); code != tt.code {
				t.Fatalf("exit code = %d, want %d", code, tt.code)
			}
			if tt.code == 0 {
				if stdout.String() != tt.want || stderr.Len() != 0 {
					t.Fatalf("stdout = %q, stderr = %q", stdout, stderr)
				}
			} else if stdout.Len() != 0 || !strings.Contains(stderr.String(), tt.want) {
				t.Fatalf("stdout = %q, stderr = %q", stdout, stderr)
			}
		})
	}
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

	outVer, errVer := executeCommand("--version")
	if errVer != nil {
		t.Fatalf("version failed: %v", errVer)
	}
	if outVer != version+"\n" {
		t.Errorf("expected '%s\\n', got %q", version, outVer)
	}
}

func TestStatusCommand_HumanMode(t *testing.T) {
	t.Setenv("AGENT", "0")
	out, err := executeCommand("status")
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !strings.Contains(out, "[OK] System status: ready") {
		t.Errorf("expected human status output, got %q", out)
	}
}

func TestStatusCommand_AgentMode(t *testing.T) {
	t.Setenv("AGENT", "1")
	out, err := executeCommand("status")
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !strings.Contains(out, "status: ready") {
		t.Errorf("expected agent status output, got %q", out)
	}
}
