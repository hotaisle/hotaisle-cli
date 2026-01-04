package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"hotaisle-cli/client"

	"github.com/urfave/cli/v3"
)

func newCommandUser(app *App) *cli.Command {
	return &cli.Command{
		Name:  "user",
		Usage: "Manage user account.",
		Commands: []*cli.Command{
			newCommandUserGet(app),
			newCommandUserUpdate(app),
			newCommandUserSSHKeys(app),
			newCommandUserAPIKeys(app),
		},
	}
}

func newCommandUserSSHKeys(app *App) *cli.Command {
	return &cli.Command{
		Name:  "ssh-keys",
		Usage: "Manage SSH keys.",
		Commands: []*cli.Command{
			newCommandUserSSHKeysList(app),
			newCommandUserSSHKeysAdd(app),
			newCommandUserSSHKeysDelete(app),
		},
	}
}

func newCommandUserAPIKeys(app *App) *cli.Command {
	return &cli.Command{
		Name:  "api-keys",
		Usage: "Manage API keys.",
		Commands: []*cli.Command{
			newCommandUserAPIKeysList(app),
			newCommandUserAPIKeysGet(app),
			newCommandUserAPIKeysCreate(app),
			newCommandUserAPIKeysUpdate(app),
			newCommandUserAPIKeysDelete(app),
		},
	}
}

func newCommandUserGet(app *App) *cli.Command {
	return &cli.Command{
		Name:  "get",
		Usage: "Get information about the currently authenticated user.",
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

func newCommandUserUpdate(app *App) *cli.Command {
	return &cli.Command{
		Name:  "update",
		Usage: "Update user profile information.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Usage:    "User's full name",
				Required: true,
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			update := client.UserUpdate{
				Name: command.String("name"),
			}

			user, err := app.Client.Api.User().Update(ctx, update)
			if err != nil {
				return err
			}

			prettyJSON, _ := json.MarshalIndent(user, "", "  ")
			fmt.Print(string(prettyJSON))
			return nil
		},
	}
}

func newCommandUserSSHKeysList(app *App) *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all SSH keys.",
		Action: func(ctx context.Context, command *cli.Command) error {
			keys, err := app.Client.Api.User().GetSSHKeys(ctx)
			if err != nil {
				return err
			}

			prettyJSON, _ := json.MarshalIndent(keys, "", "  ")
			fmt.Print(string(prettyJSON))
			return nil
		},
	}
}

func newCommandUserSSHKeysAdd(app *App) *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "Add a new SSH key.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "key",
				Usage:    "SSH public key in authorized_keys format",
				Required: true,
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			key := client.SSHKeyRequest{
				AuthorizedKey: command.String("key"),
			}

			result, err := app.Client.Api.User().AddSSHKey(ctx, key)
			if err != nil {
				return err
			}

			prettyJSON, _ := json.MarshalIndent(result, "", "  ")
			fmt.Print(string(prettyJSON))
			return nil
		},
	}
}

func newCommandUserSSHKeysDelete(app *App) *cli.Command {
	return &cli.Command{
		Name:  "delete",
		Usage: "Delete an SSH key by fingerprint.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "fingerprint",
				Usage:    "SSH key fingerprint",
				Required: true,
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			err := app.Client.Api.User().DeleteSSHKey(ctx, command.String("fingerprint"))
			if err != nil {
				return err
			}

			fmt.Println("SSH key deleted successfully")
			return nil
		},
	}
}

func newCommandUserAPIKeysList(app *App) *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all API keys.",
		Action: func(ctx context.Context, command *cli.Command) error {
			keys, err := app.Client.Api.User().GetAPIKeys(ctx)
			if err != nil {
				return err
			}

			prettyJSON, _ := json.MarshalIndent(keys, "", "  ")
			fmt.Print(string(prettyJSON))
			return nil
		},
	}
}

func newCommandUserAPIKeysGet(app *App) *cli.Command {
	return &cli.Command{
		Name:  "get",
		Usage: "Get detailed information about a specific API key.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "prefix",
				Usage:    "API key prefix identifier",
				Required: true,
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			key, err := app.Client.Api.User().GetAPIKey(ctx, command.String("prefix"))
			if err != nil {
				return err
			}

			prettyJSON, _ := json.MarshalIndent(key, "", "  ")
			fmt.Print(string(prettyJSON))
			return nil
		},
	}
}

func newCommandUserAPIKeysCreate(app *App) *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a new API key.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "label",
				Usage: "Descriptive label for the API key",
			},
			&cli.StringFlag{
				Name:  "user-role",
				Usage: "User role (owner or user)",
				Value: "user",
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			req := client.UserAPIKeyRequest{
				Label:    command.String("label"),
				UserRole: command.String("user-role"),
			}

			key, err := app.Client.Api.User().CreateAPIKey(ctx, req)
			if err != nil {
				return err
			}

			prettyJSON, _ := json.MarshalIndent(key, "", "  ")
			fmt.Print(string(prettyJSON))
			return nil
		},
	}
}

func newCommandUserAPIKeysUpdate(app *App) *cli.Command {
	return &cli.Command{
		Name:  "update",
		Usage: "Update an existing API key.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "prefix",
				Usage:    "API key prefix identifier",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "label",
				Usage: "Descriptive label for the API key",
			},
			&cli.StringFlag{
				Name:  "user-role",
				Usage: "User role (owner or user)",
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			req := client.UserAPIKeyRequest{
				Label:    command.String("label"),
				UserRole: command.String("user-role"),
			}

			key, err := app.Client.Api.User().UpdateAPIKey(ctx, command.String("prefix"), req)
			if err != nil {
				return err
			}

			prettyJSON, _ := json.MarshalIndent(key, "", "  ")
			fmt.Print(string(prettyJSON))
			return nil
		},
	}
}

func newCommandUserAPIKeysDelete(app *App) *cli.Command {
	return &cli.Command{
		Name:  "delete",
		Usage: "Delete an API key.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "prefix",
				Usage:    "API key prefix identifier",
				Required: true,
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			err := app.Client.Api.User().DeleteAPIKey(ctx, command.String("prefix"))
			if err != nil {
				return err
			}

			fmt.Println("API key deleted successfully")
			return nil
		},
	}
}
