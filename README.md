go-scaffold-cli is a command-line tool for bootstrapping, generating, and inspecting standardized Go CLI and library repositories with [Cobra](https://github.com/spf13/cobra), [cobra-help-tree](https://github.com/alexgorbatchev/cobra-help-tree), dual-mode human/agent output, Justfile task automation, and GitHub Actions CI/CD.

# What It Does

- Quickly scaffolds production-ready Go CLI applications (`cli create`) and Go library packages (`lib create`).
- Defaults project, binary, and package names intelligently from the current working directory or target directory.
- Configures Cobra CLI structures with hierarchical tree-view help screens via [cobra-help-tree](https://github.com/alexgorbatchev/cobra-help-tree) for CLI projects.
- Sets up standard Go library architecture with table-driven tests and race detection for library projects.
- Embeds dual-mode execution support (`AGENT=1`) for AI agents and human users.
- Generates `justfile` automation recipes for running, testing, vetting, formatting, and building.
- Sets up GitHub Actions CI/CD workflows and GoReleaser release configurations.
- Audits and inspects existing Go CLI and library projects for compliance with project standards (`cli inspect`, `lib inspect`).

# How It Works

- Scaffolds a new project hierarchy using embedded external template files when invoked.
- Populates Go module names, binary names, package identifiers, and author metadata automatically.
- Initializes local git repositories and resolves Go module dependencies with `go mod tidy`.
- Provides an inspection tool to verify existing repositories against CLI and library best practice standards.

# How it Really Works

- Uses Go's `embed.FS` to bundle templates for `cli` and `lib` archetypes (`go.mod`, `justfile`, `.goreleaser.yml`, GitHub Actions, Cobra entry points, library sources, and tests).
- Renders templates with `text/template` into target directories, handling custom flags and overwrite safety checks.
- Adheres to `AGENT=1` environment variable contracts to switch between human-formatted output and compact key-value agent output.
- Employs [cobra-help-tree](https://github.com/alexgorbatchev/cobra-help-tree) for terminal-width-aware hierarchical tree help formatting.

# Prerequisites

- [Go](https://go.dev/) 1.26 or higher (for development)
- [Just](https://github.com/casey/just) task runner (for task automation)

# Installation

Download the prebuilt binary for your platform from the [latest release](https://github.com/alexgorbatchev/go-scaffold-cli/releases/latest).

```bash
# macOS (Apple Silicon)
curl -sSL https://github.com/alexgorbatchev/go-scaffold-cli/releases/latest/download/go-scaffold_1.0.0_darwin_arm64.tar.gz | tar -xz -C ~/.local/bin
```

# Quick Start

```bash
# Scaffold a CLI project in the current directory (defaults name to cwd dirname)
go-scaffold cli create

# Scaffold a CLI project in a subfolder with custom name and module
go-scaffold cli create ./my-tool-cli --name my-tool-cli --module github.com/alexgorbatchev/my-tool-cli

# Scaffold a Go library project in a subfolder
go-scaffold lib create ./my-lib --name my-lib --pkg mylib

# Preview files before writing to disk
go-scaffold cli create ./my-tool-cli --dry-run

# Inspect existing CLI or library repositories for standards compliance
go-scaffold cli inspect ./
go-scaffold lib inspect ../godeps
```

# Options & Flags

| Flag | Short | Default | Description |
| :--- | :--- | :--- | :--- |
| `--help` | `-h` | `false` | Display help and command tree |
| `--version` | `-v` | `false` | Display binary version |

### `cli create` Flags

| Flag | Short | Default | Description |
| :--- | :--- | :--- | :--- |
| `--name <name>` | `-n` | `cwd dirname` | Name of the CLI project |
| `--binary <bin>` | `-b` | `<name> (without -cli)` | Executable binary name |
| `--module <path>` | `-m` | `github.com/alexgorbatchev/<name>` | Go module path |
| `--description <desc>` | `-d` | `A command-line tool...` | Short description for project documentation |
| `--go-version <ver>` | | `1.26.2` | Go version declared in go.mod |
| `--force` | `-f` | `false` | Overwrite files if target directory is not empty |
| `--dry-run` | | `false` | Preview generated files without writing to disk |
| `--no-git` | | `false` | Skip git repository initialization |
| `--no-tidy` | | `false` | Skip running go mod tidy |

### `lib create` Flags

| Flag | Short | Default | Description |
| :--- | :--- | :--- | :--- |
| `--name <name>` | `-n` | `cwd dirname` | Name of the library project |
| `--pkg <pkg>` | `-p` | `sanitized name` | Go package name |
| `--module <path>` | `-m` | `github.com/alexgorbatchev/<name>` | Go module path |
| `--description <desc>` | `-d` | `A Go library for...` | Short description for library documentation |
| `--go-version <ver>` | | `1.26.2` | Go version declared in go.mod |
| `--force` | `-f` | `false` | Overwrite files if target directory is not empty |
| `--dry-run` | | `false` | Preview generated files without writing to disk |
| `--no-git` | | `false` | Skip git repository initialization |
| `--no-tidy` | | `false` | Skip running go mod tidy |

# License

MIT License. Copyright (c) 2026 Alex Gorbatchev.
