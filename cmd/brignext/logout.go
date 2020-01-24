package main

import (
	"context"
	"fmt"

	"github.com/krancour/brignext/pkg/users"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func logout(c *cli.Context) error {
	// Connect to the API server
	conn, err := getConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := users.NewUsersClient(conn)

	if _, err := client.Logout(
		context.Background(),
		&users.LogoutRequest{},
	); err != nil {
		return errors.Wrap(err, "error logging out of the API server")
	}

	if err := deleteConfig(); err != nil {
		return errors.Wrap(err, "error deleting configuration")
	}

	fmt.Println("Logout was successful.")

	return nil
}
