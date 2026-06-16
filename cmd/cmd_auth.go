package cmd

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/MD-Mushfiqur123/lychee/api"
)

func SigninHandler(cmd *cobra.Command, args []string) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	user, err := client.Whoami(cmd.Context())
	if err != nil {
		var aErr api.AuthorizationError
		if errors.As(err, &aErr) && aErr.StatusCode == http.StatusUnauthorized {
			fmt.Println("You need to be signed in to Lychee to run Cloud models.")
			fmt.Println()

			if aErr.SigninURL != "" {
				_ = browser.OpenURL(aErr.SigninURL)
				fmt.Printf(ConnectInstructions, aErr.SigninURL)
			}
			return nil
		}
		return err
	}

	if user != nil && user.Name != "" {
		fmt.Printf("You are already signed in as user '%s'\n", user.Name)
		fmt.Println()
		return nil
	}

	return nil
}

func SignoutHandler(cmd *cobra.Command, args []string) error {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}

	err = client.Signout(cmd.Context())
	if err != nil {
		var aErr api.AuthorizationError
		if errors.As(err, &aErr) && aErr.StatusCode == http.StatusUnauthorized {
			fmt.Println("You are not signed in to lychee.com")
			fmt.Println()
			return nil
		} else {
			return err
		}
	}

	fmt.Println("You have signed out of lychee.com")
	fmt.Println()
	return nil
}
