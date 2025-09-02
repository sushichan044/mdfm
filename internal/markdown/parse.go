package markdown

import (
	"io"

	"github.com/adrg/frontmatter"
)

// Parse parses the front matter from the given markdown content
// and returns the parsed metadata and the rest of the content.
//
// The front matter is expected to be in YAML format and is unmarshalled into the
// provided type T. The rest of the content is returned as a string.
//
// If the front matter is not present, FrontMatter will be empty.
func Parse[T any](input io.Reader, output io.Writer, frontMatter *T) error {
	// Parse the front matter and require it to be present
	rest, err := frontmatter.Parse(input, frontMatter)
	if err != nil {
		return err
	}

	_, err = output.Write(rest)
	return err
}
