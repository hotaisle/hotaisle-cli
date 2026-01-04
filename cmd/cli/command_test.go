package cli

import "github.com/urfave/cli/v3"

// getCommand looks up a command by path and sets flags for testing
func getCommand(app *App, def commandDef, path string, flags map[string]string) (*cli.Command, error) {
	cmdDef := def.findCommand(path)
	if cmdDef == nil {
		return nil, nil
	}

	cmd := buildCommand(app, *cmdDef)

	// Set flags if provided
	for name, value := range flags {
		if err := cmd.Set(name, value); err != nil {
			return nil, err
		}
	}

	return cmd, nil
}
