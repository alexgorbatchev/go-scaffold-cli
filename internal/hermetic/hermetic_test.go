package hermetic

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(Run(m))
}

func TestRunStubsGitHubCLI(t *testing.T) {
	path, err := exec.LookPath("gh")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(path, "hermetic") {
		t.Fatalf("gh resolves to %s", path)
	}
	if err := exec.Command("gh", "api", "user").Run(); err == nil {
		t.Fatal("gh must fail")
	}
}

func TestRunDisablesModuleDownloads(t *testing.T) {
	out, err := exec.Command("go", "env", "GOPROXY", "GOSUMDB").Output()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Fields(string(out)); len(got) != 2 || got[0] != "off" || got[1] != "off" {
		t.Fatalf("GOPROXY and GOSUMDB = %v", got)
	}
}
