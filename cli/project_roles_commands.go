package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var projectRolesCommand = &cli.Command{
	Name:  "roles",
	Usage: "Manage project roles",
	Subcommands: []*cli.Command{
		{
			Name:  "grant",
			Usage: "Grant a project role to a user or service account",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagProject,
					Aliases:  []string{"p"},
					Usage:    "Grant the role for the specified project (required)",
					Required: true,
				},
				&cli.StringFlag{
					Name:     flagRole,
					Aliases:  []string{"r"},
					Usage:    "Grant the specified role (required)",
					Required: true,
				},
				&cli.StringFlag{
					Name:    flagServiceAccount,
					Aliases: []string{"s"},
					Usage: "Grant the role to the specified service account; mutually " +
						"exclusive with --user",
				},
				&cli.StringFlag{
					Name:    flagUser,
					Aliases: []string{"u"},
					Usage: "Grant the role to the specified user; mutually exclusive " +
						"with --servcice-account",
				},
			},
			Action: projectRolesGrant,
		},
		{
			Name:  "revoke",
			Usage: "Revoke a project role from a user or service account",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagProject,
					Aliases:  []string{"p"},
					Usage:    "Revoke the role for the specified project (required)",
					Required: true,
				},
				&cli.StringFlag{
					Name:     flagRole,
					Aliases:  []string{"r"},
					Usage:    "Revoke the specified role (required)",
					Required: true,
				},
				&cli.StringFlag{
					Name:    flagServiceAccount,
					Aliases: []string{"s"},
					Usage: "Revoke the role from the specified service account; " +
						"mutually exclusive with --user",
				},
				&cli.StringFlag{
					Name:    flagUser,
					Aliases: []string{"u"},
					Usage: "Revoke the role from the specified user; mutually " +
						"exclusive with --service-account",
				},
			},
			Action: projectRolesRevoke,
		},
	},
}

func projectRolesGrant(c *cli.Context) error {
	projectID := c.String(flagProject)
	role := c.String(flagRole)
	userID := c.String(flagUser)
	serviceAccountID := c.String(flagServiceAccount)

	if userID == "" && serviceAccountID == "" {
		return errors.New(
			"one of --user or --service-account must be specified",
		)
	}
	if userID != "" && serviceAccountID != "" {
		return errors.New(
			"only one of --user or --service-account must be specified",
		)
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brigade client")
	}

	if userID != "" {
		if err := client.Core().Projects().Roles().GrantToUser(
			c.Context,
			projectID,
			userID,
			role,
		); err != nil {
			return err
		}

		fmt.Printf(
			"Granted role %q for project %q to user %q.\n\n",
			role,
			projectID,
			userID,
		)

		return nil
	}

	if err := client.Core().Projects().Roles().GrantToServiceAccount(
		c.Context,
		projectID,
		serviceAccountID,
		role,
	); err != nil {
		return err
	}

	fmt.Printf(
		"Granted role %q for project %q to service account %q.\n\n",
		role,
		projectID,
		serviceAccountID,
	)

	return nil
}

func projectRolesRevoke(c *cli.Context) error {
	projectID := c.String(flagProject)
	role := c.String(flagRole)
	userID := c.String(flagUser)
	serviceAccountID := c.String(flagServiceAccount)

	if userID == "" && serviceAccountID == "" {
		return errors.New(
			"one of --user or --service-account must be specified",
		)
	}
	if userID != "" && serviceAccountID != "" {
		return errors.New(
			"only one of --user or --service-account must be specified",
		)
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brigade client")
	}

	if userID != "" {
		if err := client.Core().Projects().Roles().RevokeFromUser(
			c.Context,
			projectID,
			userID,
			role,
		); err != nil {
			return err
		}

		fmt.Printf(
			"Revoked role %q for project %q from user %q.\n\n",
			role,
			projectID,
			userID,
		)

		return nil
	}

	if err := client.Core().Projects().Roles().RevokeFromServiceAccount(
		c.Context,
		projectID,
		serviceAccountID,
		role,
	); err != nil {
		return err
	}

	fmt.Printf(
		"Revoked role %q for project %q from service account %q.\n\n",
		role,
		projectID,
		serviceAccountID,
	)

	return nil
}
