package brignext

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTriggeringEventsMatches(t *testing.T) {
	testCases := []struct {
		name          string
		tes           TriggeringEvents
		eventProvider string
		eventType     string
		shouldMatch   bool
	}{
		{
			// Edge case-- really this shouldn't ever happen
			name: "triggering event provider not specified",
			tes: TriggeringEvents{
				Types: []string{"push"},
			},
			eventProvider: "github",
			eventType:     "push",
			shouldMatch:   false,
		},
		{
			// Edge case-- really this shouldn't ever happen
			name: "event provider not specified",
			tes: TriggeringEvents{
				Provider: "github",
				Types:    []string{"push"},
			},
			eventType:   "push",
			shouldMatch: false,
		},
		{
			// Edge case-- really this shouldn't ever happen
			name: "neither triggering event provider nor event provider specified",
			tes: TriggeringEvents{
				Types: []string{"push"},
			},
			eventType:   "push",
			shouldMatch: false,
		},
		{
			name: "provider does not match",
			tes: TriggeringEvents{
				Provider: "github",
				Types:    []string{"push"},
			},
			eventProvider: "bitbucket",
			eventType:     "push",
			shouldMatch:   false,
		},
		{
			name: "provider matches, no triggering types specified",
			tes: TriggeringEvents{
				Provider: "github",
			},
			eventProvider: "github",
			eventType:     "push",
			shouldMatch:   true,
		},
		{
			name: "provider matches, type does not",
			tes: TriggeringEvents{
				Provider: "github",
				Types:    []string{"push"},
			},
			eventProvider: "github",
			eventType:     "issue_comment",
			shouldMatch:   false,
		},
		{
			name: "provider and type both match",
			tes: TriggeringEvents{
				Provider: "github",
				Types:    []string{"push"},
			},
			eventProvider: "github",
			eventType:     "push",
			shouldMatch:   true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(
				t,
				testCase.shouldMatch,
				testCase.tes.Matches(testCase.eventProvider, testCase.eventType),
			)
		})
	}
}
