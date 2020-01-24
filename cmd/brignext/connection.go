package main

import (
	"fmt"

	"github.com/krancour/brignext/pkg/auth"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// getRegistrationConnection gets a new connection with no credentials
// specified. It's only useful for registration.
func getRegistrationConnection(
	host string,
	port int,
) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", host, port),
		grpc.WithInsecure(), // TODO: Don't hardcode this-- make a flag for it
	)
	if err != nil {
		return nil, errors.Wrap(err, "error dialing API server")
	}
	return conn, nil
}

// getLoginConnection gets a new connection using username / password
// credentials. It's only useful for logging in. All other auth is accomplished
// with a bearer token.
func getLoginConnection(
	host string,
	port int,
	username string,
	password string,
) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", host, port),
		grpc.WithInsecure(), // TODO: Don't hardcode this-- make a flag for it
		grpc.WithPerRPCCredentials(
			auth.NewBasicAuthCredentials(
				username,
				password,
				false, // TODO: Don't hardcode this-- make a flag for it
			),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error dialing API server")
	}
	return conn, nil
}

func getConnection() (*grpc.ClientConn, error) {
	config, err := getConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "error retrieving configuration")
	}

	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", config.APIHost, config.APIPort),
		grpc.WithInsecure(), // TODO: Don't hardcode this-- make a flag for it
		grpc.WithPerRPCCredentials(
			auth.NewTokenAuthCredentials(
				config.APIToken,
				false, // TODO: Don't hardcode this-- make a flag for it
			),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error dialing API server")
	}
	return conn, nil
}
