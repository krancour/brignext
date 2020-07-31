package sdk

import (
	"encoding/json"
	"time"

	"github.com/krancour/brignext/v2/sdk/meta"
)

type LogEntryList struct {
	Items []LogEntry `json:"items"`
}

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

type LogEntry struct {
	Time    time.Time `json:"time"`
	Message string    `json:"message"`
}

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

type LogOptions struct {
	Job       string `json:"job"`
	Container string `json:"container"`
}
