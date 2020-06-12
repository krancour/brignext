package serviceaccounts

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/pkg/api"
)

type Client interface {
	Create(context.Context, brignext.ServiceAccount) (brignext.Token, error)
	List(context.Context) (brignext.ServiceAccountList, error)
	Get(context.Context, string) (brignext.ServiceAccount, error)
	Lock(context.Context, string) error
	Unlock(context.Context, string) (brignext.Token, error)
}

type client struct {
	*api.BaseClient
}

func NewClient(apiAddress string, apiToken string, allowInsecure bool) Client {
	return &client{
		BaseClient: &api.BaseClient{
			APIAddress: apiAddress,
			APIToken:   apiToken,
			HTTPClient: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: allowInsecure,
					},
				},
			},
		},
	}
}

func (c *client) Create(
	_ context.Context,
	serviceAccount brignext.ServiceAccount,
) (brignext.Token, error) {
	token := brignext.Token{}
	return token, c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/service-accounts",
			AuthHeaders: c.BearerTokenAuthHeaders(),
			ReqBodyObj:  serviceAccount,
			SuccessCode: http.StatusCreated,
			RespObj:     &token,
		},
	)
}

func (c *client) List(context.Context) (brignext.ServiceAccountList, error) {
	serviceAccountList := brignext.ServiceAccountList{}
	return serviceAccountList, c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/service-accounts",
			AuthHeaders: c.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &serviceAccountList,
		},
	)
}

func (c *client) Get(
	_ context.Context,
	id string,
) (brignext.ServiceAccount, error) {
	serviceAccount := brignext.ServiceAccount{}
	return serviceAccount, c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/service-accounts/%s", id),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &serviceAccount,
		},
	)
}

func (c *client) Lock(_ context.Context, id string) error {
	return c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/service-accounts/%s/lock", id),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}

func (c *client) Unlock(_ context.Context, id string) (brignext.Token, error) {
	token := brignext.Token{}
	return token, c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        fmt.Sprintf("v2/service-accounts/%s/lock", id),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &token,
		},
	)
}
