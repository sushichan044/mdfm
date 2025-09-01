// Package mdfm provides functionality for finding Markdown files using glob patterns
// and extracting their frontmatter metadata while respecting Git ignore rules.
//
// The main function GlobFrontMatter allows you to search for Markdown files and
// parse their YAML/TOML frontmatter in a concurrent, type-safe manner.
//
// Example usage:
//
//	type BlogPost struct {
//		Title     string   `yaml:"title"`
//		Tags      []string `yaml:"tags"`
//		Published bool     `yaml:"published"`
//	}
//
//	results, err := mdfm.GlobFrontMatter[BlogPost]("content/**/*.md")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	for _, result := range results {
//		if result.Result.Err != nil {
//			fmt.Printf("Error: %v\n", result.Result.Err)
//			continue
//		}
//
//		post := result.Result.Value
//		fmt.Printf("Title: %s\n", post.FrontMatter.Title)
//		fmt.Printf("Content: %s\n", post.Body)
//	}
package mdfm

import (
	"os"

	"github.com/basemachina/lo"
	"github.com/bmatcuk/doublestar/v4"

	"github.com/sushichan044/mdfm/internal/concurrent"
	"github.com/sushichan044/mdfm/internal/gitignore"
	"github.com/sushichan044/mdfm/internal/markdown"
)

// MarkdownDocument represents a Markdown file with its parsed frontmatter and content.
// The type parameter T specifies the structure of the frontmatter metadata.
//
// Example:
//
//	type Metadata struct {
//		Title  string `yaml:"title"`
//		Author string `yaml:"author"`
//	}
//
//	var doc MarkdownDocument[Metadata]
//	fmt.Println(doc.FrontMatter.Title) // Access frontmatter
//	fmt.Println(doc.Body)              // Access markdown content
type MarkdownDocument[T any] struct {
	// FrontMatter contains the parsed metadata from the document's frontmatter.
	// The structure depends on the type parameter T provided to GlobFrontMatter.
	FrontMatter T

	// Body contains the raw markdown content without the frontmatter.
	// This includes all content after the frontmatter delimiter (--- or +++).
	Body string
}

// MarkdownDocumentMetadata contains metadata about the processing of a Markdown file.
// This is separate from the frontmatter content and provides information about
// the file itself during processing.
type MarkdownDocumentMetadata struct {
	// Path is the file system path to the markdown file, relative to the
	// current working directory when GlobFrontMatter was called.
	Path string
}

// GlobFrontMatter finds Markdown files matching the given glob pattern and
// extracts their frontmatter metadata concurrently. It respects Git ignore rules
// and returns results for successful and failed file processing.
//
// The function uses a type parameter T to specify the expected structure of the
// frontmatter metadata. Use map[string]any for dynamic frontmatter or define
// a custom struct with YAML/TOML tags for type safety.
//
// Supported glob patterns:
//   - "*" matches any file in the current directory
//   - "**/*.md" matches all .md files recursively
//   - "content/{blog,docs}/*.md" matches .md files in blog or docs subdirectories
//
// Git integration:
// Files matching patterns in .gitignore, global Git excludes, or local Git excludes
// are automatically filtered out from the results.
//
// Error handling:
// The function returns an error only for fatal conditions (e.g., invalid glob pattern).
// Per-file errors (e.g., invalid frontmatter) are included in individual TaskResult.Err
// fields, allowing you to handle them on a case-by-case basis.
//
// Example usage:
//
//	type Article struct {
//		Title     string    `yaml:"title"`
//		Author    string    `yaml:"author"`
//		Date      time.Time `yaml:"date"`
//		Tags      []string  `yaml:"tags"`
//		Published bool      `yaml:"published"`
//	}
//
//	results, err := GlobFrontMatter[Article]("content/**/*.md")
//	if err != nil {
//		log.Fatalf("Failed to glob files: %v", err)
//	}
//
//	for _, result := range results {
//		if result.Result.Err != nil {
//			fmt.Printf("Error processing %s: %v\n",
//				result.Metadata.Path, result.Result.Err)
//			continue
//		}
//
//		article := result.Result.Value
//		if article.FrontMatter.Published {
//			fmt.Printf("Published: %s by %s\n",
//				article.FrontMatter.Title, article.FrontMatter.Author)
//		}
//	}
//
// For dynamic frontmatter (when structure is unknown):
//
//	results, err := GlobFrontMatter[map[string]any]("**/*.md")
//	// ... handle results with type assertions
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

// processMarkdownFile reads and parses a single Markdown file.
// It extracts frontmatter metadata and returns the processed document.
// This function is used internally by GlobFrontMatter for concurrent processing.
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

// runGlob executes glob pattern matching while respecting Git ignore rules.
// It filters out files that are excluded by .gitignore, global Git excludes,
// or local Git excludes, ensuring only relevant files are processed.
//
// The function uses doublestar for advanced glob pattern support and
// integrates with Git ignore functionality for seamless filtering.
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
