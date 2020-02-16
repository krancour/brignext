package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gosuri/uitable"
	"github.com/krancour/brignext/pkg/brignext"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"k8s.io/apimachinery/pkg/util/duration"
)

func eventGet(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"event get requires one argument-- an event ID",
		)
	}
	id := c.Args()[0]

	// Global flags
	allowInsecure := c.GlobalBool(flagInsecure)

	// Command-specific flags
	output := c.String(flagOutput)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	req, err := buildRequest(
		http.MethodGet,
		fmt.Sprintf("v2/events/%s", id),
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
		return errors.Errorf("Event %q not found.", id)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "error reading response body")
	}

	event := brignext.Event{}
	if err := json.Unmarshal(respBodyBytes, &event); err != nil {
		return errors.Wrap(err, "error unmarshaling response body")
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "PROJECT", "PROVIDER", "TYPE", "AGE", "STATUS")
		var age string
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
		fmt.Println(table)

		if len(event.Workers) > 0 {
			fmt.Printf("\nEvent %q workers:\n\n", event.ID)
			table = uitable.New()
			table.AddRow("NAME", "STARTED", "ENDED", "STATUS")
			for workerName, worker := range event.Workers {
				var started, ended string
				if worker.Started != nil {
					started = duration.ShortHumanDuration(time.Since(*worker.Started))
				}
				if worker.Ended != nil {
					ended = duration.ShortHumanDuration(time.Since(*worker.Ended))
				}
				table.AddRow(
					workerName,
					started,
					ended,
					worker.Status,
				)
			}
			fmt.Println(table)
		}

	case "json":
		prettyJSON, err := json.MarshalIndent(event, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get event operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}
