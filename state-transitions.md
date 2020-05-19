# BrigNext State Transitions

## Events

### Event is Deleted by an API Call

- [ ] If the event's worker is in __terminal__ state, delete the event.
- [ ] If the `--pending` flag is set and the event's worker is in a `PENDING`
  state, delete it.
- [ ] If the `--running` flag is set and the event's worker is in a `RUNNING`
  state, delete it.
- [ ] If the event was deleted, delete any Kubernetes resources associated with
  the event:
  - [ ] Pods
  - [ ] Persistent volume claims
  - [ ] Config maps
  - [ ] Secrets

__How will the controller react to this?__

- [ ] If any pods (workers or jobs) shut down as a result of this, the
  controller will attempt to invoke the API to update worker or job status and
  will receive a 404, which it will log, but we'll be done.

### Event is Canceled by an API Call

- [ ] If the event's worker is in __terminal__ state, do nothing.
- [ ] If the event's worker is in a `PENDING` state, update its state to
  `CANCELED`.
- [ ] If the `--running` flag is set and the event's worker is in a
  `RUNNING` state, update its state to `ABORTED`.
  - [ ] For each job, if the job is in a `RUNNING` state, update its state to
    `ABORTED`.
- [ ] If the event was deleted, delete any Kubernetes resources associated with
  the event, _except for the event secret_. This is where the payload is stored
  and we want to keep that around for as long as the secret exists.
  - [ ] Pods
  - [ ] Persistent volume claims
  - [ ] Config maps
  - [ ] Secrets

__How will the controller react to this?__

- [ ] If any worker or job pods shut down as a result of this, the controller
  doesn't really know the reason. It will attempt to invoke the API to update
  worker or job statuses as failed, but these API invocations will effect no
  state transitions on workers and jobs _already_ found to be in a terminal
  state.

## Workers

### Worker Success Observed by the Controller

- [ ] Invoke API to update worker status as `SUCCEEDED`.
- [ ] Delete the worker pod after 60 seconds. This gives the log aggregator time
  to capture all worker output.

__How will the API react to this?__

- [ ] This will only effect a state transition if the worker is not already in a
  __terminal state__.

### Worker Failure Observed by the Controller

The controller won't know _why_ the pod failed. i.e. If it was aborted, the
controller doesn't know that.

- [ ] Invoke API to update worker status as `FAILED`.
- [ ] Delete the worker pod after 60 seconds. This gives the log aggregator time
  to capture all worker output.

__How will the API react to this?__

- [ ] This will only effect a state transition if the worker is not already in a
  __terminal state__.

### Worker Timeout Observed by the Controller

- [ ] Invoke API to update worker status as `TIMED_OUT`.
- [ ] Delete the worker pod immediately. If we wait 60 seconds as we do upon
  worker completion, this worker might yet complete and then it won't really
  have timed out as we recorded.

__How will the API react to this?__

- [ ] This will only effect a state transition if the worker is not already in a
  __terminal state__.

### Worker Fails to Start

I'm not sure yet how to detect this condition.

### Worker Status is Updated by an API Call

- [ ] This will only effect a state transition if the worker is not already in a
  __terminal state__.
- [ ] If the worker reaches a terminal state, delete Kubernetes resources
  associated with the event, _except for pods_, which must remain around long
  enough for the log aggregator to capture all output.
  - [ ] Persistent volume claims
  - [ ] Config maps
  - [ ] Secrets

## Jobs

### Job Success Observed by the Controller

- [ ] Invoke API to update job status as `SUCCEEDED`.
- [ ] Delete the job pod after 60 seconds. This gives the log aggregator time
  to capture all job output.

__How will the API react to this?__

- [ ] This will only effect a state transition if the job is not already in a
  __terminal state__.

Note that worker state is unaffected by job state. Workers themselves define
how they will react to job success or failure. Worker outcome is thus determined
solely by the worker's own exit code.

### Job Failure Observed by the Controller

The controller won't know _why_ the pod failed. i.e. If it was aborted, the
controller doesn't know that.

- [ ] Invoke API to update job status as `FAILED`.
- [ ] Delete the job pod after 60 seconds. This gives the log aggregator time
  to capture all job output.

__How will the API react to this?__

- [ ] This will only effect a state transition if the job is not already in a
__terminal state__.

Note that worker state is unaffected by job state. Workers themselves define
how they will react to job success or failure. Worker outcome is thus determined
solely by the worker's own exit code.

### Job Timeout Observed by the Controller

- [ ] Invoke API to update job status as `TIMED_OUT`.
- [ ] Delete the job pod immediately. If we wait 60 seconds as we do upon job
  completion, this job might yet complete and then it won't really have timed
  out as we recorded.

__How will the API react to this?__

- [ ] This will only effect a state transition if the job is not already in a
__terminal state__.

Note that worker state is unaffected by job state. Workers themselves define
how they will react to job success or failure. Worker outcome is thus determined
solely by the worker's own exit code.

### Job Fails to Start

I'm not sure yet how to detect this condition.

### Job Status is Updated by an API Call

- [ ] This will only effect a state transition if the job is not already in a
  __terminal state__.
- [ ] If the job reaches a terminal state, delete Kubernetes resources
  associated with the job, _except for pods_, which must remain around long
  enough for the log aggregator to capture all output.
  - [ ] Persistent volume claims
  - [ ] Config maps
  - [ ] Secrets