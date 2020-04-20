package main

import "github.com/urfave/cli"

const (
	flagBrowse       = "browse"
	flagsBrowse      = "browse, b"
	flagDescription  = "description"
	flagsDescription = "description, d"
	flagEvent        = "event"
	flagsEvent       = "event, e"
	flagFollow       = "follow"
	flagsFollow      = "follow, f"
	flagInit         = "init"
	flagInsecure     = "insecure"
	flagsInsecure    = "insecure, k"
	flagOutput       = "output"
	flagsOutput      = "output, o"
	flagPassword     = "password"
	flagsPassword    = "password, p"
	flagPayload      = "payload"
	flagsPayload     = "payload, p"
	flagPayloadFile  = "payload-file"
	flagsPayloadFile = "payload-file"
	flagPending      = "pending"
	flagsPending     = "pending"
	flagProcessing   = "processing"
	flagsProcessing  = "processing"
	flagProject      = "project"
	flagsProject     = "project"
	flagRoot         = "root"
	flagsRoot        = "root, r"
	flagRunning      = "running"
	flagsRunning     = "running, r"
	flagSource       = "source"
	flagsSource      = "source, s"
	flagType         = "type"
	flagsType        = "type, t"
)

var (
	cliFlagOutput = cli.StringFlag{
		Name: flagsOutput,
		Usage: "Return output in another format. Supported formats: table, " +
			"yaml, json",
		Value: "table",
	}
)
