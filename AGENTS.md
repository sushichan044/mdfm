# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

mdfm is a Go CLI tool / Library that finds Markdown files using glob patterns and extracts their frontmatter metadata while respecting Git ignore rules.

## Development Commands

The project uses `mise` for task management and development workflow:

- `mise run dev` - Run the application in development mode
  - e.g. `mise run dev "**/*.md"`
- `mise run test` - Run tests using gotestsum
- `mise run test-coverage` - Run tests with coverage reporting
- `mise run lint` - Run golangci-lint for code quality checks
- `mise run lint-fix` - Auto-fix linting issues
- `mise run fmt` - Format code
- `mise run build-snapshot` - Build cross-platform binaries with goreleaser
- `mise run clean` - Remove generated files

Standard Go commands also work:

- `go run ./cmd/cli "**/*.md"` - Output results as JSON
- `go test ./...` - Run all tests
- `go mod tidy` - Clean up dependencies

## Architecture

### Core Structure

- `cmd/cli/main.go` - CLI entry point using Kong for argument parsing
- `mdfm.go` - Main library API with `Glob` and `GlobStream` functions
- `internal/gitignore/` - Git ignore handling with support for global/local ignore files
  - `matcher.go` - Main gitignore matching logic
  - `path.go` - Path resolution for various gitignore files
- `internal/markdown/` - Markdown frontmatter parsing
  - `parse.go` - Frontmatter extraction and parsing logic
- `internal/concurrent/` - Concurrent processing utilities
  - `run_all.go` - Parallel task execution with `RunAll` and `RunAllStream`
  - `options.go` - Concurrency control options and configuration
- `version/version.go` - Version constant (updated by goreleaser)

### API Design

The library provides two main processing modes:

#### Batch Processing (`Glob`)
- Processes all files and returns complete results
- Uses `RunAll` for concurrent processing with order preservation
- Suitable for smaller file sets where you need all results at once

#### Streaming Processing (`GlobStream`)  
- Streams results as they become available
- Uses `RunAllStream` for immediate result streaming
- Better performance for large file sets
- Results arrive in completion order, not input order

### Concurrency Model

- **Default Concurrency**: 10 concurrent file processors
- **Semaphore Control**: Uses `golang.org/x/sync/semaphore` for limiting concurrency
- **Panic Recovery**: All panics in file processing are caught and converted to errors
- **Error Isolation**: Individual file errors don't stop processing of other files

### Type System

The concurrent processing uses generic types:

```go
type Task[T, M any] struct {
    Metadata M
    Run      func() (T, error)
}

type TaskExecution[T, M any] struct {
    Metadata M
    Result   taskResult[T]
}
```

Where:
- `T`: The result type (`*MarkdownDocument[T]`)
- `M`: The metadata type (`MarkdownDocumentMetadata`)

### Key Dependencies

- `github.com/alecthomas/kong` - CLI argument parsing
- `github.com/bmatcuk/doublestar/v4` - Glob pattern matching
- `github.com/sabhiram/go-gitignore` - Git ignore pattern matching
- `github.com/basemachina/lo` - Utility functions (filtering)
- `github.com/Songmu/gitconfig` - Git configuration access
- `github.com/adrg/frontmatter` - YAML/TOML frontmatter parsing
- `golang.org/x/sync/semaphore` - Semaphore-based concurrency control
