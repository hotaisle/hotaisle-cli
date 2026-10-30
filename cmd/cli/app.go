package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/hotaisle/hotaisle-cli/client"
	"github.com/hotaisle/hotaisle-cli/internal/api"
	"github.com/hotaisle/hotaisle-cli/internal/config"
	"github.com/hotaisle/hotaisle-cli/internal/log"

	"github.com/urfave/cli/v3"
)

func Run() {
	// Create a base context that will be canceled on signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	code := Execute(ctx, os.Args, os.Stdout, os.Stderr)
	stop()
	if code != 0 {
		os.Exit(code)
	}
}

// Execute runs a command with explicit output streams and reports its exit code.
// Client options allow embedding the CLI with a custom HTTP transport.
func Execute(ctx context.Context, args []string, stdout, stderr io.Writer, opts ...client.Option) int {
	cmd := NewCommand(opts...)
	cmd.Writer, cmd.ErrWriter = stdout, stderr
	if err := cmd.Run(ctx, args); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrWriter, "Error: %v\n", err)
		return 1
	}
	return 0
}

// NewCommand constructs a CLI without reading credentials or writing files.
func NewCommand(opts ...client.Option) *cli.Command {
	app, _ := makeApp()
	app.Client = api.NewClient("", Version, opts...)
	return app.AppCli
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

	app.Config = config.NewConfig()
	app.Client = api.NewClient(app.Config.ApiToken, Version)

	app.AppCli = &cli.Command{
		Usage: "Manage Hot Aisle resources from your terminal.",
		Version: fmt.Sprintf("%s (commit: %s, branch: %s)\nBuilt by: %s at %s\nGo version: %s",
			Version, Commit, Branch, BuildBy, BuildTime, GoVersion),
		EnableShellCompletion: true,
		DefaultCommand:        "help",
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			if !cmd.Args().Present() || cmd.Args().First() == "help" {
				return ctx, nil
			}
			path := cmd.String("config-file")
			cfg, err := config.Load(&path)
			if err != nil {
				return ctx, fmt.Errorf("load configuration: %w", err)
			}
			app.Config = cfg
			app.Client.Api.SetToken(app.token())
			if err := setupLogging(cfg.LogLevel); err != nil {
				return ctx, err
			}
			return ctx, nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config-file",
				Aliases: []string{"c"},
				Usage:   "Path to the config file",
				Value:   config.Pretty,
				Sources: cli.EnvVars("HOTAISLE_CONFIG_FILE"),
			},
		},
		Commands: makeCommands(app),
	}

	return app, nil
}

func (app *App) token() string {
	if token := strings.TrimSpace(os.Getenv("HOTAISLE_API_TOKEN")); token != "" {
		return token
	}
	return app.Config.ApiToken
}

func (app *App) defaultTeam() string {
	if team := strings.TrimSpace(os.Getenv("HOTAISLE_DEFAULT_TEAM")); team != "" {
		return team
	}
	return app.Config.DefaultTeam
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

// printErrorf prints a formatted error message to stderr
func printErrorf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
}
