package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

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

	printer := passthroughPrinter()
	wtr := bufio.NewWriter(os.Stdout)
	defer wtr.Flush()

	for _, task := range tasks {
		if task.Result.Err != nil {
			fmt.Fprintf(os.Stderr, "error processing %s: %s", task.Metadata.Path, task.Result.Err)
			continue
		}

		payload := jsonPayload{
			Body:        task.Result.Value.BodyString(),
			Path:        task.Metadata.Path,
			FrontMatter: task.Result.Value.FrontMatter,
		}

		if fmtErr := printer(wtr, payload); fmtErr != nil {
			fmt.Fprintf(os.Stderr, "error formatting JSON for %s: %s", task.Metadata.Path, fmtErr)
			continue
		}

		if err := wtr.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "error flushing output for %s: %s", task.Metadata.Path, err)
		}
	}

	return nil
}

// jsonPrinter is a simple function type for formatting the payload to JSON.
type jsonPrinter func(output io.Writer, payload jsonPayload) error

func passthroughPrinter() jsonPrinter {
	return func(output io.Writer, payload jsonPayload) error {
		encoder := json.NewEncoder(output)
		encoder.SetIndent("", "  ")

		return encoder.Encode(payload)
	}
}

func main() {
	ctx := kong.Parse(&CLI{}, kong.Vars{
		"version": fmt.Sprintf("mdfm %s (rev: %s)", version.Version, revision),
	})
	ctx.FatalIfErrorf(ctx.Run())
}
