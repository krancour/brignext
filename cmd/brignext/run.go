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

func run(c *cli.Context) error {
	// Inputs
	projectName := c.Args()[0]
	// background := c.Bool(flagBackground)
	// commit := c.String(flagCommit)
	// configFile := c.String(flagConfig)
	eventType := c.String(flagType)
	// scriptFile := c.String(flagFile)
	// level := c.String(flagLevel)
	// payloadFile := c.String(flagPayload)
	// ref := c.String(flagRef)
	allowInsecure := c.GlobalBool(flagInsecure)

	// var configBytes []byte
	// if configFile != "" {
	// 	var err error
	// 	configBytes, err = ioutil.ReadFile(configFile)
	// 	if err != nil {
	// 		return errors.Wrapf(err, "error reading config file %s", configFile)
	// 	}
	// }

	// var scriptBytes []byte
	// if scriptFile != "" {
	// 	var err error
	// 	scriptBytes, err = ioutil.ReadFile(scriptFile)
	// 	if err != nil {
	// 		return errors.Wrapf(err, "error reading script file %s", scriptFile)
	// 	}
	// }

	// var payloadBytes []byte
	// if payloadFile != "" {
	// 	var err error
	// 	payloadBytes, err = ioutil.ReadFile(payloadFile)
	// 	if err != nil {
	// 		return errors.Wrapf(err, "error reading payload file %s", payloadFile)
	// 	}
	// }

	event := brignext.Event{
		ProjectName: projectName,
		Provider:    "brigade-cli",
		Type:        eventType,
		// Revision: &brignext.Revision{
		// 	Commit: commit,
		// 	Ref:    ref,
		// },
		// Payload:  payloadBytes,
		// Script:   scriptBytes,
		// Config:   configBytes,
		// LogLevel: level,
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
