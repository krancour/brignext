package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/krancour/brignext/pkg/builds"
	"github.com/urfave/cli"
)

func buildLogs(c *cli.Context) error {
	buildID := c.Args()[0]
	// init := c.Bool(flagInit)
	// jobs := c.Bool(flagJobs)
	// last := c.Bool(flagLast)

	conn, err := getConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := builds.NewBuildsClient(conn)
	stream, err := client.StreamBuildLogs(
		context.Background(),
		&builds.StreamBuildLogsRequest{
			Id: buildID,
		},
	)
	if err != nil {
		log.Fatalf("error getting log stream for build %q: %s", buildID, err)
	}
	for {
		logEntry, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error reading stream for build %q: %s", buildID, err)
		}
		fmt.Print(logEntry.Message)
	}

	return nil
}
