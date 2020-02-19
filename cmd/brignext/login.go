package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func login(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"login requires one argument-- the address of the API server",
		)
	}
	address := c.Args()[0]

	// Command-specific flags
	browseToAuthURL := c.Bool(flagBrowse)
	password := c.String(flagPassword)
	rootLogin := c.Bool(flagRoot)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	var token, authURL string

	if rootLogin {
		reader := bufio.NewReader(os.Stdin)
		for {
			password = strings.TrimSpace(password)
			if password != "" {
				break
			}
			fmt.Print("Root user password? ")
			if password, err = reader.ReadString('\n'); err != nil {
				return errors.Wrap(err, "error reading password from stdin")
			}
		}

		if token, err = client.CreateRootSession(context.TODO(), password); err != nil {
			return err
		}
	} else if authURL, token, err = client.CreateUserSession(context.TODO()); err != nil {
		return err
	}

	if err := saveConfig(
		&config{
			APIAddress: address,
			APIToken:   token,
		},
	); err != nil {
		return errors.Wrap(err, "error persisting configuration")
	}

	if rootLogin {
		fmt.Println("\nYou are logged in as the root user.")
		return nil
	}

	if browseToAuthURL {
		var err error
		switch runtime.GOOS {
		case "linux":
			err = exec.Command("xdg-open", authURL).Start()
		case "windows":
			err = exec.Command(
				"rundll32",
				"url.dll,FileProtocolHandler",
				authURL,
			).Start()
		case "darwin":
			err = exec.Command("open", authURL).Start()
		default:
			err = errors.New("unsupported OS")
		}
		if err != nil {
			return errors.Wrapf(
				err,
				"Error opening authentication URL using the system's default web "+
					"browser.\n\nPlease visit  %s  to complete authentication.\n",
				authURL,
			)
		}
		return nil
	}

	fmt.Printf("Please visit  %s  to complete authentication.\n", authURL)

	return nil
}
