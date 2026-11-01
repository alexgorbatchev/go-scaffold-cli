package main

import (
	"os"
	"testing"

	"github.com/alexgorbatchev/go-scaffold-cli/internal/hermetic"
)

func TestMain(m *testing.M) {
	os.Exit(hermetic.Run(m))
}
