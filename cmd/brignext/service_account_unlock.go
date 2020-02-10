package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func serviceAccountUnlock(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"service-account unlock requires one parameter-- a service account ID" +
				"(case insensitive)",
		)
	}
	id := c.Args()[0]

	// Global flags
	allowInsecure := c.GlobalBool(flagInsecure)

	req, err := buildRequest(
		http.MethodDelete,
		fmt.Sprintf("v2/service-accounts/%s/lock", id),
		nil,
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

	fmt.Printf(
		"\nService account %q unlocked and a new token has been issued:\n",
		id,
	)
	fmt.Printf("\n\t%s\n", respStruct.Token)
	fmt.Println(
		"\nStore this token someplace secure NOW. It can not be retrieved " +
			"later through any other means.",
	)

	fmt.Printf("Service account %q unlocked.\n", id)

	return nil
}
