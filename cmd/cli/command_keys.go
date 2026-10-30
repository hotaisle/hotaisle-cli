package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/hotaisle/hotaisle-cli/client"
	"github.com/urfave/cli/v3"
)

const sshKeyPropagation = "Managed instances fetch account SSH keys at the next login as hotaisle, including running instances, for teams where the account has the user role. No rebuild is required. Custom SSH configurations or an unavailable key service can prevent this; a guest may fall back to cached keys."

func newCommandWhoami(app *App) *cli.Command {
	def := *userCommands.findCommand("get")
	def.Name = "whoami"
	return buildCommand(app, def)
}

func newCommandAPIKeys(app *App) *cli.Command {
	cmd := buildCommand(app, *userCommands.findCommand("api-keys"))
	cmd.Aliases = []string{"api-key"}
	cmd.Usage = "Manage API credentials; use ssh-keys for instance access."
	return cmd
}

func newCommandSSHKeys(app *App) *cli.Command {
	cmd := buildCommand(app, *userCommands.findCommand("ssh-keys"))
	cmd.Aliases = []string{"ssh-key"}
	cmd.Usage = "Manage account SSH keys for new and running managed instances."
	cmd.Description = sshKeyPropagation + "\nAdding a key requires an API token with personal owner permission."
	for i, child := range cmd.Commands {
		if child.Name == "add" {
			cmd.Commands[i] = &cli.Command{
				Name: "add", Usage: "Register a public key once; optionally check the target team and instance.",
				ArgsUsage: "[PUBLIC_KEY_FILE]", OnUsageError: usageError,
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "file", Aliases: []string{"key-file"}, Usage: "SSH public key file (.pub)"},
					&cli.StringFlag{Name: "key", Usage: "SSH public key in authorized_keys format"},
					&cli.BoolFlag{Name: "stdin", Usage: "Read the public key from stdin"},
					&cli.StringFlag{Name: "team", Usage: "Check this team's SSH permissions (does not restrict the account key to this team)"},
					&cli.StringFlag{Name: "vm", Usage: "Check access to this existing VM and return SSH details"},
					&cli.StringFlag{Name: "server", Usage: "Check access to this existing bare metal server and return SSH details"},
				},
				Action: app.addSSHKey,
			}
		}
	}
	return cmd
}

func publicKeyInput(cmd *cli.Command) (string, error) {
	file := cmd.String("file")
	key := cmd.String("key")
	inputs := 0
	for _, set := range []bool{file != "", key != "", cmd.Bool("stdin"), cmd.Args().Present()} {
		if set {
			inputs++
		}
	}
	if inputs != 1 || cmd.NArg() > 1 {
		return "", invalid("provide exactly one public key file, --file, --key, or --stdin")
	}
	if cmd.Args().Present() {
		file = cmd.Args().First()
	}
	if file != "" {
		return readPublicKey(file)
	}
	if cmd.Bool("stdin") {
		data, err := io.ReadAll(io.LimitReader(cmd.Root().Reader, 65537))
		if err != nil {
			return "", err
		}
		if len(data) > 65536 {
			return "", invalid("public key input exceeds 64 KiB")
		}
		key = string(data)
	}
	return validatePublicKey(key)
}

func findSSHKey(keys []client.SSHKey, authorizedKey string) *client.SSHKey {
	parts := strings.Fields(authorizedKey)
	for i, key := range keys {
		stored := strings.Fields(key.PublicKey)
		if (key.Type == parts[0] && len(stored) == 1 && stored[0] == parts[1]) ||
			(len(stored) >= 2 && stored[0] == parts[0] && stored[1] == parts[1]) {
			return &keys[i]
		}
	}
	return nil
}

func (app *App) ensureSSHKey(ctx context.Context, key string) (*client.SSHKey, bool, error) {
	keys, err := app.Client.Api.User().GetSSHKeys(ctx)
	if err != nil {
		return nil, false, err
	}
	return app.registerSSHKey(ctx, key, keys)
}

