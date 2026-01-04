package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v3"
)

// flagDef defines a command flag declaratively.
//
// Example:
//
//	{Name: "name", Usage: "User's full name", Required: true}
//	{Name: "user-role", Usage: "User role (owner or user)", Value: "user"}
type flagDef struct {
	Name     string
	Usage    string
	Required bool
	Value    string // Default value
}

// commandDef defines a command and its subcommands declaratively.
//
// Example:
//
//	var myCommands = commandDef{
//	    Name:  "resource",
//	    Usage: "Manage resources",
//	    Commands: []commandDef{
//	        {
//	            Name:  "list",
//	            Usage: "List all resources",
//	            Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
//	                resources, err := app.Client.Api.Resource().List(ctx)
//	                if err != nil {
//	                    return err
//	                }
//	                return printJSON(resources)
//	            },
//	        },
//	        {
//	            Name:  "create",
//	            Usage: "Create a new resource",
//	            Flags: []flagDef{
//	                {Name: "name", Usage: "Resource name", Required: true},
//	            },
//	            Action: func(app *App, ctx context.Context, cmd *cli.Command) error {
//	                resource, err := app.Client.Api.Resource().Create(ctx, client.ResourceRequest{
//	                    Name: cmd.String("name"),
//	                })
//	                if err != nil {
//	                    return err
//	                }
//	                return printJSON(resource)
//	            },
//	        },
//	    },
//	}
//
// Then build it with:
//
//	func newCommandResource(app *App) *cli.Command {
//	    return buildCommand(app, myCommands)
//	}
type commandDef struct {
	Name     string
	Usage    string
	Flags    []flagDef
	Action   func(*App, context.Context, *cli.Command) error
	Commands []commandDef
}

// findCommand looks up a command by path (e.g., "get", "ssh-keys.list", "api-keys.create")
func (def commandDef) findCommand(path string) *commandDef {
	parts := splitPath(path)
	if len(parts) == 0 {
		return nil
	}

	// Check if this command matches the first part
	if parts[0] == def.Name {
		// If this is the last part, return this command
		if len(parts) == 1 {
			return &def
		}
		// Otherwise search in subcommands
		for _, cmd := range def.Commands {
			if result := cmd.findCommand(joinPath(parts[1:])); result != nil {
				return result
			}
		}
	}

	// If name doesn't match, but we have subcommands, search them
	for _, cmd := range def.Commands {
		if result := cmd.findCommand(path); result != nil {
			return result
		}
	}

	return nil
}

func splitPath(path string) []string {
	result := []string{}
	current := ""
	for _, ch := range path {
		if ch == '.' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func joinPath(parts []string) string {
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += "."
		}
		result += part
	}
	return result
}

// printJSON marshals a value to pretty-printed JSON and prints it
func printJSON(v any) error {
	prettyJSON, _ := json.MarshalIndent(v, "", "  ")
	fmt.Print(string(prettyJSON))
	return nil
}

// buildCommand recursively builds a cli.Command from a commandDef
func buildCommand(app *App, def commandDef) *cli.Command {
	cmd := &cli.Command{
		Name:  def.Name,
		Usage: def.Usage,
	}

	if len(def.Flags) > 0 {
		cmd.Flags = make([]cli.Flag, len(def.Flags))
		for i, flag := range def.Flags {
			cmd.Flags[i] = &cli.StringFlag{
				Name:     flag.Name,
				Usage:    flag.Usage,
				Required: flag.Required,
				Value:    flag.Value,
			}
		}
	}

	if def.Action != nil {
		action := def.Action
		cmd.Action = func(ctx context.Context, command *cli.Command) error {
			return action(app, ctx, command)
		}
	}

	if len(def.Commands) > 0 {
		cmd.Commands = make([]*cli.Command, len(def.Commands))
		for i, subCmd := range def.Commands {
			cmd.Commands[i] = buildCommand(app, subCmd)
		}
	}

	return cmd
}
