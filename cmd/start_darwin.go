package cmd

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"regexp"

	"github.com/MD-Mushfiqur123/lychee/api"
)

var errNotRunning = errors.New("could not connect to lychee server, run 'lychee serve' to start it")

func startApp(ctx context.Context, client *api.Client) error {
	exe, err := os.Executable()
	if err != nil {
		return errNotRunning
	}
	link, err := os.Readlink(exe)
	if err != nil {
		return errNotRunning
	}
	r := regexp.MustCompile(`^.*/Lychee\s?\d*.app`)
	m := r.FindStringSubmatch(link)
	if len(m) != 1 {
		return errNotRunning
	}
	if err := exec.Command("/usr/bin/open", "-j", "-a", m[0], "--args", "--fast-startup").Run(); err != nil {
		return err
	}
	return waitForServer(ctx, client)
}
