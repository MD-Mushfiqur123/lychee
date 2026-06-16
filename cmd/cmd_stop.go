package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/MD-Mushfiqur123/lychee/api"
)

func StopHandler(cmd *cobra.Command, args []string) error {
	opts := &runOptions{
		Model:     args[0],
		KeepAlive: &api.Duration{Duration: 0},
	}
	if err := loadOrUnloadModel(cmd, opts); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("couldn't find model \"%s\" to stop", args[0])
		}
		return err
	}
	return nil
}
