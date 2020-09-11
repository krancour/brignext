package system

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/brigadecore/brigade/v2/sdk/internal/restmachinery"
	"github.com/stretchr/testify/require"
)

const (
	testAPIAddress          = "localhost:8080"
	testAPIToken            = "11235813213455"
	testClientAllowInsecure = true
)

// TODO: Move this-- it's common to several tests
func requireBaseClient(t *testing.T, baseClient *restmachinery.BaseClient) {
	require.Equal(t, testAPIAddress, baseClient.APIAddress)
	require.Equal(t, testAPIToken, baseClient.APIToken)
	require.IsType(t, &http.Client{}, baseClient.HTTPClient)
	require.IsType(t, &http.Transport{}, baseClient.HTTPClient.Transport)
	require.IsType(
		t,
		&tls.Config{},
		baseClient.HTTPClient.Transport.(*http.Transport).TLSClientConfig,
	)
	require.Equal(
		t,
		testClientAllowInsecure,
		baseClient.HTTPClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify, // nolint: lll
	)
}
