package main

import "github.com/urfave/cli/v2"

var systemCommand = &cli.Command{
	Name:  "system",
	Usage: "Manage the Brigade system",
	Subcommands: []*cli.Command{
		systemRolesCommand,
	},
}
