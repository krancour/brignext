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

func buildList(c *cli.Context) error {
	// Inputs
	var projectName string
	if len(c.Args()) > 0 {
		projectName = c.Args()[0]
	}
	output := c.String(flagOutput)

	switch output {
	case "table":
	case "json":
	default:
		return errors.Errorf("unknown output format %q", output)
	}

	path := "v2/builds"
	if projectName != "" {
		path = fmt.Sprintf("v2/projects/%s/builds", projectName)
	}
	req, err := getRequest(http.MethodGet, path, nil)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "error reading response body")
	}

	builds := []*brignext.Build{}
	if err := json.Unmarshal(respBodyBytes, &builds); err != nil {
		return errors.Wrap(err, "error unmarshaling response body")
	}

	if len(builds) == 0 {
		fmt.Println("No builds found.")
		return nil
	}

	switch output {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "PROJECT", "PROVIDER", "TYPE", "STATUS", "AGE")
		for _, build := range builds {
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
		}
		fmt.Println(table)

	case "json":
		responseJSON, err := json.MarshalIndent(builds, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get builds operation",
			)
		}
		fmt.Println(string(responseJSON))
	}

	return nil
}
