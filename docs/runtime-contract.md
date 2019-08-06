similar but way simpler to this: https://github.com/knative/serving/blob/master/docs/runtime-contract.md

# Knative Sources Runtime Contract

## Background

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD",
"SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" are to be
interpreted as described in [RFC 2119](https://tools.ietf.org/html/rfc2119).

This document considers two users of a given Knative environment, and is
particularly concerned with the expectations of developers (and language and
tooling developers, by extension) running code in the environment.

 - **Developers** write code which is packaged into a container which is run on the
   Knative cluster.
   - **Language and tooling developers** typically write tools used by developers to
     package code into containers. As such, they are concerned that tooling
     which wraps developer code complies with this runtime contract.
 - **Operators** (also known as platform providers) provision the compute resources
   and manage the software configuration of Knative and the underlying
   abstractions (for example, Linux, Kubernetes, Istio, etc).


## Environment

Containers must be started with the following environment variables set:

| Name              | Value                                                |
| ---               | ---                                                  |
| `K_SINK`          | This will be a URI.                                  |
| `K_OUTPUT_FORMAT` | This will be one of either `structured` or `binary`. |

TODO: extra Sources stuff.

## Runtime & Lifecycle

Containers written by developers are subject to the following:
 - A container may be killed if the URI of the sink changes.
 - The container must send CloudEvents to the URI specified in `K_SINK`.
 - The container should send CloudEvents with structured or binary encoding
   matching `K_OUTPUT_FORMAT`.
 - The container should send CloudEvents over HTTP POST.
   - Note that the sink does not necessarily have to have the scheme `http` or
     `https`, but HTTP is the standard use case.
 - The container may use any version of CloudEvents.

TODO: fill in details, add examples.

## Sources

### JobSource

 - A JobSource will run the container as a Kubernetes Job. All configuration
   options available for Jobs are available in the JobSource spec.
 - A JobSource will run to completion (or however many completions are specified
   in the spec) and mark itself as succeeded. This state is terminal and no
   further action will take place.
 - A JobSource *will not* be killed and restarted when the sink changes after
   the Job has started.
   - Because a Job is meant to be a short-lived resource, sink changes will be
     ignored after the Job starts.

### CronJobSource

TBD

### ServiceSource

TBD

### DeploymentSource

TBD

Each source will talk about how they expect to run ? JobSource is Source contract + will run as a k8s job to completion.
