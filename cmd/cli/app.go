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
		return reportError(cmd, err)
	}
	return 0
}

// NewCommand constructs a CLI without reading credentials or writing files.
func NewCommand(opts ...client.Option) *cli.Command {
	app, _ := makeApp()
	app.Client = api.NewClient("", Version, opts...)
	addHelpCommands(app.AppCli)
	return app.AppCli
}

type App struct {
	AppCli *cli.Command
	Config *config.Config
	Client *api.Client
}

func makeCommands(app *App) []*cli.Command {
	return []*cli.Command{
		newCommandSetup(app),
		newCommandDoctor(app),
		newCommandVirtualMachine(app),
		newCommandBareMetal(app),
		newCommandTeam(app),
		newCommandUser(app),
		newCommandConfig(app),
	}
}

func makeApp() (*App, error) {
	app := &App{}

	app.Config = config.NewConfig()
	app.Client = api.NewClient(app.Config.ApiToken, Version)

	app.AppCli = &cli.Command{
		Usage:       "Manage Hot Aisle resources from your terminal.",
		Description: "Get started with 'hotaisle setup', or use HOTAISLE_API_TOKEN and run 'hotaisle doctor'.\nUse --agent for Markdown or --json for JSON automation output.",
		Version: fmt.Sprintf("%s (commit: %s, branch: %s)\nBuilt by: %s at %s\nGo version: %s",
			Version, Commit, Branch, BuildBy, BuildTime, GoVersion),
		EnableShellCompletion: true,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Present() {
				return invalid("unknown command: " + cmd.Args().First())
			}
			return cli.ShowRootCommandHelp(cmd)
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			if cmd.Bool("json") && cmd.Bool("agent") {
				return ctx, invalid("choose either --json or --agent")
			}
			if localDocumentation(cmd) {
				return ctx, nil
			}
			path := cmd.String("config-file")
			cfg, err := config.Load(&path)
			if err != nil {
				return ctx, problem("configuration", "Check --config-file or HOTAISLE_CONFIG_FILE; help works without a valid config.", err)
			}
			app.Config = cfg
			app.Client.Api.SetToken(app.token())
			if structuredOutput(cmd) {
				slog.SetDefault(slog.New(slog.NewJSONHandler(cmd.ErrWriter, &slog.HandlerOptions{Level: slog.LevelError})))
			} else if err := setupLogging(cfg.LogLevel); err != nil {
				return ctx, err
			}
			return ctx, nil
		},
		ExitErrHandler: func(context.Context, *cli.Command, error) {},
		OnUsageError:   usageError,
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "json", Usage: "Emit JSON results and structured errors with categorized exit codes"},
			&cli.BoolFlag{Name: "agent", Usage: "Emit Markdown results and errors, never prompt, and use categorized exit codes"},
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

func localDocumentation(root *cli.Command) bool {
	if !root.Args().Present() {
		return true
	}
	command := root
	for _, name := range root.Args().Slice() {
		command = command.Command(name)
		if command == nil {
			return false
		}
		if command.Name == "help" {
			return true
		}
	}
	return false
}

// Supply help commands explicitly so their errors follow the same output
// contract as every other command, including nested help and aliases.
func addHelpCommands(parent *cli.Command) {
	for _, child := range parent.Commands {
		addHelpCommands(child)
	}
	if len(parent.Commands) == 0 {
		return
	}
	parent.Commands = append(parent.Commands, &cli.Command{
		Name: "help", Aliases: []string{"h"}, Usage: "Show commands or help for a command.",
		ArgsUsage: "[command...]", OnUsageError: usageError,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			target := parent
			for _, name := range cmd.Args().Slice() {
				target = target.Command(name)
				if target == nil {
					return invalid("unknown help topic: " + name)
				}
			}
			if target == cmd.Root() {
				return cli.ShowRootCommandHelp(target)
			}
			return cli.ShowSubcommandHelp(target)
		},
	})
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
