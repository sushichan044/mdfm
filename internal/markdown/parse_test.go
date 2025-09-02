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

func TestParse(t *testing.T) {
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
			var output strings.Builder
			defer output.Reset()

			var meta testMetadata
			err := markdown.Parse[testMetadata](strings.NewReader(tt.input), &output, &meta)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedMeta, meta)
			assert.Equal(t, tt.expectedContent, output.String())
		})
	}

	t.Run("metadata value mutation doesn't affect the source", func(t *testing.T) {
		content := `---
title: Original Title
---
Content`

		var output strings.Builder
		var meta testMetadata
		err := markdown.Parse[testMetadata](strings.NewReader(content), &output, &meta)
		require.NoError(t, err)
		assert.Equal(t, "Original Title", meta.Title)

		// Modify the result metadata
		meta.Title = "Modified Title"

		// Parse again and confirm the original values are preserved
		err = markdown.Parse[testMetadata](strings.NewReader(content), &output, &meta)
		require.NoError(t, err)
		assert.Equal(t, "Original Title", meta.Title)
	})
}
