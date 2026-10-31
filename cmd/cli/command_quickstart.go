package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/hotaisle/hotaisle-cli/skills"
	"github.com/urfave/cli/v3"
)

func newCommandQuickstart(app *App) *cli.Command {
	return &cli.Command{
		Name: "quickstart", Aliases: []string{"skill"},
		Usage:        "Print the Markdown agent skill shipped with this installed CLI version; works offline.",
		OnUsageError: usageError,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Present() {
				return invalid("quickstart does not accept positional arguments")
			}
			// Keep YAML frontmatter intact, allowing the output to be saved as an
			// installable skill while adding the installed binary's build identity.
			frontmatter, body, _ := strings.Cut(strings.TrimPrefix(skills.HotAisle, "---\n"), "\n---\n")
			markdown := fmt.Sprintf("---\n%s\n---\n\nInstalled CLI version: %s. Commit: %s.\n%s", frontmatter, markdownCode(Version), markdownCode(Commit), body)
			if cmd.Bool("json") {
				return app.printResult(struct {
					Version  string `json:"version"`
					Commit   string `json:"commit"`
					Markdown string `json:"markdown"`
				}{Version, Commit, markdown})
			}
			_, err := fmt.Fprint(app.writer(), markdown)
			return err
		},
	}
}
