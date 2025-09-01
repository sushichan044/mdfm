# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

fmx is a Go CLI tool / Library that finds files matching glob patterns while respecting Git ignore rules.

## Development Commands

The project uses `mise` for task management and development workflow:

- `mise run dev` - Run the application in development mode
- `mise run test` - Run tests using gotestsum
- `mise run test-coverage` - Run tests with coverage reporting
- `mise run lint` - Run golangci-lint for code quality checks
- `mise run lint-fix` - Auto-fix linting issues
- `mise run fmt` - Format code
- `mise run build-snapshot` - Build cross-platform binaries with goreleaser
- `mise run clean` - Remove generated files

Standard Go commands also work:

- `go run . "**/*.md"` - Run with example glob pattern
- `go test ./...` - Run all tests
- `go mod tidy` - Clean up dependencies

## Architecture

### Core Structure

- `main.go` - CLI entry point using Kong for argument parsing
- `internal/gitignore/` - Git ignore handling with support for global/local ignore files
  - `matcher.go` - Main gitignore matching logic
  - `path.go` - Path resolution for various gitignore files
- `version/version.go` - Version constant (updated by goreleaser)

### Key Dependencies

- `github.com/alecthomas/kong` - CLI argument parsing
- `github.com/bmatcuk/doublestar/v4` - Glob pattern matching
- `github.com/sabhiram/go-gitignore` - Git ignore pattern matching
- `github.com/basemachina/lo` - Utility functions (filtering)
- `github.com/Songmu/gitconfig` - Git configuration access
