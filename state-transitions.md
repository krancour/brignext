# BrigNext State Transitions

Unlike Brigade 1.x, BrigNext is meant to persist event data in a database and
keep those records synced with resources in the underlying worker runtime (i.e.
Kubernetes).

When an API request creates an event, apart from storing it in a database,
certain resources must be created in the underlying runtime. Likewise, when an
event is canceled or deleted via an API request, certain resources must be
deleted from the the underlying runtime.

Conversely, when a component of the scheduler that monitors worker status
observes success, failure, timeout, etc., status updates need to be recorded in
the database and resources in the underlying runtime must be scheduled for
deletion after a short grace period that compensates for any back pressure on
the log aggregator.

Some of the logic that is described in broad terms above is not yet
implemented-- or in some cases is implemented, but poorly. This remainder of
this document describes a number of detailed scenarios to guide developers in
covering all cases.

## Event is Created by an API Call

A single event can be created by directly referencing a project. Multiple
events can be created by finding projects that are subscribed to a event
prototype and iterating over those, creating a distinct event for each.

The steps below describe the process for creating a _single_ event.

1. [ ] Configure the event's worker by copying relevant details from the
   corresponding project's worker template.
1. [ ] Set status of the event's worker to PENDING.
1. [ ] Create Kubernetes resources associated with the event. This is done prior
   to persisting the event in the database so that the scheduler component can
   augment the event with any necessary Kubernetes-related details beforehand.
    1. [ ] A secret that contains all event and project details that will be
       required by the worker pod when it is eventually created by the
       controller. This is a secret, because it contains a point-in-time COPY of
       the project's own secrets. This was a carefully considered design choice.
       Creation of the event secret, including the copy of project secrets
       _could_ have been deferred until just before the worker pod is scheduled,
       however, with the event itself having been persisted _immediately_ upon
       creation, with its worker spec based on the project's worker template _at
       that moment in time_, it would be inconsistent to use project secrets as
       they existed in _a different_ moment in time. Creating an event secret
       _immediately_ at the same time as the event itself, with all of it based
       off a single, consistent point-in-time snapshot for the project won out
       as the correct decision.
1. [ ] Store the new event in the database.
1. [ ] Schedule asynchronous execution of the event's worker using a reliable
  queue.

## Event is Canceled by an API Call

A single event can be canceled unconditionally _if_ it is in a non-terminal
state. It is a conflict (409) to cancel an event that is already in a terminal
state. PENDING events transition to CANCELED upon cancelation while RUNNING
events transition to ABORTED. Multiple events can be canceled conditionally
by matching specific criteria.

The steps below describe the process for canceling a _single_ event.

1. [ ] Cancel the event in the database.
    1. [ ] Update status of the event's worker.
    1. [ ] TODO: Do we need to update each of the worker's jobs as well?
1. [ ] Delete any Kubernetes resources associated with the event, which by
   definition would also include all resources associated with the event's
   worker or any jobs spawned by that worker. We should probably consider doing
   this asynchronously (use a reliable queue? or make a best effort with a
   goroutine?) so API clients aren't potentially waiting a long time for a
   response while the API server has a long and chatty conversation with the
   Kubernetes API server.
    1. [ ] Pods
    1. [ ] Persistent volume claims
    1. [ ] Config maps (BrigNext never creates any, but a custom worker might)
    1. [ ] Secrets

## Event is Deleted by an API Call

A single event can be deleted unconditionally. Multiple events can be deleted
conditionally by matching specific criteria.

The steps below describe the process for deleting a _single_ event.

1. [ ] Delete the event from the database.
1. [ ] Delete any Kubernetes resources associated with the event. We should
   probably consider doing this asynchronously (use a reliable queue? or make a
   best effort with a goroutine?) so API clients aren't potentially waiting a
   long time for a response while the API server has a long and chatty
   conversation with the Kubernetes API server.
    1. [ ] Pods
    1. [ ] Persistent volume claims
    1. [ ] Config maps (BrigNext never creates any, but a custom worker might)
    1. [ ] Secrets

