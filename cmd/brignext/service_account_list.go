package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/krancour/brignext/pkg/brignext"

	"github.com/gosuri/uitable"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func serviceAccountList(c *cli.Context) error {
	// Inputs
	output := c.String(flagOutput)
	allowInsecure := c.GlobalBool(flagInsecure)

	switch output {
	case "table":
	case "json":
	default:
		return errors.Errorf("unknown output format %q", output)
	}

	req, err := buildRequest(http.MethodGet, "v2/service-accounts", nil)
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

	serviceAccounts := []brignext.ServiceAccount{}
	if err := json.Unmarshal(respBodyBytes, &serviceAccounts); err != nil {
		return errors.Wrap(err, "error unmarshaling response body")
	}

	if len(serviceAccounts) == 0 {
		fmt.Println("No service accounts found.")
		return nil
	}

	switch output {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "DESCRIPTION", "CREATED")
		for _, serviceAccount := range serviceAccounts {
			table.AddRow(
				serviceAccount.ID,
				serviceAccount.Description,
				serviceAccount.Created,
			)
		}
		fmt.Println(table)

	case "json":
		responseJSON, err := json.MarshalIndent(serviceAccounts, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get service accounts operation",
			)
		}
		fmt.Println(string(responseJSON))
	}

	return nil
}
