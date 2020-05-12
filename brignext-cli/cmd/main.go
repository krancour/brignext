package main

import (
	"fmt"
	"os"

	"github.com/krancour/brignext/v2/pkg/signals"
	"github.com/krancour/brignext/v2/pkg/version"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "brignext"
	app.Usage = "Is this what Brigade 2.0 looks like?"
	app.Version = fmt.Sprintf(
		"%s -- commit %s",
		version.Version(),
		version.Commit(),
	)
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
					Name:  "cancel",
					Usage: "Cancel event(s) without deleting them",
					Description: "By default, only cancels event(s) with their worker " +
						"in a PENDING state.",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    flagID,
							Aliases: []string{"i"},
							Usage: "Cancel the specified event; mutually exclusive with " +
								"--project",
						},
						&cli.BoolFlag{
							Name:    flagRunning,
							Aliases: []string{"r"},
							Usage: "If set, will also abort event(s) with their worker in a" +
								"RUNNING state",
						},
						&cli.StringFlag{
							Name:    flagProject,
							Aliases: []string{"p"},
							Usage: "Cancel events for the specified project; mutually " +
								"exclusive with --id",
						},
					},
					Action: eventCancel,
				},
				{
					Name:  "create",
					Usage: "Create a new event",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:  flagPayload,
							Usage: "The event payload",
						},
						&cli.StringFlag{
							Name:  flagPayloadFile,
							Usage: "The location of a file containing the event payload",
						},
						&cli.StringFlag{
							Name:     flagProject,
							Aliases:  []string{"p"},
							Usage:    "Create an event for the specified project (required)",
							Required: true,
						},
						&cli.StringFlag{
							Name:    flagSource,
							Aliases: []string{"s"},
							Usage:   "Override the default event source",
							Value:   "github.com/krancour/brignext/cli",
						},
						&cli.StringFlag{
							Name:    flagType,
							Aliases: []string{"t"},
							Usage:   "Override the default event type",
							Value:   "exec",
						},
					},
					Action: eventCreate,
				},
				{
					Name:  "delete",
					Usage: "Delete event(s)",
					Description: "By default, only deletes event(s) with their worker " +
						"in a terminal state (neither PENDING nor RUNNING).",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    flagID,
							Aliases: []string{"i"},
							Usage: "Delete the specified event; mutually exclusive with " +
								" --project",
						},
						&cli.BoolFlag{
							Name: flagPending,
							Usage: "If set, will also delete event(s) with their worker " +
								"in a PENDING state",
						},
						&cli.BoolFlag{
							Name: flagRunning,
							Usage: "If set, will also abort and delete event(s) with their " +
								"worker in a RUNNING state",
						},
						&cli.StringFlag{
							Name:    flagProject,
							Aliases: []string{"p"},
							Usage: "Delete events for the specified project; mutually " +
								"exclusive with --id",
						},
					},
					Action: eventDelete,
				},
				{
					Name:  "get",
					Usage: "Retrieve an event",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     flagID,
							Aliases:  []string{"i"},
							Usage:    "Retrieve the specified event (required)",
							Required: true,
						},
						cliFlagOutput,
					},
					Action: eventGet,
				},
				{
					Name:  "list",
					Usage: "Retrieve many events",
					Flags: []cli.Flag{
						cliFlagOutput,
						&cli.StringFlag{
							Name:  flagProject,
							Usage: "Retrieve events only for the specified project",
						},
					},
					Action: eventList,
				},
			},
		},

		{
			Name:  "logs",
			Usage: "View worker or job logs",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    flagEvent,
					Aliases: []string{"e"},
					Usage:   "View logs from the specified event",
				},
				&cli.BoolFlag{
					Name:    flagFollow,
					Aliases: []string{"f"},
					Usage:   "If set, will stream logs until interrupted",
				},
				&cli.BoolFlag{
					Name:    flagInit,
					Aliases: []string{"i"},
					Usage:   "View logs from the corresponding init container",
				},
				&cli.StringFlag{
					Name:    flagJob,
					Aliases: []string{"j"},
					Usage: "View logs from the specified job; if not set, displays " +
						"worker logs",
				},
			},
			Action: logs,
		},
		{
			Name:  "login",
			Usage: "Log in to BrigNext",
			Description: "By default, initiates authentication using OpenID " +
				"Connect. This may not be supported by all BrigNext API servers.",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    flagServer,
					Aliases: []string{"s"},
					Usage: "Log into the API server at the specified address " +
						"(required)",
					Required: true,
				},
				&cli.BoolFlag{
					Name:    flagBrowse,
					Aliases: []string{"b"},
					Usage: "Use the system's default web browser to complete " +
						"authentication; not applicable when --root is used",
				},
				&cli.StringFlag{
					Name:    flagPassword,
					Aliases: []string{"p"},
					Usage: "Specify the password for non-interactive root user login; " +
						"only applicable when --root is used",
				},
				&cli.BoolFlag{
					Name:    flagRoot,
					Aliases: []string{"r"},
					Usage:   "Log in as the root user; does not use OpenID Connect",
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
					Name:  "create",
					Usage: "Create a new project",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    flagFile,
							Aliases: []string{"f"},
							Usage: "A YAML or JSON file that describes the project " +
								"(required)",
							Required:  true,
							TakesFile: true,
						},
					},
					Action: projectCreate,
				},
				{
					Name:  "delete",
					Usage: "Delete a project",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     flagID,
							Aliases:  []string{"i"},
							Usage:    "Delete the specified project (required)",
							Required: true,
						},
					},
					Action: projectDelete,
				},
				{
					Name:  "get",
					Usage: "Retrieve a project",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     flagID,
							Aliases:  []string{"i"},
							Usage:    "Retrieve the specified project (required)",
							Required: true,
						},
						cliFlagOutput,
					},
					Action: projectGet,
				},
				{
					Name:  "list",
					Usage: "Retrieve many projects",
					Flags: []cli.Flag{
						cliFlagOutput,
					},
					Action: projectList,
				},
				{
					Name:  "update",
					Usage: "Update a project",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    flagFile,
							Aliases: []string{"f"},
							Usage: "A YAML or JSON file that describes the project " +
								"(required)",
							Required:  true,
							TakesFile: true,
						},
					},
					Action: projectUpdate,
				},
			},
		},
		{
			Name:  "secrets",
			Usage: "Manage project secrets",
			Subcommands: []*cli.Command{
				{
					Name:  "list",
					Usage: "List a project's secrets; values are always redacted",
					Flags: []cli.Flag{
						cliFlagOutput,
					},
					Action: secretsList,
				},
				{
					Name:  "set",
					Usage: "Define or redefine the value of one or more secrets",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     flagProject,
							Aliases:  []string{"p"},
							Usage:    "Set secrets for the specified project (required)",
							Required: true,
						},
						&cli.StringSliceFlag{
							Name:    flagSet,
							Aliases: []string{"s"},
							Usage: "Set a secret using the specified key=value pair " +
								"(required)",
							Required: true,
						},
					},
					Action: secretsSet,
				},
				{
					Name:  "unset",
					Usage: "Clear the value of one or more secrets",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     flagProject,
							Aliases:  []string{"p"},
							Usage:    "Clear secrets for the specified project",
							Required: true,
						},
						&cli.StringSliceFlag{
							Name:     flagUnset,
							Aliases:  []string{"u"},
							Usage:    "Clear a secret haveing the specified key (required)",
							Required: true,
						},
					},
					Action: secretsUnset,
				},
			},
		},
		{
			Name:  "service-account",
			Usage: "Manage service accounts",
			Subcommands: []*cli.Command{
				{
					Name:  "create",
					Usage: "Create a new service account",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:    flagID,
							Aliases: []string{"i"},
							Usage: "Create a service account with the specified ID " +
								"(required)",
							Required: true,
						},
						&cli.StringFlag{
							Name:    flagDescription,
							Aliases: []string{"d"},
							Usage: "Create a service account with the specified " +
								"description (required)",
							Required: true,
						},
					},
					Action: serviceAccountCreate,
				},
				{
					Name:  "get",
					Usage: "Retrieve a service account",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     flagID,
							Aliases:  []string{"i"},
							Usage:    "Retrieve the specified service account (required)",
							Required: true,
						},
						cliFlagOutput,
					},
					Action: serviceAccountGet,
				},
				{
					Name:  "list",
					Usage: "Retrieve many service accounts",
					Flags: []cli.Flag{
						cliFlagOutput,
					},
					Action: serviceAccountList,
				},
				{
					Name:  "lock",
					Usage: "Lock a service account out of BrigNext",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     flagID,
							Aliases:  []string{"i"},
							Usage:    "Lock the specified service account (required)",
							Required: true,
						},
					},
					Action: serviceAccountLock,
				},
				{
					Name:  "unlock",
					Usage: "Restore a service account's access to BrigNext",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     flagID,
							Aliases:  []string{"i"},
							Usage:    "Unlock the specified service account (required)",
							Required: true,
						},
					},
					Action: serviceAccountUnlock,
				},
			},
		},
		{
			Name:  "user",
			Usage: "Manage users",
			Subcommands: []*cli.Command{
				{
					Name:  "get",
					Usage: "Retrieve a user",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     flagID,
							Aliases:  []string{"i"},
							Usage:    "Retrieve the specified user (required)",
							Required: true,
						},
						cliFlagOutput,
					},
					Action: userGet,
				},
				{
					Name:  "list",
					Usage: "Retrieve many users",
					Flags: []cli.Flag{
						cliFlagOutput,
					},
					Action: userList,
				},
				{
					Name:  "lock",
					Usage: "Lock a user out of BrigNext",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     flagID,
							Aliases:  []string{"i"},
							Usage:    "Lock the specified user (required)",
							Required: true,
						},
					},
					Action: userLock,
				},
				{
					Name:  "unlock",
					Usage: "Restore a user's access to BrigNext",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     flagID,
							Aliases:  []string{"i"},
							Usage:    "Unlock the specified user (required)",
							Required: true,
						},
					},
					Action: userUnlock,
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
