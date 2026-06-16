package main

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/MD-Mushfiqur123/lychee/cmd"
)

func main() {
	cobra.CheckErr(cmd.NewCLI().ExecuteContext(context.Background()))
}
