package cli

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/hotaisle/hotaisle-cli/client"
	"github.com/hotaisle/hotaisle-cli/internal/config"
	"github.com/urfave/cli/v3"
	"golang.org/x/term"
)

func newCommandSetup(app *App) *cli.Command {
	return &cli.Command{
		Name: "setup", Aliases: []string{"login"}, Usage: "Validate credentials, select a team, and configure SSH access.",
		Description:  "Prompts only in a terminal. Agents can use HOTAISLE_API_TOKEN with --non-interactive --team HANDLE.\nObtain your initial API key through ssh admin.hotaisle.app (personal settings > API keys).",
		OnUsageError: usageError,
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "non-interactive", Usage: "Never prompt; report missing information instead"},
			&cli.BoolFlag{Name: "token-stdin", Usage: "Read the API token from stdin instead of environment or saved credentials"},
			&cli.StringFlag{Name: "team", Usage: "Default team handle (automatically selected for a single team)"},
			&cli.StringFlag{Name: "ssh-key-file", Usage: "Register this SSH public key file if it is not already registered"},
		},
		Action: app.setup,
	}
}

func interactiveInput(cmd *cli.Command) (*os.File, bool) {
	file, ok := cmd.Root().Reader.(*os.File)
	return file, ok && term.IsTerminal(int(file.Fd())) && !cmd.Bool("non-interactive") && !structuredOutput(cmd)
}

func prompt(reader *bufio.Reader, writer io.Writer, message string) (string, error) {
	if _, err := fmt.Fprint(writer, message); err != nil {
		return "", err
	}
	line, err := reader.ReadString('\n')
	if err != nil && (!errors.Is(err, io.EOF) || len(line) == 0) {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func (app *App) setup(ctx context.Context, cmd *cli.Command) error {
	file, interactive := interactiveInput(cmd)
	reader := bufio.NewReader(cmd.Root().Reader)
	token := app.token()
	if cmd.Bool("token-stdin") {
		if file != nil && term.IsTerminal(int(file.Fd())) {
			return invalid("--token-stdin requires piped input; run setup without it for a hidden token prompt")
		}
		data, err := io.ReadAll(io.LimitReader(reader, 65537))
		if err != nil {
			return err
		}
		if len(data) > 65536 {
			return invalid("API token input exceeds 64 KiB")
		}
		token = strings.TrimSpace(string(data))
	} else if token == "" && interactive {
		_, _ = fmt.Fprint(cmd.ErrWriter, "Create an API key in personal settings via ssh admin.hotaisle.app.\nAPI token (hidden): ")
		data, err := term.ReadPassword(int(file.Fd()))
		_, _ = fmt.Fprintln(cmd.ErrWriter)
		if err != nil {
			return err
		}
		token = strings.TrimSpace(string(data))
	}
	if token == "" {
		return missingToken()
	}
	app.Client.Api.SetToken(token)
	user, err := app.Client.Api.User().Get(ctx)
	if err != nil {
		return err
	}
	teams := slices.DeleteFunc(slices.Clone(user.Teams), func(team client.UserTeam) bool { return team.Invitation })
	team := strings.TrimSpace(cmd.String("team"))
	if team == "" {
		team = app.defaultTeam()
	}
	if team == "" && len(teams) == 1 {
		team = teams[0].Handle
	}
	if team == "" && interactive && len(teams) > 0 {
		for _, t := range teams {
			_, _ = fmt.Fprintf(cmd.ErrWriter, "  %s (%s)\n", t.Handle, t.Name)
		}
		team, err = prompt(reader, cmd.ErrWriter, "Default team handle: ")
		if err != nil {
			return err
		}
	}
	if err := validateTeam(teams, team); err != nil {
		return err
	}
	keys, err := app.Client.Api.User().GetSSHKeys(ctx)
	if err != nil {
		return err
	}
	keyFile := cmd.String("ssh-key-file")
	if keyFile == "" && len(keys) == 0 && interactive {
		keyFile, err = prompt(reader, cmd.ErrWriter, "SSH public key file (Enter to skip): ")
		if err != nil {
			return err
		}
	}
	if keyFile != "" {
		key, err := readPublicKey(keyFile)
		if err != nil {
			return err
		}
		added, existing, err := app.registerSSHKey(ctx, key, keys)
		if err != nil {
			return err
		}
		if !existing {
			keys = append(keys, *added)
		}
	}
	app.Config.ApiToken, app.Config.DefaultTeam = token, team
	if err := config.Save(app.Config); err != nil {
		return err
	}
	next := []string{"hotaisle doctor", "hotaisle vm available"}
	if len(keys) == 0 {
		next = append([]string{"Register an SSH public key before provisioning: hotaisle setup --ssh-key-file ~/.ssh/id_ed25519.pub"}, next...)
	}
	if structuredOutput(cmd) {
		return app.printResult(struct {
			Success  bool     `json:"success"`
			Team     string   `json:"team"`
			SSHReady bool     `json:"ssh_ready"`
			Next     []string `json:"next_steps"`
		}{true, team, len(keys) > 0, next})
	}
	_, err = fmt.Fprintf(app.writer(), "Setup complete for team %s.\nNext steps:\n  %s\n", team, strings.Join(next, "\n  "))
	return err
}

func missingToken() error {
	return problem("authentication", "Run hotaisle setup in a terminal, set HOTAISLE_API_TOKEN, or pipe a token to hotaisle setup --token-stdin --non-interactive.", errors.New("no API token configured"))
}

func validateTeam(teams []client.UserTeam, team string) error {
	if len(teams) == 0 {
		return problem("configuration", "Accept a team invitation or check the API key's team permissions.", errors.New("no accessible teams found"))
	}
	if !slices.ContainsFunc(teams, func(t client.UserTeam) bool { return t.Handle == team && !t.Invitation }) {
		handles := make([]string, 0, len(teams))
		for _, t := range teams {
			if !t.Invitation {
				handles = append(handles, t.Handle)
			}
		}
		return problem("configuration", "Pass --team with one of: "+strings.Join(handles, ", "), fmt.Errorf("select an accessible team (requested %q)", team))
	}
	return nil
}

func readPublicKey(path string) (string, error) {
	path, err := config.ResolvePath(path)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read SSH public key: %w", err)
	}
	return validatePublicKey(string(data))
}

func validatePublicKey(value string) (string, error) {
	key := strings.TrimSpace(value)
	parts := strings.Fields(key)
	if strings.Contains(key, "PRIVATE KEY") || strings.ContainsAny(key, "\r\n") || len(parts) < 2 {
		return "", invalid("expected one SSH public key in authorized_keys format; choose the .pub file")
	}
	if _, err := base64.StdEncoding.DecodeString(parts[1]); err != nil {
		return "", invalid("SSH public key contains invalid base64")
	}
	return key, nil
}
