package main

import (
	"context"
	"flag"

	"github.com/brigadecore/brigade/pkg/storage/kube"
	"github.com/golang/glog"
	"github.com/krancour/brignext/pkg/controller"
	"github.com/krancour/brignext/pkg/kubernetes"
	mongodbUtils "github.com/krancour/brignext/pkg/mongodb"
	"github.com/krancour/brignext/pkg/storage/mongodb"
	"github.com/krancour/brignext/pkg/version"
)

func main() {
	// We need to parse flags for glog-related options to take effect
	flag.Parse()

	glog.Infof(
		"Starting BrigNext Controller -- version %s -- commit %s",
		version.Version(),
		version.Commit(),
	)

	// Old datastore (Kubernetes)
	kubeClient, err := kubernetes.Client()
	if err != nil {
		glog.Fatal(err)
	}
	brigadeNamespace, err := kubernetes.BrigadeNamespace()
	if err != nil {
		glog.Fatal(err)
	}
	oldStore := kube.New(kubeClient, brigadeNamespace)

	// New datastore (Mongo)
	database, err := mongodbUtils.Database()
	if err != nil {
		glog.Fatal(err)
	}
	projectStore := mongodb.NewProjectStore(database)

	// Run the controller
	controller.NewController(
		kubeClient,
		brigadeNamespace,
		oldStore,
		projectStore,
	).Run(context.Background())
}
