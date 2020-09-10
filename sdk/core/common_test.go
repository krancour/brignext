package core

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/brigadecore/brigade/v2/sdk/internal/restmachinery"
	"github.com/brigadecore/brigade/v2/sdk/meta"
	"github.com/stretchr/testify/require"
)

// TODO: Move these-- they're common to several tests

const (
	testAPIAddress          = "localhost:8080"
	testAPIToken            = "11235813213455"
	testClientAllowInsecure = true
)

// TODO: Move this-- it's common to several tests
func requireAPIVersionAndType(
	t *testing.T,
	obj interface{},
	expectedType string,
) {
	objJSON, err := json.Marshal(obj)
	require.NoError(t, err)
	objMap := map[string]interface{}{}
	err = json.Unmarshal(objJSON, &objMap)
	require.NoError(t, err)
	require.Equal(t, meta.APIVersion, objMap["apiVersion"])
	require.Equal(t, expectedType, objMap["kind"])
}

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
