package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MD-Mushfiqur123/lychee/api"
	"github.com/MD-Mushfiqur123/lychee/progress"
	"github.com/MD-Mushfiqur123/lychee/types/model"
)

func PushHandler(cmd *cobra.Command, args []string) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	insecure, err := cmd.Flags().GetBool("insecure")
	if err != nil {
		return err
	}

	n := model.ParseName(args[0])
	if strings.HasSuffix(n.Host, ".lychee.ai") || strings.HasSuffix(n.Host, ".lychee.com") {
		_, err := client.Whoami(cmd.Context())
		if err != nil {
			var aErr api.AuthorizationError
			if errors.As(err, &aErr) && aErr.StatusCode == http.StatusUnauthorized {
				fmt.Println("You need to be signed in to push models to lychee.com.")
				fmt.Println()

				if aErr.SigninURL != "" {
					fmt.Printf(ConnectInstructions, aErr.SigninURL)
				}
				return nil
			}

			return err
		}
	}

	p := progress.NewProgress(os.Stderr)
	defer p.Stop()

	bars := make(map[string]*progress.Bar)
	var status string
	var spinner *progress.Spinner

	fn := func(resp api.ProgressResponse) error {
		if resp.Digest != "" {
			if spinner != nil {
				spinner.Stop()
			}

			bar, ok := bars[resp.Digest]
			if !ok {
				msg := resp.Status
				if msg == "" {
					msg = fmt.Sprintf("pushing %s...", resp.Digest[7:19])
				}
				bar = progress.NewBar(msg, resp.Total, resp.Completed)
				bars[resp.Digest] = bar
				p.Add(resp.Digest, bar)
			}

			bar.Set(resp.Completed)
		} else if status != resp.Status {
			if spinner != nil {
				spinner.Stop()
			}

			status = resp.Status
			spinner = progress.NewSpinner(status)
			p.Add(status, spinner)
		}

		return nil
	}

	request := api.PushRequest{Name: args[0], Insecure: insecure}

	if err := client.Push(cmd.Context(), &request, fn); err != nil {
		if spinner != nil {
			spinner.Stop()
		}
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "access denied") || strings.Contains(errStr, "unauthorized") {
			return errors.New("you are not authorized to push to this namespace, create the model under a namespace you own")
		}
		return err
	}

	p.Stop()
	if spinner != nil {
		spinner.Stop()
	}

	destination := n.String()
	if strings.HasSuffix(n.Host, ".lychee.ai") || strings.HasSuffix(n.Host, ".lychee.com") {
		destination = "https://lychee.com/" + strings.TrimSuffix(n.DisplayShortest(), ":latest")
	}
	fmt.Printf("\nYou can find your model at:\n\n")
	fmt.Printf("\t%s\n", destination)

	return nil
}
