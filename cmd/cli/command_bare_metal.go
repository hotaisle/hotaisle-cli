package cli

import (
	"context"
	"fmt"
	"strconv"

	"hotaisle-cli/client"

	"github.com/urfave/cli/v3"
)

var bareMetalCommands = commandDef{
	Name:  "bm",
	Usage: "Manage bare metal servers.",
	Commands: []commandDef{
		{
			Name:  "list",
			Usage: "List all bare metal servers for a team.",
			Flags: []flagDef{
				{Name: "team", Usage: "Team handle", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				servers, err := app.Client.Api.BareMetal().List(ctx, cmd.String("team"))
				if err != nil {
					return err
				}
				return printJSON(servers)
			},
		},
		{
			Name:  "get",
			Usage: "Get detailed information about a specific bare metal server.",
			Flags: []flagDef{
				{Name: "team", Usage: "Team handle", Required: true},
				{Name: "server", Usage: "Server name", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				server, err := app.Client.Api.BareMetal().Get(ctx, cmd.String("team"), cmd.String("server"))
				if err != nil {
					return err
				}
				return printJSON(server)
			},
		},
		{
			Name:  "reserve",
			Usage: "Reserve a bare metal server.",
			Flags: []flagDef{
				{Name: "team", Usage: "Team handle", Required: true},
				{Name: "description", Usage: "Server description"},
				{Name: "cpu-cores", Usage: "Required CPU cores", Required: true},
				{Name: "ram-gb", Usage: "Required RAM in GB", Required: true},
				{Name: "disk-gb", Usage: "Required Disk in GB", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				cpuCores, err := strconv.ParseUint(cmd.String("cpu-cores"), 10, 64)
				if err != nil {
					return fmt.Errorf("invalid cpu-cores: %w", err)
				}
				ramGB, err := strconv.ParseUint(cmd.String("ram-gb"), 10, 64)
				if err != nil {
					return fmt.Errorf("invalid ram-gb: %w", err)
				}
				diskGB, err := strconv.ParseUint(cmd.String("disk-gb"), 10, 64)
				if err != nil {
					return fmt.Errorf("invalid disk-gb: %w", err)
				}

				resp, err := app.Client.Api.BareMetal().Reserve(ctx, cmd.String("team"), client.BareMetalServerReservation{
					Description: cmd.String("description"),
					Specs: client.BareMetalServerSpecs{
						CPUCores:     cpuCores,
						RAMCapacity:  ramGB,
						DiskCapacity: diskGB,
					},
				})
				if err != nil {
					return err
				}
				return printJSON(resp)
			},
		},
		{
			Name:  "update",
			Usage: "Update a bare metal server's description.",
			Flags: []flagDef{
				{Name: "team", Usage: "Team handle", Required: true},
				{Name: "server", Usage: "Server name", Required: true},
				{Name: "description", Usage: "New description", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				err := app.Client.Api.BareMetal().Update(ctx, cmd.String("team"), cmd.String("server"), client.BareMetalServerUpdate{
					Description: cmd.String("description"),
				})
				if err != nil {
					return err
				}
				fmt.Println("Server updated successfully")
				return nil
			},
		},
		{
			Name:  "delete",
			Usage: "Release a bare metal server back to the pool.",
			Flags: []flagDef{
				{Name: "team", Usage: "Team handle", Required: true},
				{Name: "server", Usage: "Server name", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				err := app.Client.Api.BareMetal().Delete(ctx, cmd.String("team"), cmd.String("server"))
				if err != nil {
					return err
				}
				fmt.Println("Server deleted successfully")
				return nil
			},
		},
		{
			Name:  "available",
			Usage: "List available bare metal server types.",
			Flags: []flagDef{
				{Name: "team", Usage: "Team handle", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				available, err := app.Client.Api.BareMetal().GetAvailable(ctx, cmd.String("team"))
				if err != nil {
					return err
				}
				return printJSON(available)
			},
		},
		{
			Name:  "power",
			Usage: "Manage server power state.",
			Commands: []commandDef{
				{
					Name:  "status",
					Usage: "Get current power state.",
					Flags: []flagDef{
						{Name: "team", Usage: "Team handle", Required: true},
						{Name: "server", Usage: "Server name", Required: true},
					},
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						state, err := app.Client.Api.BareMetal().GetPowerState(ctx, cmd.String("team"), cmd.String("server"))
						if err != nil {
							return err
						}
						return printJSON(state)
					},
				},
				{
					Name:  "on",
					Usage: "Power on the server.",
					Flags: []flagDef{
						{Name: "team", Usage: "Team handle", Required: true},
						{Name: "server", Usage: "Server name", Required: true},
					},
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						err := app.Client.Api.BareMetal().PowerOn(ctx, cmd.String("team"), cmd.String("server"))
						if err != nil {
							return err
						}
						fmt.Println("Power on command sent")
						return nil
					},
				},
				{
					Name:  "shutdown",
					Usage: "Gracefully shutdown the server.",
					Flags: []flagDef{
						{Name: "team", Usage: "Team handle", Required: true},
						{Name: "server", Usage: "Server name", Required: true},
					},
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						err := app.Client.Api.BareMetal().GracefulShutdown(ctx, cmd.String("team"), cmd.String("server"))
						if err != nil {
							return err
						}
						fmt.Println("Graceful shutdown command sent")
						return nil
					},
				},
				{
					Name:  "force-shutdown",
					Usage: "Immediately power off the server.",
					Flags: []flagDef{
						{Name: "team", Usage: "Team handle", Required: true},
						{Name: "server", Usage: "Server name", Required: true},
					},
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						err := app.Client.Api.BareMetal().ForceShutdown(ctx, cmd.String("team"), cmd.String("server"))
						if err != nil {
							return err
						}
						fmt.Println("Force shutdown command sent")
						return nil
					},
				},
				{
					Name:  "reboot",
					Usage: "Warm reboot the server.",
					Flags: []flagDef{
						{Name: "team", Usage: "Team handle", Required: true},
						{Name: "server", Usage: "Server name", Required: true},
					},
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						err := app.Client.Api.BareMetal().WarmReboot(ctx, cmd.String("team"), cmd.String("server"))
						if err != nil {
							return err
						}
						fmt.Println("Warm reboot command sent")
						return nil
					},
				},
				{
					Name:  "cold-reboot",
					Usage: "Cold reboot the server.",
					Flags: []flagDef{
						{Name: "team", Usage: "Team handle", Required: true},
						{Name: "server", Usage: "Server name", Required: true},
					},
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						err := app.Client.Api.BareMetal().ColdReboot(ctx, cmd.String("team"), cmd.String("server"))
						if err != nil {
							return err
						}
						fmt.Println("Cold reboot command sent")
						return nil
					},
				},
				{
					Name:  "ac-reset",
					Usage: "Perform a complete AC reset.",
					Flags: []flagDef{
						{Name: "team", Usage: "Team handle", Required: true},
						{Name: "server", Usage: "Server name", Required: true},
					},
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						err := app.Client.Api.BareMetal().ACReset(ctx, cmd.String("team"), cmd.String("server"))
						if err != nil {
							return err
						}
						fmt.Println("AC reset command sent")
						return nil
					},
				},
			},
		},
		{
			Name:  "reinstall",
			Usage: "Wipe all disks and reinstall the OS.",
			Flags: []flagDef{
				{Name: "team", Usage: "Team handle", Required: true},
				{Name: "server", Usage: "Server name", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				server, err := app.Client.Api.BareMetal().Reinstall(ctx, cmd.String("team"), cmd.String("server"))
				if err != nil {
					return err
				}
				return printJSON(server)
			},
		},
		{
			Name:  "console",
			Usage: "Get a temporary console URL.",
			Flags: []flagDef{
				{Name: "team", Usage: "Team handle", Required: true},
				{Name: "server", Usage: "Server name", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				url, err := app.Client.Api.BareMetal().GetConsoleURL(ctx, cmd.String("team"), cmd.String("server"))
				if err != nil {
					return err
				}
				return printJSON(url)
			},
		},
		{
			Name:  "support-access",
			Usage: "Manage support access.",
			Commands: []commandDef{
				{
					Name:  "enable",
					Usage: "Enable Hot Aisle support access.",
					Flags: []flagDef{
						{Name: "team", Usage: "Team handle", Required: true},
						{Name: "server", Usage: "Server name", Required: true},
					},
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						err := app.Client.Api.BareMetal().EnableSupportAccess(ctx, cmd.String("team"), cmd.String("server"))
						if err != nil {
							return err
						}
						fmt.Println("Support access enabled")
						return nil
					},
				},
				{
					Name:  "disable",
					Usage: "Disable Hot Aisle support access.",
					Flags: []flagDef{
						{Name: "team", Usage: "Team handle", Required: true},
						{Name: "server", Usage: "Server name", Required: true},
					},
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						err := app.Client.Api.BareMetal().DisableSupportAccess(ctx, cmd.String("team"), cmd.String("server"))
						if err != nil {
							return err
						}
						fmt.Println("Support access disabled")
						return nil
					},
				},
			},
		},
	},
}

func newCommandBareMetal(app *App) *cli.Command {
	return buildCommand(app, bareMetalCommands)
}
