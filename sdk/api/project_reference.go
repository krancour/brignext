package api

import (
	"encoding/json"

	"github.com/krancour/brignext/v2/sdk/meta"
)

// ProjectReference is an abridged representation of a Project useful to
// API operations that construct and return potentially large collections of
// projects. Utilizing such an abridged representation both limits response size
// and accounts for the reality that not all clients with authorization to list
// projects are authorized to view the details of every Project.
type ProjectReference struct {
	// ObjectReferenceMeta encapsulates an abridged representation of Project
	// metadata.
	meta.ObjectReferenceMeta `json:"metadata"`
	// Description is a natural language description of the Project.
	Description string `json:"description"`
}

// MarshalJSON amends ProjectReference instances with type metadata so that
// clients do not need to be concerned with the tedium of doing so.
func (p ProjectReference) MarshalJSON() ([]byte, error) {
	type Alias ProjectReference
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ProjectReference",
			},
			Alias: (Alias)(p),
		},
	)
}

// ProjectReferenceList is an ordered list of ProjectReferences.
type ProjectReferenceList struct {
	// TODO: When pagination is implemented, list metadata will need to be added
	// Items is a slice of ProjectReferences.
	Items []ProjectReference `json:"items"`
}

// MarshalJSON amends ProjectReferenceList instances with type metadata so that
// clients do not need to be concerned with the tedium of doing so.
func (p ProjectReferenceList) MarshalJSON() ([]byte, error) {
	type Alias ProjectReferenceList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ProjectReferenceList",
			},
			Alias: (Alias)(p),
		},
	)
}
