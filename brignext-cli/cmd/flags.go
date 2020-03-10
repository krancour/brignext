package main

import "github.com/urfave/cli"

const (
	flagBrowse       = "browse"
	flagsBrowse      = "browse, b"
	flagDescription  = "description"
	flagsDescription = "description, d"
	flagEvent        = "event"
	flagsEvent       = "event, e"
	flagInit         = "init"
	flagInsecure     = "insecure"
	flagsInsecure    = "insecure, k"
	flagOutput       = "output"
	flagsOutput      = "output, o"
	flagPassword     = "password"
	flagsPassword    = "password, p"
	flagPending      = "pending"
	flagsPending     = "pending"
	flagProcessing   = "processing"
	flagsProcessing  = "processing"
	flagProject      = "project"
	flagsProject     = "project"
	flagProvider     = "provider"
	flagsProvider    = "provider, p"
	flagRoot         = "root"
	flagsRoot        = "root, r"
	flagRunning      = "running"
	flagsRunning     = "running, r"
	flagType         = "type"
	flagsType        = "type, t"
)

var (
	cliFlagOutput = cli.StringFlag{
		Name:  flagsOutput,
		Usage: "Return output in another format. Supported formats: table, json",
		Value: "table",
	}
)