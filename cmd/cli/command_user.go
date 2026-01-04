package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v3"
)

func newCommandUser(app *App) *cli.Command {
	return &cli.Command{
		Name:  "user",
		Usage: "Gets the current user.",
		Action: func(ctx context.Context, command *cli.Command) error {
			user, err := app.Client.Api.User().Get(ctx)
			if err != nil {
				return err
			}

			prettyJSON, _ := json.MarshalIndent(user, "", "  ")
			fmt.Print(string(prettyJSON))
			return nil
		},
	}
}
