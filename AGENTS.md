---
created_on: 2026-08-25 14:00
last_modified: 2026-11-01 14:23
status: current
---

# go-scaffold-cli

CLI utility to quickly scaffold and bootstrap standardized Go CLI (`cli`) and library (`lib`) projects adhering to Cobra command hierarchies, cobra-help-tree, dual-mode AGENT=1 execution, Justfile task automation, and GitHub Actions CI/CD workflows.

## Commands
- **Build Local Binary:** `just build` (compiles to `bin/go-scaffold`)
- **Prepare Dependencies:** `just setup`
- **Run CLI with Arguments:** `just run [args...]`
- **Run CLI in Agent Mode:** `just run-ai [args...]` (`AGENT=1`)
- **Run Tests:** `just test` (`go test -v ./...`)
- **Run Static Analysis & Tests:** `just check` (`go vet ./... && go test -v ./...`)
- **Run Linter / Static Analysis:** `just vet` or `just lint` (`go vet ./...`)
- **Format Source Code:** `just fmt` (`go fmt ./...`)
- **Print Embedded Agent Skill:** `bin/go-scaffold skill`
- **Scaffold New CLI Project:** `bin/go-scaffold cli create [path] [--name <name>] [--module <mod>] [--binary <bin>]`
- **Inspect CLI Repository:** `bin/go-scaffold cli inspect [path]`
- **Scaffold New Library Project:** `bin/go-scaffold lib create [path] [--name <name>] [--pkg <pkg>] [--module <mod>]`
- **Inspect Library Repository:** `bin/go-scaffold lib inspect [path]`

## Setup & Environment
- **Prerequisites:** Go 1.26+, `just`.
- **Temporary Files:** Temporary file operations use `.tmp/` within the project root.

## Conventions
- **Output Formatting & No Decorative Headers:** All CLI output must be plain text without emojis across all modes.
- **Help Screens & Terminal Width:** CLI help output (`--help`) must display an aligned hierarchical tree view with `├─` and `╰─` glyphs powered by `github.com/alexgorbatchev/cobra-help-tree/v4` (`cobrahelptree.Setup(rootCmd)`), with long help text wrapped to the terminal width and never truncated.
- **Agent Mode (`AGENT=1`):** When `AGENT=1` is set, output compact key-value pairs or bullets.
- **External Templates:** All scaffolding templates reside as external files in `internal/scaffold/templates/` (`cli/` and `lib/`) and are embedded via `embed.FS`.
- **Embedded CLI Skill:** `go-scaffold` ships a top-level `skill` command embedding `cmd/go-scaffold/SKILL.md`, and its agent-mode help screens start with its reading alert.
- **Generated CLI Skill:** CLI templates generate `skills/<binary>/SKILL.md` at the project root and never embed it. Generated help uses `cobra-help-tree/v4` through `HelpOptions.Skill` and `TreeOptions.HideGeneratedCommands`, generated code calls `cobrahelptree.IsAgentMode`, and generated tests use `cobrahelptree.SkillOmissions`. `Options.Repository` supplies the `SkillSource` repository, derived from the module path. For a module in a repository subdirectory, `Options.RepositoryRoot` places the skill at the repository root, and `Generate` rejects a target directory that does not end in that subdirectory. Keep the guide and generated `AGENTS.md` maintenance instructions synchronized with the generated CLI. Cover template changes in `TestGenerate_ScaffoldedCLICompilesAndRuns`, including `skill` execution outside the repository.
- **Skill Maintenance:** Update `cmd/go-scaffold/SKILL.md` in the same change whenever any command, argument, option, default, environment variable, output format, or side effect changes. Verify documented behavior against the implementation and update `last_modified` metadata.
- **Hermetic Unit Tests:** All unit tests must remain 100% offline and hermetic. Each test package runs through `hermetic.Run` in `TestMain`, which stubs `gh` and turns off `GOPROXY` and `GOSUMDB`, so tests that run `go mod tidy` read only the local module cache. Run `just setup` before testing, and add the same `TestMain` to any new test package that generates a project.

## Boundaries
- **Always:** Maintain >= 90% statement code coverage across Go packages.
- **Always:** Run `just test` and `just vet` before committing changes.
- **Never:** Commit compiled Go binaries (e.g. `bin/go-scaffold-cli`) to git.
