package main

import "github.com/urfave/cli/v2"

const (
	flagBrowse      = "browse"
	flagDescription = "description"
	flagFollow      = "follow"
	flagInit        = "init"
	flagInsecure    = "insecure"
	flagOutput      = "output"
	flagPassword    = "password"
	flagPayload     = "payload"
	flagPayloadFile = "payload-file"
	flagPending     = "pending"
	flagProject     = "project"
	flagRoot        = "root"
	flagRunning     = "running"
	flagSource      = "source"
	flagType        = "type"
)

var (
	cliFlagOutput = &cli.StringFlag{
		Name:    flagOutput,
		Aliases: []string{"o"},
		Usage: "Return output in another format. Supported formats: table, " +
			"yaml, json",
		Value: "table",
	}
)
