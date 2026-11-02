`go-scaffold` generates and audits standardized Go CLI and library repositories, so a new project starts with command help, dual human and agent output, task automation, tests, and release automation already wired together. It is for developers who start Go projects often and want each one to meet the same standards from the first commit.

# What It Does

- **Scaffolds CLI projects** with `cli create`: a [Cobra](https://github.com/spf13/cobra) command tree with [cobra-help-tree](https://github.com/alexgorbatchev/cobra-help-tree) help, `AGENT=1` output, a `justfile`, GitHub Actions, GoReleaser, and offline tests.
- **Scaffolds Go libraries** with `lib create`: a root package, table-driven tests, a `justfile`, and CI.
- **Generates an agent skill** for each CLI: a `skills/<binary>/SKILL.md` guide, a `skill` command that prints how to install it, and an alert at the top of agent help.
- **Fills in names for you** from the target directory, your GitHub account, and your git identity.
- **Audits existing repositories** with `cli inspect` and `lib inspect`, scoring them against the same standards out of 10.

# How It Works

- Takes a target directory, defaulting to the current one, and derives the project, binary, package, and module names from it unless you pass your own.
- Writes the project files, then initializes a git repository and resolves dependencies.
- Prints the next commands to run, along with any warnings about what it could not finish.
- For an existing repository, `inspect` reports which standard files and features are missing or out of date and gives a score.

# How it Really Works

- `create` refuses a non-empty target directory unless you pass `--force`, which overwrites any file it generates and leaves the rest alone.
- The default module owner comes from `gh api user`, then the `github.user` and `user.username` git settings, then your system username. Passing `--module` skips the lookup.
- The author comes from `--author`, then the `user.name` git setting, then your system user.
- `git init` runs only when the target has no `.git`, and `go mod tidy` runs after the files are written. Failures from either are warnings: the files stay and the exit code is 0, so read the warnings before treating a project as ready.
- The generated `skill` command names the `owner/repo` from a `github.com` module path. Any other module falls back to `<owner>/<name>`, and `create` warns and names the constant to correct.
- A module below its repository root, such as `github.com/owner/repo/tools/demo`, gets its skill at the repository root, because that is the only place the Skills CLI looks. The target directory must end in `tools/demo`, so this is the one case where `create` writes outside the target. It skips `git init`, and warns that the generated workflows must move to the repository root to run.
- `inspect` checks for files and dependencies and reads no build output, so a score of 10 means the pieces exist, not that they work.
- With `AGENT=1`, output is compact key-value lines. Errors exit 1 and go to stderr in both modes.

# Installation

Download the prebuilt binary for your platform from the [latest release](https://github.com/alexgorbatchev/go-scaffold-cli/releases/latest), replacing `X.X.X` with the version shown on that page.

```bash
# macOS (Apple Silicon)
curl -sSL https://github.com/alexgorbatchev/go-scaffold-cli/releases/latest/download/go-scaffold_X.X.X_darwin_arm64.tar.gz | tar -xz -C ~/.local/bin
```

# Setup

- [Go](https://go.dev/dl/) - Required by `go mod tidy` after `create`. Skip it with `--no-tidy`.
- [Git](https://git-scm.com/) - Required to initialize the repository after `create`. Skip it with `--no-git`.
- [GitHub CLI](https://cli.github.com/) - Optional. `create` uses your logged-in account as the default module owner.
- [just](https://github.com/casey/just) - Required to run the recipes in generated projects.

# Quick Start

```bash
# Scaffold a CLI project in the current directory
go-scaffold cli create

# Scaffold a CLI project in a new directory with a custom name and module
go-scaffold cli create ./my-tool-cli --name my-tool-cli --module github.com/alexgorbatchev/my-tool-cli

# Scaffold a Go library
go-scaffold lib create ./my-lib --name my-lib --pkg mylib

# Preview the files without writing them
go-scaffold cli create ./my-tool-cli --dry-run

# Audit existing repositories
go-scaffold cli inspect ./
go-scaffold lib inspect ../my-lib

# Print the command that installs the agent skill
go-scaffold skill
```

Sample Output:

```
[DRY-RUN] Created CLI project my-tool-cli (github.com/alexgorbatchev/my-tool-cli) in /home/dev/projects/my-tool-cli
[DRY-RUN] Generated 15 files:
  - .github/workflows/ci.yml
  - .github/workflows/release.yml
  - .gitignore
  ...
```

# Options & Flags

### `cli create`

| Flag | Short | Default | Description |
| :--- | :--- | :--- | :--- |
| `--name <name>` | `-n` | target directory name | Name of the CLI project |
| `--binary <bin>` | `-b` | `<name>` without `-cli` | Executable binary name |
| `--module <path>` | `-m` | `github.com/<owner>/<name>` | Go module path |
| `--description <desc>` | `-d` | `A command-line tool for <name>` | Short description for project documentation |
| `--author <name>` | `-a` | git `user.name` or system user | Author name for the license and skill |
| `--go-version <ver>` | | `1.26.2` | Go version declared in `go.mod` |
| `--force` | `-f` | `false` | Overwrite files if the target directory is not empty |
| `--dry-run` | | `false` | Preview generated files without writing to disk |
| `--no-git` | | `false` | Skip git repository initialization |
| `--no-tidy` | | `false` | Skip running `go mod tidy` |

### `lib create`

| Flag | Short | Default | Description |
| :--- | :--- | :--- | :--- |
| `--name <name>` | `-n` | target directory name | Name of the library project |
| `--pkg <pkg>` | `-p` | sanitized `<name>` | Go package name |
| `--module <path>` | `-m` | `github.com/<owner>/<name>` | Go module path |
| `--description <desc>` | `-d` | `A Go library for <name>` | Short description for library documentation |
| `--author <name>` | `-a` | git `user.name` or system user | Author name for the license |
| `--go-version <ver>` | | `1.26.2` | Go version declared in `go.mod` |
| `--force` | `-f` | `false` | Overwrite files if the target directory is not empty |
| `--dry-run` | | `false` | Preview generated files without writing to disk |
| `--no-git` | | `false` | Skip git repository initialization |
| `--no-tidy` | | `false` | Skip running `go mod tidy` |

# License

MIT License (c) 2026 Alex Gorbatchev
