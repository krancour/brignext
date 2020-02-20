package main

import (
	"strings"

	"github.com/pkg/errors"
)

func validateOutputFormat(outputFormat string) error {
	switch strings.ToLower(outputFormat) {
	case "table":
	case "json":
	default:
		return errors.Errorf("unknown output format %q", outputFormat)
	}
	return nil
}
