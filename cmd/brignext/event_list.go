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
	// Args
	if len(c.Args()) != 0 {
		return errors.New(
			"event list requires no arguments",
		)
	}

	// GobalFlags
	allowInsecure := c.GlobalBool(flagInsecure)

	// Command-specific flags
	output := c.String(flagOutput)
	projectID := c.String(flagProject)

	switch output {
	case "table":
	case "json":
	default:
		return errors.Errorf("unknown output format %q", output)
	}

	req, err := buildRequest(http.MethodGet, "v2/events", nil)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}
	if projectID != "" {
		q := req.URL.Query()
		q.Set("projectID", projectID)
		req.URL.RawQuery = q.Encode()
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
		table.AddRow("ID", "PROJECT", "PROVIDER", "TYPE", "AGE", "STATUS")
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
