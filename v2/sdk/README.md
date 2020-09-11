# Brigade 2 SDK for Go

The Brigade 2 SDK for Go provides Go language bindings for the new Brigade 2
API.

This SDK remains under development, but has been made publicly availably sooner
rather than later in order to unblock Brigade contibutors who wish to work on
Brigade 2 compatible gateways (or other tools). It should be _relatively_
stable, but please do expect occassional, minor breakages at this juncture.

Brigade 2's own non-API components, including its scheduler and CLI, are also
consumers of this SDK.

## Quickstart

```console
$ go get github.com/brigadecore/brigade/v2/sdk
```

In your gateway code:

```golang
import "github.com/brigadecore/brigade/v2/sdk"

// ...

client, err := sdk.NewAPIClient(
	apiAddress, // The address of the Brigade 2 Prototpye API server, beginning with http:// or https//
	apiToken, // An API token obtained using the Brigade 2 Prototpye CLI
	insecure, // boolean indicating whether TLS errors (if applicable) are tolerated
)
if err != nil {
  // ...
}
```

`client` is an entrypoint into a _tree_ of specialized clients. At its highest
level, it's broken into:

| Function | Returns | Purpose |
|----------|---------|---------|
| `Authx()`| `authx.APIClient` | Manages `User`s, `ServiceAccount`s, and related concerns. |
| `Core()` | `core.APIClient` | Manages "core" Brigade componnets such as `Project`s and `Event`s. |
| `System()` | `system.APIClient` | Manages miscellaneous system-wide concerns. |

Each of these, in turn, provides access to even more specialized clients.

When developing code that integrates with a specific aspect of the Brigade API,
it makes sense to directly create the specific client you need, avoiding the
things you do not need. For instance, instantiating a `core.EventClient` only
would make sense for gateways (whose job is simply to broker events from
upstream systems):

```golang
import "github.com/brigadecore/brigade/v2/sdk/core"

// ...

client, err := core.NewEventsClient(
  apiAddress, // The address of the Brigade 2 Prototpye API server, beginning with http:// or https//
	apiToken, // An API token obtained using the Brigade 2 Prototpye CLI
	insecure, // boolean indicating whether TLS errors (if applicable) are tolerated
)
if err != nil {
  // ...
}
```

The SDK's godocs are quite thorough. Please explore those for further details.

## Using with the Brigade 2 Prototype

Visit [brigadecore/brigade](../README.md) for instructions
on standing up your own instance of the Brigade 2 Prototype, installing the
Brigade 2 Prototype CLI (`brig`), and authenticating.

Once you are set up and have authenticated, you may create a service account
whose token can be used in testing your new gateway (or other tool).

```console
$ brig service-account create --id <name> --description <description>
```

The command will return the token.

# Contributing

The Brigade 2 SDK for Go accepts contributions via GitHub pull requests. The
[Contributing](CONTRIBUTING.md) document outlines the process to help get your
contribution accepted.

# Support & Feedback

We have a slack channel!
[Kubernetes/#brigade](https://kubernetes.slack.com/messages/C87MF1RFD) Feel free
to join for any support questions or feedback, we are happy to help. To report
an issue or to request a feature open an issue
[here](https://github.com/brigadecore/brigade/issues)
