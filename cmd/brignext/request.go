package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

func getRequest(method, path string, body []byte) (*http.Request, error) {
	config, err := getConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "error retrieving configuration")
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewBuffer(body)
	}
	req, err := http.NewRequest(
		method,
		fmt.Sprintf("%s/%s", config.APIAddress, path),
		bodyReader,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating request %s %s", method, path)
	}

	req.Header.Add(
		"Authorization",
		fmt.Sprintf("Bearer %s", config.APIToken),
	)

	return req, nil
}
