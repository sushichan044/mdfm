package main

import (
	"encoding/json"
	"fmt"
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
		Path        string `json:"path"`
		FrontMatter any    `json:"frontMatter"`
	}
)

func (cmd *CLI) Run() error {
	globResult, err := mdfm.GlobFrontMatter[map[string]any](cmd.Pattern)

	if err != nil {
		return err
	}

	for _, m := range globResult {
		if m.Result.Err != nil {
			fmt.Fprintf(os.Stderr, "error processing %s: %s", m.Metadata.Path, m.Result.Err)
			continue
		}

		payload := jsonPayload{
			Path:        m.Metadata.Path,
			FrontMatter: m.Result.Value.FrontMatter,
		}

		jsonData, marshalErr := json.MarshalIndent(payload, "", "  ")
		if marshalErr != nil {
			fmt.Fprintf(os.Stderr, "error marshaling JSON for %s: %s", m.Metadata.Path, marshalErr)
			continue
		}

		//nolint:forbidigo // This is fine
		fmt.Printf("%s\n", jsonData)
	}

	return nil
}

func main() {
	ctx := kong.Parse(&CLI{}, kong.Vars{
		"version": fmt.Sprintf("mdfm %s (rev: %s)", version.Version, revision),
	})
	ctx.FatalIfErrorf(ctx.Run())
}
