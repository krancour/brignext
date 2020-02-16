package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/krancour/brignext/pkg/brignext"
	"k8s.io/apimachinery/pkg/util/duration"

	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func serviceAccountGet(c *cli.Context) error {
	// Args
	if len(c.Args()) != 1 {
		return errors.New(
			"service-account get requires one argument-- a service account ID " +
				"(case insensitive)",
		)
	}
	id := c.Args()[0]

	// Global flags
	allowInsecure := c.GlobalBool(flagInsecure)

	// Command-specific flags
	output := c.String(flagOutput)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	req, err := buildRequest(
		http.MethodGet,
		fmt.Sprintf("v2/service-accounts/%s", id),
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

	if resp.StatusCode == http.StatusNotFound {
		return errors.Errorf("Service account %q not found.", id)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("received %d from API server", resp.StatusCode)
	}

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "error reading response body")
	}

	serviceAccount := brignext.ServiceAccount{}
	if err := json.Unmarshal(respBodyBytes, &serviceAccount); err != nil {
		return errors.Wrap(err, "error unmarshaling response body")
	}

	switch strings.ToLower(output) {
	case "table":
		table := uitable.New()
		table.AddRow("ID", "DESCRIPTION", "AGE", "LOCKED?")
		var age string
		if serviceAccount.Created != nil {
			age = duration.ShortHumanDuration(time.Since(*serviceAccount.Created))
		}
		table.AddRow(
			serviceAccount.ID,
			serviceAccount.Description,
			age,
			serviceAccount.Locked != nil && *serviceAccount.Locked,
		)
		fmt.Println(table)

	case "json":
		prettyJSON, err := json.MarshalIndent(serviceAccount, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get service account operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}
