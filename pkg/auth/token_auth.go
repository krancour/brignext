package auth

import (
	"context"

	"google.golang.org/grpc/credentials"
)

type tokenAuthCredentials struct {
	token      string
	requireTLS bool
}

func NewTokenAuthCredentials(
	token string,
	requireTLS bool,
) credentials.PerRPCCredentials {
	return &tokenAuthCredentials{
		token:      token,
		requireTLS: requireTLS,
	}
}

func (t *tokenAuthCredentials) GetRequestMetadata(
	context.Context,
	...string,
) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.token,
	}, nil
}

func (t *tokenAuthCredentials) RequireTransportSecurity() bool {
	return t.requireTLS
}
