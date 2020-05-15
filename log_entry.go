package brignext

import "time"

type LogEntryList struct {
	TypeMeta `json:",inline"`
	ListMeta `json:"metadata"`
	Items    []LogEntry `json:"items"`
}

type LogEntry struct {
	Time    time.Time `json:"time" bson:"time"`
	Message string    `json:"message" bson:"log"`
}
