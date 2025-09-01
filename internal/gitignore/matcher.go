package gitignore

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	ignore "github.com/sabhiram/go-gitignore"
)

// Matcher wraps a compiled .gitignore and its root directory.
// Thread-safe after construction.
type Matcher struct {
	root    string
	matcher *ignore.GitIgnore
}

// New creates a Matcher for the given root directory by reading `.gitignore` there.
// If no .gitignore exists, it returns an empty matcher that never matches.
func New(root string) (*Matcher, error) {
	//nolint:mnd // local,infoExclude,global
	giPaths := make([]string, 0, 3)
	giPaths = append(giPaths, filepath.Join(root, ".gitignore"))

	if globalGi, err := getGlobalGitIgnorePath(); err != nil {
		return nil, fmt.Errorf("failed to get global gitignore path: %w", err)
	} else if globalGi != "" {
		giPaths = append(giPaths, globalGi)
	}

	if localGi, err := getLocalGitIgnorePath(); err != nil {
		return nil, fmt.Errorf("failed to get local gitignore path: %w", err)
	} else if localGi != "" {
		giPaths = append(giPaths, localGi)
	}

	var lines []string
	for _, giPath := range giPaths {
		if file, fErr := os.Open(giPath); fErr == nil {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
		}
	}

	return &Matcher{root: root, matcher: ignore.CompileIgnoreLines(lines...)}, nil
}

// NewFromCWD builds a Matcher using the current working directory as root.
func NewFromCWD() (*Matcher, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}
	return New(cwd)
}

// IsIgnored reports whether path is ignored by this matcher.
// The path can be absolute or relative; it will be normalized relative to root.
func (m *Matcher) IsIgnored(path string) bool {
	if m == nil || m.matcher == nil {
		return false
	}
	rel := path
	if m.root != "" {
		if r, err := filepath.Rel(m.root, path); err == nil {
			rel = r
		}
	}
	rel = filepath.ToSlash(rel)
	return m.matcher.MatchesPath(rel)
}
