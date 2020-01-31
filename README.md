# BrigNext

This is a **_VERY_** rough around the edges prototype to showcase some of my own
ideas that _could_ be included in Brigade 2.0 if they gain traction.

**_Do not run this in production._**

## Installing

BrigNext runs on top of plain old Brigade.

Start by installing Brigade as normal.

The following commands all use Helm 3.

```
$ helm repo add brigade https://brigadecore.github.io/charts
$ helm repo update
$ kubectl create namespace brigade
$ helm install brigade brigade/brigade -n brigade
```

Now install BrigNext on top of it. You can use all the default settings.

We assume you have Go installed. You're going to need it anyway in order to
build the CLI.

```
$ go get github.com/krancour/brignext
$ cd $GOAPTH/src/github.com/krancour/brignext
$ helm install brignext charts/brignext -n brigade
```

## Building the BrigNext CLI

Again, we assume you have Go installed.

```
$ go build -o bin/brignext ./cmd/brignext/
```

## Try BrigNext

The `brignext` CLI works, _mostly_ like the `brig` CLI you might already be used
to. The biggest difference is that it no longer communicates directly with
Kubernetes. Instead, it talks to an API server. This API server requires you to
authenticate. Fortunately creating and using credentials is easy.

(Note that authentication is implemented, but authorization is not. Any
registered user can do anything! **_We DID warn you not to run this in
production, didn't we?_**)

Find the public IP the API server:

```
$ kubectl get svc brignext-apiserver -n brigade \
  -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

```
$ bin/brignext register <public IP here> -u <username> -p <password>
```

Besides registering you, this will also automatically log you in.

Unlike the `brig` you may be used to, `brignext` does not create projects
using an interactive process. You create projects by submitting JSON. (Perhaps
we'll add YAML support soon as well.) This repo contains an example project
you can use:

```
$ bin/brignext project create projects/demo.json
```

From this point, everything else works (more or less) as you are accustomed to.

## How it Works

BrigNext _augments_ the behavior of the Brigade you're used to without actually
modifying it, and in doing so, highlights certain changes that are under
consideration to be formally implemented in Brigade 2.0.

* BrigNext introduces MongoDB as a new (tentative replacement) datastore.

    * When the BrigNext API writes anything to the new datastore, it is also
      writes it to the old datastore (Kubernetes secrets). This allows Brigade
      to continue functioning as normal.
    * All BrigNext API read operations only consult the new datastore.
    * When Brigade writes anything to the old datastore, the BrigNext controller
      detects the changes and syncs them to the new datastore.

* BrigNext tracks your jobs better.

    * Brigade, on its own, doesn't keep good track of jobs. A job is simply a
      Kubernetes pod and if and when a pod goes away, every trace of the job
      is lost-- including all of its logs.
    * When Brigade launches a job pod, the BrigNext controller writes job data
      to the new datastore-- meaning there is now a record of all jobs that
      persists beyond the lifetime of the corresponding pod.

* BrigNext persists your job logs.

    * BrigNext uses fluentd as a logging agent running on every node. This
      agent filters out logs that aren't from Brigade workers or jobs and
      stashes the ones that are in Mongo. This means worker and job logs now
      persist beyond the lifetimes of their corresponding pods.

* BrigNext cleans up your cluser.

    * Since BrigNext keeps better records of everything Brigade has done, there
      is no longer a need for build secrets, completed worker pods, and
      completed job pods to hang around in your cluster indefinitely. The
      BrigNext controller terminates each of these shortly after they outlive
      their usefulness.
