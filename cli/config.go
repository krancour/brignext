package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/brigadecore/brigade/v2/internal/file"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

type config struct {
	APIAddress string `json:"apiAddress"`
	APIToken   string `json:"apiToken"`
}

func getConfig() (*config, error) {
	brigadeHome, err := getBrigadeHome()
	if err != nil {
		return nil, errors.Wrapf(err, "error finding brigade home")
	}
	brigadeConfigFile := path.Join(brigadeHome, "config")
	if !file.Exists(brigadeConfigFile) {
		return nil, errors.Errorf(
			"no brigade configuration was found at %s; please use "+
				"`brig login` to continue\n",
			brigadeConfigFile,
		)
	}

	configBytes, err := ioutil.ReadFile(brigadeConfigFile)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error reading brigade config file at %s",
			brigadeConfigFile,
		)
	}

	config := &config{}
	if err := json.Unmarshal(configBytes, config); err != nil {
		return nil, errors.Wrapf(
			err,
			"error parsing brigade config file at %s",
			brigadeConfigFile,
		)
	}

	return config, nil
}

func saveConfig(config *config) error {
	brigadeHome, err := getBrigadeHome()
	if err != nil {
		return errors.Wrapf(err, "error finding brigade home")
	}
	if _, err = os.Stat(brigadeHome); err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(
				err,
				"error checking for existence of brigade home at %s",
				brigadeHome,
			)
		}
		// The directory doesn't exist-- create it
		if err = os.MkdirAll(brigadeHome, 0755); err != nil {
			return errors.Wrapf(
				err,
				"error creating brigade home at %s",
				brigadeHome,
			)
		}
	}
	brigadeConfigFile := path.Join(brigadeHome, "config")

	configBytes, err := json.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "error marshaling config")
	}
	if err :=
		ioutil.WriteFile(brigadeConfigFile, configBytes, 0644); err != nil {
		return errors.Wrapf(err, "error writing to %s", brigadeConfigFile)
	}
	return nil
}

func deleteConfig() error {
	brigadeHome, err := getBrigadeHome()
	if err != nil {
		return errors.Wrapf(err, "error finding brigade home")
	}
	brigadeConfigFile := path.Join(brigadeHome, "config")

	if err := os.Remove(brigadeConfigFile); err != nil {
		return errors.Wrap(err, "error deleting configuration")
	}

	return nil
}

func getBrigadeHome() (string, error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		return "", errors.Wrap(err, "error locating user's home directory")
	}

	return path.Join(homeDir, ".brigade"), nil
}
