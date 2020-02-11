package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func eventCreate(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"event create requires one argument-- a project ID (case insensitive) " +
				"for for which an event should be created",
		)
	}
	projectID := c.Args()[0]

	// Global flags
	allowInsecure := c.GlobalBool(flagInsecure)

	// Command-specific flags
	eventType := c.String(flagType)
	provider := c.String(flagProvider)

	event := brignext.Event{
		ProjectID: projectID,
		Provider:  provider,
		Type:      eventType,
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return errors.Wrap(err, "error marshaling event")
	}

	req, err := buildRequest(http.MethodPost, "v2/events", eventBytes)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := getHTTPClient(allowInsecure).Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "error reading response body")
	}

	respStruct := struct {
		ID string `json:"id"`
	}{}
	if err := json.Unmarshal(respBodyBytes, &respStruct); err != nil {
		return errors.Wrap(err, "error unmarshaling response body")
	}
	eventID := respStruct.ID

	fmt.Printf("Created event %q.\n\n", eventID)

	fmt.Println("Streaming event logs...\n")

	// Now stream the logs

	if req, err = buildRequest(
		http.MethodGet,
		fmt.Sprintf("v2/events/%s/logs", eventID),
		nil,
	); err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	logsResp, err := getHTTPClient(allowInsecure).Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer logsResp.Body.Close()

	bufferedReader := bufio.NewReader(logsResp.Body)
	logsBuffer := make([]byte, 4*1024)
	for {
		len, err := bufferedReader.Read(logsBuffer)
		if err != nil {
			return errors.Wrap(err, "error streaming logs from API")
		}
		if len > 0 {
			log.Println(string(logsBuffer[:len]))
		}
	}
}
