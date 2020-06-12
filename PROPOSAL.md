# Toward a Better Brigade

In the time since its first major release, Brigade has proven itself a useful
and [moderately
popular](https://www.cncf.io/wp-content/uploads/2020/03/CNCF_Survey_Report.pdf)
platform for achieving _"event-driven scripting for Kubernetes."_ Over that same
period, issues opened by our community and the maintainers' own efforts to "dog
food" Brigade have exposed certain issues and feature requests which cannot be
addressed without re-architecting the product and incurring some breaking
changes. _In other words, it is time to talk about Brigade 2.0._

In this document, the project's maintainers wish to present a formal proposal
for Brigade 2.0 and solicit feedback from and ratification by the broader
Brigade community.

## TL;DR

Brigade 2.0 is proposed to introduce a less Kubernetes-centric experience,
wherein users having little or no Kubernetes experience, or even those lacking
direct access to a Kubernetes cluster, can quickly become productive with
Brigade. Breaking changes are on the docket, but Brigade 2.0 should feel
familiar to anyone with previous Brigade experience. In short, the maintainers
are proposing Brigade's nuanced transition from "event-driven scripting for
Kubernetes" to "event driven scripting (for Kubernetes)."

## Motivations

To understand both the philosophical and technical drivers of the proposed
changes, a brief examination of Brigade's early design decisions, its
architecture, and a selection of existing issues is in order.

### Early Design Decisions & Architecture

Brigade was designed to be as lightweight and "cloud native" as possible,
consciously shunning third-party dependencies and relying solely on Kubernetes
wherever practical. To be sure, this was not without merit. By leveraging
Kubernetes `Secret` resources as a sort of makeshift document store and message
bus, Brigade's developers were spared from integrating with third-party
alternatives. Meanwhile, Brigade operators were spared from deploying and such
dependencies.

These principles resulted in a _very thin_ layer of abstraction between
Brigade's users and Kubernetes. The `brig` CLI, for instance, communicates
_directly_ with the Kubernetes API server using the user's own Kubernetes
cluster credentials. Similarly, gateways that broker events from external
systems like GitHub, Docker Hub, or Slack (to name a few), communicate directly
with the Kubernetes API server using Kubernetes cluster service accounts.

Brigade owes much of its early success to its lightweight nature, but many
outstanding issues can also be traced back to these early design decisions,
which the maintainers now propose to revist.

### Usability Issues

Usability issues have been a common complaint. By way of example, anyone who has
created a Brigade project through the interactive `brig project create` command
is likely to have found that experience clumsy. In the event of later wishing to
revise one's project definition, users discover that the `brig` CLI exposes no
`brig project update` command. A novice user might issue a `brig project create
--replace` command to repeat the onerous creation process. Meanwhile, an
intrepid Kubernetes user is apt to directly edit the `Secret` resource backing
the project definition. Maintainers are aware of many such users having fallen
back on Helm or loose Kubernetes manifests to manage Brigade projects.

To be fair, this issue is a UX one and could possibly be remediated without
re-architecting the entire product, but it is noteworthy because it illustrates
just how thin Brigade's abstraction between the user and Kubernetes is. While a
seasoned Kubernaut might perceive the ability to "drop down" to Kubernetes on an
as-needed basis to be a feature, a novice Kubernetes user is more likely to
perceive that UX in more perilous terms-- falling through thin ice.

In the view of the maintainers, permitting Kubernetes expertise to remain a _de
facto_ prerequisite for success with Brigade creates an undesired barrier to
adoption by a broader population of developers who may lack Kubernetes expertise
but could otherwise find value in Brigade.

Refer to:

* [Issue #1044](https://github.com/brigadecore/brigade/issues/1044)

### Security Risks

Brigade's abstraction between users and Kubernetes being as thin as it is poses
a security risk. Because the `brig` CLI is useless without Kubernetes cluster
credentials, all Brigade users must have such credentials and, minimally, be
granted read access to `Secret` resources within Brigade's namespace.

With the maintainers wishing for Kubernetes expertise not to remain a _de facto_
prerequisite for success with Brigade, it is prudent to contemplate the folly of
granting any level of cluster access to a novice user. Even if Kubernetes
expertise were to remain a given, there is obvious risk in granting even _read_
access to `Secret` resources within Brigade's namespace, since it is likely to
host projects _other than the users' own._ With such permissions, users too
easily gain access to one another's project-level secrets and the inherent risk
in that cannot be overstated.

N.B. 1: The current workaround for this is to run a separate Brigade instance
for each project or logical group of projects.

N.B. 2: Isolating each Brigade project to its own namespace has been among
Brigade's most requested features. By the maintainers' own admission, this
feature was intended to be included in the first major release. The fact that it
does not work is a persistent bug that cannot likely be remediated without
breaking changes.

Refer to:

* [Issue #755](https://github.com/brigadecore/brigade/issues/755)

### Data Loss

Reiterating that Brigade leverages Kubernetes as a makeshift data store, it is
worth examining whether it fills that role well or whether possible alteratives
are due some consideration.

Usability issues and security risks notwithstanding, persisting Brigade projects
as `Secret` resources seems sensible. Yet, projects are not the only Brigade
objects that are backed by a Kubernetes resources. Like projects, events are
backed by `Secret` resources. (This is discussed in more detail in the next
section.) _Workers_ that process those events, as well as the _jobs_ those
workers may fan out to, are backed by the Kubernetes pods in which they were
executed.

A worker or job's logs are found nowhere except within the corresponding pod.
CLI commands such as `brig logs` merely retrieve those logs via the Kubernetes
API. Even a worker or job's state is determined solely by the corresponding
pod's state, and a worker or job's very existence is coupled to the existence of
the corresponding pod. Should a pod be deleted, all record of the corresponding
worker or job is deleted with it.

The above is especially problematic when considering the array of circumstances
beyond any user's control in which a pod might be deleted. If the Kubernetes
node that hosted a given pod were to be decomissioned, for instance, all record
of the corresponding worker or job would vanish without a trace. A pod evicted
from its node by the Kubelet, for any number of reasons, would have the same
result.

While it is easy to imagine how the potential for data loss might preclude
Brigade's use by enterprises with strict data retention policies, all users
would do well to consider the ramifications of the above before committing to
use Brigade for any critical process.

Refer to:

* [Issue #960](https://github.com/brigadecore/brigade/issues/960)
* [Issue #1034](https://github.com/brigadecore/brigade/issues/1034)
* [Issue #1035](brigade/issues/1035)
* [Issue #1037](https://github.com/brigadecore/brigade/issues/1037)

### Resource Contention

Use of Kubernetes as a makeshift message bus for event delivery has also been
found to be problematic. Brigade's controller component monitors Brigade's
namespace for `Secret` resources that represent new events. When found, it
launches a worker pod in response. If a large volume of events are created at
once, a large volume of worker pods are also launched in rapid succession-- each
potentially fanning out and creating multiple job pods.

How events unfold depends on the extent to which the good practice of specifying
resource limits for each worker and job has been observed. If not observed
widely, the amount of work scheduled in the cluster may effectively be
unbounded. This may result in numerous pod evictions as resources become scarce.
(The previous section describes how that is problematic in its own right.)

If resource limits have been specified, situations may be encountered wherein
worker pods launch, consume all available resources, and cannot spawn job pods
due to resource scarcity. This can deadlock a cluster until workers begin timing
out-- perhaps only to be replaced with new workers that will encounter the same
conditions.

It is clear that the Brigade controller would benefit from the ability to
_throttle_ the number of concurrent events that may be processed, perhaps on a
per-project basis as well as cluster-wide. It is difficult to imagine how this
might be accomplished whilst utilizing Kubernetes as a makeshift message bus,
but is easy to implement using any of several alternatives.

Refer to:

* [Issue #1034](https://github.com/brigadecore/brigade/issues/1034)
* [Issue #1036](https://github.com/brigadecore/brigade/issues/1036)

### No Formal Project Discovery

Brigade does not currently provide gateways with any explicit method of
discovering projects that are subscribed to an inbound event. A project's
subscription to any particular set of events is implied by its name. For
instance, the GitHub gateway emits events from GitHub into Brigade for (at most)
one project whose name must precisely match the fully qualified name of the
repository. For example, the Brigade project named `krancour/demo` is implicitly
subscribed to events originating from the `krancour/demo` repository on GitHub.
This unfortunately precludes multiple projects from subscribing to events from a
single GitHub repository. It also precludes a single project from subscribing to
events brokered by multiple gateways that make differing assumptions about how
projects are named.

Refer to:

* [Issue #500](https://github.com/brigadecore/brigade/issues/500)

### Limited Support for Alternative Worker Images

The question of support for workflows defined using something other than
JavaScript (such as [YAML](https://github.com/brigadecore/brigade/issues/1024)
or [TypeScript](https://github.com/brigadecore/brigade/issues/527)) is a
perrenial favorite among community members and maintainers alike. This has
already been shown to be possible simply by utilizing a custom worker image that
behaves similarly enough to Brigade's default, but whose behavior is defined
differently. This concept was even demonstrated live at KubeCon 2019 talk.

Though it has been proven possible, the method of achieving this remains
cumbersome and unintuitive from a UX standpoint because Brigade makes
unnecessary assumptions that do not hold for all custom worker images. It is,
for instance, confusing to encounter a project that embeds as default
`brigade.js` file that actually contains YAML or TypeScript.

### The GitHub Tangle

Brigade has always intended to remain entirely agnostic with respect to both
event gateways and upstream event sources-- treating all events the same and
implementing no gateway-specific or source-specific support in its core.

While this principle was and remains commendable, Brigade has failed to enforce
a clean separation between itself and GitHub. GitHub-specific fields exist
within the Brigade `Project` type. The maintainers do wish to see this undone.

Refer to:

* [project.go](https://github.com/brigadecore/brigade/blob/8ffeb29d3e8133826851d8c7471915b7c50fe412/pkg/brigade/project.go#L31-L34)

### Job Sidecars

Support for "sidecar" containers in job pods is a feature the maintainers have
wished to add for some time. To date, this has not been possible. Kubernetes
supports multiple containers per pod, but does not permit any one container to
be designated as "primary" and treated differently than the others. Thus,
there's no way to shut down all containers and conclude a pod's lifecycle once
one particular container has exited. A workaround for this is to monitor pods
for completion of a particular container and then delete the entire pod. To
date, Brigade has been unable to take advantage of that workaround since all job
state and logs are coupled to the continued existence of its corresponding pod.
If the decision to utilize Kubernetes as a makeshift data store is revisited, so
too can the possibility of support for job sidecars.

Refer to:

* [Issue #340](https://github.com/brigadecore/brigade/issues/340)

### Philosophical Issues

Our final section outlining the motivations driving Brigade 2.0 tackles two
purlely philosophical course corrections.

First, Brigade currently refers to individual executions of user-defined
workflows as "builds." This is a misnomer which the maintainers regret, as it
strongly suggests that Brigade is a CI platform, which it is not (although it
_can_ be used to effect CI). Equally regretable is that the term "build" has too
often _also_ been used synonamously with "event." The maintainers propose to
strike "build" from Brigade's lexicon, speaking generically, only in terms of
events, each of which has a worker that _handles_ the event.

Second, prior guidance from Brigade documentation and maintainers has been that
Brigade workers should ideally perform no work beyond coordinating the execution
of jobs, to which actual work is delegated. While this is a useful pattern for
many use cases that are well served by serialized or concurrent execution of
multiple containerized tasks, it is a pattern whose additional overhead (in the
form of container start time) probably underserves very simple use cases or use
cases requiring minimized latency. To that end, the maintainers move to no
longer discourage Brigade workers from doing actual work.

## Proposed Path Forward

### Guiding Principles

The maintainers propose several guiding principles for the design and
development of Brigade 2.0:

1. Neither Kubernetes expertise nor cluster credentials must be a prerequisite
   for success.
    1. Some degree of freedom to "drop down" to Kubernetes may be retained to
       enable advanced / fringe use cases, but the need to utilize this should
       be _exceptional._

1. The UX must be both simplified and improved while striving to remain
   recognizable and comfortable for experienced Brigade users.

1. Security must be improved beyond the status quo.

1. Product evaluation (aka "tire kicking") / local use for development must not
   be onerous.
    1. For any new dependencies, a default Brigade installation must provision
       _free, in-cluster instances running in configurations that limit resource
       use_. Documentation may offer guidance on optimizing installations for
       production use.
    1. Security improvements must not encumber evaluation.

1. Small compromises are deemed acceptable if they remediate larger issues. By
   way of example, logs that stream more slowly than Brigade 1.x's are
   acceptable if they're less susceptible to data loss.

1. Design cues should be taken from Kubernetes (or other familiar sources)
   wherever possible in order to produce architecture, tooling, APIs, etc. that
   feel familiar to both contributors and users and incur minimal learning
   curve.

### High Level Architecture

With a few cross-cutting concerns such as access control notwithstanding, the
proposed architecture fundamentally decomposes Brigade into three distinct
sub-systems-- record keeping, event handling, and log aggregation with a service
layer to coordinate among these three. The service layer will be exposed to
clients via secure HTTP endpoints that implement a RESTful API. An API client
will be made available as a Go package.

N.B.: Brigade 1.x has a read-only API that is utilized by Kashti only. (Kashti
is the web-based, read-only Brigade dashboard.) The API referenced above is a
_new_ one for use by all Brigade 2.0 components, gateways, and clients so as to
abstract all of those away from underlying technology choices.

### Record Keeping

### Event Handling

### Log Aggregation

### Securing the API

### Non-Goals

### Unknowns

### Early Prototype

### Development Approach

### Roadmap