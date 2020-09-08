package core

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/crypto"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
)

// Event represents an occurrence in some upstream system. Once accepted into
// the system, Brigade amends each Event with a plan for handling it in the
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

// EventsSelector represents useful filter criteria when selecting multiple
// Events for API group operations like list, cancel, or delete.
type EventsSelector struct {
	// ProjectID specifies that Events belonging to the indicated Project should
	// be selected.
	ProjectID string
	// WorkerPhases specifies that Events with their Worker's in any of the
	// indicated phases should be selected.
	WorkerPhases []WorkerPhase
}

// EventList is an ordered and pageable list of Events.
type EventList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of Events.
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

type CancelManyEventsResult struct {
	Count int64 `json:"count"`
}

func (c CancelManyEventsResult) MarshalJSON() ([]byte, error) {
	type Alias CancelManyEventsResult
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "CancelManyEventsResult",
			},
			Alias: (Alias)(c),
		},
	)
}

type DeleteManyEventsResult struct {
	Count int64 `json:"count"`
}

func (d DeleteManyEventsResult) MarshalJSON() ([]byte, error) {
	type Alias DeleteManyEventsResult
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "DeleteManyEventsResult",
			},
			Alias: (Alias)(d),
		},
	)
}

// EventsService is the specialized interface for managing Events. It's
// decoupled from underlying technology choices (e.g. data store, message bus,
// etc.) to keep business logic reusable and consistent while the underlying
// tech stack remains free to change.
type EventsService interface {
	// Create creates a new Event.
	Create(context.Context, Event) (
		EventList,
		error,
	)
	// List returns an EventList, with its Items (Events) ordered by
	// age, newest first. Criteria for which Events should be retrieved can be
	// specified using the EventListOptions parameter.
	List(
		context.Context,
		EventsSelector,
		meta.ListOptions,
	) (EventList, error)
	// Get retrieves a single Event specified by its identifier.
	Get(context.Context, string) (Event, error)
	// GetByWorkerToken retrieves a single Event specified by its Worker's token.
	GetByWorkerToken(context.Context, string) (Event, error)
	// Cancel cancels a single Event specified by its identifier.
	Cancel(context.Context, string) error
	// CancelMany cancels multiple Events specified by the EventListOptions
	// parameter.
	CancelMany(
		context.Context,
		EventsSelector,
	) (CancelManyEventsResult, error)
	// Delete deletes a single Event specified by its identifier.
	Delete(context.Context, string) error
	// DeleteMany deletes multiple Events specified by the EventListOptions
	// parameter.
	DeleteMany(
		context.Context,
		EventsSelector,
	) (DeleteManyEventsResult, error)
}

type eventsService struct {
	authorize     authx.AuthorizeFn
	projectsStore ProjectsStore
	eventsStore   EventsStore
	substrate     Substrate
}

// NewEventsService returns a specialized interface for managing Events.
func NewEventsService(
	projectsStore ProjectsStore,
	eventsStore EventsStore,
	substrate Substrate,
) EventsService {
	return &eventsService{
		authorize:     authx.Authorize,
		projectsStore: projectsStore,
		eventsStore:   eventsStore,
		substrate:     substrate,
	}
}

