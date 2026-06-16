package main

import (
	"os"

	"github.com/MD-Mushfiqur123/lychee/cmd"
)

func main() {
	rootCmd := cmd.NewCLI()
	// Insert "bench" as the first argument to run the Cobra subcommand directly
	newArgs := make([]string, 0, len(os.Args)+1)
	newArgs = append(newArgs, os.Args[0], "bench")
	newArgs = append(newArgs, os.Args[1:]...)
	os.Args = newArgs

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
