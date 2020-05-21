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
	"github.com/krancour/brignext/v2/internal/common/file"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"k8s.io/apimachinery/pkg/util/duration"
)

var eventCommand = &cli.Command{
	Name:  "event",
	Usage: "Manage events",
	Subcommands: []*cli.Command{
		{
			Name:  "cancel",
			Usage: "Cancel event(s) without deleting them",
			Description: "By default, only cancels event(s) with their worker " +
				"in a PENDING state.",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    flagID,
					Aliases: []string{"i"},
					Usage: "Cancel the specified event; mutually exclusive with " +
						"--project",
				},
				&cli.BoolFlag{
					Name:    flagRunning,
					Aliases: []string{"r"},
					Usage: "If set, will also abort event(s) with their worker in a" +
						"RUNNING state",
				},
				&cli.StringFlag{
					Name:    flagProject,
					Aliases: []string{"p"},
					Usage: "Cancel events for the specified project; mutually " +
						"exclusive with --id",
				},
			},
			Action: eventCancel,
		},
		{
			Name:  "create",
			Usage: "Create a new event",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  flagPayload,
					Usage: "The event payload",
				},
				&cli.StringFlag{
					Name:  flagPayloadFile,
					Usage: "The location of a file containing the event payload",
				},
				&cli.StringFlag{
					Name:     flagProject,
					Aliases:  []string{"p"},
					Usage:    "Create an event for the specified project (required)",
					Required: true,
				},
				&cli.StringFlag{
					Name:    flagSource,
					Aliases: []string{"s"},
					Usage:   "Override the default event source",
					Value:   "github.com/krancour/brignext/cli",
				},
				&cli.StringFlag{
					Name:    flagType,
					Aliases: []string{"t"},
					Usage:   "Override the default event type",
					Value:   "exec",
				},
			},
			Action: eventCreate,
		},
		{
			Name:  "delete",
			Usage: "Delete event(s)",
			Description: "By default, only deletes event(s) with their worker " +
				"in a terminal state (neither PENDING nor RUNNING).",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    flagID,
					Aliases: []string{"i"},
					Usage: "Delete the specified event; mutually exclusive with " +
						" --project",
				},
				&cli.BoolFlag{
					Name: flagPending,
					Usage: "If set, will also delete event(s) with their worker " +
						"in a PENDING state",
				},
				&cli.BoolFlag{
					Name: flagRunning,
					Usage: "If set, will also abort and delete event(s) with their " +
						"worker in a RUNNING state",
				},
				&cli.StringFlag{
					Name:    flagProject,
					Aliases: []string{"p"},
					Usage: "Delete events for the specified project; mutually " +
						"exclusive with --id",
				},
			},
			Action: eventDelete,
		},
		{
			Name:  "get",
			Usage: "Retrieve an event",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i"},
					Usage:    "Retrieve the specified event (required)",
					Required: true,
				},
				cliFlagOutput,
			},
			Action: eventGet,
		},
		{
			Name:  "list",
			Usage: "Retrieve many events",
			Flags: []cli.Flag{
				cliFlagOutput,
				&cli.StringFlag{
					Name:    flagProject,
					Aliases: []string{"p"},
					Usage:   "Retrieve events only for the specified project",
				},
			},
			Action: eventList,
		},
	},
}

func eventCreate(c *cli.Context) error {
	payload := c.String(flagPayload)
	payloadFile := c.String(flagPayloadFile)
	projectID := c.String(flagProject)
	source := c.String(flagSource)
	eventType := c.String(flagType)

	if payload != "" && payloadFile != "" {
		return errors.New(
			"only one of --payload or --payload-file may be specified",
		)
	}
	if payloadFile != "" {
		if !file.Exists(payloadFile) {
			return errors.Errorf("no event payload was found at %s", payloadFile)
		}
		payloadBytes, err := ioutil.ReadFile(payloadFile)
		if err != nil {
			return errors.Wrapf(
				err,
				"error reading event payload from %s",
				payloadFile,
			)
		}
		payload = string(payloadBytes)
	}

	event := brignext.NewEvent()
	event.ProjectID = projectID
	event.Source = source
	event.Type = eventType
	event.Payload = payload

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	eventRefList, err := client.Events().Create(c.Context, event)
	if err != nil {
		return err
	}
	fmt.Printf("Created event %q.\n\n", eventRefList.Items[0].ID)

	return nil
}

