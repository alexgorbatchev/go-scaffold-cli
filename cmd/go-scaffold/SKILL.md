---
name: "go-scaffold"
description: "Use when operating the go-scaffold CLI."
author: "Alex Gorbatchev"
metadata:
  created_on: 2026-08-25
  last_modified: 2026-09-30 16:30
  status: current
---

## Execution rules

Read this entire skill before running operational commands. Set `AGENT=1`
on every invocation, or export it for the session. Select commands and
options from this reference. Read exit status as well as stdout; command
errors exit nonzero and print `ERR:` diagnostics to stderr in agent mode.
For `cli create` and `lib create`, git initialization and `go mod tidy`
errors are nonfatal: generated files remain, exit status is zero, and
stdout includes `status: ok`, `warnings_count: <count>`, and one `warning:`
line per error. Inspect these warnings before treating bootstrap as complete;
do not infer successful git initialization or tidy from exit status alone.

Set `AGENT=0` or leave it unset for human output. Values `1`, `true`, and
`yes` enable agent output, with case and surrounding whitespace ignored.
Operational commands use the arguments and options documented below.

## `go-scaffold`

Invoke without a subcommand to print root help. All commands accept
`--help` / `-h`; the root also accepts `--version` / `-v`.
Agent help starts with an alert to read this skill. Root and operational
commands accept no positional arguments. Create runs `go mod tidy` by
default, which may download dependencies and require network access and
credentials for private modules. Use `--no-tidy` for offline scaffolding;
`--dry-run` also skips tidy. Help, skill, version, completion, and inspect
commands require no network access or credentials.

| Flag | Short | Type | Default | Behavior |
| --- | --- | --- | --- | --- |
| `--help` | `-h` | `bool` | `false` | Print help for the selected command. |
| `--version` | `-v` | `bool` | `false` | Print only the raw version and a newline. |

## `go-scaffold skill`

Print this embedded SKILL.md verbatim to stdout, including frontmatter,
in both human and agent modes. Accept no command-specific flags or
positional arguments. Work from any directory without repository files.
Output-write failures return an error.

## `go-scaffold cli`

Manage, scaffold, and inspect Go CLI applications. Accept no positional
arguments. Print CLI command group help when invoked directly.

## `go-scaffold cli create [path]`

Scaffold a new standardized Go CLI project at the target path. Defaults
to the current directory if path is omitted. Generates a production-ready
repository structure with Cobra command hierarchies, cobra-help-tree,
AGENT=1 dual-mode support, Justfile task automation, GitHub Actions CI/CD,
GoReleaser configuration, embedded skill guide, and offline unit tests.
Initializes a git repository and runs `go mod tidy` unless skipped.
Generated CI runs on pushes to `main` and pull requests targeting `main`,
and requires at least 90% combined statement coverage across Go packages.

| Flag | Short | Type | Default | Behavior |
| --- | --- | --- | --- | --- |
| `--author` | `-a` | `string` | `""` | Author name (defaults to git config user.name or system user). |
| `--binary` | `-b` | `string` | `""` | Executable binary name (defaults to project name without -cli). |
| `--description` | `-d` | `string` | `""` | Short description for project documentation. |
| `--dry-run` | — | `bool` | `false` | Preview generated files without writing them to disk. |
| `--force` | `-f` | `bool` | `false` | Overwrite existing files if target directory is not empty. |
| `--go-version` | — | `string` | `""` | Go version declared in go.mod (default: 1.26.2). |
| `--module` | `-m` | `string` | `""` | Go module path (defaults to github.com/<owner>/<name>). |
| `--name` | `-n` | `string` | `""` | Project name (defaults to target directory name). |
| `--no-git` | — | `bool` | `false` | Skip git repository initialization. |
| `--no-tidy` | — | `bool` | `false` | Skip running go mod tidy. |

## `go-scaffold cli inspect [path]`

Inspect and audit an existing Go CLI repository for compliance with best
practices and standards. Checks for `go.mod`, `justfile`, `.gitignore`,
`.goreleaser.yml`, GitHub Actions CI/CD workflows, `LICENSE`, `README.md`,
`AGENTS.md`, `AGENT=1` dual-mode helper, and `cobra-help-tree` integration.
Defaults to the current working directory if path is omitted.

