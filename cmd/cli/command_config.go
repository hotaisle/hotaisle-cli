package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/hotaisle/hotaisle-cli/internal/config"

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
							return invalid("missing token, set HOTAISLE_API_TOKEN")
						}
						app.Config.ApiToken = token
						err := config.Save(app.Config)
						if err != nil {
							return err
						}
						slog.Debug("Config set", "token", partialToken(token))
						return app.configSaved()
					},
				},
				{
					Name:  "log-level",
					Usage: "Set the log-level. Valid values are: debug, info, warn, error, fatal, panic.",
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						logLevel := strings.TrimSpace(cmd.Args().First())
						if len(logLevel) == 0 {
							return invalid("missing log-level")
						}
						app.Config.LogLevel = logLevel
						err := config.Save(app.Config)
						if err != nil {
							return err
						}
						slog.Info("Config set", "log-level", logLevel)
						return app.configSaved()
					},
				},
				{
					Name:  "default-team",
					Usage: "Set the default team.",
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						team := strings.TrimSpace(cmd.Args().First())
						if len(team) == 0 {
							return invalid("missing default-team")
						}
						app.Config.DefaultTeam = team
						err := config.Save(app.Config)
						if err != nil {
							return err
						}
						slog.Info("Config set", "default-team", team)
						return app.configSaved()
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
						return app.configValue("api_token", app.Config.ApiToken)
					},
				},
				{
					Name:  "log-level",
					Usage: "Get the log-level. Valid values are: debug, info, warn, error, fatal, panic.",
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						return app.configValue("log_level", app.Config.LogLevel)
					},
				},
				{
					Name:  "default-team",
					Usage: "Get the default team.",
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						return app.configValue("default_team", app.Config.DefaultTeam)
					},
				},
			},
		},
	},
}

func partialToken(token string) string {
	if token == "" {
		return ""
	}
	return "[redacted]"
}

func newCommandConfig(app *App) *cli.Command {
	return buildCommand(app, configCommands)
}

func (app *App) configValue(name, value string) error {
	if app.AppCli != nil && app.AppCli.Bool("json") {
		return app.printJSON(map[string]string{name: value})
	}
	_, err := fmt.Fprint(app.writer(), value)
	return err
}

func (app *App) configSaved() error {
	if app.AppCli != nil && app.AppCli.Bool("json") {
		return app.success("Configuration saved")
	}
	return nil
}
