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
					Name:      "create",
					Usage:     "Create a new event",
					ArgsUsage: "PROJECT_ID",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  flagsProvider,
							Usage: "The event provider",
							Value: "brignext-cli",
						},
						cli.StringFlag{
							Name:  flagsType,
							Usage: "The event type",
							Value: "exec",
						},
					},
					Action: eventCreate,
				},
				{
					Name:      "delete",
					Usage:     "Delete an event",
					ArgsUsage: "[EVENT_ID]",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name: flagsForce,
							Usage: "If set, will also delete events with running workers. " +
								"Default: false",
						},
						cli.StringFlag{
							Name:  flagsProject,
							Usage: "Delete all events for the specified project",
						},
					},
					Action: eventDelete,
				},
				{
					Name:      "get",
					Usage:     "Get an event",
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
					Name:  "list",
					Usage: "List events, optionally filtered by project",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: flagsOutput,
							Usage: "Return output in another format. Supported formats: " +
								"table, json",
							Value: "table",
						},
						cli.StringFlag{
							Name:  flagsProject,
							Usage: "Return events only for the specified project",
						},
					},
					Action: eventList,
				},
			},
		},
		{
			Name:      "login",
			Usage:     "Log in to BrigNext",
			ArgsUsage: "API_SERVER_ADDRESS",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: flagsBrowse,
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
			Usage:  "Log out of BrigNext",
			Action: logout,
		},
		{
			Name:  "project",
			Usage: "Manage projects",
			Subcommands: []cli.Command{
				{
					Name:      "create",
					Usage:     "Create a new project",
					ArgsUsage: "FILE",
					Action:    projectCreate,
				},
				{
					Name:      "delete",
					Usage:     "Delete a project",
					ArgsUsage: "PROJECT_ID",
					Action:    projectDelete,
				},
				{
					Name:      "get",
					Usage:     "Get a project",
					ArgsUsage: "PROJECT_ID",
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
					Usage: "List projects",
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
					Usage:     "Update a project",
					ArgsUsage: "FILE",
					Action:    projectUpdate,
				},
			},
		},
		{
			Name:  "service-account",
			Usage: "Manage service accounts",
			Subcommands: []cli.Command{
				{
					Name:      "create",
					Usage:     "Create a new service account",
					ArgsUsage: "[SERVICE_ACCOUNT_ID]",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  flagsDescription,
							Usage: "A description of the service account",
						},
					},
					Action: serviceAccountCreate,
				},
				{
					Name:      "get",
					Usage:     "Get a service account",
					ArgsUsage: "SERVICE_ACCOUNT_ID",
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
					Usage: "List service accounts",
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
				{
					Name:      "lock",
					Usage:     "Lock a service account out of Brigade",
					ArgsUsage: "SERVICE_ACCOUNT_ID",
					Action:    serviceAccountLock,
				},
				{
					Name:      "unlock",
					Usage:     "Restore a service account's access to Brigade",
					ArgsUsage: "SERVICE_ACCOUNT_ID",
					Action:    serviceAccountUnlock,
				},
			},
		},
		{
			Name:  "user",
			Usage: "Manage users",
			Subcommands: []cli.Command{
				{
					Name:      "get",
					Usage:     "Get a user",
					ArgsUsage: "USER_ID",
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
					Usage: "List users",
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
				{
					Name:      "lock",
					Usage:     "Lock a user out of Brigade",
					ArgsUsage: "USER_ID",
					Action:    userLock,
				},
				{
					Name:      "unlock",
					Usage:     "Restore a user's access to Brigade",
					ArgsUsage: "USER_ID",
					Action:    userUnlock,
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
