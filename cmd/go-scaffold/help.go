package main

import (
	cobrahelptree "github.com/alexgorbatchev/cobra-help-tree"
	"github.com/spf13/cobra"
)

var techCatalog = cobrahelptree.TechCatalog{
	"go-scaffold": {
		Summary:     "Go CLI and library scaffolding utility",
		Description: "CLI utility to generate standardized Go CLI and library repositories with Justfile, CI/CD, and tests.",
	},
	"go-scaffold cli": {
		Summary:     "CLI application scaffolding and compliance commands",
		Description: "Command group for creating new Go CLI projects and auditing CLI repository compliance.",
	},
	"go-scaffold cli create": {
		Summary:     "Scaffold a new standardized Go CLI project",
		Description: "Renders templates for Cobra CLI hierarchy, Justfile, GoReleaser, GitHub Actions, AGENT=1 dual mode, and offline tests.",
		Args:        "[path]",
	},
	"go-scaffold cli inspect": {
		Summary:     "Audit repository for CLI best practices and standards",
		Description: "Checks for go.mod, justfile, .gitignore, .goreleaser.yml, workflows, AGENTS.md, README.md, LICENSE, and dual-mode support.",
		Args:        "[path]",
	},
	"go-scaffold lib": {
		Summary:     "Library project scaffolding and compliance commands",
		Description: "Command group for creating new Go library packages and auditing library repository compliance.",
	},
	"go-scaffold lib create": {
		Summary:     "Scaffold a new standardized Go library project",
		Description: "Renders templates for root Go package, table-driven unit tests, Justfile, CI workflows, and documentation.",
		Args:        "[path]",
	},
	"go-scaffold lib inspect": {
		Summary:     "Audit repository for Go library best practices and standards",
		Description: "Checks for go.mod, justfile, .gitignore, CI workflow, LICENSE, README.md, AGENTS.md, root Go source, and tests.",
		Args:        "[path]",
	},
}

func setupHelp(cmd *cobra.Command) {
	cobrahelptree.Setup(cmd, cobrahelptree.TreeOptions{
		TechCatalog: techCatalog,
	})
}
