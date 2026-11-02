package scaffold

import (
	"os"
	"testing"

	"github.com/alexgorbatchev/go-scaffold-cli/v2/internal/hermetic"
)

func TestMain(m *testing.M) {
	os.Exit(hermetic.Run(m))
}