func eventList(c *cli.Context) error {
	output := c.String(flagOutput)
	projectID := c.String(flagProject)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	var eventList brignext.EventList
	if projectID == "" {
		eventList, err = client.Events().List(c.Context)
	} else {
		eventList, err = client.Events().ListByProject(c.Context, projectID)
	}
	if err != nil {
		return err
	}

	if len(eventList.Items) == 0 {
		fmt.Println("No events found.")
		return nil
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "PROJECT", "SOURCE", "TYPE", "AGE", "WORKER PHASE")
		for _, event := range eventList.Items {
			var age string
			if event.Created != nil {
				age = duration.ShortHumanDuration(time.Since(*event.Created))
			}
			table.AddRow(
				event.ID,
				event.ProjectID,
				event.Source,
				event.Type,
				age,
				event.Status.WorkerStatus.Phase,
			)
		}
		fmt.Println(table)

	case "yaml":
		yamlBytes, err := yaml.Marshal(eventList)
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get events operation",
			)
		}
		fmt.Println(string(yamlBytes))

	case "json":
		prettyJSON, err := json.MarshalIndent(eventList, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get events operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}

func eventGet(c *cli.Context) error {
	id := c.String(flagID)
	output := c.String(flagOutput)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	event, err := client.Events().Get(c.Context, id)
	if err != nil {
		return err
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "PROJECT", "SOURCE", "TYPE", "AGE", "WORKER PHASE")
		var age string
		if event.Created != nil {
			age = duration.ShortHumanDuration(time.Since(*event.Created))
		}
		table.AddRow(
			event.ID,
			event.ProjectID,
			event.Source,
			event.Type,
			age,
			event.Status.WorkerStatus.Phase,
		)
		fmt.Println(table)

		if len(event.Status.JobStatuses) > 0 {
			fmt.Printf("\nEvent %q worker jobs:\n\n", event.ID)
			table = uitable.New()
			table.AddRow("NAME", "STARTED", "ENDED", "PHASE")
			for jobName, jobStatus := range event.Status.JobStatuses {
				var started, ended string
				if jobStatus.Started != nil {
					started =
						duration.ShortHumanDuration(time.Since(*jobStatus.Started))
				}
				if jobStatus.Ended != nil {
					ended =
						duration.ShortHumanDuration(time.Since(*jobStatus.Ended))
				}
				table.AddRow(
					jobName,
					started,
					ended,
					jobStatus.Phase,
				)
			}
			fmt.Println(table)
		}

	case "yaml":
		yamlBytes, err := yaml.Marshal(event)
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get event operation",
			)
		}
		fmt.Println(string(yamlBytes))

	case "json":
		prettyJSON, err := json.MarshalIndent(event, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get event operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}

func eventCancel(c *cli.Context) error {
	eventID := c.String(flagID)
	projectID := c.String(flagProject)
	cancelRunning := c.Bool(flagRunning)

	if eventID == "" && projectID == "" {
		return errors.New("one of --id or --project must be set")
	}

	if eventID != "" && projectID != "" {
		return errors.New("--id and --project are mutually exclusive")
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if eventID != "" {
		var eventRefList brignext.EventReferenceList
		if eventRefList, err = client.Events().Cancel(
			c.Context,
			eventID,
			cancelRunning,
		); err != nil {
			return err
		}
		if len(eventRefList.Items) != 0 {
			fmt.Printf("Event %q canceled.\n", eventID)
			return nil
		}
		return errors.Errorf(
			"event %q was not canceled because specified conditions were not "+
				"satisfied",
			eventID,
		)
	}

	eventRefList, err := client.Events().CancelByProject(
		c.Context,
		projectID,
		cancelRunning,
	)
	if err != nil {
		return err
	}
	fmt.Printf(
		"Canceled %d events for project %q.\n",
		len(eventRefList.Items),
		projectID,
	)

	return nil
}

func eventDelete(c *cli.Context) error {
	eventID := c.String(flagID)
	deletePending := c.Bool(flagPending)
	projectID := c.String(flagProject)
	deleteRunning := c.Bool(flagRunning)

	if eventID == "" && projectID == "" {
		return errors.New("one of --id or --project must be set")
	}

	if eventID != "" && projectID != "" {
		return errors.New("--id and --project are mutually exclusive")
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if eventID != "" {
		var eventRefList brignext.EventReferenceList
		if eventRefList, err = client.Events().Delete(
			c.Context,
			eventID,
			deletePending,
			deleteRunning,
		); err != nil {
			return err
		}
		if len(eventRefList.Items) != 0 {
			fmt.Printf("Event %q deleted.\n", eventID)
			return nil
		}
		return errors.Errorf(
			"event %q was not deleted because specified conditions were not "+
				"satisfied",
			eventID,
		)
	}

	eventRefList, err := client.Events().DeleteByProject(
		c.Context,
		projectID,
		deletePending,
		deleteRunning,
	)
	if err != nil {
		return err
	}
	fmt.Printf(
		"Deleted %d events for project %q.\n",
		len(eventRefList.Items),
		projectID,
	)

	return nil
}
