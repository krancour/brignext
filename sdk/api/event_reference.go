package api

import (
	"encoding/json"

	"github.com/krancour/brignext/v2/sdk"
	"github.com/krancour/brignext/v2/sdk/meta"
)

// EventReference is an abridged representation of an Event useful to
// API operations that construct and return potentially large collections of
// events. Utilizing such an abridged representation limits response size
// significantly as Events have the potentia to be quite large.
type EventReference struct {
	meta.ObjectReferenceMeta `json:"metadata"`
	ProjectID                string          `json:"projectID"`
	Source                   string          `json:"source"`
	Type                     string          `json:"type"`
	WorkerPhase              sdk.WorkerPhase `json:"workerPhase"`
}

// MarshalJSON amends EventReference instances with type metadata so that
// clients do not need to be concerned with the tedium of doing so.
func (e EventReference) MarshalJSON() ([]byte, error) {
	type Alias EventReference
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "EventReference",
			},
			Alias: (Alias)(e),
		},
	)
}

// EventReferenceList is an ordered list of EventtReferences.
type EventReferenceList struct {
	// TODO: When pagination is implemented, list metadata will need to be added
	// Items is a slice of EventReferences.
	Items []EventReference `json:"items"`
}

// MarshalJSON amends EventReferenceList instances with type metadata so that
// clients do not need to be concerned with the tedium of doing so.
func (e EventReferenceList) MarshalJSON() ([]byte, error) {
	type Alias EventReferenceList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "EventReferenceList",
			},
			Alias: (Alias)(e),
		},
	)
}
