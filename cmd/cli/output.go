package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/hotaisle/hotaisle-cli/client"
	"github.com/urfave/cli/v3"
)

type commandError struct {
	kind string
	hint string
	err  error
}

func (e *commandError) Error() string { return e.err.Error() }
func (e *commandError) Unwrap() error { return e.err }

func problem(kind, hint string, err error) error {
	return &commandError{kind: kind, hint: hint, err: err}
}

func invalid(message string) error {
	return problem("invalid_arguments", "Run the command with --help for usage.", errors.New(message))
}

func usageError(ctx context.Context, cmd *cli.Command, err error, _ bool) error {
	if !structuredOutput(cmd) {
		_, _ = fmt.Fprintf(cmd.Root().ErrWriter, "Incorrect Usage: %s\n\n", err)
		_ = cli.ShowSubcommandHelp(cmd)
	}
	return problem("invalid_arguments", "Run the command with --help for usage.", err)
}

func structuredOutput(cmd *cli.Command) bool {
	return cmd.Bool("json") || cmd.Bool("agent")
}

type errorDetails struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Hint       string `json:"hint,omitempty"`
	HTTPStatus int    `json:"http_status,omitzero"`
	Retryable  bool   `json:"retryable"`
}

func describeError(err error) (errorDetails, int) {
	detail := errorDetails{Code: "error", Message: err.Error()}
	exit := 1
	if _, ok := errors.AsType[*strconv.NumError](err); ok {
		detail.Code, exit = "invalid_arguments", 2
		detail.Hint = "Run the command with --help for valid argument values."
	}
	if e, ok := errors.AsType[*commandError](err); ok {
		detail.Code, detail.Hint = e.kind, e.hint
		if e.kind == "invalid_arguments" || e.kind == "configuration" {
			exit = 2
		}
		if e.kind == "authentication" {
			exit = 3
		}
	}
	if e, ok := errors.AsType[*client.APIError](err); ok {
		detail.HTTPStatus = e.StatusCode
		switch {
		case e.StatusCode == 401 || e.StatusCode == 403:
			detail.Code, exit = "authentication", 3
			detail.Hint = "Run hotaisle setup; check API token and team permissions."
		case e.StatusCode == 404:
			detail.Code, exit = "not_found", 4
		case e.StatusCode == 409:
			detail.Code, exit = "conflict", 5
			detail.Hint = "Check resource state and available capacity before trying again."
		case e.StatusCode == 429 || e.StatusCode >= 500:
			detail.Code, exit, detail.Retryable = "temporary_failure", 6, true
			detail.Hint = "Retry read-only requests with backoff. Check whether a mutation succeeded before retrying it."
		case e.StatusCode == 400 || e.StatusCode == 422:
			detail.Code, exit = "invalid_arguments", 2
		}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		detail.Code, exit = "timeout", 7
		detail.Hint = "Inspect resource state before retrying; a timeout does not cancel provisioning or billing."
	} else if errors.Is(err, context.Canceled) {
		detail.Code, exit = "canceled", 130
	} else if _, ok := errors.AsType[net.Error](err); ok {
		detail.Code, exit, detail.Retryable = "temporary_failure", 6, true
	}
	return detail, exit
}

func reportError(cmd *cli.Command, err error) int {
	detail, code := describeError(err)
	if cmd.Bool("json") {
		_ = writeJSON(cmd.ErrWriter, struct {
			Error errorDetails `json:"error"`
		}{detail}, true)
		return code
	}
	if cmd.Bool("agent") {
		_ = writeMarkdown(cmd.ErrWriter, "Error", detail)
		return code
	}
	_, _ = fmt.Fprintf(cmd.ErrWriter, "Error: %s\n", detail.Message)
	if detail.Hint != "" {
		_, _ = fmt.Fprintln(cmd.ErrWriter, detail.Hint)
	}
	return 1
}

func (app *App) writer() io.Writer {
	if app.AppCli != nil && app.AppCli.Writer != nil {
		return app.AppCli.Writer
	}
	return os.Stdout
}

func (app *App) printJSON(value any) error {
	if app.AppCli != nil && app.AppCli.Bool("agent") {
		return writeMarkdown(app.writer(), "Result", value)
	}
	return writeJSON(app.writer(), value, app.AppCli != nil && app.AppCli.Bool("json"))
}

func (app *App) success(message string) error {
	if app.AppCli != nil && structuredOutput(app.AppCli) {
		return app.printJSON(struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}{true, message})
	}
	_, err := fmt.Fprintln(app.writer(), message)
	return err
}
