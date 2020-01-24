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

func buildList(c *cli.Context) error {
	// Inputs
	var projectName string
	if len(c.Args()) > 0 {
		projectName = c.Args()[0]
	}
	// count := c.Int(flagCount)
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
	response, err := client.GetBuilds(
		context.Background(),
		&builds.GetBuildsRequest{
			ProjectName: projectName,
		},
	)
	if err != nil {
		return errors.Wrap(err, "error listing builds")
	}

	bs := make([]*brignext.Build, len(response.Builds))
	for i, wireBuild := range response.Builds {
		bs[i] = builds.WireBuildToBrignextBuild(wireBuild)
	}

	switch output {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "PROJECT", "PROVIDER", "TYPE", "STATUS", "AGE")
		for _, build := range bs {
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
		responseJSON, err := json.MarshalIndent(bs, "", "  ")
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
