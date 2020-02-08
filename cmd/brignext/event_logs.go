package main

// import (
// 	"context"
// 	"fmt"
// 	"io"
// 	"log"

// 	"github.com/krancour/brignext/pkg/builds"
// 	"github.com/urfave/cli"
// )

// func eventLogs(c *cli.Context) error {
// 	eventID := c.Args()[0]
// 	// init := c.Bool(flagInit)
// 	// jobs := c.Bool(flagJobs)
// 	// last := c.Bool(flagLast)

// 	conn, err := getConnection()
// 	if err != nil {
// 		return err
// 	}
// 	defer conn.Close()
// 	client := events.NewBuildsClient(conn)
// 	stream, err := client.StreamBuildLogs(
// 		context.Background(),
// 		&events.StreamBuildLogsRequest{
// 			Id: eventID,
// 		},
// 	)
// 	if err != nil {
// 		log.Fatalf("error getting log stream for event %q: %s", eventID, err)
// 	}
// 	for {
// 		logEntry, err := stream.Recv()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			log.Fatalf("error reading stream for event %q: %s", eventID, err)
// 		}
// 		fmt.Print(logEntry.Message)
// 	}

// 	return nil
// }