## Worker is Scheduled by the Controller

1. [ ] Receive a reference to an event via persistent queue.
1. [ ] Wait for available capacity.
1. [ ] Retrieve the event using the API.
    1. [ ] If a 404 is received, presume the event to have been deleted. Go to 6.
    1. [ ] If the event's worker is in a RUNNING state, go to 5.
    1. [ ] If the event's worker is in a terminal state, go to 6.
1. [ ] Create Kubernetes resources associated with the event.
    1. [ ] Persistent volume claim for event workspace
    1. [ ] Worker pod
1. [ ] Wait for worker pod completion.
1. [ ] Signal available capacity.

## Worker Pod State is Observed by the Controller

1. [ ] Use API to update worker state.
    1. [ ] If a 404 is received, presume the event to have been deleted. Treat
       it as such.
    1. [ ] If a 409 is received, the worker was already in a terminal state.
       Treat it as if it had been canceled/aborted.
1. [ ] If state is terminal, defer deletion of Kubernetes resources 60 seconds
   to compensate for any back pressure on the log aggregator _unless_ the event
   is presumed deleted or aborted, then delete Kubernetes resources immediately.
    1. [ ] Pods
    1. [ ] Persistent volume claims
    1. [ ] Config maps (BrigNext never creates any, but a custom worker might)
    1. [ ] Secrets

## Worker Pod Timeout is Observed by the Controller

- [ ] Use API to update worker state.
    1. [ ] If a 404 is received, presume the event to have been deleted.
    1. [ ] If a 409 is received, the worker was already in a terminal state.
       Treat it as if it had been canceled/aborted.
- [ ] Immediately delete Kubernetes resources.
  - [ ] Pods
  - [ ] Persistent volume claims
  - [ ] Config maps
  - [ ] Secrets

## Worker Fails to Start

- [ ] Use API to update worker state.
    1. [ ] If known, include an explanation for why the worker wouldn't start.
    1. [ ] If a 404 is received, presume the event to have been deleted.
    1. [ ] If a 409 is received, the worker was already in a terminal state
       (such as ABORTED). Treat it as such going forward.
- [ ] Immediately delete Kubernetes resources.
  - [ ] Pods
  - [ ] Persistent volume claims
  - [ ] Config maps
  - [ ] Secrets

## Job State is Observed by the Controller

NB: The controller does not start jobs. _Workers_ start jobs. In fact,
controllers do not even know of any given job's existence until its pod first
appears. From that point on, the controller monitors jobs only for the sake of
reporting their status to the API server to clean up after job completion is
detected. Both of these lower the bar for the implementation of custom worker
images. It is also worth noting that from BrigNext's own perspective, job status
has no bearing on worker status. The worker that _created_ the job determines
how success, failure, or timeout impact the worker's eventual exit code.

1. [ ] Use API to update job state.
    1. [ ] If a 404 is received, presume the event to have been deleted.
    1. [ ] If a 409 is received, the job was already in a terminal state. Treat
       it as such going forward.
1. [ ] If state is terminal, defer deletion of Kubernetes resources 60 seconds
   to compensate for any back pressure on the log aggregator.
    1. [ ] Pods
    1. [ ] Config maps
    1. [ ] Secrets

## Job Timeout is Observed by the Controller

NB: Monitoring for job timeout is the concern of the worker that created the
job because the controller has no way of knowing what the job's configured
timeout value is.

## Job Fails to Start

- [ ] Use API to update job state.
    1. [ ] If known, include an explanation for why the job wouldn't start.
    1. [ ] If a 404 is received, presume the event to have been deleted.
    1. [ ] If a 409 is received, the job was already in a terminal state
       (such as ABORTED). Treat it as such going forward.
- [ ] Immediately delete Kubernetes resources.
  - [ ] Pods
  - [ ] Persistent volume claims
  - [ ] Config maps
  - [ ] Secrets
  