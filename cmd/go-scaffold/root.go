package main

import (
	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
)

func newRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "go-scaffold",
		Short:        "Scaffolding CLI to quickly bootstrap standardized Go CLI and library projects",
		Long: `go-scaffold is a utility to quickly scaffold and bootstrap production-ready Go CLI
and library projects adhering to modern Go standards, Cobra command hierarchies, cobra-help-tree,
dual-mode human/agent output via AGENT=1, Justfile automation, and GitHub Actions CI/CD.`,
		Version:      version,
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.SetVersionTemplate("{{.Version}}\n")

	rootCmd.AddCommand(newCLICommand())
	rootCmd.AddCommand(newLibCommand())

	setupHelp(rootCmd)

	return rootCmd
}
