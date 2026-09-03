---
created_on: 2026-08-25 14:00
last_modified: 2026-08-25 15:30
status: current
---

# go-scaffold-cli

CLI utility to quickly scaffold and bootstrap standardized Go CLI (`cli`) and library (`lib`) projects adhering to Cobra command hierarchies, cobra-help-tree, dual-mode AGENT=1 execution, Justfile task automation, and GitHub Actions CI/CD workflows.

## Commands
- **Build Local Binary:** `just build` (compiles to `bin/go-scaffold-cli`)
- **Run CLI with Arguments:** `just run [args...]`
- **Run CLI in Agent Mode:** `just run-ai [args...]` (`AGENT=1`)
- **Run Tests:** `just test` (`go test -v ./...`)
- **Run Static Analysis & Tests:** `just check` (`go vet ./... && go test -v ./...`)
- **Run Linter / Static Analysis:** `just vet` or `just lint` (`go vet ./...`)
- **Format Source Code:** `just fmt` (`go fmt ./...`)
- **Scaffold New CLI Project:** `bin/go-scaffold-cli cli create [path] [--name <name>] [--module <mod>] [--binary <bin>]`
- **Inspect CLI Repository:** `bin/go-scaffold-cli cli inspect [path]`
- **Scaffold New Library Project:** `bin/go-scaffold-cli lib create [path] [--name <name>] [--pkg <pkg>] [--module <mod>]`
- **Inspect Library Repository:** `bin/go-scaffold-cli lib inspect [path]`

## Setup & Environment
- **Prerequisites:** Go 1.26+, `just`.
- **Temporary Files:** Temporary file operations use `.tmp/` within the project root.

## Conventions
- **Output Formatting & No Decorative Headers:** All CLI output must be plain text without emojis across all modes.
- **Tree Rendering:** Help screens and hierarchies render using `cobra-help-tree`.
- **Agent Mode (`AGENT=1`):** When `AGENT=1` is set, output compact key-value pairs or bullets.
- **External Templates:** All scaffolding templates reside as external files in `internal/scaffold/templates/` (`cli/` and `lib/`) and are embedded via `embed.FS`.
- **Hermetic Unit Tests:** All unit tests must remain 100% offline and hermetic.

## Boundaries
- **Always:** Maintain >= 90% statement code coverage across Go packages.
- **Always:** Run `just test` and `just vet` before committing changes.
- **Never:** Commit compiled Go binaries (e.g. `bin/go-scaffold-cli`) to git.
