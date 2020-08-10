package sdk

import (
	"encoding/json"
	"time"

	"github.com/krancour/brignext/v2/sdk/meta"
)

// Event represents an occurrence in some upstream system. Once accepted into
// the system, BrigNext amends each Event with a plan for handling it in the
// form of a Worker. An Event's status is the status of its Worker.
type Event struct {
	// ObjectMeta contains Event metadata.
	meta.ObjectMeta `json:"metadata"`
	// ProjectID specifies the Project this Event is for. Often, this field will
	// be left blank, in which case the Event is matched against subscribed
	// Projects on the basis of the Source, Type, and Labels fields, then used as
	// a template to create a discrete Event for each subscribed Project.
	ProjectID string `json:"projectID,omitempty"`
	// Source specifies the source of the event, e.g. what gateway created it.
	Source string `json:"source,omitempty"`
	// Type specifies the exact event that has occurred in the upstream system.
	// These are source-specific.
	Type string `json:"type,omitempty"`
	// Labels convey additional event details for the purposes of matching Events
	// to subscribed projects. For instance, no subscribers to the "GitHub" Source
	// and the "push" Type are likely to want to hear about push events for ALL
	// repositories. If the "GitHub" gateway labels events with the name of the
	// repository from which the event originated (e.g. "repo=github.com/foo/bar")
	// then subscribers can utilize those same criteria to narrow their
	// subscription from all push events emitted by the GitHub gateway to just
	// those having originated from a specific repository.
	Labels Labels `json:"labels,omitempty"`
	// ShortTitle is an optional, succinct title for the Event, ideal for use in
	// lists or in scenarios where UI real estate is constrained.
	ShortTitle string `json:"shortTitle,omitempty"`
	// LongTitle is an optional, detailed title for the Event.
	LongTitle string `json:"longTitle,omitempty"`
	// Git contains git-specific Event details. These can be used to override
	// similar details defined at the Project level. This is useful for scenarios
	// wherein an Event may need to convey an alternative source, branch, etc.
	Git *EventGitConfig `json:"git,omitempty"`
	// Payload optionally contains Event details provided by the upstream system
	// that was the original source of the event. Payloads MUST NOT contain
	// sensitive information. Since Projects SUBSCRIBE to Events, the potential
	// exists for any Project to express an interest in any or all Events. This
	// being the case, sensitive details must never be present in payloads. The
	// common workaround work this constraint is that event payloads may contain
	// REFERENCES to sensitive details that are useful to properly configured
	// Workers only.
	Payload string `json:"payload,omitempty"`
	// Kubernetes contains Kubernetes-specific details of the Event's Worker's
	// environment. These details are populated by BrigNext. Clients MUST leave
	// the value of this field nil when using the API to create an Event.
	Kubernetes *KubernetesConfig `json:"kubernetes,omitempty"`
	// Worker contains details of the worker that will/is/has handle(d) the Event.
	Worker *Worker `json:"worker,omitempty"`
}

// MarshalJSON amends Event instances with type metadata so that clients do not
// need to be concerned with the tedium of doing so.
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

// EventGitConfig represents git-specific Event details. These may override
// similar details set at the Project level.
type EventGitConfig struct {
	// CloneURL specifies the location from where a source code repository may
	// be cloned.
	CloneURL string `json:"cloneURL,omitempty"`
	// Commit specifies a commit (by sha) to be checked out.
	Commit string `json:"commit,omitempty"`
	// Ref specifies a tag or branch to be checked out. If left blank, this will
	// default to "master" at runtime.
	Ref string `json:"ref,omitempty"`
}

// LogEntry represents one line of output from an OCI container.
type LogEntry struct {
	// Time is the time the line was written.
	Time *time.Time `json:"time,omitempty"`
	// Message is a single line of log output from an OCI container.
	Message string `json:"message,omitempty"`
}

// LogEntryList is an ordered list of LogEntries.
type LogEntryList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of LogEntries.
	Items []LogEntry `json:"items"`
}
