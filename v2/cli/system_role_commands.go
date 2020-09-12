package main

import (
	"fmt"

	"github.com/brigadecore/brigade/v2/sdk/authx"
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
		return err
	}

	roleAssignment := authx.RoleAssignment{
		Role: authx.RoleName(role),
	}

	var readablePrincipalType string
	if userID != "" {
		readablePrincipalType = "user"
		roleAssignment.PrincipalType = authx.PrincipalTypeUser
		roleAssignment.PrincipalID = userID
	} else {
		readablePrincipalType = "service account"
		roleAssignment.PrincipalType = authx.PrincipalTypeServiceAccount
		roleAssignment.PrincipalID = serviceAccountID
	}

	if err :=
		client.System().Roles().Grant(c.Context, roleAssignment); err != nil {
		return err
	}

	fmt.Printf(
		"Granted system role %q to %s %q.\n\n",
		role,
		readablePrincipalType,
		userID,
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
		return err
	}

	roleAssignment := authx.RoleAssignment{
		Role: authx.RoleName(role),
	}

	var readablePrincipalType string
	if userID != "" {
		readablePrincipalType = "user"
		roleAssignment.PrincipalType = authx.PrincipalTypeUser
		roleAssignment.PrincipalID = userID
	} else {
		readablePrincipalType = "service account"
		roleAssignment.PrincipalType = authx.PrincipalTypeServiceAccount
		roleAssignment.PrincipalID = serviceAccountID
	}

	if err :=
		client.System().Roles().Revoke(c.Context, roleAssignment); err != nil {
		return err
	}

	fmt.Printf(
		"Revoked system role %q from %s %q.\n\n",
		role,
		readablePrincipalType,
		userID,
	)

	return nil
}
