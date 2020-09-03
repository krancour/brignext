package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var systemRolesCommand = &cli.Command{
	Name:  "roles",
	Usage: "Manage system roles",
	Subcommands: []*cli.Command{
		{
			Name:  "grant",
			Usage: "Grant a system role to a user or service account",
			Flags: []cli.Flag{
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
			Action: systemRolesGrant,
		},
		{
			Name:  "revoke",
			Usage: "Revoke a system role from a user or service account",
			Flags: []cli.Flag{
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
			Action: systemRolesRevoke,
		},
	},
}

func systemRolesGrant(c *cli.Context) error {
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
		return errors.Wrap(err, "error getting brignext client")
	}

	if userID != "" {
		if err := client.System().Roles().GrantToUser(
			c.Context,
			userID,
			role,
		); err != nil {
			return err
		}

		fmt.Printf(
			"Granted system role %q to user %q.\n\n",
			role,
			userID,
		)

		return nil
	}

	if err := client.System().Roles().GrantToServiceAccount(
		c.Context,
		serviceAccountID,
		role,
	); err != nil {
		return err
	}

	fmt.Printf(
		"Granted system role %q to service account %q.\n\n",
		role,
		serviceAccountID,
	)

	return nil
}

func systemRolesRevoke(c *cli.Context) error {
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
		return errors.Wrap(err, "error getting brignext client")
	}

	if userID != "" {
		if err := client.System().Roles().RevokeFromUser(
			c.Context,
			userID,
			role,
		); err != nil {
			return err
		}

		fmt.Printf(
			"Revoked system role %q from user %q.\n\n",
			role,
			userID,
		)

		return nil
	}

	if err := client.System().Roles().RevokeFromServiceAccount(
		c.Context,
		serviceAccountID,
		role,
	); err != nil {
		return err
	}

	fmt.Printf(
		"Revoked system role %q for from service account %q.\n\n",
		role,
		serviceAccountID,
	)

	return nil
}