func (app *App) registerSSHKey(ctx context.Context, key string, keys []client.SSHKey) (*client.SSHKey, bool, error) {
	if existing := findSSHKey(keys, key); existing != nil {
		return existing, true, nil
	}
	added, err := app.Client.Api.User().AddSSHKey(ctx, client.SSHKeyRequest{AuthorizedKey: key})
	if apiErr, ok := errors.AsType[*client.APIError](err); ok && apiErr.StatusCode == http.StatusConflict {
		// A concurrent registration is successful only if this account can read
		// the same key. A key belonging to a different account remains a conflict.
		if keys, readErr := app.Client.Api.User().GetSSHKeys(ctx); readErr == nil {
			if existing := findSSHKey(keys, key); existing != nil {
				return existing, true, nil
			}
		}
	}
	if err != nil {
		if apiErr, ok := errors.AsType[*client.APIError](err); ok && apiErr.StatusCode == http.StatusForbidden {
			return nil, false, problem("authentication", "Use an API token with personal owner permission to register SSH keys.", err)
		}
		return nil, false, err
	}
	return added, false, nil
}

type sshKeyTarget struct {
	Team       string         `json:"team"`
	Kind       string         `json:"kind"`
	Name       string         `json:"name"`
	Connection *sshConnection `json:"connection"`
}

func (app *App) addSSHKey(ctx context.Context, cmd *cli.Command) error {
	if cmd.String("vm") != "" && cmd.String("server") != "" {
		return invalid("choose either --vm or --server")
	}
	key, err := publicKeyInput(cmd)
	if err != nil {
		return err
	}
	var target *sshKeyTarget
	team := strings.TrimSpace(cmd.String("team"))
	if cmd.String("vm") != "" || cmd.String("server") != "" {
		team, err = app.resourceTeam(cmd)
		if err != nil {
			return err
		}
		target = &sshKeyTarget{Team: team, Kind: "vm", Name: cmd.String("vm")}
		if cmd.String("server") != "" {
			target.Kind, target.Name = "bm", cmd.String("server")
		}
	}
	if team != "" {
		user, err := app.Client.Api.User().Get(ctx)
		if err != nil {
			return err
		}
		if err := validateTeam(user.Teams, team); err != nil {
			return err
		}
		for _, membership := range user.Teams {
			if membership.Handle == team && !membership.Invitation && !canUseSSH(membership) {
				return problem("authentication", "Ask a team owner for the user role; registering a key does not grant team permissions.", fmt.Errorf("account lacks SSH access to team %q", team))
			}
		}
	}
	if target != nil {
		// Validate the named instance before mutating the account. SSH metadata
		// may legitimately be absent until the account has a registered key.
		if _, err := app.sshAccess(ctx, team, target.Name, target.Kind == "bm"); err != nil {
			return err
		}
	}
	registered, existing, err := app.ensureSSHKey(ctx, key)
	if err != nil {
		return err
	}
	if target != nil {
		access, err := app.sshAccess(ctx, team, target.Name, target.Kind == "bm")
		if err != nil {
			return fmt.Errorf("key registered, but could not refresh instance SSH details; run hotaisle ssh again: %w", err)
		}
		if access != nil {
			target.Connection, err = connectionFor(access, "hotaisle")
			if err != nil {
				return fmt.Errorf("key registered, but instance SSH details are invalid: %w", err)
			}
		}
	}
	return app.printResult(struct {
		Key                *client.SSHKey `json:"key"`
		AlreadyRegistered  bool           `json:"already_registered"`
		Scope              string         `json:"scope"`
		Propagation        string         `json:"propagation"`
		Target             *sshKeyTarget  `json:"target,omitzero"`
		GuestLoginVerified bool           `json:"guest_login_verified"`
	}{registered, existing, "account", sshKeyPropagation, target, false})
}

func canUseSSH(team client.UserTeam) bool {
	if len(team.EffectiveRoles) > 0 {
		return slices.Contains(team.EffectiveRoles, "user")
	}
	return slices.ContainsFunc(team.Roles, func(role string) bool { return role == "user" || role == "operator" || role == "owner" })
}
