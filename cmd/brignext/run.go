package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"github.com/krancour/brignext/pkg/builds"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func run(c *cli.Context) error {
	// Inputs
	projectName := c.Args()[0]
	background := c.Bool(flagBackground)
	commit := c.String(flagCommit)
	configFile := c.String(flagConfig)
	event := c.String(flagEvent)
	scriptFile := c.String(flagFile)
	level := c.String(flagLevel)
	payloadFile := c.String(flagPayload)
	ref := c.String(flagRef)

	var configBytes []byte
	if configFile != "" {
		var err error
		configBytes, err = ioutil.ReadFile(configFile)
		if err != nil {
			return errors.Wrapf(err, "error reading config file %s", configFile)
		}
	}

	var scriptBytes []byte
	if scriptFile != "" {
		var err error
		scriptBytes, err = ioutil.ReadFile(scriptFile)
		if err != nil {
			return errors.Wrapf(err, "error reading script file %s", scriptFile)
		}
	}

	var payloadBytes []byte
	if payloadFile != "" {
		var err error
		payloadBytes, err = ioutil.ReadFile(payloadFile)
		if err != nil {
			return errors.Wrapf(err, "error reading payload file %s", payloadFile)
		}
	}

	conn, err := getConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := builds.NewBuildsClient(conn)
	response, err := client.CreateBuild(
		context.Background(),
		&builds.CreateBuildRequest{
			Build: &builds.Build{
				ProjectName: projectName,
				Type:        event,
				Provider:    "brigade-cli",
				Revision: &builds.Revision{
					Commit: commit,
					Ref:    ref,
				},
				Payload:  payloadBytes,
				Script:   scriptBytes,
				Config:   configBytes,
				LogLevel: level,
			},
		},
	)
	if err != nil {
		return errors.Wrap(err, "error creating build")
	}

	buildID := response.Build.Id

	if !background {
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
	}

	return nil
}
