package main

import (
	"fmt"
	"os"

	"github.com/krancour/brignext/v2/pkg/signals"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "brignext"
	app.Usage = "Is this what Brigade 2.0 looks like?"
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    flagInsecure,
			Aliases: []string{"k"},
			Usage:   "Allow insecure API server connections when using TLS",
		},
	}
	app.Commands = []*cli.Command{
		{
			Name:  "event",
			Usage: "Manage events",
			Subcommands: []*cli.Command{
				{
					Name:        "cancel",
					Usage:       "Cancel event(s) without deleting them",
					Description: "By default, only cancels events in a PENDING state.",
					ArgsUsage:   "[EVENT_ID]",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  flagProcessing,
							Usage: "If set, will also abort events in a PROCESSING state",
						},
						&cli.StringFlag{
							Name: flagProject,
							Usage: "Cancel all events for the specified project " +
								"(ignores EVENT_ID argument)",
						},
					},
					Action: eventCancel,
				},
				{
					Name:      "create",
					Usage:     "Create a new event",
					ArgsUsage: "PROJECT_ID",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    flagPayload,
							Aliases: []string{"p"},
							Usage:   "The event payload",
						},
						&cli.StringFlag{
							Name:  flagPayloadFile,
							Usage: "The location of a file containing the event payload",
						},
						&cli.StringFlag{
							Name:    flagSource,
							Aliases: []string{"s"},
							Usage:   "The event source",
							Value:   "brignext-cli",
						},
						&cli.StringFlag{
							Name:    flagType,
							Aliases: []string{"t"},
							Usage:   "The event type",
							Value:   "exec",
						},
					},
					Action: eventCreate,
				},
				{
					Name:  "delete",
					Usage: "Delete event(s)",
					Description: "By default, only deletes events in a terminal state " +
						"(neither PENDING nor PROCESSING).",
					ArgsUsage: "[EVENT_ID]",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  flagPending,
							Usage: "If set, will also delete events in a PENDING state",
						},
						&cli.BoolFlag{
							Name: flagProcessing,
							Usage: "If set, will also abort and delete events in a " +
								"PROCESSING state",
						},
						&cli.StringFlag{
							Name: flagProject,
							Usage: "Delete all events for the specified project " +
								"(ignores EVENT_ID argument)",
						},
					},
					Action: eventDelete,
				},
				{
					Name:      "get",
					Usage:     "Get an event",
					ArgsUsage: "EVENT_ID",
					Flags: []cli.Flag{
						cliFlagOutput,
					},
					Action: eventGet,
				},
				{
					Name:  "list",
					Usage: "List events",
					Flags: []cli.Flag{
						cliFlagOutput,
						&cli.StringFlag{
							Name:  flagProject,
							Usage: "Return events only for the specified project",
						},
					},
					Action: eventList,
				},
			},
		},
		{
			Name:  "job",
			Usage: "Manage jobs",
			Subcommands: []*cli.Command{
				{
					Name:      "get",
					Usage:     "Get a job",
					ArgsUsage: "EVENT_ID WORKER_NAME JOB_NAME",
					Flags: []cli.Flag{
						cliFlagOutput,
					},
					Action: jobGet,
				},
				{
					Name:      "logs",
					Usage:     "Get job logs",
					ArgsUsage: "EVENT_ID WORKER_NAME JOB_NAME",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:    flagFollow,
							Aliases: []string{"f"},
							Usage:   "If set, will stream job logs until interrupted",
						},
					},
					Action: jobLogs,
				},
			},
		},
		{
			Name:      "login",
			Usage:     "Log in to BrigNext",
			ArgsUsage: "API_SERVER_ADDRESS",
			Description: "By default, initiates authentication using OpenID " +
				"Connect. This may not be supported by all BrigNext API servers.",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    flagBrowse,
					Aliases: []string{"b"},
					Usage: "Use the system's default web browser to complete " +
						"authentication (not applicable when --root is used)",
				},
				&cli.StringFlag{
					Name:    flagPassword,
					Aliases: []string{"p"},
					Usage: "Specify the password for root user login " +
						"non-interactively (only applicable when --root is used)",
				},
				&cli.BoolFlag{
					Name:    flagRoot,
					Aliases: []string{"r"},
					Usage:   "Log in as the root user (does not use OpenID Connect)",
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
			Subcommands: []*cli.Command{
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
						cliFlagOutput,
					},
					Action: projectGet,
				},
				{
					Name:  "list",
					Usage: "List projects",
					Flags: []cli.Flag{
						cliFlagOutput,
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
			Name:  "secrets",
			Usage: "Manage worker secrets",
			Subcommands: []*cli.Command{
				{
					Name:      "list",
					Usage:     "List a worker's secrets",
					ArgsUsage: "PROJECT_ID WORKER_NAME",
					Flags: []cli.Flag{
						cliFlagOutput,
					},
					Action: secretsList,
				},
				{
					Name:  "set",
					Usage: "Define or redefine the value of a secret",
					ArgsUsage: "PROJECT_ID WORKER_NAME KEY_0=VALUE_0 " +
						"[KEY_1=VALUE_1 .. KEY_N=VALUE_N]",
					Action: secretsSet,
				},
				{
					Name:      "unset",
					Usage:     "Clear the value of a secret",
					ArgsUsage: "PROJECT_ID WORKER_NAME KEY_0 [KEY_1 .. KEY_N]",
					Action:    secretsUnset,
				},
			},
		},
		{
			Name:  "service-account",
			Usage: "Manage service accounts",
			Subcommands: []*cli.Command{
				{
					Name:      "create",
					Usage:     "Create a new service account",
					ArgsUsage: "[SERVICE_ACCOUNT_ID]",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    flagDescription,
							Aliases: []string{"d"},
							Usage:   "A description of the service account",
						},
					},
					Action: serviceAccountCreate,
				},
				{
					Name:      "get",
					Usage:     "Get a service account",
					ArgsUsage: "SERVICE_ACCOUNT_ID",
					Flags: []cli.Flag{
						cliFlagOutput,
					},
					Action: serviceAccountGet,
				},
				{
					Name:  "list",
					Usage: "List service accounts",
					Flags: []cli.Flag{
						cliFlagOutput,
					},
					Action: serviceAccountList,
				},
				{
					Name:      "lock",
					Usage:     "Lock a service account out of BrigNext",
					ArgsUsage: "SERVICE_ACCOUNT_ID",
					Action:    serviceAccountLock,
				},
				{
					Name:      "unlock",
					Usage:     "Restore a service account's access to BrigNext",
					ArgsUsage: "SERVICE_ACCOUNT_ID",
					Action:    serviceAccountUnlock,
				},
			},
		},
		{
			Name:  "user",
			Usage: "Manage users",
			Subcommands: []*cli.Command{
				{
					Name:      "get",
					Usage:     "Get a user",
					ArgsUsage: "USER_ID",
					Flags: []cli.Flag{
						cliFlagOutput,
					},
					Action: userGet,
				},
				{
					Name:  "list",
					Usage: "List users",
					Flags: []cli.Flag{
						cliFlagOutput,
					},
					Action: userList,
				},
				{
					Name:      "lock",
					Usage:     "Lock a user out of BrigNext",
					ArgsUsage: "USER_ID",
					Action:    userLock,
				},
				{
					Name:      "unlock",
					Usage:     "Restore a user's access to BrigNext",
					ArgsUsage: "USER_ID",
					Action:    userUnlock,
				},
			},
		},
		{
			Name:  "worker",
			Usage: "Manage workers",
			Subcommands: []*cli.Command{
				{
					Name:        "cancel",
					Usage:       "Cancel a worker",
					Description: "By default, only cancels workers in a PENDING state.",
					ArgsUsage:   "WORKER_ID",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:    flagRunning,
							Aliases: []string{"r"},
							Usage:   "If set, will abort a worker in a RUNNING state",
						},
					},
					Action: workerCancel,
				},
				{
					Name:      "get",
					Usage:     "Get a worker",
					ArgsUsage: "EVENT_ID WORKER_NAME",
					Flags: []cli.Flag{
						cliFlagOutput,
					},
					Action: workerGet,
				},
				{
					Name:      "logs",
					Usage:     "Get worker logs",
					ArgsUsage: "EVENT_ID WORKER_NAME",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:    flagFollow,
							Aliases: []string{"f"},
							Usage:   "If set, will stream worker logs until interrupted",
						},
					},
					Action: workerLogs,
				},
			},
		},
	}
	fmt.Println()
	if err := app.RunContext(signals.Context(), os.Args); err != nil {
		fmt.Printf("\n%s\n\n", err)
		os.Exit(1)
	}
	fmt.Println()
}
