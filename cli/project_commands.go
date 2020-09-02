package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/krancour/brignext/v2/sdk"
	"github.com/krancour/brignext/v2/sdk/api"
	"github.com/krancour/brignext/v2/sdk/meta"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
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
			Usage: "Delete a single project",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i"},
					Usage:    "Delete the specified project (required)",
					Required: true,
				},
				&cli.BoolFlag{
					Name:    flagYes,
					Aliases: []string{"y"},
					Usage:   "Non-interactively confirm deletion",
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
			Name:  "grant",
			Usage: "Grant a project-scoped role to a user or service account",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i"},
					Usage:    "Grant the role for the specified project (required)",
					Required: true,
				},
				&cli.StringFlag{
					Name:     flagRole,
					Aliases:  []string{"r"},
					Usage:    "Grant the specified role (required)",
					Required: true,
				},
				&cli.StringFlag{
					Name:    flagServiceAccount,
					Aliases: []string{"s"},
					Usage:   "Grant the role to the specified service account",
				},
				&cli.StringFlag{
					Name:    flagUser,
					Aliases: []string{"u"},
					Usage:   "Grant the role to the specified user",
				},
			},
			Action: projectGrant,
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
			Name:  "revoke",
			Usage: "Revoke a project-scoped role from a user or service account",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i"},
					Usage:    "Revoke the role for the specified project (required)",
					Required: true,
				},
				&cli.StringFlag{
					Name:     flagRole,
					Aliases:  []string{"r"},
					Usage:    "Revoke the specified role (required)",
					Required: true,
				},
				&cli.StringFlag{
					Name:    flagServiceAccount,
					Aliases: []string{"s"},
					Usage:   "Revoke the role from the specified service account",
				},
				&cli.StringFlag{
					Name:    flagUser,
					Aliases: []string{"i"},
					Usage:   "Revoke the role from the specified user",
				},
			},
			Action: projectRevoke,
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

	if strings.HasSuffix(filename, ".yaml") ||
		strings.HasSuffix(filename, ".yml") {
		if projectBytes, err = yaml.YAMLToJSON(projectBytes); err != nil {
			return errors.Wrapf(err, "error converting file %s to JSON", filename)
		}
	}

	// We unmarshal just so that we can get the project ID. Otherwise, we wouldn't
	// need to do this, because we pass raw JSON to the API so that server-side
	// JSON schema validation is applied to what's in the file and NOT to a
	// project description that was inadvertently scrubbed of non-permitted fields
	// during client-side unmarshaling.
	project := sdk.Project{}
	if err = json.Unmarshal(projectBytes, &project); err != nil {
		return errors.Wrapf(err, "error unmarshaling project file %s", filename)
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if _, err :=
		client.Projects().CreateFromBytes(c.Context, projectBytes); err != nil {
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

	opts := meta.ListOptions{}

	for {
		projects, err :=
			client.Projects().List(c.Context, api.ProjectsSelector{}, opts)
		if err != nil {
			return err
		}

		if len(projects.Items) == 0 {
			fmt.Println("No projects found.")
			return nil
		}

		switch strings.ToLower(output) {
		case "table":
			table := uitable.New()
			table.AddRow("ID", "DESCRIPTION", "AGE")
			for _, project := range projects.Items {
				table.AddRow(
					project.ID,
					project.Description,
					duration.ShortHumanDuration(time.Since(*project.Created)),
				)
			}
			fmt.Println(table)

		case "yaml":
			yamlBytes, err := yaml.Marshal(projects)
			if err != nil {
				return errors.Wrap(
					err,
					"error formatting output from get projects operation",
				)
			}
			fmt.Println(string(yamlBytes))

		case "json":
			prettyJSON, err := json.MarshalIndent(projects, "", "  ")
			if err != nil {
				return errors.Wrap(
					err,
					"error formatting output from get projects operation",
				)
			}
			fmt.Println(string(prettyJSON))
		}

		if projects.RemainingItemCount < 1 || projects.Continue == "" {
			break
		}

		// Exit after one page of output if this isn't a terminal
		if !terminal.IsTerminal(int(os.Stdout.Fd())) {
			break
		}

		// TODO: DRY this up
		var shouldContinue bool
		fmt.Println()
		if err := survey.AskOne(
			&survey.Confirm{
				Message: fmt.Sprintf(
					"%d results remain. Fetch more?",
					projects.RemainingItemCount,
				),
			},
			&shouldContinue,
		); err != nil {
			return errors.Wrap(
				err,
				"error confirming if user wishes to continue",
			)
		}
		fmt.Println()
		if !shouldContinue {
			break
		}

		opts.Continue = projects.Continue
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
			project.Description,
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

	if strings.HasSuffix(filename, ".yaml") ||
		strings.HasSuffix(filename, ".yml") {
		if projectBytes, err = yaml.YAMLToJSON(projectBytes); err != nil {
			return errors.Wrapf(err, "error converting file %s to JSON", filename)
		}
	}

	// We unmarshal just so that we can get the project ID. Otherwise, we wouldn't
	// need to do this, because we pass raw JSON to the API so that server-side
	// JSON schema validation is applied to what's in the file and NOT to a
	// project description that was inadvertently scrubbed of non-permitted fields
	// during client-side unmarshaling.
	project := sdk.Project{}
	if err = json.Unmarshal(projectBytes, &project); err != nil {
		return errors.Wrapf(err, "error unmarshaling project file %s", filename)
	}

	// If the project ID is missing, we can go no further. All other validation
	// occurs server-side, but without an ID, we cannot even construct the URL
	// that we need to PUT to.
	if project.ID == "" {
		return errors.New("project definition does not specify an ID")
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if _, err = client.Projects().UpdateFromBytes(
		c.Context,
		project.ID,
		projectBytes,
	); err != nil {
		return err
	}

	fmt.Printf("Updated project %q.\n", project.ID)

	return nil
}

func projectDelete(c *cli.Context) error {
	id := c.String(flagID)

	confirmed, err := confirmed(c)
	if err != nil {
		return err
	}
	if !confirmed {
		return nil
	}

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

func projectGrant(c *cli.Context) error {
	id := c.String(flagID)
	role := c.String(flagRole)
	userID := c.String(flagUser)
	serviceAccountID := c.String(flagServiceAccount)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.Projects().GrantRole(
		c.Context,
		id,
		api.RoleAssignment{
			Role:             role,
			UserID:           userID,
			ServiceAccountID: serviceAccountID,
		},
	); err != nil {
		return err
	}

	if userID != "" {
		fmt.Printf(
			"Granted role %q for project %q to user %q.\n\n",
			role,
			id,
			userID,
		)
	} else {
		fmt.Printf(
			"Granted role %q for project %q to service account %q.\n\n",
			role,
			id,
			serviceAccountID,
		)
	}

	return nil
}

func projectRevoke(c *cli.Context) error {
	id := c.String(flagID)
	role := c.String(flagRole)
	userID := c.String(flagUser)
	serviceAccountID := c.String(flagServiceAccount)

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if err := client.Projects().RevokeRole(
		c.Context,
		id,
		api.RoleAssignment{
			Role:             role,
			UserID:           userID,
			ServiceAccountID: serviceAccountID,
		},
	); err != nil {
		return err
	}

	if userID != "" {
		fmt.Printf(
			"Revoked role %q for project %q from user %q.\n\n",
			role,
			id,
			userID,
		)
	} else {
		fmt.Printf(
			"Revoked role %q for project %q from service account %q.\n\n",
			role,
			id,
			serviceAccountID,
		)
	}

	return nil
}
