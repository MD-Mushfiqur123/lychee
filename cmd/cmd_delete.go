package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/MD-Mushfiqur123/lychee/api"
)

func DeleteHandler(cmd *cobra.Command, args []string) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	for _, arg := range args {
		// Unload the model if it's running before deletion
		if err := loadOrUnloadModel(cmd, &runOptions{
			Model:     arg,
			KeepAlive: &api.Duration{Duration: 0},
		}); err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "not found") {
				fmt.Fprintf(os.Stderr, "Warning: unable to stop model '%s'\n", arg)
			}
		}

		if err := client.Delete(cmd.Context(), &api.DeleteRequest{Name: arg}); err != nil {
			return err
		}
		fmt.Printf("deleted '%s'\n", arg)
	}
	return nil
}
