package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func serviceAccountCreate(c *cli.Context) error {
	// Inputs
	var name string
	if len(c.Args()) > 0 {
		name = c.Args()[0]
	}
	description := c.String(flagDescription)
	allowInsecure := c.GlobalBool(flagInsecure)

	reader := bufio.NewReader(os.Stdin)

	for {
		name = strings.TrimSpace(name)
		if name != "" {
			break
		}
		fmt.Print("Service account name? ")
		var err error
		if name, err = reader.ReadString('\n'); err != nil {
			return errors.Wrap(err, "error reading service account name from stdin")
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
		struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}{
			Name:        name,
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

	if resp.StatusCode != http.StatusOK {
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

	fmt.Printf("\nService account %q created with token:\n", name)
	fmt.Printf("\n\t%s\n", respStruct.Token)
	fmt.Println(
		"\nStore this token someplace secure NOW. It can not be retrieved " +
			"later through any other means.",
	)

	return nil
}