// TODO: There's a lot of stuff that happens in this function that maybe we
// should defer until later-- like when the worker pod actually gets created.
func (e *eventsService) Create(
	ctx context.Context,
	event Event,
) (EventList, error) {
	events := EventList{}

	if event.ProjectID == "" {
		if err := e.authorize(
			ctx,
			authx.RoleEventCreator(event.Source),
		); err != nil {
			return events, err
		}
	} else {
		if err := e.authorize(
			ctx,
			authx.RoleProjectUser(event.ProjectID),
			authx.RoleEventCreator(event.Source),
		); err != nil {
			return events, err
		}
	}

	now := time.Now()
	event.Created = &now

	// If no project ID is specified, we use other criteria to locate projects
	// that are subscribed to this event. We iterate over all of those and create
	// an event for each of these by making a recursive call to this same
	// function.
	if event.ProjectID == "" {
		projects, err := e.projectsStore.ListSubscribers(ctx, event)
		if err != nil {
			return events, errors.Wrap(
				err,
				"error retrieving subscribed projects from store",
			)
		}
		events.Items = make([]Event, len(projects.Items))
		for i, project := range projects.Items {
			event.ProjectID = project.ID
			projectEvents, err := e.Create(ctx, event)
			if err != nil {
				return events, err
			}
			// projectEvents.Items will always contain precisely one element
			events.Items[i] = projectEvents.Items[0]
		}
		return events, nil
	}

	// Make sure the project exists
	project, err := e.projectsStore.Get(ctx, event.ProjectID)
	if err != nil {
		return events, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			event.ProjectID,
		)
	}

	event.ID = uuid.NewV4().String()

	workerSpec := project.Spec.WorkerTemplate

	if workerSpec.WorkspaceSize == "" {
		workerSpec.WorkspaceSize = "10Gi"
	}

	if event.Git != nil {
		if workerSpec.Git == nil {
			workerSpec.Git = &WorkerGitConfig{}
		}
		// VCS details from the event override project-level details
		// TODO: Might need some nil checks below
		if event.Git.CloneURL != "" {
			workerSpec.Git.CloneURL = event.Git.CloneURL
		}
		if event.Git.Commit != "" {
			workerSpec.Git.Commit = event.Git.Commit
		}
		if event.Git.Ref != "" {
			workerSpec.Git.Ref = event.Git.Ref
		}
	}
	if workerSpec.Git != nil {
		if workerSpec.Git.CloneURL != "" &&
			workerSpec.Git.Commit == "" &&
			workerSpec.Git.Ref == "" {
			workerSpec.Git.Ref = "master"
		}
	}

	if workerSpec.LogLevel == "" {
		workerSpec.LogLevel = LogLevelInfo
	}

	if workerSpec.ConfigFilesDirectory == "" {
		workerSpec.ConfigFilesDirectory = "."
	}

	token := crypto.NewToken(256)

	event.Worker = Worker{
		Spec: workerSpec,
		Status: WorkerStatus{
			Phase: WorkerPhasePending,
		},
		Token:       token,
		HashedToken: crypto.ShortSHA("", token),
	}

	// Amend the Event with substrate-specific details before we persist.
	if event, err = e.substrate.PreCreateEvent(ctx, project, event); err != nil {
		return events, errors.Wrapf(
			err,
			"error pre-creating event %q on the substrate",
			event.ID,
		)
	}

	// Persist the Event
	if err = e.eventsStore.Create(ctx, event); err != nil {
		return events, errors.Wrapf(
			err,
			"error storing new event %q",
			event.ID,
		)
	}

	// Prepare the substrate for the Worker
	if err = e.substrate.PreScheduleWorker(ctx, event); err != nil {
		return events, errors.Wrapf(
			err,
			"error pre-scheduling event %q worker on the substrate",
			event.ID,
		)
	}

	// Schedule the Worker for async / eventual execution
	if err = e.substrate.ScheduleWorker(ctx, event); err != nil {
		return events, errors.Wrapf(
			err,
			"error scheduling event %q worker on the substrate",
			event.ID,
		)
	}

	events.Items = []Event{event}
	return events, nil
}

func (e *eventsService) List(
	ctx context.Context,
	selector EventsSelector,
	opts meta.ListOptions,
) (EventList, error) {
	if err := e.authorize(ctx, authx.RoleReader()); err != nil {
		return EventList{}, err
	}

	// If no worker phase filters were applied, retrieve all phases
	if len(selector.WorkerPhases) == 0 {
		selector.WorkerPhases = WorkerPhasesAll()
	}
	if opts.Limit == 0 {
		opts.Limit = 20
	}

	events, err := e.eventsStore.List(ctx, selector, opts)
	if err != nil {
		return events, errors.Wrap(err, "error retrieving events from store")
	}
	return events, nil
}

func (e *eventsService) Get(
	ctx context.Context,
	id string,
) (Event, error) {
	if err := e.authorize(ctx, authx.RoleReader()); err != nil {
		return Event{}, err
	}

	event, err := e.eventsStore.Get(ctx, id)
	if err != nil {
		return event, errors.Wrapf(err, "error retrieving event %q from store", id)
	}
	return event, nil
}

func (e *eventsService) GetByWorkerToken(
	ctx context.Context,
	workerToken string,
) (Event, error) {
	// No authz is required here because this is only ever called by the system
	// itself.

	event, err := e.eventsStore.GetByHashedWorkerToken(
		ctx,
		crypto.ShortSHA("", workerToken),
	)
	if err != nil {
		return event, errors.Wrap(err, "error retrieving event from store")
	}
	return event, nil
}

func (e *eventsService) Cancel(ctx context.Context, id string) error {
	event, err := e.eventsStore.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", id)
	}

	if err = e.authorize(
		ctx,
		authx.RoleProjectUser(event.ProjectID),
	); err != nil {
		return err
	}

	if err = e.eventsStore.Cancel(ctx, id); err != nil {
		return errors.Wrapf(err, "error canceling event %q in store", id)
	}

	if err = e.substrate.DeleteWorkerAndJobs(ctx, event); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q worker and jobs from the substrate",
			id,
		)
	}

	return nil
}

