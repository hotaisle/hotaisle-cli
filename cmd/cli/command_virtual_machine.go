package cli

import (
	"context"
	"fmt"
	"strconv"

	"hotaisle-cli/client"

	"github.com/urfave/cli/v3"
)

var virtualMachineCommands = commandDef{
	Name:  "vm",
	Usage: "Manage virtual machines.",
	Commands: []commandDef{
		{
			Name:  "list",
			Usage: "List all virtual machines for a team.",
			Flags: []flagDef{
				{Name: "team", Usage: "Team handle", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				vms, err := app.Client.Api.VirtualMachines().List(ctx, cmd.String("team"))
				if err != nil {
					return err
				}
				return printJSON(vms)
			},
		},
		{
			Name:  "get",
			Usage: "Get detailed information about a specific virtual machine.",
			Flags: []flagDef{
				{Name: "team", Usage: "Team handle", Required: true},
				{Name: "vm", Usage: "VM name", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				vm, err := app.Client.Api.VirtualMachines().Get(ctx, cmd.String("team"), cmd.String("vm"))
				if err != nil {
					return err
				}
				return printJSON(vm)
			},
		},
		{
			Name:  "provision",
			Usage: "Provision a new virtual machine.",
			Flags: []flagDef{
				{Name: "team", Usage: "Team handle", Required: true},
				{Name: "cpu-cores", Usage: "Required CPU cores"},
				{Name: "ram-gb", Usage: "Required RAM in GB"},
				{Name: "disk-gb", Usage: "Required Disk in GB"},
				{Name: "gpu-model", Usage: "GPU model"},
				{Name: "gpu-count", Usage: "GPU count"},
				{Name: "user-data-url", Usage: "URL for cloud-init user data"},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				var req client.VMProvisionRequest
				req.UserDataURL = cmd.String("user-data-url")

				if cpuCoresStr := cmd.String("cpu-cores"); cpuCoresStr != "" {
					cpuCores, err := strconv.ParseUint(cpuCoresStr, 10, 64)
					if err != nil {
						return fmt.Errorf("invalid cpu-cores: %w", err)
					}
					req.CPUCores = &cpuCores
				}

				if ramGBStr := cmd.String("ram-gb"); ramGBStr != "" {
					ramGB, err := strconv.ParseUint(ramGBStr, 10, 64)
					if err != nil {
						return fmt.Errorf("invalid ram-gb: %w", err)
					}
					req.RAMCapacity = &ramGB
				}

				if diskGBStr := cmd.String("disk-gb"); diskGBStr != "" {
					diskGB, err := strconv.ParseUint(diskGBStr, 10, 64)
					if err != nil {
						return fmt.Errorf("invalid disk-gb: %w", err)
					}
					req.DiskCapacity = &diskGB
				}

				if gpuModel := cmd.String("gpu-model"); gpuModel != "" {
					gpuCountStr := cmd.String("gpu-count")
					if gpuCountStr == "" {
						return fmt.Errorf("gpu-count is required when gpu-model is specified")
					}
					gpuCount, err := strconv.ParseUint(gpuCountStr, 10, 64)
					if err != nil {
						return fmt.Errorf("invalid gpu-count: %w", err)
					}
					req.GPUs = []client.GPUs{{
						Components: client.Components{
							Model: gpuModel,
							Count: gpuCount,
						},
					}}
				}

				if req.CPUCores == nil && req.RAMCapacity == nil && req.DiskCapacity == nil && len(req.GPUs) == 0 {
					return fmt.Errorf("at least one specification must be provided (cpu-cores, ram-gb, disk-gb, or gpu-model)")
				}

				resp, err := app.Client.Api.VirtualMachines().Provision(ctx, cmd.String("team"), req)
				if err != nil {
					return err
				}
				return printJSON(resp)
			},
		},
		{
			Name:  "update",
			Usage: "Update a virtual machine's description.",
			Flags: []flagDef{
				{Name: "team", Usage: "Team handle", Required: true},
				{Name: "vm", Usage: "VM name", Required: true},
				{Name: "description", Usage: "New description", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				err := app.Client.Api.VirtualMachines().Update(ctx, cmd.String("team"), cmd.String("vm"), client.VirtualMachineUpdate{
					Description: cmd.String("description"),
				})
				if err != nil {
					return err
				}
				fmt.Println("VM updated successfully")
				return nil
			},
		},
		{
			Name:  "delete",
			Usage: "Delete a virtual machine and its resources.",
			Flags: []flagDef{
				{Name: "team", Usage: "Team handle", Required: true},
				{Name: "vm", Usage: "VM name", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				err := app.Client.Api.VirtualMachines().Delete(ctx, cmd.String("team"), cmd.String("vm"))
				if err != nil {
					return err
				}
				fmt.Println("VM deleted successfully")
				return nil
			},
		},
		{
			Name:  "available",
			Usage: "List available virtual machine types.",
			Flags: []flagDef{
				{Name: "team", Usage: "Team handle", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				available, err := app.Client.Api.VirtualMachines().GetAvailable(ctx, cmd.String("team"))
				if err != nil {
					return err
				}
				return printJSON(available)
			},
		},
		{
			Name:  "state",
			Usage: "Get current state of a virtual machine.",
			Flags: []flagDef{
				{Name: "team", Usage: "Team handle", Required: true},
				{Name: "vm", Usage: "VM name", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				state, err := app.Client.Api.VirtualMachines().GetState(ctx, cmd.String("team"), cmd.String("vm"))
				if err != nil {
					return err
				}
				return printJSON(state)
			},
		},
		{
			Name:  "start",
			Usage: "Start a stopped virtual machine.",
			Flags: []flagDef{
				{Name: "team", Usage: "Team handle", Required: true},
				{Name: "vm", Usage: "VM name", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				err := app.Client.Api.VirtualMachines().Start(ctx, cmd.String("team"), cmd.String("vm"))
				if err != nil {
					return err
				}
				fmt.Println("VM start command sent")
				return nil
			},
		},
		{
			Name:  "stop",
			Usage: "Forcefully stop a running virtual machine.",
			Flags: []flagDef{
				{Name: "team", Usage: "Team handle", Required: true},
				{Name: "vm", Usage: "VM name", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				err := app.Client.Api.VirtualMachines().Stop(ctx, cmd.String("team"), cmd.String("vm"))
				if err != nil {
					return err
				}
				fmt.Println("VM stop command sent")
				return nil
			},
		},
		{
			Name:  "shutdown",
			Usage: "Gracefully shutdown a virtual machine.",
			Flags: []flagDef{
				{Name: "team", Usage: "Team handle", Required: true},
				{Name: "vm", Usage: "VM name", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				err := app.Client.Api.VirtualMachines().Shutdown(ctx, cmd.String("team"), cmd.String("vm"))
				if err != nil {
					return err
				}
				fmt.Println("VM shutdown command sent")
				return nil
			},
		},
		{
			Name:  "reboot",
			Usage: "Gracefully reboot a virtual machine.",
			Flags: []flagDef{
				{Name: "team", Usage: "Team handle", Required: true},
				{Name: "vm", Usage: "VM name", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				err := app.Client.Api.VirtualMachines().Reboot(ctx, cmd.String("team"), cmd.String("vm"))
				if err != nil {
					return err
				}
				fmt.Println("VM reboot command sent")
				return nil
			},
		},
		{
			Name:  "hard-reset",
			Usage: "Forcefully reset a virtual machine.",
			Flags: []flagDef{
				{Name: "team", Usage: "Team handle", Required: true},
				{Name: "vm", Usage: "VM name", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				err := app.Client.Api.VirtualMachines().HardReset(ctx, cmd.String("team"), cmd.String("vm"))
				if err != nil {
					return err
				}
				fmt.Println("VM hard-reset command sent")
				return nil
			},
		},
		{
			Name:  "rebuild",
			Usage: "Rebuild the virtual machine to its initial state.",
			Flags: []flagDef{
				{Name: "team", Usage: "Team handle", Required: true},
				{Name: "vm", Usage: "VM name", Required: true},
				{Name: "user-data-url", Usage: "New URL for cloud-init user data"},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				err := app.Client.Api.VirtualMachines().Rebuild(ctx, cmd.String("team"), cmd.String("vm"), client.VMResetRequest{
					UserDataURL: cmd.String("user-data-url"),
				})
				if err != nil {
					return err
				}
				fmt.Println("VM rebuild command sent")
				return nil
			},
		},
	},
}

func newCommandVirtualMachine(app *App) *cli.Command {
	return buildCommand(app, virtualMachineCommands)
}
