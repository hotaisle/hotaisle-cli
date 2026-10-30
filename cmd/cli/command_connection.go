package cli

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
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

var (
	sshHost = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.-]*$`)
	sshUser = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`)
)

func newCommandSSH(app *App, bareMetal bool) *cli.Command {
	name := "vm"
	if bareMetal {
		name = "server"
	}
	flags := append(resourceFlags(name), sshOutputFlags()...)
	return &cli.Command{
		Name: "ssh", Usage: "Print an SSH command, or JSON connection details; does not open a connection.",
		Flags: flags, OnUsageError: usageError,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			team, err := app.resourceTeam(cmd)
			if err != nil {
				return err
			}
			return app.printSSH(ctx, cmd, team, cmd.String(name), bareMetal)
		},
	}
}

func sshOutputFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "user", Value: "hotaisle", Usage: "SSH login user"},
		&cli.BoolFlag{Name: "address", Usage: "Print an ssh:// address instead of an SSH command"},
	}
}

func newCommandDirectSSH(app *App) *cli.Command {
	return &cli.Command{
		Name: "ssh", Usage: "Get an SSH command or --address for a VM or bare metal server.",
		Description:  "Examples: hotaisle ssh --vm NAME | hotaisle ssh --server NAME --address\nPrints connection information without opening a connection.",
		OnUsageError: usageError,
		Flags: append([]cli.Flag{
			&cli.StringFlag{Name: "team", Usage: "Team handle (uses the configured default)"},
			&cli.StringFlag{Name: "vm", Usage: "Virtual machine name"},
			&cli.StringFlag{Name: "server", Usage: "Bare metal server name"},
		}, sshOutputFlags()...),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if (cmd.String("vm") == "") == (cmd.String("server") == "") {
				return invalid("provide exactly one of --vm or --server")
			}
			team, err := app.resourceTeam(cmd)
			if err != nil {
				return err
			}
			if cmd.String("server") != "" {
				return app.printSSH(ctx, cmd, team, cmd.String("server"), true)
			}
			return app.printSSH(ctx, cmd, team, cmd.String("vm"), false)
		},
	}
}

type sshConnection struct {
	Host string   `json:"host"`
	Port int64    `json:"port"`
	User string   `json:"user"`
	Argv []string `json:"argv"`
}

func connectionFor(access *client.ExternalService, user string) (*sshConnection, error) {
	if access == nil {
		return nil, problem("resource_state", "Register a key with hotaisle ssh-keys add and check instance readiness.", errors.New("SSH connection details are not available yet"))
	}
	host := access.DNSName
	if host == "" {
		host = access.IPAddress
	}
	if (!sshHost.MatchString(host) && net.ParseIP(host) == nil) || access.Port < 1 || access.Port > 65535 || !sshUser.MatchString(user) {
		return nil, invalid("invalid SSH host, port, or user; inspect the resource's ssh_access field")
	}
	return &sshConnection{host, access.Port, user, []string{"ssh", "-p", strconv.FormatInt(access.Port, 10), "--", user + "@" + host}}, nil
}

func (app *App) sshAccess(ctx context.Context, team, name string, bareMetal bool) (*client.ExternalService, error) {
	if bareMetal {
		server, err := app.Client.Api.BareMetal().Get(ctx, team, name)
		if err != nil {
			return nil, err
		}
		return server.SSHAccess, nil
	}
	vm, err := app.Client.Api.VirtualMachines().Get(ctx, team, name)
	if err != nil {
		return nil, err
	}
	return vm.SSHAccess, nil
}

func (app *App) printSSH(ctx context.Context, cmd *cli.Command, team, name string, bareMetal bool) error {
	access, err := app.sshAccess(ctx, team, name, bareMetal)
	if err != nil {
		return err
	}
	connection, err := connectionFor(access, cmd.String("user"))
	if err != nil {
		return err
	}
	if cmd.Bool("address") {
		address := (&url.URL{Scheme: "ssh", User: url.User(connection.User), Host: net.JoinHostPort(connection.Host, strconv.FormatInt(connection.Port, 10))}).String()
		if structuredOutput(cmd) {
			return app.printJSON(map[string]string{"address": address})
		}
		_, err := fmt.Fprintln(app.writer(), address)
		return err
	}
	if structuredOutput(cmd) {
		return app.printJSON(connection)
	}
	_, err = fmt.Fprintln(app.writer(), strings.Join(connection.Argv, " "))
	return err
}
