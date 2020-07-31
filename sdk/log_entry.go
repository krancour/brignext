package sdk

import (
	"time"

	"github.com/krancour/brignext/v2/sdk/meta"
)

type LogEntryList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata"`
	Items         []LogEntry `json:"items"`
}

func NewLogEntryList() LogEntryList {
	return LogEntryList{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "LogEntryList",
		},
		Items: []LogEntry{},
	}
}

type LogEntry struct {
	meta.TypeMeta `json:",inline"`
	Time          time.Time `json:"time"`
	Message       string    `json:"message"`
}

func NewLogEntry() LogEntry {
	return LogEntry{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "LogEntry",
		},
	}
}

type LogOptions struct {
	Job       string `json:"job"`
	Container string `json:"container"`
}
