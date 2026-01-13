package cli

import (
	"context"
	"errors"
	"fmt"
	"hotaisle-cli/internal/config"
	"log/slog"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

var configCommands = commandDef{
	Name:  "config",
	Usage: "Config File Management",
	Commands: []commandDef{
		{
			Name:  "set",
			Usage: "Set a configuration value.",
			Commands: []commandDef{
				{
					Name:  "token",
					Usage: "Set the API token from the HOTAISLE_API_TOKEN environment variable.",
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						token := strings.TrimSpace(os.Getenv("HOTAISLE_API_TOKEN"))
						if len(token) == 0 {
							return errors.New("missing token, set HOTAISLE_API_TOKEN")
						}
						app.Config.ApiToken = token
						err := config.Save(app.Config)
						if err != nil {
							return err
						}
						slog.Debug("Config set", "token", partialToken(token))
						return nil
					},
				},
				{
					Name:  "log-level",
					Usage: "Set the log-level. Valid values are: debug, info, warn, error, fatal, panic.",
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						logLevel := strings.TrimSpace(cmd.Args().First())
						if len(logLevel) == 0 {
							return errors.New("missing log-level")
						}
						app.Config.LogLevel = logLevel
						err := config.Save(app.Config)
						if err != nil {
							return err
						}
						slog.Info("Config set", "log-level", logLevel)
						return nil
					},
				},
				{
					Name:  "default-team",
					Usage: "Set the default team.",
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						team := strings.TrimSpace(cmd.Args().First())
						if len(team) == 0 {
							return errors.New("missing default-team")
						}
						app.Config.DefaultTeam = team
						err := config.Save(app.Config)
						if err != nil {
							return err
						}
						slog.Info("Config set", "default-team", team)
						return nil
					},
				},
			},
		},
		{
			Name:  "get",
			Usage: "Get a configuration value.",
			Commands: []commandDef{
				{
					Name:  "token",
					Usage: "Get the API token.",
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						fmt.Print(app.Config.ApiToken)
						return nil
					},
				},
				{
					Name:  "log-level",
					Usage: "Get the log-level. Valid values are: debug, info, warn, error, fatal, panic.",
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						fmt.Print(app.Config.LogLevel)
						return nil
					},
				},
				{
					Name:  "default-team",
					Usage: "Get the default team.",
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						fmt.Print(app.Config.DefaultTeam)
						return nil
					},
				},
			},
		},
	},
}

func partialToken(token string) string {
	tokens := strings.Split(token, ".")
	if len(tokens) > 0 {
		return tokens[0]
	}
	return "Not a valid token"
}

func newCommandConfig(app *App) *cli.Command {
	return buildCommand(app, configCommands)
}
