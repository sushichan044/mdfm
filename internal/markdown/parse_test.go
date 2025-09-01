package markdown_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/mdfm/internal/markdown"
)

type testMetadata struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Version     int    `yaml:"version"`
}

func TestParseMarkdownWithMetadata(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedMeta    testMetadata
		expectedContent string
		expectError     bool
	}{
		{
			name: "successful parsing with metadata",
			input: `---
title: Test Title
description: Test Description
version: 1
---
# Test Content

This is test content.`,
			expectedMeta: testMetadata{
				Title:       "Test Title",
				Description: "Test Description",
				Version:     1,
			},
			expectedContent: "# Test Content\n\nThis is test content.",
			expectError:     false,
		},
		{
			name: "successful parsing with empty metadata",
			input: `---
---
Content only`,
			expectedMeta: testMetadata{
				Title:       "",
				Description: "",
				Version:     0,
			},
			expectedContent: "Content only",
			expectError:     false,
		},
		{
			name: "parsing with no frontmatter",
			input: `# No Frontmatter
Just content`,
			expectedMeta:    testMetadata{},
			expectedContent: "# No Frontmatter\nJust content",
			expectError:     false,
		},
		{
			name: "parsing with invalid frontmatter",
			input: `---
title: "Unclosed quote
---
Content`,
			expectedMeta:    testMetadata{},
			expectedContent: "",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := markdown.ParseMarkdownWithMetadata[testMetadata](strings.NewReader(tt.input))

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedMeta, result.FrontMatter)
			assert.Equal(t, tt.expectedContent, result.Content)
		})
	}

	t.Run("metadata value mutation doesn't affect the source", func(t *testing.T) {
		content := `---
title: Original Title
---
Content`

		result, err := markdown.ParseMarkdownWithMetadata[testMetadata](strings.NewReader(content))
		require.NoError(t, err)
		assert.Equal(t, "Original Title", result.FrontMatter.Title)

		// Modify the result metadata
		result.FrontMatter.Title = "Modified Title"

		// Parse again and confirm the original values are preserved
		secondResult, err := markdown.ParseMarkdownWithMetadata[testMetadata](strings.NewReader(content))
		require.NoError(t, err)
		assert.Equal(t, "Original Title", secondResult.FrontMatter.Title)
	})
}
