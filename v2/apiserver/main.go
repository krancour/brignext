package main

import (
	"log"

	"github.com/brigadecore/brigade/v2/internal/version"
)

func main() {
	log.Printf(
		"Starting Brigade API Server -- version %s -- commit %s",
		version.Version(),
		version.Commit(),
	)

	apiServer, err := getAPIServerFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(apiServer.ListenAndServe())
}
