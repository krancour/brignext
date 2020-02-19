package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func projectUpdate(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"project update requires one argument-- a path to a file containing a " +
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

	if err := client.UpdateProject(context.TODO(), project); err != nil {
		return err
	}

	fmt.Printf("Updated project %q.\n", project.ID)

	return nil
}
