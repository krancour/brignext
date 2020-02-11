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

func workerList(c *cli.Context) error {
	// Args
	if len(c.Args()) != 0 {
		return errors.New(
			"worker list requires no arguments",
		)
	}

	// GobalFlags
	allowInsecure := c.GlobalBool(flagInsecure)

	// Command-specific flags
	eventID := c.String(flagEvent)
	output := c.String(flagOutput)
	projectID := c.String(flagProject)

	switch output {
	case "table":
	case "json":
	default:
		return errors.Errorf("unknown output format %q", output)
	}

	req, err := buildRequest(http.MethodGet, "v2/workers", nil)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}
	q := req.URL.Query()
	if eventID != "" {
		q.Set("eventID", eventID)
	}
	if projectID != "" {
		q.Set("projectID", projectID)
	}
	req.URL.RawQuery = q.Encode()

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

	workers := []brignext.Worker{}
	if err := json.Unmarshal(respBodyBytes, &workers); err != nil {
		return errors.Wrap(err, "error unmarshaling response body")
	}

	if len(workers) == 0 {
		fmt.Println("No workers found.")
		return nil
	}

	switch output {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "PROJECT", "EVENT", "PROVIDER", "TYPE", "CREATED", "STARTED", "ENDED", "STATUS") // nolint: lll
		for _, worker := range workers {
			var started, ended string
			if worker.Started != nil {
				started = duration.ShortHumanDuration(time.Since(*worker.Started))
			}
			if worker.Ended != nil {
				ended = duration.ShortHumanDuration(time.Since(*worker.Ended))
			}
			table.AddRow(
				worker.ID,
				worker.ProjectID,
				worker.EventID,
				worker.EventProvider,
				worker.EventType,
				duration.ShortHumanDuration(time.Since(worker.Created)),
				started,
				ended,
				worker.Status,
			)
		}
		fmt.Println(table)

	case "json":
		prettyJSON, err := json.MarshalIndent(workers, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get workers operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}
