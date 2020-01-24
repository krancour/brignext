package main

import (
	"context"
	"fmt"

	"github.com/krancour/brignext/pkg/users"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func login(c *cli.Context) error {
	// Inputs
	host := c.Args()[0]
	port := c.Int(flagPort)
	username := c.String(flagUsername)
	password := c.String(flagPassword)

	// TODO: There should be an option to get username and password interactively
	// if not specified, otherwise username and password could show up in shell
	// history, which users may not want in some cases.

	// TODO: Log out of any API server we're already logged into

	return doLogin(host, port, username, password)
}

func doLogin(host string, port int, username, password string) error {
	// Connect to the API server
	conn, err := getLoginConnection(host, port, username, password)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := users.NewUsersClient(conn)

	resp, err := client.Login(
		context.Background(),
		&users.LoginRequest{},
	)
	if err != nil {
		return errors.Wrap(err, "error logging in to the API server")
	}

	if err := saveConfig(
		&config{
			APIHost:  host,
			APIPort:  port,
			APIToken: resp.Token,
		},
	); err != nil {
		return errors.Wrap(err, "error persisting configuration")
	}

	fmt.Println("Login was successful.")

	return nil
}
