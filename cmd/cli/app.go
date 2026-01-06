package cli

import (
	"context"
	"fmt"
	"hotaisle-cli/internal/api"
	"hotaisle-cli/internal/config"
	"hotaisle-cli/internal/log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"
)

func Run() {
	// Create a base context that will be canceled on signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	if err := runApp(ctx); err != nil {
		printError(err)
		stop()
		os.Exit(1)
	}
}

func runApp(ctx context.Context) error {
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	app, err := makeApp()
	if err != nil {
		return err
	}
	err = app.AppCli.Run(ctx, os.Args)

	return err
}

type App struct {
	AppCli *cli.Command
	Config *config.Config
	Client *api.Client
}

func makeCommands(app *App) []*cli.Command {
	return []*cli.Command{
		newCommandConfig(app),
		newCommandUser(app),
		newCommandTeam(app),
		newCommandBareMetal(app),
		newCommandVirtualMachine(app),
	}
}

func makeApp() (*App, error) {
	app := &App{}

	cfg, err := config.LoadDefault()
	if err != nil {
		if err := config.Save(cfg); err != nil {
			return nil, err
		}
	}
	app.Config = cfg

	level := app.Config.LogLevel
	err = setupLogging(level)
	if err != nil {
		return nil, err
	}

	app.Client = api.NewClient(app.Config.ApiToken, Version)

	app.AppCli = &cli.Command{
		Usage: "Manage Hot Aisle resources from your terminal.",
		Version: fmt.Sprintf("%s (commit: %s, branch: %s)\nBuilt by: %s at %s\nGo version: %s",
			Version, Commit, Branch, BuildBy, BuildTime, GoVersion),
		EnableShellCompletion: true,
		DefaultCommand:        "help",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config-file",
				Aliases: []string{"c"},
				Usage:   "Path to the config file",
				Value:   config.Pretty,
				Sources: cli.EnvVars("HOTAISLE_CONFIG_FILE"),
				Action: func(ctx context.Context, cmd *cli.Command, s string) error {
					configFile := cmd.String("config-file")
					cfg, err := config.Load(configFile)
					if err != nil {
						return err
					}
					app.Config = cfg
					slog.Info("Loaded config", "file", configFile)

					// commands have a dependency on app.Config
					app.AppCli.Commands = makeCommands(app)

					return nil
				},
			},
		},
		Commands: makeCommands(app),
	}

	return app, nil
}

// setupLogging initializes the logging configuration
func setupLogging(level string) error {
	// Set up logging
	logLevel, err := log.ParseLevel(level)
	if err != nil {
		printErrorf("Invalid log level: %s\n", level)
		logLevel = log.LevelInfo
	}
	slog.SetDefault(log.NewConsoleHandler(logLevel))
	slog.SetLogLoggerLevel(logLevel)
	return nil
}

// printError prints an error message to stderr
func printError(err error) {
	printErrorf("Error: %v\n", err)
}

// printErrorf prints a formatted error message to stderr
func printErrorf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
}
