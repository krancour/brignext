package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/krancour/brignext/v2/pkg/file"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

type config struct {
	APIAddress string `json:"apiAddress"`
	APIToken   string `json:"apiToken"`
}

func getConfig() (*config, error) {
	brignextHome, err := getBrignextHome()
	if err != nil {
		return nil, errors.Wrapf(err, "error finding brignext home")
	}
	brignextConfigFile := path.Join(brignextHome, "config")
	if !file.Exists(brignextConfigFile) {
		return nil, errors.Errorf(
			"no brignext configuration was found at %s; please use "+
				"`brignext login` to continue\n",
			brignextConfigFile,
		)
	}

	configBytes, err := ioutil.ReadFile(brignextConfigFile)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error reading brignext config file at %s",
			brignextConfigFile,
		)
	}

	config := &config{}
	if err := json.Unmarshal(configBytes, config); err != nil {
		return nil, errors.Wrapf(
			err,
			"error parsing brignext config file at %s",
			brignextConfigFile,
		)
	}

	return config, nil
}

func saveConfig(config *config) error {
	brignextHome, err := getBrignextHome()
	if err != nil {
		return errors.Wrapf(err, "error finding brignext home")
	}
	if _, err := os.Stat(brignextHome); err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(
				err,
				"error checking for existence of brignext home at %s",
				brignextHome,
			)
		}
		// The directory doesn't exist-- create it
		if err := os.MkdirAll(brignextHome, 0755); err != nil {
			return errors.Wrapf(
				err,
				"error creating brignext home at %s",
				brignextHome,
			)
		}
	}
	brignextConfigFile := path.Join(brignextHome, "config")

	configBytes, err := json.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "error marshaling config")
	}
	if err :=
		ioutil.WriteFile(brignextConfigFile, configBytes, 0644); err != nil {
		return errors.Wrapf(err, "error writing to %s", brignextConfigFile)
	}
	return nil
}

func deleteConfig() error {
	brignextHome, err := getBrignextHome()
	if err != nil {
		return errors.Wrapf(err, "error finding brignext home")
	}
	brignextConfigFile := path.Join(brignextHome, "config")

	if err := os.Remove(brignextConfigFile); err != nil {
		return errors.Wrap(err, "error deleting configuration")
	}

	return nil
}

func getBrignextHome() (string, error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		return "", errors.Wrap(err, "error locating user's home directory")
	}

	return path.Join(homeDir, ".brignext"), nil
}
