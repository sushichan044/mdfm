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
		// Metadata extracted from the markdown file.
		FrontMatter T

		// Raw markdown content except for the front matter.
		Body string
	}

	MarkdownDocumentMetadata struct {
		// The path to the markdown file.
		Path string
	}
)

func GlobFrontMatter[T any](
	glob string,
) ([]concurrent.TaskResult[*MarkdownDocument[T], MarkdownDocumentMetadata], error) {
	matched, err := runGlob(glob)
	if err != nil {
		return nil, err
	}

	tasks := lo.Map(matched, func(path string) concurrent.Task[*MarkdownDocument[T], MarkdownDocumentMetadata] {
		return concurrent.Task[*MarkdownDocument[T], MarkdownDocumentMetadata]{
			Metadata: MarkdownDocumentMetadata{Path: path},
			Run: func() (*MarkdownDocument[T], error) {
				return processMarkdownFile[T](path)
			},
		}
	})

	results := concurrent.RunAll(tasks...)
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
