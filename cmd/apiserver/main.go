package main

import (
	"flag"
	"net"

	"github.com/brigadecore/brigade/pkg/storage/kube"
	"github.com/golang/glog"
	grpcAuth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/krancour/brignext/pkg/auth"
	"github.com/krancour/brignext/pkg/builds"
	"github.com/krancour/brignext/pkg/kubernetes"
	mongodbUtils "github.com/krancour/brignext/pkg/mongodb"
	"github.com/krancour/brignext/pkg/projects"
	"github.com/krancour/brignext/pkg/storage/mongodb"
	"github.com/krancour/brignext/pkg/users"
	"github.com/krancour/brignext/pkg/version"
	"google.golang.org/grpc"
)

func main() {
	// We need to parse flags for glog-related options to take effect
	flag.Parse()

	glog.Infof(
		"Starting BrigNext API Server -- version %s -- commit %s",
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

	// New datastores (Mongo)
	database, err := mongodbUtils.Database()
	if err != nil {
		glog.Fatal(err)
	}
	projectStore := mongodb.NewProjectStore(database)
	logStore := mongodb.NewLogStore(database)
	userStore := mongodb.NewUserStore(database)

	// gRPC server

	authenticator := auth.NewAuthenticator(userStore)

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(
			grpcAuth.StreamServerInterceptor(authenticator.Authenticate),
		),
		grpc.UnaryInterceptor(
			grpcAuth.UnaryServerInterceptor(authenticator.Authenticate),
		),
	)

	users.RegisterUsersServer(
		grpcServer,
		users.NewServer(userStore),
	)
	projects.RegisterProjectsServer(
		grpcServer,
		projects.NewServer(oldStore, projectStore),
	)
	builds.RegisterBuildsServer(
		grpcServer,
		builds.NewServer(oldStore, projectStore, logStore),
	)

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		glog.Fatalf("failed to listen: %s", err)
	}

	grpcServer.Serve(lis)
}
