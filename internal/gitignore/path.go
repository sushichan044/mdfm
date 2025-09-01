package gitignore

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/Songmu/gitconfig"
)

func getGlobalGitIgnorePath() (string, error) {
	val, err := gitconfig.Get("core.excludesFile")
	if err != nil && !gitconfig.IsNotFound(err) {
		return "", err
	}

	if val != "" {
		return val, nil
	}

	return getDefaultExcludesFilePath()
}

func getDefaultExcludesFilePath() (string, error) {
	if xdgCfgHome := os.Getenv("XDG_CONFIG_HOME"); xdgCfgHome != "" {
		return filepath.Join(xdgCfgHome, "git", "ignore"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if home == "" {
		return "", errors.New("cannot determine user home directory to find default gitignore file")
	}

	return filepath.Join(home, ".config", "git", "ignore"), nil
}

func getLocalGitIgnorePath() (string, error) {
	gitDir := os.Getenv("GIT_DIR")
	if gitDir == "" {
		if wd, err := os.Getwd(); err == nil {
			gitDir = filepath.Join(wd, ".git")
		}
	}

	if gitDir == "" {
		return "", errors.New("cannot resolve git directory")
	}

	return filepath.Join(gitDir, "info", "exclude"), nil
}