## `go-scaffold lib`

Manage, scaffold, and inspect Go library projects. Accept no positional
arguments. Print library command group help when invoked directly.

## `go-scaffold lib create [path]`

Scaffold a new standardized Go library package at the target path. Defaults
to the current directory if path is omitted. Generates root package source,
table-driven unit tests with race detection, Justfile task automation,
GitHub Actions CI workflows, and documentation. Initializes a git repository
and runs `go mod tidy` unless skipped.
Generated CI runs on pushes to `main` and pull requests targeting `main`,
and requires at least 90% combined statement coverage across Go packages.

| Flag | Short | Type | Default | Behavior |
| --- | --- | --- | --- | --- |
| `--author` | `-a` | `string` | `""` | Author name (defaults to git config user.name or system user). |
| `--description` | `-d` | `string` | `""` | Short description for library documentation. |
| `--dry-run` | — | `bool` | `false` | Preview generated files without writing them to disk. |
| `--force` | `-f` | `bool` | `false` | Overwrite existing files if target directory is not empty. |
| `--go-version` | — | `string` | `""` | Go version declared in go.mod (default: 1.26.2). |
| `--module` | `-m` | `string` | `""` | Go module path (defaults to github.com/<owner>/<name>). |
| `--name` | `-n` | `string` | `""` | Project name (defaults to target directory name). |
| `--no-git` | — | `bool` | `false` | Skip git repository initialization. |
| `--no-tidy` | — | `bool` | `false` | Skip running go mod tidy. |
| `--pkg` | `-p` | `string` | `""` | Go package name (defaults to sanitized project name). |

## `go-scaffold lib inspect [path]`

Inspect and audit an existing Go library repository for compliance with best
practices and standards. Checks for `go.mod`, `justfile`, `.gitignore`,
CI workflow, `LICENSE`, `README.md`, `AGENTS.md`, root package Go source files,
and unit tests. Defaults to the current working directory if path is omitted.

## `go-scaffold help [command]`

Print root help with no argument, or supply a space-separated command path,
such as `go-scaffold help cli` or `go-scaffold help lib`. Accept no command-specific flags.

## `go-scaffold completion`

Print completion group help with no verb. Use one of the four shell commands
below to print a completion script to stdout. Accept no positional arguments
or command-specific flags on the group. Redirect a script to a file to save it.

## `go-scaffold completion bash`

Print a Bash completion script. Accept no positional arguments. Load with
`source <(go-scaffold completion bash)` in Bash with `bash-completion` installed.

| Flag | Short | Type | Default | Behavior |
| --- | --- | --- | --- | --- |
| `--no-descriptions` | — | `bool` | `false` | Omit completion descriptions. |

## `go-scaffold completion zsh`

Print a Zsh completion script. Accept no positional arguments. Enable
completion with `autoload -U compinit; compinit`, then load with
`source <(go-scaffold completion zsh)`.

| Flag | Short | Type | Default | Behavior |
| --- | --- | --- | --- | --- |
| `--no-descriptions` | — | `bool` | `false` | Omit completion descriptions. |

## `go-scaffold completion fish`

Print a Fish completion script. Accept no positional arguments. Load with
`go-scaffold completion fish | source`.

| Flag | Short | Type | Default | Behavior |
| --- | --- | --- | --- | --- |
| `--no-descriptions` | — | `bool` | `false` | Omit completion descriptions. |

## `go-scaffold completion powershell`

Print a PowerShell completion script. Accept no positional arguments. Load
with `go-scaffold completion powershell | Out-String | Invoke-Expression`.

| Flag | Short | Type | Default | Behavior |
| --- | --- | --- | --- | --- |
| `--no-descriptions` | — | `bool` | `false` | Omit completion descriptions. |

## Workflow

```sh
export AGENT=1
go-scaffold skill
go-scaffold --version
go-scaffold cli create ./my-cli --name my-cli
go-scaffold cli inspect ./my-cli
go-scaffold lib create ./my-lib --name my-lib
go-scaffold lib inspect ./my-lib
```
