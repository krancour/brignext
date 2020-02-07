package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gosuri/uitable"
	"github.com/krancour/brignext/pkg/brignext"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"k8s.io/apimachinery/pkg/util/duration"
)

func buildGet(c *cli.Context) error {
	// Inputs
	if len(c.Args()) != 1 {
		return errors.New(
			"build get requires one parameter-- a build ID",
		)
	}
	id := c.Args()[0]
	output := c.String(flagOutput)
	allowInsecure := c.GlobalBool(flagInsecure)

	switch output {
	case "table":
	case "json":
	default:
		return errors.Errorf("unknown output format %q", output)
	}

	req, err := buildRequest(
		http.MethodGet,
		fmt.Sprintf("v2/builds/%s", id),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := getHTTPClient(allowInsecure).Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return errors.Errorf("Build %q not found.", id)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "error reading response body")
	}

	build := brignext.Build{}
	if err := json.Unmarshal(respBodyBytes, &build); err != nil {
		return errors.Wrap(err, "error unmarshaling response body")
	}

	switch output {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "PROJECT", "PROVIDER", "TYPE", "STATUS", "AGE")
		var status brignext.JobStatus = "???"
		since := "???"
		if build.Worker != nil {
			status = build.Worker.Status
			if status == brignext.JobSucceeded || status == brignext.JobFailed {
				since = duration.ShortHumanDuration(
					time.Since(build.Worker.StartTime),
				)
			}
		}
		table.AddRow(
			build.ID,
			build.ProjectName,
			build.Provider,
			build.Type,
			status,
			since,
		)
		fmt.Println(table)

	case "json":
		buildJSON, err := json.MarshalIndent(build, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get build operation",
			)
		}
		fmt.Println(string(buildJSON))
	}

	return nil
}
