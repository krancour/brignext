package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gosuri/uitable"
	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/builds"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"k8s.io/apimachinery/pkg/util/duration"
)

func buildGet(c *cli.Context) error {
	// Inputs
	id := c.Args()[0]
	output := c.String(flagOutput)

	switch output {
	case "table":
	case "json":
	default:
		return errors.Errorf("unknown output format %q", output)
	}

	conn, err := getConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := builds.NewBuildsClient(conn)
	response, err := client.GetBuild(
		context.Background(),
		&builds.GetBuildRequest{
			Id: id,
		},
	)
	if err != nil {
		return errors.Wrap(err, "error getting build")
	}

	if response.Build == nil {
		return nil
	}

	build := builds.WireBuildToBrignextBuild(response.Build)

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
