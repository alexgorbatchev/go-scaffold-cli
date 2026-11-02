package displaypath

import (
	"path/filepath"
	"testing"
)

func TestFold(t *testing.T) {
	t.Setenv("HOME", "/home/dev")
	tests := []struct {
		name string
		path string
		want string
	}{
		{"home", "/home/dev", "~"},
		{"nested", "/home/dev/projects/app", filepath.Join("~", "projects", "app")},
		{"sibling sharing the prefix", "/home/developer/app", "/home/developer/app"},
		{"outside home", "/var/tmp/app", "/var/tmp/app"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Fold(tt.path); got != tt.want {
				t.Fatalf("Fold(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestFoldText(t *testing.T) {
	t.Setenv("HOME", "/home/dev")
	tests := []struct {
		name string
		text string
		want string
	}{
		{"quoted path", `target "/home/dev/app" is not empty`, `target "~/app" is not empty`},
		{"bare home at the end", "stat /home/dev", "stat ~"},
		{"two paths", "/home/dev/a and /home/dev/b", "~/a and ~/b"},
		{"sibling sharing the prefix", "open /home/developer/app", "open /home/developer/app"},
		{"sibling with a hyphen", "open /home/dev-tools/app", "open /home/dev-tools/app"},
		{"home nested in another path", "open /mnt/home/dev/app", "open /mnt/home/dev/app"},
		{"no path", "nothing to fold", "nothing to fold"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FoldText(tt.text); got != tt.want {
				t.Fatalf("FoldText(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}

func TestFoldWithoutHome(t *testing.T) {
	t.Setenv("HOME", "")
	if got := Fold("/home/dev/app"); got != "/home/dev/app" {
		t.Fatalf("Fold = %q", got)
	}
	if got := FoldText("/home/dev/app"); got != "/home/dev/app" {
		t.Fatalf("FoldText = %q", got)
	}
}
