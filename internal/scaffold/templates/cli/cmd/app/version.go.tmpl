package main

import (
	"runtime/debug"
	"strings"
)

// version is set by release builds through -ldflags.
var version string

// currentVersion returns the release version, or the module version go install
// records in the binary when no release build set one.
func currentVersion() string {
	info, _ := debug.ReadBuildInfo()
	return resolveVersion(version, info)
}

func resolveVersion(injected string, info *debug.BuildInfo) string {
	if injected != "" {
		return injected
	}
	if info != nil && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return strings.TrimPrefix(info.Main.Version, "v")
	}
	return "dev"
}
