package fmx

import (
	"os"

	"github.com/basemachina/lo"
	"github.com/bmatcuk/doublestar/v4"

	"github.com/sushichan044/fmx/internal/concurrent"
	"github.com/sushichan044/fmx/internal/gitignore"
	"github.com/sushichan044/fmx/internal/markdown"
)

type (
	MarkdownDocument[T any] struct {
		// File path of the markdown file.
		Path string `json:"path"`

		// Metadata extracted from the markdown file.
		FrontMatter T `json:"frontMatter"`

		// Raw markdown content except for the front matter.
		Body string `json:"-"`
	}
)

func GlobFrontMatter[T any](glob string) ([]concurrent.TaskResult[*MarkdownDocument[T]], error) {
	matched, err := runGlob(glob)
	if err != nil {
		return nil, err
	}

	results := concurrent.RunAll(
		lo.Map(matched, func(path string) func() (*MarkdownDocument[T], error) {
			return func() (*MarkdownDocument[T], error) {
				return processMarkdownFile[T](path)
			}
		})...,
	)

	return results, nil
}

func processMarkdownFile[T any](path string) (*MarkdownDocument[T], error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	md, err := markdown.ParseMarkdownWithMetadata[T](body)
	if err != nil {
		return nil, err
	}

	return &MarkdownDocument[T]{
		Path:        path,
		FrontMatter: md.FrontMatter,
		Body:        md.Content,
	}, nil
}

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
