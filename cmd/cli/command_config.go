package cli

import (
	"context"
	"fmt"
	"hotaisle-cli/internal/config"
	"log/slog"

	"github.com/urfave/cli/v3"
)

func newCommandConfig(app *App) *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Config File Management",
		Commands: []*cli.Command{{
			Name:  "set",
			Usage: "Set a configuration value.",
			Commands: []*cli.Command{{
				Name:  "token",
				Usage: "Set the API token.",
				Action: func(ctx context.Context, command *cli.Command) error {
					token := command.Args().First()
					if len(token) > 0 {
						app.Config.ApiToken = token
						err := config.Save(app.Config)
						if err != nil {
							return err
						}
						slog.Info("Config set", "token", token)
					}
					return nil
				},
			}, {
				Name:  "log-level",
				Usage: "Set the log-level. Valid values are: debug, info, warn, error, fatal, panic.",
				Action: func(ctx context.Context, command *cli.Command) error {
					token := command.Args().First()
					if len(token) > 0 {
						app.Config.LogLevel = token
						err := config.Save(app.Config)
						if err != nil {
							return err
						}
						slog.Info("Config set", "log-level", token)
					}
					return nil
				},
			}, {
				Name:  "default-team",
				Usage: "Set the default team.",
				Action: func(ctx context.Context, command *cli.Command) error {
					token := command.Args().First()
					if len(token) > 0 {
						app.Config.DefaultTeam = token
						err := config.Save(app.Config)
						if err != nil {
							return err
						}
						slog.Info("Config set", "default-team", token)
					}
					return nil
				},
			}},
		}, {
			Name:  "get",
			Usage: "Get a configuration value.",
			Commands: []*cli.Command{{
				Name:  "token",
				Usage: "Get the API token.",
				Action: func(ctx context.Context, command *cli.Command) error {
					fmt.Print(app.Config.ApiToken)
					return nil
				},
			}, {
				Name:  "log-level",
				Usage: "Get the log-level. Valid values are: debug, info, warn, error, fatal, panic.",
				Action: func(ctx context.Context, command *cli.Command) error {
					fmt.Print(app.Config.LogLevel)
					return nil
				},
			}, {
				Name:  "default-team",
				Usage: "Get the default team.",
				Action: func(ctx context.Context, command *cli.Command) error {
					fmt.Print(app.Config.DefaultTeam)
					return nil
				},
			}},
		}},
	}
}
