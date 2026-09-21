package agent

import (
	cobrahelptree "github.com/alexgorbatchev/cobra-help-tree/v2"
)

// IsAgentMode checks if the environment variable AGENT is set to 1, true, or yes.
func IsAgentMode() bool {
	return cobrahelptree.IsAgentMode()
}
