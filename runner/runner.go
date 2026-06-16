//go:build darwin

package runner

import (
	"fmt"

	"github.com/lychee/lychee/x/imagegen"
	"github.com/lychee/lychee/x/mlxrunner"
)

func Execute(args []string) error {
	if args[0] == "runner" {
		args = args[1:]
	}

	if len(args) > 0 {
		switch args[0] {
		case "--imagegen-engine":
			return imagegen.Execute(args[1:])
		case "--mlx-engine":
			return mlxrunner.Execute(args[1:])
		}
	}
	return fmt.Errorf("unknown runner engine, expected --imagegen-engine or --mlx-engine")
}
