package core

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAPIClient(t *testing.T) {
	client := NewAPIClient(testAPIAddress, testAPIToken, testClientAllowInsecure)
	require.IsType(t, &apiClient{}, client)
	require.NotNil(t, client.(*apiClient).projectsClient)
	require.Equal(t, client.(*apiClient).projectsClient, client.Projects())
	require.NotNil(t, client.(*apiClient).eventsClient)
	require.Equal(t, client.(*apiClient).eventsClient, client.Events())
}
