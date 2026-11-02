package main

import (
	"fmt"
	"os"

	cobrahelptree "github.com/alexgorbatchev/cobra-help-tree/v4"
	"github.com/spf13/cobra"

	"github.com/alexgorbatchev/go-scaffold-cli/v2/internal/displaypath"
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
		message := displaypath.FoldText(err.Error())
		if cobrahelptree.IsAgentMode() {
			fmt.Fprintf(cmd.ErrOrStderr(), "ERR: %s\n", message)
		} else {
			fmt.Fprintf(cmd.ErrOrStderr(), "[ERROR] %s\n", message)
		}
		return 1
	}
	return 0
}
