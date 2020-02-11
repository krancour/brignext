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

func workerGet(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"worker get requires one argument-- a worker ID",
		)
	}
	id := c.Args()[0]

	// Global flags
	allowInsecure := c.GlobalBool(flagInsecure)

	// Command-specific flags
	output := c.String(flagOutput)

	switch output {
	case "table":
	case "json":
	default:
		return errors.Errorf("unknown output format %q", output)
	}

	req, err := buildRequest(
		http.MethodGet,
		fmt.Sprintf("v2/workers/%s", id),
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
		return errors.Errorf("Worker %q not found.", id)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "error reading response body")
	}

	worker := brignext.Worker{}
	if err := json.Unmarshal(respBodyBytes, &worker); err != nil {
		return errors.Wrap(err, "error unmarshaling response body")
	}

	switch output {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "PROJECT", "EVENT", "PROVIDER", "TYPE", "CREATED", "STARTED", "ENDED", "STATUS") // nolint: lll
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
		fmt.Println(table)

	case "json":
		prettyJSON, err := json.MarshalIndent(worker, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get worker operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}
