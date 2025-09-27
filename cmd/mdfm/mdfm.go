package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/alecthomas/kong"

	"github.com/sushichan044/mdfm"
	"github.com/sushichan044/mdfm/version"
)

var (
	//nolint:gochecknoglobals // This value is overridden by goreleaser.
	revision = "dev"
)

type (
	CLI struct {
		Pattern string `arg:"" name:"pattern" help:"Glob pattern to match (eg. '**/*.md')"`

		Version kong.VersionFlag `short:"v"`

		JQ string `name:"jq" help:"jq filter to apply to each JSON object."`
	}
)

func (cmd *CLI) Run() error {
	resultChan, globErr := mdfm.GlobStream[map[string]any](cmd.Pattern)
	if globErr != nil {
		return fmt.Errorf("error during glob %s: %w", cmd.Pattern, globErr)
	}

	wtr := bufio.NewWriter(os.Stdout)
	defer func() {
		if err := wtr.Flush(); err != nil {
			if !errors.Is(err, syscall.EPIPE) {
				fmt.Fprintf(os.Stderr, "error flushing output on exit: %s", err)
			}
		}
	}()

	printer, printerErr := NewAppropriatePrinter(wtr, cmd.JQ)
	if printerErr != nil {
		return printerErr
	}

	var hasErrors bool
	for task := range resultChan {
		if task.Result.Err != nil {
			hasErrors = true
			fmt.Fprintf(os.Stderr, "error processing %s: %v\n", task.Metadata.Path, task.Result.Err)
			continue
		}

		payload := jsonPayload{
			Body:        task.Result.Value.BodyString(),
			Path:        task.Metadata.Path,
			FrontMatter: task.Result.Value.FrontMatter,
		}

		if fmtErr := printer(payload); fmtErr != nil {
			hasErrors = true
			fmt.Fprintf(os.Stderr, "error formatting JSON for %s: %v\n", task.Metadata.Path, fmtErr)
			continue
		}

		if err := wtr.Flush(); err != nil {
			if errors.Is(err, syscall.EPIPE) {
				return nil
			}
			hasErrors = true
			fmt.Fprintf(os.Stderr, "error flushing output for %s: %v\n", task.Metadata.Path, err)
		}
	}

	if hasErrors {
		return errors.New("errors occurred during processing markdown files")
	}

	return nil
}

func main() {
	ctx := kong.Parse(&CLI{}, kong.Vars{
		"version": fmt.Sprintf("mdfm %s (rev: %s)", version.Version, revision),
	})
	ctx.FatalIfErrorf(ctx.Run())
}
