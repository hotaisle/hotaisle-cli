package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hotaisle/hotaisle-cli/client"
	"github.com/urfave/cli/v3"
)

func resourceFlags(name string) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "team", Usage: "Team handle (uses the configured default)"},
		&cli.StringFlag{Name: name, Usage: "Resource name", Required: true},
	}
}

func (app *App) resourceTeam(cmd *cli.Command) (string, error) {
	team := strings.TrimSpace(cmd.String("team"))
	if team == "" {
		team = app.defaultTeam()
	}
	if team == "" {
		return "", invalid("provide --team or run hotaisle setup to select a default team")
	}
	return team, nil
}

func newCommandVMWait(app *App) *cli.Command {
	flags := append(resourceFlags("vm"),
		&cli.StringFlag{Name: "state", Value: "running", Usage: "Power state to wait for"},
		&cli.DurationFlag{Name: "timeout", Value: 5 * time.Minute, Usage: "Maximum total wait, including API requests"},
		&cli.DurationFlag{Name: "interval", Value: 2 * time.Second, Usage: "Time between state checks"},
	)
	return &cli.Command{
		Name: "wait", Usage: "Wait for a VM power state with a deadline; does not verify SSH, cloud-init, or inference readiness.",
		Flags: flags, OnUsageError: usageError,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			team, err := app.resourceTeam(cmd)
			if err != nil {
				return err
			}
			if cmd.Duration("timeout") <= 0 || cmd.Duration("interval") <= 0 || strings.TrimSpace(cmd.String("state")) == "" {
				return invalid("--timeout and --interval must be positive, and --state must not be empty")
			}
			ctx, cancel := context.WithTimeout(ctx, cmd.Duration("timeout"))
			defer cancel()
			state, err := app.waitForVM(ctx, team, cmd.String("vm"), cmd.String("state"), cmd.Duration("interval"))
			if err != nil {
				return err
			}
			return app.printJSON(state)
		},
	}
}

func (app *App) waitForVM(ctx context.Context, team, name, target string, interval time.Duration) (*client.VirtualMachineState, error) {
	ticks := time.Tick(interval)
	last := "not checked"
	for {
		state, err := app.Client.Api.VirtualMachines().GetState(ctx, team, name)
		if err != nil {
			return nil, fmt.Errorf("waiting for VM %s (last state: %s): %w", name, last, err)
		}
		last = state.State
		if strings.EqualFold(last, target) {
			return state, nil
		}
		switch strings.ToLower(last) {
		case "crashed", "failed", "error", "shut off", "shutdown":
			return nil, problem("resource_state", "Inspect hotaisle vm get and vm state before starting or rebuilding the VM.", fmt.Errorf("VM %s entered state %q while waiting for %q", name, last, target))
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("waiting for VM %s (last state: %s): %w", name, last, ctx.Err())
		case <-ticks:
		}
	}
}
