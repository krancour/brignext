package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/krancour/brignext"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func projectCreate(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"project create requires one argument-- a path to a file containing a " +
				"project definition",
		)
	}
	filename := c.Args()[0]

	// Read and parse the file
	projectBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrapf(err, "error reading project file %s", filename)
	}

	project := brignext.Project{}
	if err := json.Unmarshal(projectBytes, &project); err != nil {
		return errors.Wrapf(err, "error unmarshaling project file %s", filename)
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.CreateProject(context.TODO(), project); err != nil {
		return err
	}

	fmt.Printf("Created project %q.\n", project.ID)

	return nil
}
