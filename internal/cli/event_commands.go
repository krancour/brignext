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
			Usage: "Cancel events without deleting them",
			Description: "Use the --id flag to unconditionally cancel a single " +
				"event in a non-terminal state or use other flags to describe " +
				"criteria for conditionally canceling multiple events",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    flagID,
					Aliases: []string{"i"},
					Usage: "Cancel (and abort if applicable) the specified event only; " +
						"mutually exclusive with all other flags",
				},
				&cli.BoolFlag{
					Name:    flagRunning,
					Aliases: []string{"r"},
					Usage: "If set, will abort events with their worker in a RUNNING " +
						"state; mutually exclusive with --id",
				},
				&cli.StringFlag{
					Name:    flagProject,
					Aliases: []string{"p"},
					Usage: "Cancel events for the specified project only; mutually " +
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
			Usage: "Delete events",
			Description: "Use the --id flag to unconditionally delete a single " +
				"event or use other flags to describe criteria for conditionally " +
				"deleting multiple events",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name: flagAborted,
					Usage: "If set, will delete events with their worker in an ABORTED " +
						"state; mutually exclusive with --id, --any-state, and --terminal",
				},
				&cli.BoolFlag{
					Name: flagAnyState,
					Usage: "If set, will delete events with their worker in any state; " +
						"mutually exclusive with --id, --terminal, and all other state " +
						"flags",
				},
				&cli.BoolFlag{
					Name: flagCanceled,
					Usage: "If set, will delete events with their worker in a CANCELED " +
						"state; mutually exclusive with --id, --any-state, and --terminal",
				},
				&cli.BoolFlag{
					Name: flagFailed,
					Usage: "If set, will delete events with their worker in a FAILED " +
						"state; mutually exclusive with --id, --any-state, and --terminal",
				},
				&cli.StringFlag{
					Name:    flagID,
					Aliases: []string{"i"},
					Usage: "Delete the specified event only; mutually exclusive with " +
						"all other flags",
				},
				&cli.BoolFlag{
					Name: flagPending,
					Usage: "If set, will delete events with their worker in a PENDING " +
						"state; mutually exclusive with --id, --any-state, and --terminal",
				},
				&cli.StringFlag{
					Name:    flagProject,
					Aliases: []string{"p"},
					Usage: "Delete events for the specified project only; mutually " +
						"exclusive with --id",
				},
				&cli.BoolFlag{
					Name: flagRunning,
					Usage: "If set, will abort and delete events with their worker in " +
						"a RUNNING state; mutually exclusive with --id, --any-state, and " +
						"--terminal",
				},
				&cli.BoolFlag{
					Name: flagSucceeded,
					Usage: "If set, will delete events with their worker in a " +
						"SUCCEEDED state; mutually exclusive with --id, --any-state, " +
						"and --terminal",
				},
				&cli.BoolFlag{
					Name: flagTerminal,
					Usage: "If set, will delete events with their worker in any " +
						"terminal state; mutually exclusive with --id, --any-state, and " +
						"all other state flags",
				},
				&cli.BoolFlag{
					Name: flagTimedOut,
					Usage: "If set, will delete events with their worker in a " +
						"TIMED_OUT state; mutually exclusive with --id, --any-state, and " +
						"--terminal",
				},
				&cli.BoolFlag{
					Name: flagUnknown,
					Usage: "If set, will delete events with their worker in an UNKNOWN " +
						"state; mutually exclusive with --id, --any-state, and --terminal",
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
			fmt.Printf("\nEvent %q jobs:\n\n", event.ID)
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
	id := c.String(flagID)
	projectID := c.String(flagProject)
	cancelRunning := c.Bool(flagRunning)

	if id != "" && (projectID != "" || cancelRunning) {
		return errors.New("--id is mutually exclusive with all other flags")
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if id != "" {
		if err = client.Events().Cancel(c.Context, id); err != nil {
			return err
		}
		fmt.Printf("Event %q canceled.\n", id)
		return nil
	}

	opts := brignext.EventListOptions{
		ProjectID:    projectID,
		WorkerPhases: []brignext.WorkerPhase{brignext.WorkerPhasePending},
	}
	if cancelRunning {
		opts.WorkerPhases = append(opts.WorkerPhases, brignext.WorkerPhaseRunning)
	}

	eventRefList, err := client.Events().CancelCollection(c.Context, opts)
	if err != nil {
		return err
	}
	fmt.Printf("Canceled %d events.\n", len(eventRefList.Items))

	return nil
}

func eventDelete(c *cli.Context) error {
	id := c.String(flagID)
	projectID := c.String(flagProject)

	workerPhases := []brignext.WorkerPhase{}

	if c.Bool(flagAborted) {
		workerPhases = append(workerPhases, brignext.WorkerPhaseAborted)
	}
	if c.Bool(flagCanceled) {
		workerPhases = append(workerPhases, brignext.WorkerPhaseCanceled)
	}
	if c.Bool(flagFailed) {
		workerPhases = append(workerPhases, brignext.WorkerPhaseFailed)
	}
	if c.Bool(flagPending) {
		workerPhases = append(workerPhases, brignext.WorkerPhasePending)
	}
	if c.Bool(flagRunning) {
		workerPhases = append(workerPhases, brignext.WorkerPhaseRunning)
	}
	if c.Bool(flagSucceeded) {
		workerPhases = append(workerPhases, brignext.WorkerPhaseSucceeded)
	}
	if c.Bool(flagTimedOut) {
		workerPhases = append(workerPhases, brignext.WorkerPhaseTimedOut)
	}
	if c.Bool(flagUnknown) {
		workerPhases = append(workerPhases, brignext.WorkerPhaseUnknown)
	}

	if c.Bool(flagAnyState) {
		if len(workerPhases) > 0 {
			return errors.New(
				"--any-state is mutually exclusive with all other state flags",
			)
		}
		workerPhases = brignext.WorkerPhasesAll()
	}

	if c.Bool(flagTerminal) {
		if len(workerPhases) > 0 {
			return errors.New(
				"--terminal is mutually exclusive with all other state flags",
			)
		}
		workerPhases = brignext.WorkerPhasesTerminal()
	}

	if id != "" && (projectID != "" || len(workerPhases) > 0) {
		return errors.New("--id is mutually exclusive with all other flags")
	}

	client, err := getClient(c)
	if err != nil {
		return errors.Wrap(err, "error getting brignext client")
	}

	if id != "" {
		if err = client.Events().Delete(c.Context, id); err != nil {
			return err
		}
		fmt.Printf("Event %q deleted.\n", id)
		return nil
	}

	opts := brignext.EventListOptions{
		ProjectID:    projectID,
		WorkerPhases: workerPhases,
	}

	eventRefList, err := client.Events().DeleteCollection(c.Context, opts)
	if err != nil {
		return err
	}
	fmt.Printf("Deleted %d events.\n", len(eventRefList.Items))

	return nil
}
