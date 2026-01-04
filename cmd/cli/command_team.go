package cli

import (
	"context"
	"fmt"

	"hotaisle-cli/client"

	"github.com/urfave/cli/v3"
)

var teamCommands = commandDef{
	Name:  "team",
	Usage: "Manage teams.",
	Commands: []commandDef{
		{
			Name:  "list",
			Usage: "List all teams you belong to.",
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				teams, err := app.Client.Api.Teams().List(ctx)
				if err != nil {
					return err
				}
				return printJSON(teams)
			},
		},
		{
			Name:  "create",
			Usage: "Create a new team.",
			Flags: []flagDef{
				{Name: "handle", Usage: "Team handle (slug)", Required: true},
				{Name: "name", Usage: "Team name", Required: true},
				{Name: "description", Usage: "Team description"},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				team, err := app.Client.Api.Teams().Create(ctx, client.Team{
					Handle:      cmd.String("handle"),
					Name:        cmd.String("name"),
					Description: cmd.String("description"),
				})
				if err != nil {
					return err
				}
				return printJSON(team)
			},
		},
		{
			Name:  "get",
			Usage: "Get detailed information about a team.",
			Flags: []flagDef{
				{Name: "handle", Usage: "Team handle", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				team, err := app.Client.Api.Teams().Get(ctx, cmd.String("handle"))
				if err != nil {
					return err
				}
				return printJSON(team)
			},
		},
		{
			Name:  "update",
			Usage: "Update team information.",
			Flags: []flagDef{
				{Name: "handle", Usage: "Team handle", Required: true},
				{Name: "name", Usage: "Team name"},
				{Name: "description", Usage: "Team description"},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				team, err := app.Client.Api.Teams().Update(ctx, cmd.String("handle"), client.TeamUpdate{
					Handle:      cmd.String("handle"),
					Name:        cmd.String("name"),
					Description: cmd.String("description"),
				})
				if err != nil {
					return err
				}
				return printJSON(team)
			},
		},
		{
			Name:  "invitations",
			Usage: "List your pending team invitations.",
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				invitations, err := app.Client.Api.Teams().GetInvitations(ctx)
				if err != nil {
					return err
				}
				return printJSON(invitations)
			},
		},
		{
			Name:  "accept",
			Usage: "Accept a team invitation.",
			Flags: []flagDef{
				{Name: "handle", Usage: "Team handle", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				team, err := app.Client.Api.Teams().AcceptInvitation(ctx, cmd.String("handle"))
				if err != nil {
					return err
				}
				return printJSON(team)
			},
		},
		{
			Name:  "balance",
			Usage: "Get team balance information.",
			Flags: []flagDef{
				{Name: "handle", Usage: "Team handle", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				balance, err := app.Client.Api.Teams().GetBalance(ctx, cmd.String("handle"))
				if err != nil {
					return err
				}
				return printJSON(balance)
			},
		},
		{
			Name:  "purchase-credits",
			Usage: "Create a checkout session to purchase team credits.",
			Flags: []flagDef{
				{Name: "handle", Usage: "Team handle", Required: true},
				{Name: "cents", Usage: "Amount in cents", Required: true},
			},
			Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
				resp, err := app.Client.Api.Teams().PurchaseCredits(ctx, cmd.String("handle"), client.PurchaseTeamCreditsRequest{
					Cents: int64(cmd.Int("cents")),
				})
				if err != nil {
					return err
				}
				return printJSON(resp)
			},
		},
		{
			Name:  "members",
			Usage: "Manage team members.",
			Commands: []commandDef{
				{
					Name:  "list",
					Usage: "List team members and pending invitations.",
					Flags: []flagDef{
						{Name: "handle", Usage: "Team handle", Required: true},
					},
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						team, err := app.Client.Api.Teams().Get(ctx, cmd.String("handle"))
						if err != nil {
							return err
						}
						return printJSON(team.Members)
					},
				},
				{
					Name:  "invitations",
					Usage: "List pending team invitations.",
					Flags: []flagDef{
						{Name: "handle", Usage: "Team handle", Required: true},
					},
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						invitations, err := app.Client.Api.Teams().GetTeamInvitations(ctx, cmd.String("handle"))
						if err != nil {
							return err
						}
						return printJSON(invitations)
					},
				},
				{
					Name:  "invite",
					Usage: "Invite a new member to the team.",
					Flags: []flagDef{
						{Name: "handle", Usage: "Team handle", Required: true},
						{Name: "email", Usage: "User email", Required: true},
						{Name: "name", Usage: "User name", Required: true},
						{Name: "role", Usage: "User role (owner, admin, member, etc.)", Value: "member"},
					},
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						err := app.Client.Api.Teams().InviteMember(ctx, cmd.String("handle"), client.TeamInvitationRequest{
							Email: cmd.String("email"),
							Name:  cmd.String("name"),
							Roles: []string{cmd.String("role")},
						})
						if err != nil {
							return err
						}
						fmt.Println("Invitation sent successfully")
						return nil
					},
				},
				{
					Name:  "update",
					Usage: "Update team member roles.",
					Flags: []flagDef{
						{Name: "handle", Usage: "Team handle", Required: true},
						{Name: "email", Usage: "User email", Required: true},
						{Name: "role", Usage: "User role", Required: true},
					},
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						member, err := app.Client.Api.Teams().UpdateMember(ctx, cmd.String("handle"), cmd.String("email"), client.TeamMemberUpdate{
							Roles: []string{cmd.String("role")},
						})
						if err != nil {
							return err
						}
						return printJSON(member)
					},
				},
				{
					Name:  "remove",
					Usage: "Remove a member from the team.",
					Flags: []flagDef{
						{Name: "handle", Usage: "Team handle", Required: true},
						{Name: "email", Usage: "User email", Required: true},
					},
					Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
						err := app.Client.Api.Teams().RemoveMember(ctx, cmd.String("handle"), cmd.String("email"))
						if err != nil {
							return err
						}
						fmt.Println("Member removed successfully")
						return nil
					},
				},
			},
		},
	},
}

func newCommandTeam(app *App) *cli.Command {
	return buildCommand(app, teamCommands)
}
