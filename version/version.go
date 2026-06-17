package version

import "fmt"

// These are set at build time via ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// Info returns a formatted version string
func Info() string {
	return fmt.Sprintf("Lychee %s (commit %s, built %s)", Version, Commit, BuildDate)
}