func (e *eventsService) CancelMany(
	ctx context.Context,
	selector EventsSelector,
) (CancelManyEventsResult, error) {
	result := CancelManyEventsResult{}

	// Refuse requests not qualified by project
	if selector.ProjectID == "" {
		return result, &meta.ErrBadRequest{
			Reason: "Requests to cancel multiple events must be qualified by " +
				"project.",
		}
	}

	if err := e.authorize(
		ctx,
		authx.RoleProjectUser(selector.ProjectID),
	); err != nil {
		return CancelManyEventsResult{}, err
	}

	// Refuse requests not qualified by worker phases
	if len(selector.WorkerPhases) == 0 {
		return result, &meta.ErrBadRequest{
			Reason: "Requests to cancel multiple events must be qualified by " +
				"worker phase(s).",
		}
	}

	if selector.ProjectID != "" {
		// Make sure the project exists
		_, err := e.projectsStore.Get(ctx, selector.ProjectID)
		if err != nil {
			return result, errors.Wrapf(
				err,
				"error retrieving project %q from store",
				selector.ProjectID,
			)
		}
	}

	events, err := e.eventsStore.CancelMany(ctx, selector)
	if err != nil {
		return result, errors.Wrap(err, "error canceling events in store")
	}

	result.Count = int64(len(events.Items))

	// TODO: Can we find a quicker, more efficient way to do this?
	go func() {
		for _, event := range events.Items {
			if err := e.substrate.DeleteWorkerAndJobs(
				context.Background(), // Deliberately not using request context
				event,
			); err != nil {
				log.Println(
					errors.Wrapf(
						err,
						"error deleting event %q worker and jobs from the substrate",
						event.ID,
					),
				)
			}
		}
	}()

	return result, nil
}

func (e *eventsService) Delete(ctx context.Context, id string) error {
	event, err := e.eventsStore.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", id)
	}

	if err = e.authorize(
		ctx,
		authx.RoleProjectUser(event.ProjectID),
	); err != nil {
		return err
	}

	if err = e.eventsStore.Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "error deleting event %q from store", id)
	}

	if err = e.substrate.DeleteWorkerAndJobs(ctx, event); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q worker and jobs from the substrate",
			id,
		)
	}

	return nil
}

func (e *eventsService) DeleteMany(
	ctx context.Context,
	selector EventsSelector,
) (DeleteManyEventsResult, error) {
	result := DeleteManyEventsResult{}

	// Refuse requests not qualified by project
	if selector.ProjectID == "" {
		return result, &meta.ErrBadRequest{
			Reason: "Requests to delete multiple events must be qualified by " +
				"project.",
		}
	}

	if err := e.authorize(
		ctx,
		authx.RoleProjectUser(selector.ProjectID),
	); err != nil {
		return DeleteManyEventsResult{}, err
	}

	// Refuse requests not qualified by worker phases
	if len(selector.WorkerPhases) == 0 {
		return result, &meta.ErrBadRequest{
			Reason: "Requests to delete multiple events must be qualified by " +
				"worker phase(s).",
		}
	}

	if selector.ProjectID != "" {
		// Make sure the project exists
		_, err := e.projectsStore.Get(ctx, selector.ProjectID)
		if err != nil {
			return result, errors.Wrapf(
				err,
				"error retrieving project %q from store",
				selector.ProjectID,
			)
		}
	}

	events, err := e.eventsStore.DeleteMany(ctx, selector)
	if err != nil {
		return result, errors.Wrap(err, "error deleting events from store")
	}

	result.Count = int64(len(events.Items))

	// TODO: Can we find a quicker, more efficient way to do this?
	go func() {
		for _, event := range events.Items {
			if err := e.substrate.DeleteWorkerAndJobs(
				context.Background(), // Deliberately not using request context
				event,
			); err != nil {
				log.Println(
					errors.Wrapf(
						err,
						"error deleting event %q worker and jobs from the substrate",
						event.ID,
					),
				)
			}
		}
	}()

	return result, nil
}

type EventsStore interface {
	Create(context.Context, Event) error
	List(
		context.Context,
		EventsSelector,
		meta.ListOptions,
	) (EventList, error)
	Get(context.Context, string) (Event, error)
	GetByHashedWorkerToken(context.Context, string) (Event, error)
	Cancel(context.Context, string) error
	CancelMany(
		context.Context,
		EventsSelector,
	) (EventList, error)
	Delete(context.Context, string) error
	DeleteMany(
		context.Context,
		EventsSelector,
	) (EventList, error)
}
