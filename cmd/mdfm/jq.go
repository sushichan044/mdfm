package main

import (
	"fmt"
	"os"

	"github.com/itchyny/gojq"
)

func prepareJQ(filter string) (*gojq.Code, error) {
	jqQuery, jqErr := gojq.Parse(filter)
	if jqErr != nil {
		return nil, fmt.Errorf("error parsing jq filter %q: %w", filter, jqErr)
	}
	jqCode, jqCompileErr := gojq.Compile(jqQuery, gojq.WithEnvironLoader(os.Environ))
	if jqCompileErr != nil {
		return nil, fmt.Errorf("error compiling jq filter %q: %w", filter, jqCompileErr)
	}

	return jqCode, nil
}
