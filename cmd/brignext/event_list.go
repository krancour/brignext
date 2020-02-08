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

func eventList(c *cli.Context) error {
	// Inputs
	var projectName string
	if len(c.Args()) > 0 {
		projectName = c.Args()[0]
	}
	output := c.String(flagOutput)
	allowInsecure := c.GlobalBool(flagInsecure)

	switch output {
	case "table":
	case "json":
	default:
		return errors.Errorf("unknown output format %q", output)
	}

	path := "v2/events"
	if projectName != "" {
		path = fmt.Sprintf("v2/projects/%s/events", projectName)
	}
	req, err := buildRequest(http.MethodGet, path, nil)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := getHTTPClient(allowInsecure).Do(req)
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

	events := []brignext.Event{}
	if err := json.Unmarshal(respBodyBytes, &events); err != nil {
		return errors.Wrap(err, "error unmarshaling response body")
	}

	if len(events) == 0 {
		fmt.Println("No events found.")
		return nil
	}

	switch output {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "PROJECT", "PROVIDER", "TYPE", "STATUS", "AGE")
		for _, event := range events {
			var status brignext.JobStatus = "???"
			since := "???"
			if event.Worker != nil {
				status = event.Worker.Status
				if status == brignext.JobSucceeded || status == brignext.JobFailed {
					since = duration.ShortHumanDuration(
						time.Since(event.Worker.StartTime),
					)
				}
			}
			table.AddRow(
				event.ID,
				event.ProjectName,
				event.Provider,
				event.Type,
				status,
				since,
			)
		}
		fmt.Println(table)

	case "json":
		responseJSON, err := json.MarshalIndent(events, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get events operation",
			)
		}
		fmt.Println(string(responseJSON))
	}

	return nil
}
