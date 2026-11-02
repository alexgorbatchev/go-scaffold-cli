package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	skillAlert   = "ALERT: Agents must read `AGENT=1 go-scaffold skill` before using this tool.\n"
	skillInstall = "Install the go-scaffold skill using the Skills CLI:\n\n" +
		"  npx skills add alexgorbatchev/go-scaffold-cli --skill go-scaffold\n\n" +
		"Skill:\n\n" +
		"  https://github.com/alexgorbatchev/go-scaffold-cli/tree/main/skills/go-scaffold\n"
)

func TestSkillCommand(t *testing.T) {
	for _, mode := range []string{"0", "1"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("AGENT", mode)
			t.Chdir(t.TempDir())
			out, err := executeCommand("skill")
			if err != nil {
				t.Fatalf("skill failed: %v", err)
			}
			if out != skillInstall {
				t.Fatalf("skill output = %q, want %q", out, skillInstall)
			}
		})
	}
}

func TestSkillCommandRejectsArguments(t *testing.T) {
	if out, err := executeCommand("skill", "extra"); err == nil {
		t.Fatalf("skill must reject arguments, got: %s", out)
	}
}

func commandTree(t *testing.T) *cobra.Command {
	t.Helper()
	cmd := newRootCommand()
	cmd.SetOut(new(bytes.Buffer))
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultCompletionCmd()
	return cmd
}

func walkCommands(cmd *cobra.Command, visit func(*cobra.Command)) {
	if cmd.Hidden {
		return
	}
	visit(cmd)
	for _, child := range cmd.Commands() {
		walkCommands(child, visit)
	}
}

func TestHelpSkillAlert(t *testing.T) {
	for _, mode := range []string{"0", "1", "true", "yes"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("AGENT", mode)
			walkCommands(commandTree(t), func(cmd *cobra.Command) {
				args := strings.Fields(cmd.CommandPath())[1:]
				out, err := executeCommand(append(args, "--help")...)
				if err != nil {
					t.Fatalf("%s help: %v", cmd.CommandPath(), err)
				}
				if mode == "0" {
					if strings.Contains(out, skillAlert) {
						t.Errorf("%s human help contains agent alert", cmd.CommandPath())
					}
					return
				}
				if !strings.HasPrefix(out, skillAlert) || strings.Count(out, skillAlert) != 1 {
					t.Errorf("%s help must start with one skill alert: %q", cmd.CommandPath(), out)
				}
				if !strings.Contains(out, "command: "+cmd.CommandPath()+"\n") {
					t.Errorf("%s lost its agent help content: %s", cmd.CommandPath(), out)
				}
			})
			out, err := executeCommand()
			if err != nil || (mode != "0" && !strings.HasPrefix(out, skillAlert)) {
				t.Fatalf("bare invocation help: %v: %s", err, out)
			}
			out, err = executeCommand("help", "cli")
			if err != nil || (mode != "0" && !strings.HasPrefix(out, skillAlert)) {
				t.Fatalf("explicit help command: %v: %s", err, out)
			}
			out, err = executeCommand("--version")
			if err != nil || out != currentVersion()+"\n" {
				t.Fatalf("version output changed: %v: %q", err, out)
			}
		})
	}
}

func TestHumanHelpHidesGeneratedCommands(t *testing.T) {
	t.Setenv("AGENT", "0")
	out, err := executeCommand("--help")
	if err != nil {
		t.Fatal(err)
	}
	for _, generated := range []string{"completion", "├─ skill", "╰─ skill"} {
		if strings.Contains(out, generated) {
			t.Errorf("human help lists %s:\n%s", generated, out)
		}
	}
	if !strings.Contains(out, "inspect") {
		t.Errorf("human help lost its commands:\n%s", out)
	}
}

func TestSkillDocumentsCommandInterface(t *testing.T) {
	printed, err := executeCommand("skill")
	if err != nil {
		t.Fatal(err)
	}
	_, skillPath, ok := strings.Cut(strings.TrimSpace(printed), "/tree/"+skillBranch+"/")
	if !ok {
		t.Fatalf("skill output has no repository path: %s", printed)
	}
	data, err := os.ReadFile(filepath.Join("..", "..", filepath.FromSlash(skillPath), "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	if !strings.Contains(out, `name: "go-scaffold"`) {
		t.Fatal("the skill's frontmatter name must equal the --skill value")
	}
	walkCommands(commandTree(t), func(cmd *cobra.Command) {
		cmd.InitDefaultHelpFlag()
		cmd.InitDefaultVersionFlag()
		usage := cmd.Use
		if cmd.Parent() != nil {
			usage = cmd.Parent().CommandPath() + " " + usage
		}
		heading := "## `" + usage + "`\n"
		_, section, ok := strings.Cut(out, heading)
		if !ok {
			t.Errorf("skill is missing command %s", usage)
			return
		}
		section, _, _ = strings.Cut(section, "\n## ")
		checkFlags := func(f *pflag.Flag) {
			if f.Hidden {
				return
			}
			if f.Name == "help" && cmd.Parent() != nil {
				return // The shared help flag is documented once in the root section.
			}
			shorthand := "—"
			if f.Shorthand != "" {
				shorthand = "`-" + f.Shorthand + "`"
			}
			def := f.DefValue
			if def == "" {
				def = `""`
			}
			row := fmt.Sprintf("| `--%s` | %s | `%s` | `%s` |", f.Name, shorthand, f.Value.Type(), def)
			if !strings.Contains(section, row) {
				t.Errorf("%s skill is missing flag/type/default: %s", usage, row)
			}
		}
		cmd.LocalNonPersistentFlags().VisitAll(checkFlags)
		cmd.PersistentFlags().VisitAll(checkFlags)
	})
}

func TestSkillOutputErrors(t *testing.T) {
	for _, args := range [][]string{
		{"skill"},
		{"--help"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			t.Setenv("AGENT", "1")
			f, err := os.Create(filepath.Join(t.TempDir(), "closed"))
			if err != nil {
				t.Fatal(err)
			}
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
			cmd := newRootCommand()
			stderr := new(bytes.Buffer)
			cmd.SetOut(f)
			cmd.SetErr(stderr)
			cmd.SetArgs(args)
			if args[0] == "skill" {
				if err := cmd.Execute(); !errors.Is(err, os.ErrClosed) {
					t.Fatalf("expected closed output error, got %v", err)
				}
				return
			}
			t.Cleanup(func() { helpWriteError = nil })
			if code := run(cmd); code != 1 {
				t.Fatalf("exit code = %d, want 1", code)
			}
			if !strings.HasPrefix(stderr.String(), "ERR: ") || strings.Count(stderr.String(), "file already closed") != 1 {
				t.Fatalf("expected one help output error on stderr, got %q", stderr.String())
			}
		})
	}
}
