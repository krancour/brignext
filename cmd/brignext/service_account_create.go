package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/krancour/brignext/pkg/brignext"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func serviceAccountCreate(c *cli.Context) error {
	// Args
	var id string
	if len(c.Args()) == 1 {
		id = c.Args()[0]
	} else if len(c.Args()) != 0 {
		return errors.New(
			"service-account create requires, at most, one parameter-- the new " +
				"service account ID",
		)
	}

	// Global flags
	allowInsecure := c.GlobalBool(flagInsecure)

	// Command-specific flags
	description := c.String(flagDescription)

	reader := bufio.NewReader(os.Stdin)

	for {
		id = strings.TrimSpace(id)
		if id != "" {
			break
		}
		fmt.Print("Service account ID? ")
		var err error
		if id, err = reader.ReadString('\n'); err != nil {
			return errors.Wrap(err, "error reading service account ID from stdin")
		}
	}

	for {
		description = strings.TrimSpace(description)
		if description != "" {
			break
		}
		fmt.Print("Service account description? ")
		var err error
		if description, err = reader.ReadString('\n'); err != nil {
			return errors.Wrap(
				err,
				"error reading service account description from stdin",
			)
		}
	}

	serviceAccountBytes, err := json.Marshal(
		brignext.ServiceAccount{
			ID:          id,
			Description: description,
		},
	)
	if err != nil {
		return errors.Wrap(err, "error marshaling service account")
	}

	req, err := buildRequest(
		http.MethodPost,
		"v2/service-accounts",
		serviceAccountBytes,
	)
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	resp, err := getHTTPClient(allowInsecure).Do(req)
	if err != nil {
		return errors.Wrap(err, "error invoking API")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusConflict {
		return errors.Errorf(
			"a service account with the ID %q already exists",
			id,
		)
	}
	if resp.StatusCode != http.StatusCreated {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "error reading response body")
	}

	respStruct := struct {
		Token string `json:"token"`
	}{}
	if err := json.Unmarshal(respBodyBytes, &respStruct); err != nil {
		return errors.Wrap(err, "error unmarshaling response body")
	}

	fmt.Printf("\nService account %q created with token:\n", id)
	fmt.Printf("\n\t%s\n", respStruct.Token)
	fmt.Println(
		"\nStore this token someplace secure NOW. It can not be retrieved " +
			"later through any other means.",
	)

	return nil
}
