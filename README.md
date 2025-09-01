# fmx

[![CI](https://github.com/sushichan044/fmx/actions/workflows/ci.yml/badge.svg)](https://github.com/sushichan044/fmx/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/sushichan044/fmx.svg)](https://pkg.go.dev/github.com/sushichan044/fmx)
[![Go Report Card](https://goreportcard.com/badge/github.com/sushichan044/fmx)](https://goreportcard.com/report/github.com/sushichan044/fmx)

**fmx** is a Go library and CLI tool that finds Markdown files using glob patterns and extracts their frontmatter metadata as JSON.

## Features

- 🔍 **Glob Pattern Matching**: Find Markdown files using powerful glob patterns like `**/*.md`
- 📄 **Frontmatter Extraction**: Parse YAML, TOML, JSON frontmatter from Markdown files
- 🚫 **Git Integration**: Automatically respects `.gitignore`, global Git excludes, and local Git excludes
- 🛡️ **Type Safety**: Generic type support for strongly-typed frontmatter structures
- 📦 **Both Library & CLI**: Use as a Go library or standalone command-line tool

## Installation

### CLI Installation

Install the CLI tool using Go:

```bash
go install github.com/sushichan044/fmx/cmd/cli@latest
```

Or download pre-built binaries from the [releases page](https://github.com/sushichan044/fmx/releases).

### Library Installation

Add fmx to your Go project:

```bash
go get github.com/sushichan044/fmx
```

## CLI Usage

### Basic Usage

Find all Markdown files and extract their frontmatter:

```bash
fmx "**/*.md"
```

### Examples

```bash
# Find all Markdown files in the docs directory
fmx "docs/**/*.md"

# Find all blog posts
fmx "content/posts/*.md"

# Find specific file
fmx "README.md"

# Find files with specific pattern
fmx "content/**/{blog,docs}/*.md"
```

### Output Format

The CLI outputs JSON lines, with each line containing:

```json
{
  "path": "content/posts/my-post.md",
  "frontMatter": {
    "title": "My Blog Post",
    "date": "2023-12-01",
    "tags": ["golang", "markdown"]
  }
}
```

### Options

```bash
# Show version information
fmx --version

# Show help
fmx --help
```

## Library Usage

### Basic Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/sushichan044/fmx"
)

type BlogPost struct {
    Title       string   `yaml:"title"`
    Date        string   `yaml:"date"`
    Tags        []string `yaml:"tags"`
    Published   bool     `yaml:"published"`
}

func main() {
    results, err := fmx.GlobFrontMatter[BlogPost]("content/**/*.md")
    if err != nil {
        log.Fatal(err)
    }

    for _, result := range results {
        if result.Result.Err != nil {
            fmt.Printf("Error processing %s: %v\n",
                result.Metadata.Path, result.Result.Err)
            continue
        }

        post := result.Result.Value
        fmt.Printf("Title: %s\n", post.FrontMatter.Title)
        fmt.Printf("Path: %s\n", result.Metadata.Path)
        fmt.Printf("Published: %t\n", post.FrontMatter.Published)
        fmt.Printf("Content preview: %.100s...\n", post.Body)
        fmt.Println("---")
    }
}
```

### Using Generic Types

You can use any type for frontmatter extraction:

```go
// Use map for dynamic frontmatter
results, err := fmx.GlobFrontMatter[map[string]any]("**/*.md")

// Use a custom struct for type safety
type Metadata struct {
    Title       string    `yaml:"title"`
    Author      string    `yaml:"author"`
    CreatedAt   time.Time `yaml:"created_at"`
}

results, err := fmx.GlobFrontMatter[Metadata]("**/*.md")
```

### Error Handling

The library uses a concurrent processing model where individual file processing errors don't stop the entire operation:

```go
results, err := fmx.GlobFrontMatter[MyType]("**/*.md")
if err != nil {
    // This is a fatal error (e.g., invalid glob pattern)
    log.Fatal(err)
}

for _, result := range results {
    if result.Result.Err != nil {
        // This is a per-file error (e.g., invalid frontmatter)
        fmt.Printf("Error processing %s: %v\n",
            result.Metadata.Path, result.Result.Err)
        continue
    }

    // Process successful result
    doc := result.Result.Value
    // ... use doc.FrontMatter and doc.Body
}
```

## Supported Frontmatter Formats

See here for details:

<https://github.com/adrg/frontmatter?tab=readme-ov-file#supported-formats>

## Git Integration

fmx automatically respects Git ignore rules from:

- **Local `.gitignore`**: Project-specific ignore patterns
- **Global Git excludes**: User's global `~/.config/git/ignore` (or `$XDG_CONFIG_HOME/git/ignore`)
- **Repository excludes**: Local `.git/info/exclude` file

This means fmx will automatically skip files that Git would ignore, making it perfect for processing only the files that are part of your project.

## Development

### Prerequisites

- Go 1.24 or later
- [mise](https://mise.jdx.dev/) (optional, for development tasks)

### Building

```bash
# Build the CLI
go build ./cmd/cli

# Build with mise
mise run build-snapshot
```

### Testing

```bash
# Run tests
go test ./...
# or you can use `mise run test`

# Run tests with coverage
mise run test-coverage

# Run linting
mise run lint
```

### Available Development Commands

The project uses `mise` for task management:

```bash
mise run dev "**/*.md" --json  # Run CLI locally
mise run test                  # Run tests
mise run test-coverage         # Run tests with coverage
mise run lint                  # Run linter
mise run lint-fix             # Auto-fix linting issues
mise run fmt                   # Format code
mise run build-snapshot       # Build cross-platform binaries
mise run clean                # Clean generated files
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

For major changes, please open an issue first to discuss what you would like to change.

## License

This project is licensed under the MIT License - see the [LICENSE](/LICENSE) file for details.

---

Made with ❤️ by [sushichan044](https://github.com/sushichan044)
