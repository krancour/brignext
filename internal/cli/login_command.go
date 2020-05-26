package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var loginCommand = &cli.Command{
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
}

func login(c *cli.Context) error {
	address := c.String(flagServer)
	browseToAuthURL := c.Bool(flagBrowse)
	password := c.String(flagPassword)
	rootLogin := c.Bool(flagRoot)

	client := brignext.NewClient(
		address,
		"",
		c.Bool(flagInsecure),
	)

	var tokenStr, authURL string

	var err error
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

		var token brignext.Token
		if token, err =
			client.Sessions().CreateRootSession(c.Context, password); err != nil {
			return err
		}
		tokenStr = token.Value
	} else if authURL, tokenStr, err =
		client.Sessions().CreateUserSession(c.Context); err != nil {
		return err
	}

	if err := saveConfig(
		&config{
			APIAddress: address,
			APIToken:   tokenStr,
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