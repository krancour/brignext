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
	var projectID string
	if len(c.Args()) > 0 {
		projectID = c.Args()[0]
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
	if projectID != "" {
		path = fmt.Sprintf("v2/projects/%s/events", projectID)
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
		table.AddRow("ID", "PROJECT ID", "PROVIDER", "TYPE", "AGE", "STATUS")
		for _, event := range events {
			age := "???"
			if event.Created != nil {
				age = duration.ShortHumanDuration(time.Since(*event.Created))
			}
			table.AddRow(
				event.ID,
				event.ProjectID,
				event.Provider,
				event.Type,
				age,
				event.Status,
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
