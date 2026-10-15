package utils

import (
	"fmt"
	"os"
	"runtime/debug"
)

// Set at build time by GoReleaser via -X main.version=... etc.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// versionInfo prefers ldflags values. When installed with `go install`,
// ldflags aren't set, so it falls back to the module version from build info.
func versionInfo() (string, string, string) {
	v, c, d := version, commit, date
	if v == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			v = info.Main.Version
		}
	}
	return v, c, d
}

func RunVersion() {
	v, c, d := versionInfo()
	fmt.Fprintf(os.Stdout, "sealed %s (commit %s, built %s)\n", v, c, d)
}
