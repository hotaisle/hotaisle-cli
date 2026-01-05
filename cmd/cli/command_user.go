package cli

import (
	"context"
	"fmt"

	"hotaisle-cli/client"

	"github.com/urfave/cli/v3"
)

var userCommands = commandDef{
	Name:  "user",
	Usage: "Manage user account.",
	Commands: []commandDef{
		{
			Name:  "get",
			Usage: "Get information about the currently authenticated user.",
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				user, err := app.Client.Api.User().Get(ctx)
				if err != nil {
					return err
				}
				return printJSON(user)
			},
		},
		{
			Name:  "update",
			Usage: "Update user profile information.",
			Flags: []flagDef{
				{Name: "name", Usage: "User's full name", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				user, err := app.Client.Api.User().Update(ctx, client.UserUpdate{
					Name: cmd.String("name"),
				})
				if err != nil {
					return err
				}
				return printJSON(user)
			},
		},
		{
			Name:  "ssh-keys",
			Usage: "Manage SSH keys.",
			Commands: []commandDef{
				{
					Name:  "list",
					Usage: "List all SSH keys.",
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						keys, err := app.Client.Api.User().GetSSHKeys(ctx)
						if err != nil {
							return err
						}
						return printJSON(keys)
					},
				},
				{
					Name:  "add",
					Usage: "Add a new SSH key.",
					Flags: []flagDef{
						{Name: "key", Usage: "SSH public key in authorized_keys format", Required: true},
					},
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						result, err := app.Client.Api.User().AddSSHKey(ctx, client.SSHKeyRequest{
							AuthorizedKey: cmd.String("key"),
						})
						if err != nil {
							return err
						}
						return printJSON(result)
					},
				},
				{
					Name:  "delete",
					Usage: "Delete an SSH key by fingerprint.",
					Flags: []flagDef{
						{Name: "fingerprint", Usage: "SSH key fingerprint", Required: true},
					},
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						err := app.Client.Api.User().DeleteSSHKey(ctx, cmd.String("fingerprint"))
						if err != nil {
							return err
						}
						fmt.Println("SSH key deleted successfully")
						return nil
					},
				},
			},
		},
		{
			Name:  "api-keys",
			Usage: "Manage API keys.",
			Commands: []commandDef{
				{
					Name:  "list",
					Usage: "List all API keys.",
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						keys, err := app.Client.Api.User().GetAPIKeys(ctx)
						if err != nil {
							return err
						}
						return printJSON(keys)
					},
				},
				{
					Name:  "get",
					Usage: "Get detailed information about a specific API key.",
					Flags: []flagDef{
						{Name: "prefix", Usage: "API key prefix identifier", Required: true},
					},
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						key, err := app.Client.Api.User().GetAPIKey(ctx, cmd.String("prefix"))
						if err != nil {
							return err
						}
						return printJSON(key)
					},
				},
				{
					Name:  "create",
					Usage: "Create a new API key.",
					Flags: []flagDef{
						{Name: "label", Usage: "Descriptive label for the API key"},
						{Name: "user-role", Usage: "User role (owner or user)", Value: "user"},
					},
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						key, err := app.Client.Api.User().CreateAPIKey(ctx, client.UserAPIKeyRequest{
							Label:    cmd.String("label"),
							UserRole: cmd.String("user-role"),
						})
						if err != nil {
							return err
						}
						return printJSON(key)
					},
				},
				{
					Name:  "update",
					Usage: "Update an existing API key.",
					Flags: []flagDef{
						{Name: "prefix", Usage: "API key prefix identifier", Required: true},
						{Name: "label", Usage: "Descriptive label for the API key"},
						{Name: "user-role", Usage: "User role (owner or user)"},
					},
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						key, err := app.Client.Api.User().UpdateAPIKey(ctx, cmd.String("prefix"), client.UserAPIKeyRequest{
							Label:    cmd.String("label"),
							UserRole: cmd.String("user-role"),
						})
						if err != nil {
							return err
						}
						return printJSON(key)
					},
				},
				{
					Name:  "delete",
					Usage: "Delete an API key.",
					Flags: []flagDef{
						{Name: "prefix", Usage: "API key prefix identifier", Required: true},
					},
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						err := app.Client.Api.User().DeleteAPIKey(ctx, cmd.String("prefix"))
						if err != nil {
							return err
						}
						fmt.Println("API key deleted successfully")
						return nil
					},
				},
			},
		},
	},
}

func newCommandUser(app *App) *cli.Command {
	return buildCommand(app, userCommands)
}
