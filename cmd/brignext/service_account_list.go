package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/krancour/brignext/pkg/brignext"
	"k8s.io/apimachinery/pkg/util/duration"

	"github.com/gosuri/uitable"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func serviceAccountList(c *cli.Context) error {
	// Args
	if len(c.Args()) != 0 {
		return errors.New("service-account list requires no arguments")
	}

	// Global flags
	allowInsecure := c.GlobalBool(flagInsecure)

	// Command-specific flags
	output := c.String(flagOutput)

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
		table.AddRow("ID", "DESCRIPTION", "AGE", "LOCKED?")
		for _, serviceAccount := range serviceAccounts {
			age := "???"
			if serviceAccount.Created != nil {
				age = duration.ShortHumanDuration(time.Since(*serviceAccount.Created))
			}
			table.AddRow(
				serviceAccount.ID,
				serviceAccount.Description,
				age,
				serviceAccount.Locked != nil && *serviceAccount.Locked,
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
