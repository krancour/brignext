package sdk

import (
	"encoding/json"
	"time"

	"github.com/krancour/brignext/v2/sdk/meta"
)

// LogEntry represents one line of output from an OCI container.
type LogEntry struct {
	// Time is the time the line was written.
	Time time.Time `json:"time"`
	// Message is a single line of log output from an OCI container.
	Message string `json:"message"`
}

// MarshalJSON amends LogEntry instances with type metadata so that clients do
// not need to be concerned with the tedium of doing so.
func (l LogEntry) MarshalJSON() ([]byte, error) {
	type Alias LogEntry
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "LogEntry",
			},
			Alias: (Alias)(l),
		},
	)
}

type LogEntryList struct {
	// TODO: When pagination is implemented, list metadata will need to be added
	// Items is a slice of LogEntries.
	Items []LogEntry `json:"items"`
}

// MarshalJSON amends LogEntryList instances with type metadata so that clients
// do not need to be concerned with the tedium of doing so.
func (l LogEntryList) MarshalJSON() ([]byte, error) {
	type Alias LogEntryList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "LogEntryList",
			},
			Alias: (Alias)(l),
		},
	)
}
