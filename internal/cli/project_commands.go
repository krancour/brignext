package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"k8s.io/apimachinery/pkg/util/duration"
)

var projectCommand = &cli.Command{
	Name:  "project",
	Usage: "Manage projects",
	Subcommands: []*cli.Command{
		{
			Name:  "create",
			Usage: "Create a new project",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    flagFile,
					Aliases: []string{"f"},
					Usage: "A YAML or JSON file that describes the project " +
						"(required)",
					Required:  true,
					TakesFile: true,
				},
			},
			Action: projectCreate,
		},
		{
			Name:  "delete",
			Usage: "Delete a project",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i"},
					Usage:    "Delete the specified project (required)",
					Required: true,
				},
			},
			Action: projectDelete,
		},
		{
			Name:  "get",
			Usage: "Retrieve a project",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i"},
					Usage:    "Retrieve the specified project (required)",
					Required: true,
				},
				cliFlagOutput,
			},
			Action: projectGet,
		},
		{
			Name:  "list",
			Usage: "Retrieve many projects",
			Flags: []cli.Flag{
				cliFlagOutput,
			},
			Action: projectList,
		},
		{
			Name:  "update",
			Usage: "Update a project",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    flagFile,
					Aliases: []string{"f"},
					Usage: "A YAML or JSON file that describes the project " +
						"(required)",
					Required:  true,
					TakesFile: true,
				},
			},
			Action: projectUpdate,
		},
	},
}

func projectCreate(c *cli.Context) error {
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

	if err := client.Projects().Create(c.Context, project); err != nil {
		return err
	}

	fmt.Printf("Created project %q.\n", project.ID)

	return nil
}

func projectList(c *cli.Context) error {
	output := c.String(flagOutput)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	projectList, err := client.Projects().List(c.Context)
	if err != nil {
		return err
	}

	if len(projectList.Items) == 0 {
		fmt.Println("No projects found.")
		return nil
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "DESCRIPTION", "AGE")
		for _, project := range projectList.Items {
			var age string
			if project.Created != nil {
				age = duration.ShortHumanDuration(time.Since(*project.Created))
			}
			table.AddRow(
				project.ID,
				project.Spec.Description,
				age,
			)
		}
		fmt.Println(table)

	case "yaml":
		yamlBytes, err := yaml.Marshal(projectList)
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get projects operation",
			)
		}
		fmt.Println(string(yamlBytes))

	case "json":
		prettyJSON, err := json.MarshalIndent(projectList, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get projects operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}

func projectGet(c *cli.Context) error {
	id := c.String(flagID)
	output := c.String(flagOutput)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	project, err := client.Projects().Get(c.Context, id)
	if err != nil {
		return err
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "DESCRIPTION", "AGE")
		var age string
		if project.Created != nil {
			age = duration.ShortHumanDuration(time.Since(*project.Created))
		}
		table.AddRow(
			project.ID,
			project.Spec.Description,
			age,
		)
		fmt.Println(table)

	case "yaml":
		yamlBytes, err := yaml.Marshal(project)
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get project operation",
			)
		}
		fmt.Println(string(yamlBytes))

	case "json":
		prettyJSON, err := json.MarshalIndent(project, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get project operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}

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

func projectDelete(c *cli.Context) error {
	id := c.String(flagID)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.Projects().Delete(c.Context, id); err != nil {
		return err
	}

	fmt.Printf("Project %q deleted.\n", id)

	return nil
}
