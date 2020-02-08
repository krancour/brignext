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
			Name:  "event",
			Usage: "Manage events",
			Subcommands: []cli.Command{
				{
					Name:      "delete",
					Usage:     "deletes an event by ID",
					ArgsUsage: "EVENT_ID",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name: flagsForce,
							Usage: "If set, will also delete events with running workers. " +
								"Default: false",
						},
					},
					Action: eventDelete,
				},
				{
					Name:      "delete-all",
					Usage:     "deletes all events for a given project",
					ArgsUsage: "PROJECT_NAME",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name: flagsForce,
							Usage: "If set, will also delete events with running workers. " +
								"Default: false",
						},
					},
					Action: eventDeleteAll,
				},
				{
					Name:      "get",
					Usage:     "get an event",
					ArgsUsage: "EVENT_ID",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: flagsOutput,
							Usage: "Return output in another format. Supported formats: " +
								"table, json",
							Value: "table",
						},
					},
					Action: eventGet,
				},
				{
					Name:      "list",
					Usage:     "list all events or events for a given project",
					ArgsUsage: "[PROJECT_NAME]",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: flagsOutput,
							Usage: "Return output in another format. Supported formats: " +
								"table, json",
							Value: "table",
						},
					},
					Action: eventList,
				},
				// {
				// 	Name:      "logs",
				// 	Usage:     "show event logs",
				// 	ArgsUsage: "EVENT_ID",
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
				// 			Usage: "Show last event's log (ignores EVENT_ID)",
				// 		},
				// 	},
				// 	Action: eventLogs,
				// },
			},
		},
		{
			Name:      "login",
			Usage:     "Log in to Brigade",
			ArgsUsage: "API_SERVER_ADDRESS",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: flagsOpen,
					Usage: "Use the system's default web browser to navigate to the " +
						"URL to begin authentication using OpenID Connect " +
						"(if supported); not applicable when --root is used",
				},
				cli.StringFlag{
					Name: flagsPassword,
					Usage: "Specify the password for root user login " +
						"non-interactively; only applicaple when --root is used",
				},
				cli.BoolFlag{
					Name:  flagsRoot,
					Usage: "Log in as the root user",
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
					Action:    projectCreate,
				},
				{
					Name:      "delete",
					Usage:     "delete a project",
					ArgsUsage: "PROJECT_NAME",
					Action:    projectDelete,
				},
				{
					Name:      "get",
					Usage:     "get a project",
					ArgsUsage: "PROJECT_NAME",
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
					Action:    projectUpdate,
				},
			},
		},
		{
			Name:      "run",
			Usage:     "Crete a new event",
			ArgsUsage: "PROJECT_NAME",
			Flags: []cli.Flag{
				// cli.BoolFlag{
				// 	Name: flagsBackground,
				// 	Usage: "Trigger the event and exit. Let the job run in the " +
				// 		"background.",
				// },
				// cli.StringFlag{
				// 	Name:  flagsCommit,
				// 	Usage: "A VCS (git) commit",
				// },
				// cli.StringFlag{
				// 	Name:  flagConfig,
				// 	Usage: "The brigade.json config file",
				// },
				// cli.StringFlag{
				// 	Name:  flagsFile,
				// 	Usage: "The JavaScript file to execute",
				// },
				// cli.StringFlag{
				// 	Name:  flagsLevel,
				// 	Usage: "Specified log level: log, info, warn, error",
				// 	Value: "log",
				// },
				// cli.StringFlag{
				// 	Name:  flagsPayload,
				// 	Usage: "The path to a payload file",
				// },
				// cli.StringFlag{
				// 	Name:  flagsRef,
				// 	Usage: "A VCS (git) version, tag, or branch",
				// 	Value: "master",
				// },
				cli.StringFlag{
					Name:  flagsType,
					Usage: "The event type fire",
					Value: "exec",
				},
			},
			Action: run,
		},
		{
			Name:  "service-account",
			Usage: "Manage service accounts",
			Subcommands: []cli.Command{
				{
					Name:      "create",
					Usage:     "create a new service account",
					ArgsUsage: "[SERVICE_ACCOUNT_NAME]",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  flagsDescription,
							Usage: "A description of the service account",
						},
					},
					Action: serviceAccountCreate,
				},
				{
					Name:      "delete",
					Usage:     "delete a service account",
					ArgsUsage: "SERVICE_ACCOUNT_NAME",
					Action:    serviceAccountDelete,
				},
				{
					Name:      "get",
					Usage:     "get a service account",
					ArgsUsage: "SERVICE_ACCOUNT_NAME",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: flagsOutput,
							Usage: "Return output in another format. Supported formats: " +
								"table, json",
							Value: "table",
						},
					},
					Action: serviceAccountGet,
				},
				{
					Name:  "list",
					Usage: "list service accounts",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: flagsOutput,
							Usage: "Return output in another format. Supported formats: " +
								"table, json",
							Value: "table",
						},
					},
					Action: serviceAccountList,
				},
			},
		},
		{
			Name:  "user",
			Usage: "Manage users",
			Subcommands: []cli.Command{
				{
					Name:      "delete",
					Usage:     "delete a user",
					ArgsUsage: "USERNAME",
					Action:    userDelete,
				},
				{
					Name:      "get",
					Usage:     "get a user",
					ArgsUsage: "USERNAME",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: flagsOutput,
							Usage: "Return output in another format. Supported formats: " +
								"table, json",
							Value: "table",
						},
					},
					Action: userGet,
				},
				{
					Name:  "list",
					Usage: "list users",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: flagsOutput,
							Usage: "Return output in another format. Supported formats: " +
								"table, json",
							Value: "table",
						},
					},
					Action: userList,
				},
			},
		},
	}
	fmt.Println()
	if err := app.Run(os.Args); err != nil {
		fmt.Printf("\n%s\n\n", err)
		os.Exit(1)
	}
	fmt.Println()
}
