package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	}

	jsonPayload struct {
		Body        string `json:"body"`
		Path        string `json:"path"`
		FrontMatter any    `json:"frontMatter"`
	}
)

func (cmd *CLI) Run() error {
	tasks, globErr := mdfm.Glob[map[string]any](cmd.Pattern)
	if globErr != nil {
		return fmt.Errorf("error during glob %s: %w", cmd.Pattern, globErr)
	}

	wtr := bufio.NewWriter(os.Stdout)
	printer := newPassthroughPrinter(wtr)

	defer func() {
		if err := wtr.Flush(); err != nil {
			if !errors.Is(err, syscall.EPIPE) {
				fmt.Fprintf(os.Stderr, "error flushing output on exit: %s", err)
			}
		}
	}()

	var hasErrors bool
	for _, task := range tasks {
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

// jsonPrinter writes a payload as JSON using a captured encoder.
type jsonPrinter func(payload jsonPayload) error

func newPassthroughPrinter(output io.Writer) jsonPrinter {
	enc := json.NewEncoder(output)
	enc.SetIndent("", "  ")

	return func(payload jsonPayload) error {
		return enc.Encode(payload)
	}
}

func main() {
	ctx := kong.Parse(&CLI{}, kong.Vars{
		"version": fmt.Sprintf("mdfm %s (rev: %s)", version.Version, revision),
	})
	ctx.FatalIfErrorf(ctx.Run())
}
