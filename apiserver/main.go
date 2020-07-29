package main

import (
	"log"

	"github.com/krancour/brignext/v2/internal/version"
)

func main() {
	log.Printf(
		"Starting BrigNext API Server -- version %s -- commit %s",
		version.Version(),
		version.Commit(),
	)

	apiServer, err := getAPIServerFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(apiServer.ListenAndServe())
}
