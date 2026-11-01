package main

import (
	"fmt"
	"io"

	"github.com/alexgorbatchev/go-scaffold-cli/internal/scaffold"
)

// printSetupResult reports warnings and the setup steps a human-mode create finished.
func printSetupResult(w io.Writer, res *scaffold.Result) {
	for _, warning := range res.Warnings {
		fmt.Fprintf(w, "[WARN] %s\n", warning)
	}
	if res.GitInitialized {
		fmt.Fprintln(w, "[OK] Initialized git repository")
	}
	if res.TidyCompleted {
		fmt.Fprintln(w, "[OK] Ran go mod tidy")
	}
}
