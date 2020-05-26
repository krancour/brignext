package brignext

import "time"

type LogEntryList struct {
	TypeMeta `json:",inline"`
	ListMeta `json:"metadata"`
	Items    []LogEntry `json:"items"`
}

func NewLogEntryList() LogEntryList {
	return LogEntryList{
		TypeMeta: TypeMeta{
			APIVersion: APIVersion,
			Kind:       "LogEntryList",
		},
		Items: []LogEntry{},
	}
}

type LogEntry struct {
	TypeMeta `json:",inline"`
	Time     time.Time `json:"time" bson:"time"`
	Message  string    `json:"message" bson:"log"`
}

func NewLogEntry() LogEntry {
	return LogEntry{
		TypeMeta: TypeMeta{
			APIVersion: APIVersion,
			Kind:       "LogEntry",
		},
	}
}
