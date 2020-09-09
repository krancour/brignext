package system

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAPIClient(t *testing.T) {
	client := NewAPIClient(testAPIAddress, testAPIToken, testClientAllowInsecure)
	require.IsType(t, &apiClient{}, client)
	require.NotNil(t, client.(*apiClient).rolesClient)
	require.NotNil(t, client.Roles())
}
