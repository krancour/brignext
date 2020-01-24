package auth

import (
	"context"
	"encoding/base64"
	"fmt"

	"google.golang.org/grpc/credentials"
)

type basicAuthCredentials struct {
	username   string
	password   string
	requireTLS bool
}

func NewBasicAuthCredentials(
	username string,
	password string,
	requireTLS bool,
) credentials.PerRPCCredentials {
	return &basicAuthCredentials{
		username:   username,
		password:   password,
		requireTLS: requireTLS,
	}
}

func (b *basicAuthCredentials) GetRequestMetadata(
	context.Context,
	...string,
) (map[string]string, error) {
	return map[string]string{
		"authorization": fmt.Sprintf(
			"Basic %s",
			base64.StdEncoding.EncodeToString(
				[]byte(
					fmt.Sprintf("%s:%s", b.username, b.password),
				),
			),
		),
	}, nil
}

func (b *basicAuthCredentials) RequireTransportSecurity() bool {
	return b.requireTLS
}
