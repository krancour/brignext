package main

import "github.com/urfave/cli/v2"

const (
	flagAborted        = "aborted"
	flagAnyPhase       = "any-phase"
	flagBrowse         = "browse"
	flagCanceled       = "canceled"
	flagContainer      = "container"
	flagDescription    = "description"
	flagEvent          = "event"
	flagFailed         = "failed"
	flagFile           = "file"
	flagFollow         = "follow"
	flagID             = "id"
	flagInsecure       = "insecure"
	flagJob            = "job"
	flagOutput         = "output"
	flagPassword       = "password"
	flagPayload        = "payload"
	flagPayloadFile    = "payload-file"
	flagPending        = "pending"
	flagProject        = "project"
	flagRole           = "role"
	flagNonTerminal    = "non-terminal"
	flagRoot           = "root"
	flagRunning        = "running"
	flagServer         = "server"
	flagServiceAccount = "service-account"
	flagSet            = "set"
	flagSource         = "source"
	flagSucceeded      = "succeeded"
	flagTerminal       = "terminal"
	flagTimedOut       = "timedout"
	flagType           = "type"
	flagUnknown        = "unknown"
	flagUnset          = "unset"
	flagUser           = "user"
	flagYes            = "yes"
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
