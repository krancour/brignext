package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/krancour/brignext"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func serviceAccountCreate(c *cli.Context) error {
	// Args
	var id string
	if len(c.Args()) == 1 {
		id = c.Args()[0]
	} else if len(c.Args()) != 0 {
		return errors.New(
			"service-account create requires, at most, one argument-- the new " +
				"service account ID",
		)
	}

	// Command-specific flags
	description := c.String(flagDescription)

	reader := bufio.NewReader(os.Stdin)

	for {
		id = strings.TrimSpace(id)
		if id != "" {
			break
		}
		fmt.Print("Service account ID? ")
		var err error
		if id, err = reader.ReadString('\n'); err != nil {
			return errors.Wrap(err, "error reading service account ID from stdin")
		}
	}

	for {
		description = strings.TrimSpace(description)
		if description != "" {
			break
		}
		fmt.Print("Service account description? ")
		var err error
		if description, err = reader.ReadString('\n'); err != nil {
			return errors.Wrap(
				err,
				"error reading service account description from stdin",
			)
		}
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	token, err := client.CreateServiceAccount(
		context.TODO(),
		brignext.ServiceAccount{
			ID:          id,
			Description: description,
		},
	)
	if err != nil {
		return err
	}

	fmt.Printf("\nService account %q created with token:\n", id)
	fmt.Printf("\n\t%s\n", token)
	fmt.Println(
		"\nStore this token someplace secure NOW. It cannot be retrieved " +
			"later through any other means.",
	)

	return nil
}
