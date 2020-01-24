package main

import (
	"context"
	"fmt"

	"github.com/krancour/brignext/pkg/users"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func register(c *cli.Context) error {
	// Inputs
	host := c.Args()[0]
	port := c.Int(flagPort)
	username := c.String(flagUsername)
	password := c.String(flagPassword)

	// TODO: There should be an option to get username and password interactively
	// if not specified, otherwise username and password could show up in shell
	// history, which users may not want in some cases.

	// Connect to the API server
	conn, err := getRegistrationConnection(host, port)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := users.NewUsersClient(conn)

	if _, err = client.Register(
		context.Background(),
		&users.RegistrationRequest{
			Username: username,
			Password: password,
		},
	); err != nil {
		return errors.Wrap(err, "error registering with the API server")
	}

	fmt.Println("Registration was successful.")

	return doLogin(host, port, username, password)
}
