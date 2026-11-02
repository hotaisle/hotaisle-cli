// release-plan previews release metadata. Publishing belongs to GitHub Actions.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/hotaisle/hotaisle-cli/internal/release"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	repository := flag.String("repository", "hotaisle/hotaisle-cli", "GitHub owner/repository used in release links")
	notesFile := flag.String("notes-file", "", "write Markdown release notes to this file")
	outputFile := flag.String("github-output", "", "append metadata to this GitHub Actions output file")
	flag.Parse()
	if flag.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	plan, err := release.PlanRepository(ctx, ".", *repository)
	if err != nil {
		return err
	}
	if *notesFile != "" {
		if err := os.WriteFile(*notesFile, []byte(plan.Notes), 0o600); err != nil {
			return err
		}
	}
	if *outputFile == "" {
		return plan.WriteOutputs(os.Stdout)
	}
	file, err := os.OpenFile(*outputFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	return errors.Join(plan.WriteOutputs(file), file.Close())
}
