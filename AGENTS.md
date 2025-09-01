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
- `mdfm.go` - Main library API with `GlobFrontMatter` function
- `internal/gitignore/` - Git ignore handling with support for global/local ignore files
  - `matcher.go` - Main gitignore matching logic
  - `path.go` - Path resolution for various gitignore files
- `internal/markdown/` - Markdown frontmatter parsing
  - `parse.go` - Frontmatter extraction and parsing logic
- `internal/concurrent/` - Concurrent processing utilities
  - `run_all.go` - Parallel task execution
- `version/version.go` - Version constant (updated by goreleaser)

### Key Dependencies

- `github.com/alecthomas/kong` - CLI argument parsing
- `github.com/bmatcuk/doublestar/v4` - Glob pattern matching
- `github.com/sabhiram/go-gitignore` - Git ignore pattern matching
- `github.com/basemachina/lo` - Utility functions (filtering)
- `github.com/Songmu/gitconfig` - Git configuration access
- `github.com/adrg/frontmatter` - YAML/TOML frontmatter parsing
- `github.com/yuin/goldmark` - Markdown parsing
