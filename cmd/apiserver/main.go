package main

import (
	"flag"
	"log"

	"github.com/brigadecore/brigade/pkg/storage/kube"
	"github.com/golang/glog"
	"github.com/krancour/brignext/pkg/api"
	"github.com/krancour/brignext/pkg/kubernetes"
	mongodbUtils "github.com/krancour/brignext/pkg/mongodb"
	"github.com/krancour/brignext/pkg/oidc"
	"github.com/krancour/brignext/pkg/storage/mongodb"
	"github.com/krancour/brignext/pkg/version"
)

func main() {
	// We need to parse flags for glog-related options to take effect
	flag.Parse()

	glog.Infof(
		"Starting BrigNext API Server -- version %s -- commit %s",
		version.Version(),
		version.Commit(),
	)

	// API server config
	apiConfig, err := api.GetConfigFromEnvironment()
	if err != nil {
		glog.Fatal(err)
	}

	// OpenID Connect config
	oauth2Config, oidcIdentityVerifier, err :=
		oidc.GetConfigAndVerifierFromEnvironment()
	if err != nil {
		glog.Fatal(err)
	}

	// Old datastore (Kubernetes)
	kubeClient, err := kubernetes.Client()
	if err != nil {
		glog.Fatal(err)
	}
	brigadeNamespace, err := kubernetes.BrigadeNamespace()
	if err != nil {
		glog.Fatal(err)
	}
	oldProjectStore := kube.New(kubeClient, brigadeNamespace)

	// New datastores (Mongo)
	database, err := mongodbUtils.Database()
	if err != nil {
		glog.Fatal(err)
	}
	projectStore := mongodb.NewProjectStore(database)
	logStore := mongodb.NewLogStore(database)
	userStore := mongodb.NewUserStore(database)
	sessionStore := mongodb.NewSessionStore(database)

	log.Println(
		api.NewServer(
			apiConfig,
			oauth2Config,
			oidcIdentityVerifier,
			userStore,
			sessionStore,
			oldProjectStore,
			projectStore,
			logStore,
		).ListenAndServe(),
	)
}
