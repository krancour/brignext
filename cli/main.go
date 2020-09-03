package main

import (
	"fmt"
	"os"

	"github.com/krancour/brignext/v2/internal/signals"
	"github.com/krancour/brignext/v2/internal/version"
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
		eventCommand,
		loginCommand,
		logoutCommand,
		projectCommand,
		serviceAccountCommand,
		systemCommand,
		userCommand,
	}
	fmt.Println()
	if err := app.RunContext(signals.Context(), os.Args); err != nil {
		fmt.Printf("\n%s\n\n", err)
		os.Exit(1)
	}
	fmt.Println()
}
