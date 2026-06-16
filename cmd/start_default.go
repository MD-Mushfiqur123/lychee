//go:build !windows && !darwin

package cmd

import (
	"context"
	"errors"

	"github.com/MD-Mushfiqur123/lychee/api"
)

func startApp(ctx context.Context, client *api.Client) error {
	return errors.New("could not connect to lychee server, run 'lychee serve' to start it")
}
