package sdk

import (
	"encoding/json"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"
	"go.mongodb.org/mongo-driver/bson"
)

// Event represents an occurrence in some upstream system. Once accepted into
// the system, BrigNext amends each Event with a plan for handling it in the
// form of a Worker. An Event's status is the status of its Worker.
type Event struct {
	// ObjectMeta contains Event metadata.
	meta.ObjectMeta `json:"metadata" bson:",inline"`
	// ProjectID specifies the Project this Event is for. Often, this field will
	// be left blank, in which case the Event is matched against subscribed
	// Projects on the basis of the Source, Type, and Labels fields, then used as
	// a template to create a discrete Event for each subscribed Project.
	ProjectID string `json:"projectID,omitempty" bson:"projectID,omitempty"`
	// Source specifies the source of the event, e.g. what gateway created it.
	Source string `json:"source,omitempty" bson:"source,omitempty"`
	// Type specifies the exact event that has occurred in the upstream system.
	// These are source-specific.
	Type string `json:"type,omitempty" bson:"type,omitempty"`
	// Labels convey additional event details for the purposes of matching Events
	// to subscribed projects. For instance, no subscribers to the "GitHub" Source
	// and the "push" Type are likely to want to hear about push events for ALL
	// repositories. If the "GitHub" gateway labels events with the name of the
	// repository from which the event originated (e.g. "repo=github.com/foo/bar")
	// then subscribers can utilize those same criteria to narrow their
	// subscription from all push events emitted by the GitHub gateway to just
	// those having originated from a specific repository.
	Labels Labels `json:"labels,omitempty" bson:"labels,omitempty"`
	// ShortTitle is an optional, succinct title for the Event, ideal for use in
	// lists or in scenarios where UI real estate is constrained.
	ShortTitle string `json:"shortTitle,omitempty" bson:"shortTitle,omitempty"`
	LongTitle  string `json:"longTitle,omitempty" bson:"longTitle,omitempty"`
	// Git contains git-specific Event details. These can be used to override
	// similar details defined at the Project level. This is useful for scenarios
	// wherein an Event may need to convey an alternative source, branch, etc.
	Git *EventGitConfig `json:"git,omitempty" bson:"git,omitempty"`
	// Payload optionally contains Event details provided by the upstream system
	// that was the original source of the event. Payloads MUST NOT contain
	// sensitive information. Since Projects SUBSCRIBE to Events, the potential
	// exists for any Project to express an interest in any or all Events. This
	// being the case, sensitive details must never be present in payloads. The
	// common workaround work this constraint is that event payloads may contain
	// REFERENCES to sensitive details that are useful to properly configured
	// Workers only.
	Payload string `json:"payload,omitempty" bson:"payload,omitempty"`
	// Kubernetes contains Kubernetes-specific details of the Event's Worker's
	// environment.
	Kubernetes *KubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes,omitempty"` // nolint: lll
	// Worker contains details of the worker that will/is/has handle(d) the Event.
	Worker Worker `json:"worker" bson:"worker"`
}

// MarshalJSON amends Event instances with type metadata.
func (e Event) MarshalJSON() ([]byte, error) {
	type Alias Event
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "Event",
			},
			Alias: (Alias)(e),
		},
	)
}

// UnmarshalBSON implements custom BSON unmarshaling for the Event type.
// This does little more than guarantees that the Labels field isn't nil so that
// custom unmarshaling of the Labels (which is more involved) can succeed.
func (e *Event) UnmarshalBSON(bytes []byte) error {
	if e.Labels == nil {
		e.Labels = Labels{}
	}
	type EventAlias Event
	return bson.Unmarshal(
		bytes,
		&struct {
			*EventAlias `bson:",inline"`
		}{
			EventAlias: (*EventAlias)(e),
		},
	)
}

// EventGitConfig represents git-specific Event details. These may override
// similar details set at the Project level.
type EventGitConfig struct {
	CloneURL string `json:"cloneURL,omitempty" bson:"cloneURL,omitempty"`
	Commit   string `json:"commit,omitempty" bson:"commit,omitempty"`
	Ref      string `json:"ref,omitempty" bson:"ref,omitempty"`
}

// EventListOptions represents useful filter criteria when selecting multiple
// Events for API group operations like list, cancel, or delete.
type EventListOptions struct {
	// ProjectID specifies that Events belonging to the indicated Project should
	// be selected.
	ProjectID string
	// WorkerPhases specifies that Events with their Worker's in any of the
	// indicated phases should be selected.
	WorkerPhases []WorkerPhase
}

// EventList is an ordered and pageable list of Evens.
type EventList struct {
	// Items is a slice of Events.
	//
	// TODO: When pagination is implemented, list metadata will need to be added
	Items []Event `json:"items,omitempty"`
}

// MarshalJSON amends EventList instances with type metadata.
func (e EventList) MarshalJSON() ([]byte, error) {
	type Alias EventList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "EventList",
			},
			Alias: (Alias)(e),
		},
	)
}

// LogOptions represents useful criteria for identifying a specific container
// of a specific Job when requesting Event logs.
type LogOptions struct {
	// Job specifies, by name, a Job spawned by the Worker. If this field is
	// left blank, it is presumed logs are desired for the Worker itself.
	Job string `json:"job,omitempty"`
	// Container specifies, by name, a container belonging to the Worker or Job
	// whose logs are being retrieved. If left blank, a container with the same
	// name as the Worker or Job is assumed.
	Container string `json:"container,omitempty"`
}

// LogEntry represents one line of output from an OCI container.
type LogEntry struct {
	// Time is the time the line was written.
	Time *time.Time `json:"time,omitempty" bson:"time,omitempty"`
	// Message is a single line of log output from an OCI container.
	Message string `json:"message,omitempty" bson:"message,omitempty"`
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

// LogEntryList is an ordered list of LogEntries.
type LogEntryList struct {
	// Items is a slice of LogEntries.
	//
	// TODO: When pagination is implemented, list metadata will need to be added
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
