package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "brignext"
	app.Usage = "Is this what Brigade 2.0 looks like?"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  flagsInsecure,
			Usage: "Allow insecure API server connections when using TLS",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "build",
			Usage: "Manage builds",
			Subcommands: []cli.Command{
				{
					Name:      "delete",
					Usage:     "deletes build",
					ArgsUsage: "BUILD",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  flagsForce,
							Usage: "If set, will also delete running builds. Default: false",
						},
					},
					Action: buildDelete,
				},
				{
					Name:      "delete-all",
					Usage:     "deletes all builds for a project",
					ArgsUsage: "PROJECT",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  flagsForce,
							Usage: "If set, will also delete running builds. Default: false",
						},
					},
					Action: buildDeleteAll,
				},
				{
					Name:      "get",
					Usage:     "get a build",
					ArgsUsage: "BUILD",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: flagsOutput,
							Usage: "Return output in another format. Supported formats: " +
								"table, json",
							Value: "table",
						},
					},
					Action: buildGet,
				},
				{
					Name:      "list",
					Usage:     "list builds",
					ArgsUsage: "[PROJECT]",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: flagsOutput,
							Usage: "Return output in another format. Supported formats: " +
								"table, json",
							Value: "table",
						},
					},
					Action: buildList,
				},
				// {
				// 	Name:      "logs",
				// 	Usage:     "show build logs",
				// 	ArgsUsage: "BUILD",
				// 	Flags: []cli.Flag{
				// 		cli.BoolFlag{
				// 			Name:  flagsInit,
				// 			Usage: "Show init container logs as well as the worker log",
				// 		},
				// 		cli.BoolFlag{
				// 			Name:  flagsJobs,
				// 			Usage: "Show job logs as well as the worker log",
				// 		},
				// 		cli.BoolFlag{
				// 			Name:  flagsLast,
				// 			Usage: "Show last build's log (ignores BUILD_ID)",
				// 		},
				// 	},
				// 	Action: buildLogs,
				// },
			},
		},
		{
			Name:      "login",
			Usage:     "Log in to Brigade",
			ArgsUsage: "HOST",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  flagsRoot,
					Usage: "Log in as the root user",
				},
				cli.StringFlag{
					Name: flagsPassword,
					Usage: "Specify the password for root user login " +
						"non-interactively; only applicaple when --root is used",
				},
				cli.BoolFlag{
					Name: flagsOpen,
					Usage: "Use the system's default web browser to navigate to the " +
						"authentication URL; not applicable when --root is used",
				},
			},
			Action: login,
		},
		{
			Name:   "logout",
			Usage:  "Log out of Brigade",
			Action: logout,
		},
		{
			Name:  "project",
			Usage: "Manage projects",
			Subcommands: []cli.Command{
				{
					Name:      "create",
					Usage:     "create a new project",
					ArgsUsage: "FILE",
					Flags:     []cli.Flag{},
					Action:    projectCreate,
				},
				{
					Name:      "delete",
					Usage:     "delete a project",
					ArgsUsage: "PROJECT",
					Flags:     []cli.Flag{},
					Action:    projectDelete,
				},
				{
					Name:      "get",
					Usage:     "get a project",
					ArgsUsage: "PROJECT",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: flagsOutput,
							Usage: "Return output in another format. Supported formats: " +
								"table, json",
							Value: "table",
						},
					},
					Action: projectGet,
				},
				{
					Name:  "list",
					Usage: "list projects",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: flagsOutput,
							Usage: "Return output in another format. Supported formats: " +
								"table, json",
							Value: "table",
						},
					},
					Action: projectList,
				},
				{
					Name:      "update",
					Usage:     "update a project",
					ArgsUsage: "FILE",
					Flags:     []cli.Flag{},
					Action:    projectUpdate,
				},
			},
		},
		{
			Name:      "run",
			Usage:     "Kick off a build",
			ArgsUsage: "PROJECT",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: flagsBackground,
					Usage: "Trigger the event and exit. Let the job run in the " +
						"background.",
				},
				cli.StringFlag{
					Name:  flagsCommit,
					Usage: "A VCS (git) commit",
				},
				cli.StringFlag{
					Name:  flagConfig,
					Usage: "The brigade.json config file",
				},
				cli.StringFlag{
					Name:  flagsEvent,
					Usage: "The name of the event to fire",
					Value: "exec",
				},
				cli.StringFlag{
					Name:  flagsFile,
					Usage: "The JavaScript file to execute",
				},
				cli.StringFlag{
					Name:  flagsLevel,
					Usage: "Specified log level: log, info, warn, error",
					Value: "log",
				},
				cli.StringFlag{
					Name:  flagsPayload,
					Usage: "The path to a payload file",
				},
				cli.StringFlag{
					Name:  flagsRef,
					Usage: "A VCS (git) version, tag, or branch",
					Value: "master",
				},
			},
			Action: run,
		},
	}
	fmt.Println()
	if err := app.Run(os.Args); err != nil {
		fmt.Printf("\n%s\n\n", err)
		os.Exit(1)
	}
	fmt.Println()
}
