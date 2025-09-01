package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	"github.com/sushichan044/fmx"
	"github.com/sushichan044/fmx/version"
)

var (
	//nolint:gochecknoglobals // This value is overridden by goreleaser.
	revision = "dev"
)

type CLI struct {
	Pattern string `arg:"" name:"pattern" help:"Glob pattern to match (eg. '**/*.md')"`

	JSON bool `help:"Output as JSON"`

	Version kong.VersionFlag `short:"v"`
}

func (cmd *CLI) Run() error {
	globResult, err := fmx.GlobFrontMatter[map[string]any](cmd.Pattern)

	if err != nil {
		return err
	}

	for _, m := range globResult {
		if m.Result.Err != nil {
			return fmt.Errorf("error processing %s: %w", m.Metadata.Path, m.Result.Err)
		}

		if cmd.JSON {
			// print as json (one line per file)
			jsonData, marshalErr := json.Marshal(m.Result.Value)
			if marshalErr != nil {
				return fmt.Errorf("error marshaling JSON for %s: %w", m.Metadata.Path, marshalErr)
			}
			fmt.Fprintln(os.Stdout, string(jsonData))
			continue
		}

		// default: just print the path
		fmt.Fprintln(os.Stdout, m.Metadata.Path)
	}

	return nil
}

func main() {
	ctx := kong.Parse(&CLI{}, kong.Vars{
		"version": fmt.Sprintf("fmx %s (rev: %s)", version.Version, revision),
	})
	ctx.FatalIfErrorf(ctx.Run())
}
