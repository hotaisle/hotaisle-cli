package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hotaisle/hotaisle-cli/client"
	"github.com/urfave/cli/v3"
)

func newCommandDoctor(app *App) *cli.Command {
	return &cli.Command{
		Name: "doctor", Usage: "Check authentication, team access, balance, and registered SSH keys without changing anything.",
		OnUsageError: usageError,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if app.token() == "" {
				return missingToken()
			}
			user, err := app.Client.Api.User().Get(ctx)
			if err != nil {
				return err
			}
			if err := validateTeam(user.Teams, app.defaultTeam()); err != nil {
				return err
			}
			keys, err := app.Client.Api.User().GetSSHKeys(ctx)
			if err != nil {
				return err
			}
			balance, err := app.Client.Api.Teams().GetBalance(ctx, app.defaultTeam())
			if err != nil {
				return err
			}
			warnings := []string{}
			if len(keys) == 0 {
				warnings = append(warnings, "Register an SSH public key before provisioning: hotaisle setup --ssh-key-file ~/.ssh/id_ed25519.pub")
			}
			if balance.AvailableBalance <= balance.MinimumBalance {
				warnings = append(warnings, "Team balance is at or below the minimum; provisioning may be unavailable. Check hotaisle team balance --handle "+app.defaultTeam())
			}
			if structuredOutput(cmd) {
				return app.printResult(struct {
					Authenticated bool                `json:"authenticated"`
					Team          string              `json:"team"`
					SSHKeyCount   int                 `json:"ssh_key_count"`
					Balance       *client.BalanceInfo `json:"balance"`
					Warnings      []string            `json:"warnings"`
				}{true, app.defaultTeam(), len(keys), balance, warnings})
			}
			if _, err := fmt.Fprintf(app.writer(), "Authentication: OK\nTeam: %s\nRegistered SSH keys: %d\nBalance: %s USD\nCurrent spend: %s USD/hour\n", app.defaultTeam(), len(keys), textMoney(strconv.FormatInt(balance.AvailableBalance, 10)), textMoney(strconv.FormatInt(balance.HourlyRate, 10))); err != nil {
				return err
			}
			for _, warning := range warnings {
				if _, err := fmt.Fprintln(app.writer(), "Warning:", warning); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
