package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/alexgorbatchev/go-scaffold-cli/internal/agent"
	"github.com/alexgorbatchev/go-scaffold-cli/internal/scaffold"
)

type libCreateFlags struct {
	name        string
	pkg         string
	module      string
	description string
	author      string
	goVersion   string
	force       bool
	dryRun      bool
	noGit       bool
	noTidy      bool
}

func newLibCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lib",
		Short: "Scaffold, inspect, and manage Go library projects",
		Long:  "Commands for generating standardized Go library repositories and auditing library project compliance.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newLibCreateCommand())
	cmd.AddCommand(newLibInspectCommand())

	return cmd
}

func newLibCreateCommand() *cobra.Command {
	var flags libCreateFlags

	cmd := &cobra.Command{
		Use:     "create [path]",
		Aliases: []string{"init", "new"},
		Short:   "Scaffold a new standardized Go library project",
		Long: `Generates a production-ready Go library repository according to best practices,
including root package implementation, table-driven unit tests with race detection,
justfile task automation, GitHub Actions CI workflows, and documentation.

The project name defaults to the directory name of the target path (or the current working directory).`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir := "."
			if len(args) > 0 {
				targetDir = args[0]
			}

			opts := scaffold.DefaultOptions(targetDir, scaffold.ProjectTypeLib)
			if flags.name != "" {
				opts.Name = flags.name
			}
			if flags.pkg != "" {
				opts.PkgName = flags.pkg
			}
			if flags.module != "" {
				opts.Module = flags.module
			}
			if flags.description != "" {
				opts.Description = flags.description
			}
			if flags.author != "" {
				opts.Author = flags.author
			}
			if flags.goVersion != "" {
				opts.GoVersion = flags.goVersion
			}
			opts.Force = flags.force
			opts.DryRun = flags.dryRun
			opts.InitGit = !flags.noGit
			opts.RunTidy = !flags.noTidy

			res, err := scaffold.Generate(opts)
			if err != nil {
				return err
			}

			if agent.IsAgentMode() {
				fmt.Fprintln(cmd.OutOrStdout(), "status: ok")
				fmt.Fprintf(cmd.OutOrStdout(), "type: %s\n", res.Type)
				fmt.Fprintf(cmd.OutOrStdout(), "project: %s\n", res.ProjectName)
				fmt.Fprintf(cmd.OutOrStdout(), "package: %s\n", res.PkgName)
				fmt.Fprintf(cmd.OutOrStdout(), "module: %s\n", res.ModuleName)
				fmt.Fprintf(cmd.OutOrStdout(), "target_dir: %s\n", res.ProjectDir)
				fmt.Fprintf(cmd.OutOrStdout(), "dry_run: %t\n", res.DryRun)
				fmt.Fprintf(cmd.OutOrStdout(), "files_created: %d\n", len(res.Files))
				for _, f := range res.Files {
					fmt.Fprintf(cmd.OutOrStdout(), "file: %s\n", f)
				}
				if len(res.Warnings) > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "warnings_count: %d\n", len(res.Warnings))
					for _, w := range res.Warnings {
						fmt.Fprintf(cmd.OutOrStdout(), "warning: %s\n", w)
					}
				}
				return nil
			}

			prefix := "[OK]"
			if res.DryRun {
				prefix = "[DRY-RUN]"
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s Created library project %s (pkg %s, %s) in %s\n", prefix, res.ProjectName, res.PkgName, res.ModuleName, res.ProjectDir)
			fmt.Fprintf(cmd.OutOrStdout(), "%s Generated %d files:\n", prefix, len(res.Files))
			for _, f := range res.Files {
				fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", f)
			}

			if !res.DryRun {
				if opts.InitGit {
					if res.GitError != nil {
						fmt.Fprintf(cmd.OutOrStdout(), "[WARN] git init skipped: %v\n", res.GitError)
					} else {
						fmt.Fprintln(cmd.OutOrStdout(), "[OK] Initialized git repository")
					}
				}
				if opts.RunTidy {
					if res.TidyError != nil {
						fmt.Fprintf(cmd.OutOrStdout(), "[WARN] go mod tidy skipped: %v\n", res.TidyError)
					} else {
						fmt.Fprintln(cmd.OutOrStdout(), "[OK] Ran go mod tidy")
					}
				}

				relPath, err := filepath.Rel(".", res.ProjectDir)
				if err != nil || relPath == "" {
					relPath = res.ProjectDir
				}

				fmt.Fprintln(cmd.OutOrStdout(), "\nNext steps:")
				if relPath != "." {
					fmt.Fprintf(cmd.OutOrStdout(), "  cd %s\n", relPath)
				}
				fmt.Fprintln(cmd.OutOrStdout(), "  just check")
				fmt.Fprintf(cmd.OutOrStdout(), "  just test\n")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&flags.name, "name", "n", "", "Library project name (defaults to target directory name)")
	cmd.Flags().StringVarP(&flags.pkg, "pkg", "p", "", "Go package name (defaults to sanitized project name)")
	cmd.Flags().StringVarP(&flags.module, "module", "m", "", "Go module path (defaults to github.com/<owner>/<name>)")
	cmd.Flags().StringVarP(&flags.description, "description", "d", "", "Short description for library documentation")
	cmd.Flags().StringVarP(&flags.author, "author", "a", "", "Author name (defaults to git config user.name or system user)")
	cmd.Flags().StringVar(&flags.goVersion, "go-version", "", "Go version declared in go.mod (default: 1.26.2)")
	cmd.Flags().BoolVarP(&flags.force, "force", "f", false, "Overwrite existing files if target directory is not empty")
	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false, "Preview generated files without writing them to disk")
	cmd.Flags().BoolVar(&flags.noGit, "no-git", false, "Skip git repository initialization")
	cmd.Flags().BoolVar(&flags.noTidy, "no-tidy", false, "Skip running go mod tidy")

	return cmd
}

func newLibInspectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect [path]",
		Short: "Inspect a library repository for standards compliance",
		Long:  "Audits a directory against standard Go library scaffolding, tests, and CI conventions.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir := "."
			if len(args) > 0 {
				targetDir = args[0]
			}

			res, err := scaffold.Inspect(targetDir, scaffold.ProjectTypeLib)
			if err != nil {
				return err
			}

			if agent.IsAgentMode() {
				fmt.Fprintln(cmd.OutOrStdout(), "status: ok")
				fmt.Fprintf(cmd.OutOrStdout(), "type: %s\n", res.Type)
				fmt.Fprintf(cmd.OutOrStdout(), "target_dir: %s\n", res.ProjectDir)
				fmt.Fprintf(cmd.OutOrStdout(), "score: %.1f\n", res.Score)
				fmt.Fprintf(cmd.OutOrStdout(), "has_go_mod: %t\n", res.HasGoMod)
				if res.ModuleName != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "module: %s\n", res.ModuleName)
				}
				if res.GoVersion != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "go_version: %s\n", res.GoVersion)
				}
				fmt.Fprintf(cmd.OutOrStdout(), "has_justfile: %t\n", res.HasJustfile)
				fmt.Fprintf(cmd.OutOrStdout(), "has_gitignore: %t\n", res.HasGitignore)
				fmt.Fprintf(cmd.OutOrStdout(), "has_ci_workflow: %t\n", res.HasCIWorkflow)
				fmt.Fprintf(cmd.OutOrStdout(), "has_license: %t\n", res.HasLicense)
				fmt.Fprintf(cmd.OutOrStdout(), "has_readme: %t\n", res.HasReadme)
				fmt.Fprintf(cmd.OutOrStdout(), "has_agents_doc: %t\n", res.HasAgentsDoc)
				fmt.Fprintf(cmd.OutOrStdout(), "has_go_source: %t\n", res.HasGoSource)
				fmt.Fprintf(cmd.OutOrStdout(), "has_go_test: %t\n", res.HasGoTest)
				fmt.Fprintf(cmd.OutOrStdout(), "issues_count: %d\n", len(res.Issues))
				for _, issue := range res.Issues {
					fmt.Fprintf(cmd.OutOrStdout(), "issue: %s\n", issue)
				}
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Library Project: %s\n", res.ProjectDir)
			fmt.Fprintf(cmd.OutOrStdout(), "Standards Score: %.1f / 10\n\n", res.Score)

			checks := []struct {
				name   string
				passed bool
				detail string
			}{
				{"go.mod", res.HasGoMod, res.ModuleName},
				{"justfile", res.HasJustfile, ""},
				{".gitignore", res.HasGitignore, ""},
				{".github/workflows/ci.yml", res.HasCIWorkflow, ""},
				{"LICENSE", res.HasLicense, ""},
				{"README.md", res.HasReadme, ""},
				{"AGENTS.md", res.HasAgentsDoc, ""},
				{"Root package Go source (*.go)", res.HasGoSource, ""},
				{"Root package Go unit test (*_test.go)", res.HasGoTest, ""},
			}

			for _, c := range checks {
				if c.passed {
					if c.detail != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "  [OK] %s (%s)\n", c.name, c.detail)
					} else {
						fmt.Fprintf(cmd.OutOrStdout(), "  [OK] %s\n", c.name)
					}
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "  [WARN] Missing %s\n", c.name)
				}
			}

			if len(res.Issues) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "\nIssues found (%d):\n", len(res.Issues))
				for _, issue := range res.Issues {
					fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", issue)
				}
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "\nAll standard library components are present and verified!")
			}

			return nil
		},
	}

	return cmd
}
