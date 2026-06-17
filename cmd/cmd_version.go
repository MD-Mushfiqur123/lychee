package cmd

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MD-Mushfiqur123/lychee/version"
)

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show detailed version information",
		Long:  "Display the Lychee version along with Go version, platform, commit, and build date.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ver := strings.TrimPrefix(version.Version, "v")
			fmt.Printf("Lychee  v%s\n", ver)
			fmt.Printf("  Commit:   %s\n", version.Commit)
			fmt.Printf("  Build:    %s\n", version.BuildDate)
			fmt.Printf("  Go:       %s\n", runtime.Version())
			fmt.Printf("  Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
}
