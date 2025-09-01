package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/basemachina/lo"
	"github.com/bmatcuk/doublestar/v4"

	"github.com/sushichan044/glob-frontmatter/internal/gitignore"
	"github.com/sushichan044/glob-frontmatter/version"
)

var (
	//nolint:gochecknoglobals // This value is overridden by goreleaser.
	revision = "dev"
)

type CLI struct {
	Pattern string `arg:"" name:"pattern" help:"Glob pattern to match (eg. '**/*.md')"`

	Version kong.VersionFlag `short:"v"`
}

// runGlob searches for files matching the given glob pattern.
// It ignores files that are excluded by Git.
func runGlob(pattern string) ([]string, error) {
	matched, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return nil, err
	}

	gi, err := gitignore.NewFromCWD()
	if err != nil {
		return nil, err
	}
	if gi == nil {
		return matched, nil
	}

	nonIgnoredFiles := lo.Filter(matched, func(p string) bool {
		// gi is non-nil
		return !gi.IsIgnored(p)
	})
	return nonIgnoredFiles, nil
}

func (cmd *CLI) Run() error {
	matched, err := runGlob(cmd.Pattern)
	if err != nil {
		return err
	}
	// TODO: read markdown frontmatter
	for _, p := range matched {
		fmt.Fprintln(os.Stdout, p)
	}
	return nil
}

func main() {
	ctx := kong.Parse(&CLI{}, kong.Vars{
		"version": fmt.Sprintf("glob-frontmatter %s (rev: %s)", version.Version, revision),
	})
	ctx.FatalIfErrorf(ctx.Run())
}
