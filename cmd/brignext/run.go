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
	commit := c.String(flagCommit)
	configFile := c.String(flagConfig)
	event := c.String(flagEvent)
	scriptFile := c.String(flagFile)
	level := c.String(flagLevel)
	payloadFile := c.String(flagPayload)
	ref := c.String(flagRef)

	var configBytes []byte
	if configFile != "" {
		var err error
		configBytes, err = ioutil.ReadFile(configFile)
		if err != nil {
			return errors.Wrapf(err, "error reading config file %s", configFile)
		}
	}

	var scriptBytes []byte
	if scriptFile != "" {
		var err error
		scriptBytes, err = ioutil.ReadFile(scriptFile)
		if err != nil {
			return errors.Wrapf(err, "error reading script file %s", scriptFile)
		}
	}

	var payloadBytes []byte
	if payloadFile != "" {
		var err error
		payloadBytes, err = ioutil.ReadFile(payloadFile)
		if err != nil {
			return errors.Wrapf(err, "error reading payload file %s", payloadFile)
		}
	}

	build := &brignext.Build{
		ProjectName: projectName,
		Type:        event,
		Provider:    "brigade-cli",
		Revision: &brignext.Revision{
			Commit: commit,
			Ref:    ref,
		},
		Payload:  payloadBytes,
		Script:   scriptBytes,
		Config:   configBytes,
		LogLevel: level,
	}

	buildBytes, err := json.Marshal(build)
	if err != nil {
		return errors.Wrap(err, "error marshaling build")
	}

	req, err := getRequest(http.MethodPost, "v2/builds", buildBytes)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "error reading response body")
	}

	build = &brignext.Build{}
	if err := json.Unmarshal(respBodyBytes, build); err != nil {
		return errors.Wrap(err, "error unmarshaling response body")
	}

	// Pretty print the response
	buildJSON, err := json.MarshalIndent(build, "", "  ")
	if err != nil {
		return errors.Wrap(
			err,
			"error marshaling output from project creation operation",
		)
	}
	fmt.Println(string(buildJSON))

	// Now stream the logs

	if req, err = getRequest(
		http.MethodGet,
		fmt.Sprintf("v2/builds/%s/logs", build.ID),
		nil,
	); err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	logsResp, err := client.Do(req)
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
