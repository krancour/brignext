package brignext

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTriggeringEventsMatches(t *testing.T) {
	testCases := []struct {
		name        string
		tes         TriggeringEvents
		eventSource string
		eventType   string
		shouldMatch bool
	}{
		{
			// Edge case-- really this shouldn't ever happen
			name: "triggering event source not specified",
			tes: TriggeringEvents{
				Types: []string{"push"},
			},
			eventSource: "github",
			eventType:   "push",
			shouldMatch: false,
		},
		{
			// Edge case-- really this shouldn't ever happen
			name: "event source not specified",
			tes: TriggeringEvents{
				Source: "github",
				Types:  []string{"push"},
			},
			eventType:   "push",
			shouldMatch: false,
		},
		{
			// Edge case-- really this shouldn't ever happen
			name: "neither triggering event source nor event source specified",
			tes: TriggeringEvents{
				Types: []string{"push"},
			},
			eventType:   "push",
			shouldMatch: false,
		},
		{
			name: "source does not match",
			tes: TriggeringEvents{
				Source: "github",
				Types:  []string{"push"},
			},
			eventSource: "bitbucket",
			eventType:   "push",
			shouldMatch: false,
		},
		{
			name: "source matches, no triggering types specified",
			tes: TriggeringEvents{
				Source: "github",
			},
			eventSource: "github",
			eventType:   "push",
			shouldMatch: true,
		},
		{
			name: "source matches, type does not",
			tes: TriggeringEvents{
				Source: "github",
				Types:  []string{"push"},
			},
			eventSource: "github",
			eventType:   "issue_comment",
			shouldMatch: false,
		},
		{
			name: "source and type both match",
			tes: TriggeringEvents{
				Source: "github",
				Types:  []string{"push"},
			},
			eventSource: "github",
			eventType:   "push",
			shouldMatch: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(
				t,
				testCase.shouldMatch,
				testCase.tes.Matches(testCase.eventSource, testCase.eventType),
			)
		})
	}
}
