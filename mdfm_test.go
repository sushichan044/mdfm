package mdfm_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/mdfm"
)

type testMetadata struct {
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
	Published   bool     `yaml:"published"`
}

func setupTestFiles(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	testFiles := map[string]string{
		"blog/post1.md": `---
title: First Post
description: This is the first post
tags: [golang, testing]
published: true
---
# First Post

This is the content of the first post.`,

		"blog/post2.md": `---
title: Second Post
description: Another post
tags: [golang, api]
published: false
---
# Second Post

Content here.`,

		"docs/readme.md": `---
title: README
description: Documentation
tags: [docs]
published: true
---
# Documentation

Some documentation content.`,

		"blog/draft.md": `---
title: Draft Post
published: false
---
# Draft

Work in progress.`,

		"no-frontmatter.md": `# No Frontmatter

Just content without frontmatter.`,

		"empty.md": ``,

		"invalid-frontmatter.md": `---
title: "Unclosed quote
---
Content`,
	}

	for relPath, content := range testFiles {
		fullPath := filepath.Join(tmpDir, relPath)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
	}

	t.Chdir(tmpDir)

	return tmpDir
}

func TestGlobFrontMatter_BasicFunctionality(t *testing.T) {
	setupTestFiles(t)

	tests := []struct {
		name           string
		pattern        string
		expectedCount  int
		expectedTitles []string
	}{
		{
			name:           "all markdown files",
			pattern:        "**/*.md",
			expectedCount:  7,
			expectedTitles: []string{"First Post", "Second Post", "README", "Draft Post"},
		},
		{
			name:           "blog posts only",
			pattern:        "blog/*.md",
			expectedCount:  3,
			expectedTitles: []string{"First Post", "Second Post", "Draft Post"},
		},
		{
			name:           "specific file",
			pattern:        "docs/readme.md",
			expectedCount:  1,
			expectedTitles: []string{"README"},
		},
		{
			name:          "no matches",
			pattern:       "nonexistent/*.md",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks, err := mdfm.Glob[testMetadata](tt.pattern)
			require.NoError(t, err)
			assert.Len(t, tasks, tt.expectedCount)

			var actualTitles []string
			for _, task := range tasks {
				if task.Result.Err == nil && task.Result.Value.FrontMatter.Title != "" {
					actualTitles = append(actualTitles, task.Result.Value.FrontMatter.Title)
				}
			}

			if len(tt.expectedTitles) > 0 {
				assert.ElementsMatch(t, tt.expectedTitles, actualTitles)
			}
		})
	}
}

func TestGlobFrontMatter_ErrorHandling(t *testing.T) {
	setupTestFiles(t)

	tasks, err := mdfm.Glob[testMetadata]("**/*.md")
	require.NoError(t, err)

	var errorCount int
	var successCount int

	for _, task := range tasks {
		if task.Result.Err != nil {
			errorCount++
			t.Logf("Error processing %s: %v", task.Metadata.Path, task.Result.Err)
		} else {
			successCount++
		}
	}

	assert.Positive(t, successCount, "Should have some successful results")
	assert.Positive(t, errorCount, "Should have some error results (invalid frontmatter)")
}

func TestGlobFrontMatter_FrontMatterParsing(t *testing.T) {
	setupTestFiles(t)

	tasks, err := mdfm.Glob[testMetadata]("blog/post1.md")
	require.NoError(t, err)
	require.Len(t, tasks, 1)

	task := tasks[0]
	require.NoError(t, task.Result.Err)
	require.NotNil(t, task.Result.Value)

	fm := task.Result.Value.FrontMatter
	assert.Equal(t, "First Post", fm.Title)
	assert.Equal(t, "This is the first post", fm.Description)
	assert.Equal(t, []string{"golang", "testing"}, fm.Tags)
	assert.True(t, fm.Published)

	assert.Contains(t, string(task.Result.Value.Body), "This is the content of the first post.")
	assert.Equal(t, "blog/post1.md", task.Metadata.Path)
}

