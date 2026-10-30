package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hotaisle/hotaisle-cli/client"
	"github.com/urfave/cli/v3"
)

type rentalOption struct {
	Kind                      string        `json:"kind"`
	GPURequirement            string        `json:"gpu_requirement"`
	GPUs                      []client.GPUs `json:"gpus"`
	QuantityAvailable         int64         `json:"quantity_available"`
	CPUCores                  *uint64       `json:"cpu_cores"`
	RAMBytes                  *uint64       `json:"ram_bytes"`
	DiskBytes                 *uint64       `json:"disk_bytes"`
	HourlyPriceCents          *int64        `json:"hourly_price_cents"`
	MinimumReservationMinutes int64         `json:"minimum_reservation_minutes"`
	Specs                     any           `json:"specs"`
}

func newCommandRentals(app *App) *cli.Command {
	flags := func() []cli.Flag {
		return []cli.Flag{
			&cli.StringFlag{Name: "team", Usage: "Team handle (uses the configured default)"},
			&cli.StringFlag{Name: "kind", Value: "all", Usage: "Resource kind: all, vm, or bm"},
		}
	}
	return &cli.Command{
		Name: "rentals", Aliases: []string{"offers"}, Usage: "List available VM and bare metal rental bundles, prices, and minimum reservations.",
		Description: "Each row is a complete offered configuration, not a per-GPU listing. Counts such as 8x MI300X describe the whole bundle. Availability and pricing are scoped to your team.",
		Flags:       flags(), OnUsageError: usageError, Action: app.listRentals,
		Commands: []*cli.Command{{Name: "list", Usage: "List all currently available rental options.", Flags: flags(), OnUsageError: usageError, Action: app.listRentals}},
	}
}

func gpuRequirement(gpus []client.GPUs) string {
	parts := make([]string, 0, len(gpus))
	for _, gpu := range gpus {
		model := gpu.Model
		if model == "" {
			model = "unspecified GPU"
		}
		parts = append(parts, fmt.Sprintf("%dx %s", gpu.Count, model))
	}
	if len(parts) == 0 {
		return "not specified"
	}
	return strings.Join(parts, " + ")
}

func quotedPrice(cents int64) *int64 {
	// The API omits zero-valued prices. Do not present a missing quote as free.
	if cents <= 0 {
		return nil
	}
	return new(cents)
}

func (app *App) listRentals(ctx context.Context, cmd *cli.Command) error {
	kind := cmd.String("kind")
	if kind != "all" && kind != "vm" && kind != "bm" {
		return invalid("--kind must be all, vm, or bm")
	}
	team, err := app.resourceTeam(cmd)
	if err != nil {
		return err
	}
	options := []rentalOption{}
	if kind != "bm" {
		available, err := app.Client.Api.VirtualMachines().GetAvailable(ctx, team)
		if err != nil {
			return err
		}
		for _, vm := range available {
			options = append(options, rentalOption{
				Kind: "vm", GPURequirement: gpuRequirement(vm.Specs.GPUs), GPUs: vm.Specs.GPUs,
				QuantityAvailable: vm.Quantity, CPUCores: vm.Specs.CPUCores, RAMBytes: vm.Specs.RAMCapacity, DiskBytes: vm.Specs.DiskCapacity,
				HourlyPriceCents: quotedPrice(vm.OnDemandPrice), MinimumReservationMinutes: vm.MinimumReservationMinutes, Specs: vm.Specs,
			})
		}
	}
	if kind != "vm" {
		available, err := app.Client.Api.BareMetal().GetAvailable(ctx, team)
		if err != nil {
			return err
		}
		for _, server := range available {
			options = append(options, rentalOption{
				Kind: "bm", GPURequirement: gpuRequirement(server.Specs.GPUs), GPUs: server.Specs.GPUs,
				QuantityAvailable: server.Quantity, CPUCores: new(server.Specs.CPUCores), RAMBytes: new(server.Specs.RAMCapacity), DiskBytes: new(server.Specs.DiskCapacity),
				HourlyPriceCents: quotedPrice(server.OnDemandPrice), MinimumReservationMinutes: server.MinimumReservationMinutes, Specs: server.Specs,
			})
		}
	}
	if cmd.Bool("json") {
		return app.printJSON(struct {
			Team     string         `json:"team"`
			Currency string         `json:"currency"`
			Options  []rentalOption `json:"options"`
		}{team, "USD", options})
	}
	if !cmd.Bool("agent") {
		return app.printRentalTable(team, options)
	}
	var out strings.Builder
	fmt.Fprintf(&out, "# Rental options\n\nTeam: %s. Prices are USD per hour. Each row is one complete rental bundle.\n\n", markdownCode(team))
	if len(options) == 0 {
		out.WriteString("No rental options are currently available for this team.\n")
	} else {
		out.WriteString("| Kind | GPU requirement | Available bundles | CPU cores | RAM (GiB) | Disk (GiB) | USD/hour | Minimum reservation (minutes) |\n| --- | --- | --- | --- | --- | --- | --- | --- |\n")
		for _, option := range options {
			cells := option.cells()
			cells[1] = markdownCode(cells[1])
			fmt.Fprintf(&out, "| %s |\n", strings.Join(cells, " | "))
		}
	}
	out.WriteString("\nThese are currently available configurations, not a catalog of sold-out types. Do not assume an 8-GPU bundle can be rented one GPU at a time.\n")
	_, err = fmt.Fprint(app.writer(), out.String())
	return err
}

func (option rentalOption) cells() []string {
	price := "not quoted"
	if option.HourlyPriceCents != nil {
		price = fmt.Sprintf("%d.%02d", *option.HourlyPriceCents/100, *option.HourlyPriceCents%100)
	}
	return []string{
		option.Kind, option.GPURequirement, strconv.FormatInt(option.QuantityAvailable, 10), optionalUint(option.CPUCores),
		gibibytes(option.RAMBytes), gibibytes(option.DiskBytes), price, strconv.FormatInt(option.MinimumReservationMinutes, 10),
	}
}

func (app *App) printRentalTable(team string, options []rentalOption) error {
	var out strings.Builder
	fmt.Fprintf(&out, "Rental options\nTeam: %s\n\n", terminalCell(team))
	if len(options) == 0 {
		out.WriteString("No rental options are currently available for this team.\n")
	} else {
		rows := make([][]string, 0, len(options))
		for _, option := range options {
			rows = append(rows, option.cells())
		}
		table := newTextTable(
			[]string{"Kind", "GPU bundle", "Avail.", "CPUs", "RAM GiB", "Disk GiB", "USD/hr", "Min (min)"},
			rows, []bool{false, false, true, true, true, true, true, true},
		)
		table.write(&out, "")
	}
	out.WriteString("\nPrices are USD per hour for one complete rental bundle.\nAvailability excludes sold-out configurations; GPU bundles are rented whole.\n")
	_, err := fmt.Fprint(app.writer(), out.String())
	return err
}

func optionalUint(value *uint64) string {
	if value == nil {
		return "not specified"
	}
	return strconv.FormatUint(*value, 10)
}

func gibibytes(value *uint64) string {
	if value == nil {
		return "not specified"
	}
	return strconv.FormatFloat(float64(*value)/(1<<30), 'f', 2, 64)
}
