// Package displaypath formats filesystem paths for terminal output.
package displaypath

import (
	"os"
	"path/filepath"
	"strings"
)

// Fold replaces a leading home directory with ~ for display. Only whole path
// segments match, so a sibling such as /home/developer is left unchanged when
// the home directory is /home/dev. Keep absolute paths for filesystem calls.
func Fold(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if path == home {
		return "~"
	}
	if rest, ok := strings.CutPrefix(path, home+string(filepath.Separator)); ok {
		return filepath.Join("~", rest)
	}
	return path
}

// FoldText folds every home directory path inside text such as an error message
// or a tool's output, where paths arrive already formatted.
func FoldText(text string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return text
	}
	var folded strings.Builder
	for {
		start := strings.Index(text, home)
		if start < 0 {
			folded.WriteString(text)
			return folded.String()
		}
		end := start + len(home)
		folded.WriteString(text[:start])
		startsPath := start == 0 || !isPathByte(text[start-1])
		endsSegment := end == len(text) || !isNameByte(text[end])
		if startsPath && endsSegment {
			folded.WriteByte('~')
		} else {
			folded.WriteString(home)
		}
		text = text[end:]
	}
}

func isPathByte(b byte) bool {
	return b == filepath.Separator || isNameByte(b)
}

func isNameByte(b byte) bool {
	switch {
	case b >= 'a' && b <= 'z', b >= 'A' && b <= 'Z', b >= '0' && b <= '9':
		return true
	}
	return b == '.' || b == '-' || b == '_'
}
