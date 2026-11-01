package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alexgorbatchev/go-scaffold-cli/internal/agent"
)

func main() {
	os.Exit(run(newRootCommand()))
}

func run(cmd *cobra.Command) int {
	err := cmd.Execute()
	if err == nil {
		err = helpWriteError
	}
	if err != nil {
		if agent.IsAgentMode() {
			fmt.Fprintf(cmd.ErrOrStderr(), "ERR: %v\n", err)
		} else {
			fmt.Fprintf(cmd.ErrOrStderr(), "[ERROR] %v\n", err)
		}
		return 1
	}
	return 0
}
