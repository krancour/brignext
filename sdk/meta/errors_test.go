package meta

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testType        = "User"
	testUserID      = "tony@starkindustries.com"
	testErrorReason = "i don't have to answer to you"
)

var testErrorDetails = []string{"the", "devil", "is", "in", "the", "details"}

func TestErrAuthentication(t *testing.T) {
	err := &ErrAuthentication{
		Reason: testErrorReason,
	}
	require.Contains(t, err.Error(), testErrorReason)
}

func TestErrAuthorization(t *testing.T) {
	err := &ErrAuthorization{}
	require.Contains(t, err.Error(), "not authorized")
}

func TestErrBadRequest(t *testing.T) {
	testCases := []struct {
		name       string
		err        *ErrBadRequest
		assertions func(t *testing.T, err *ErrBadRequest)
	}{
		{
			name: "without details",
			err: &ErrBadRequest{
				Reason: testErrorReason,
			},
			assertions: func(t *testing.T, err *ErrBadRequest) {
				require.Contains(t, err.Error(), testErrorReason)
				for _, detail := range err.Details {
					require.NotContains(t, err.Error(), detail)
				}
			},
		},
		{
			name: "with details",
			err: &ErrBadRequest{
				Reason:  testErrorReason,
				Details: testErrorDetails,
			},
			assertions: func(t *testing.T, err *ErrBadRequest) {
				require.Contains(t, err.Error(), testErrorReason)
				for _, detail := range err.Details {
					require.Contains(t, err.Error(), detail)
				}
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.assertions(t, testCase.err)
		})
	}
}

func TestErrNotFound(t *testing.T) {
	err := &ErrNotFound{
		Type: testType,
		ID:   testUserID,
	}
	require.Contains(t, err.Error(), "not found")
	require.Contains(t, err.Error(), testType)
	require.Contains(t, err.Error(), testUserID)
}

func TestErrConflict(t *testing.T) {
	err := &ErrConflict{
		Type:   testType,
		ID:     testUserID,
		Reason: testErrorReason,
	}
	require.Contains(t, err.Error(), testErrorReason)
}

func TestErrInternalServer(t *testing.T) {
	err := &ErrInternalServer{}
	require.Contains(t, err.Error(), "internal server error")
}

func TestErrNotSupported(t *testing.T) {
	err := &ErrNotSupported{
		Details: testErrorReason,
	}
	require.Contains(t, err.Error(), testErrorReason)
}
