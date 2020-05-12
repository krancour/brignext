package main

import "github.com/urfave/cli/v2"

const (
	flagBrowse      = "browse"
	flagDescription = "description"
	flagEvent       = "event"
	flagFile        = "file"
	flagFollow      = "follow"
	flagID          = "id"
	flagInit        = "init"
	flagInsecure    = "insecure"
	flagJob         = "job"
	flagOutput      = "output"
	flagPassword    = "password"
	flagPayload     = "payload"
	flagPayloadFile = "payload-file"
	flagPending     = "pending"
	flagProject     = "project"
	flagRoot        = "root"
	flagRunning     = "running"
	flagServer      = "server"
	flagSet         = "set"
	flagSource      = "source"
	flagType        = "type"
	flagUnset       = "unset"
)

var (
	cliFlagOutput = &cli.StringFlag{
		Name:    flagOutput,
		Aliases: []string{"o"},
		Usage: "Return output in the specified format; supported formats: table, " +
			"yaml, json",
		Value: "table",
	}
)
