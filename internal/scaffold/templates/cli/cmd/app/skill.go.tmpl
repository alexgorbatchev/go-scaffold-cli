package main

import (
	_ "embed"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

//go:embed SKILL.md
var skillContent string

func newSkillCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "skill",
		Short: "Print the embedded SKILL.md usage guide for agents",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := io.WriteString(cmd.OutOrStdout(), skillContent); err != nil {
				return fmt.Errorf("printing skill: %w", err)
			}
			return nil
		},
	}
}
