package main

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/itchyny/gojq"
)

type (
	jsonPayload struct {
		Body        string `json:"body"`
		Path        string `json:"path"`
		FrontMatter any    `json:"frontMatter"`
	}

	// payloadPrinter writes a payload as JSON using a captured encoder.
	payloadPrinter func(payload jsonPayload) error
)

// AsMap converts the jsonPayload struct to a map[string]any.
func (p jsonPayload) AsMap() map[string]any {
	jsonData, marshalErr := json.Marshal(p)
	if marshalErr != nil {
		return map[string]any{}
	}

	var result map[string]any
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return map[string]any{}
	}

	return result
}

func NewAppropriatePrinter(output io.Writer, jqFilter string) (payloadPrinter, error) {
	enc := json.NewEncoder(output)
	enc.SetIndent("", "  ")

	var jqCode *gojq.Code
	if jqFilter != "" {
		var jqErr error
		jqCode, jqErr = prepareJQ(jqFilter)
		if jqErr != nil {
			return nil, jqErr
		}
	}

	if jqCode != nil {
		return newJQPrinter(enc, jqCode)
	}
	return newPassthroughPrinter(enc), nil
}

func newPassthroughPrinter(enc *json.Encoder) payloadPrinter {
	return func(payload jsonPayload) error {
		return enc.Encode(payload)
	}
}

func newJQPrinter(enc *json.Encoder, jqCode *gojq.Code) (payloadPrinter, error) {
	if jqCode == nil {
		return nil, errors.New("jqCode must not be nil")
	}

	return func(payload jsonPayload) error {
		iter := jqCode.Run(payload.AsMap())

		for {
			v, ok := iter.Next()
			if !ok {
				break
			}
			if err, isErr := v.(error); isErr {
				var e *gojq.HaltError
				if errors.As(err, &e) && e.Value() == nil {
					break
				}
				return err
			}
			if err := enc.Encode(v); err != nil {
				return err
			}
		}
		return nil
	}, nil
}
