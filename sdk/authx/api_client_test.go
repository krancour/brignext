package authx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAPIClient(t *testing.T) {
	client := NewAPIClient(testAPIAddress, testAPIToken, testClientAllowInsecure)
	require.IsType(t, &apiClient{}, client)
	require.NotNil(t, client.(*apiClient).serviceAccountsClient)
	require.NotNil(t, client.ServiceAccounts())
	require.NotNil(t, client.(*apiClient).sessionsClient)
	require.NotNil(t, client.Sessions())
	require.NotNil(t, client.(*apiClient).usersClient)
	require.NotNil(t, client.Users())
}
