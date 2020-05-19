package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func projectUpdate(c *cli.Context) error {
	filename := c.String(flagFile)

	// Read and parse the file
	projectBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrapf(err, "error reading project file %s", filename)
	}

	project := brignext.Project{}
	if strings.HasSuffix(filename, ".yaml") ||
		strings.HasSuffix(filename, ".yml") {
		err = yaml.Unmarshal(projectBytes, &project)
	} else {
		err = json.Unmarshal(projectBytes, &project)
	}
	if err != nil {
		return errors.Wrapf(err, "error unmarshaling project file %s", filename)
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.Projects().Update(c.Context, project); err != nil {
		return err
	}

	fmt.Printf("Updated project %q.\n", project.ID)

	return nil
}