func TestGlobFrontMatter_NoFrontMatter(t *testing.T) {
	setupTestFiles(t)

	tasks, err := mdfm.Glob[testMetadata]("no-frontmatter.md")
	require.NoError(t, err)
	require.Len(t, tasks, 1)

	task := tasks[0]
	require.NoError(t, task.Result.Err)
	require.NotNil(t, task.Result.Value)

	fm := task.Result.Value.FrontMatter
	assert.Empty(t, fm.Title)
	assert.Empty(t, fm.Description)
	assert.Empty(t, fm.Tags)
	assert.False(t, fm.Published)

	assert.Contains(t, string(task.Result.Value.Body), "# No Frontmatter")
}

func TestGlobFrontMatter_EmptyFile(t *testing.T) {
	setupTestFiles(t)

	tasks, err := mdfm.Glob[testMetadata]("empty.md")
	require.NoError(t, err)
	require.Len(t, tasks, 1)

	task := tasks[0]
	require.NoError(t, task.Result.Err)
	require.NotNil(t, task.Result.Value)

	assert.Empty(t, task.Result.Value.Body)
}

func TestGlobFrontMatter_GitIgnoreRespect(t *testing.T) {
	tmpDir := setupTestFiles(t)

	gitignoreContent := `blog/draft.md
*.tmp
ignored/
`
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	require.NoError(t, os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644))

	ignoredFile := filepath.Join(tmpDir, "blog", "ignored.tmp")
	require.NoError(t, os.WriteFile(ignoredFile, []byte("ignored content"), 0644))

	ignoredDir := filepath.Join(tmpDir, "ignored")
	require.NoError(t, os.MkdirAll(ignoredDir, 0755))
	ignoredInDir := filepath.Join(ignoredDir, "test.md")
	require.NoError(t, os.WriteFile(ignoredInDir, []byte("# Ignored"), 0644))

	tasks, err := mdfm.Glob[testMetadata]("**/*")
	require.NoError(t, err)

	var paths []string
	for _, task := range tasks {
		paths = append(paths, task.Metadata.Path)
	}

	assert.NotContains(t, paths, "blog/draft.md")
	assert.NotContains(t, paths, "blog/ignored.tmp")
	assert.NotContains(t, paths, "ignored/test.md")
}

func TestGlobFrontMatter_InvalidGlobPattern(t *testing.T) {
	setupTestFiles(t)

	_, err := mdfm.Glob[testMetadata]("[invalid")
	assert.Error(t, err)
}

func TestGlobFrontMatter_ConcurrentProcessing(t *testing.T) {
	setupTestFiles(t)

	tasks, err := mdfm.Glob[testMetadata]("**/*.md")
	require.NoError(t, err)

	assert.Greater(t, len(tasks), 1, "Should process multiple files")

	for _, task := range tasks {
		assert.NotEmpty(t, task.Metadata.Path, "Each result should have a path")
	}
}

func TestGlobFrontMatter_DifferentMetadataTypes(t *testing.T) {
	setupTestFiles(t)

	t.Run("map[string]any", func(t *testing.T) {
		tasks, err := mdfm.Glob[map[string]any]("blog/post1.md")
		require.NoError(t, err)
		require.Len(t, tasks, 1)

		task := tasks[0]
		require.NoError(t, task.Result.Err)

		fm := task.Result.Value.FrontMatter
		assert.Equal(t, "First Post", fm["title"])
		assert.Equal(t, true, fm["published"])
	})

	t.Run("struct with different fields", func(t *testing.T) {
		type simpleMetadata struct {
			Title string `yaml:"title"`
		}

		tasks, err := mdfm.Glob[simpleMetadata]("blog/post1.md")
		require.NoError(t, err)
		require.Len(t, tasks, 1)

		task := tasks[0]
		require.NoError(t, task.Result.Err)

		fm := task.Result.Value.FrontMatter
		assert.Equal(t, "First Post", fm.Title)
	})
}
