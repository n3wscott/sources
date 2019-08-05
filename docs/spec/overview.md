# Sources

A **source** is any resource that generates or imports an event and relays that
event to another resource on the cluster via CloudEvent. Sourcing events is
critical to developing a distributed system that reacts to events.

A Source:
 - Produces CloudEvents in some way
 - Sends CloudEvents
   - to a **sink** that it is told about
   - with a requested encoding (structured or binary)

Having many Custom Resource Definitions that satisfy the Source spec allows
developers and operators to easily generate events using any language or
container. Each CRD can extend the interface of a Source in a unique way to add
functionality.

## JobSource

A JobSource uses a Kubernetes Job to send events. The Job generates events and
runs until completion. This is useful in scenarios where an operator might want
to replay past events, manually trigger an event, or run some batch processing
job that emits events.

## CronJobSource

A CronJobSource is similar to a JobSource with the added feature of running on a
schedule (described by a cron schedule expression). An operator might want to
use a CronJobSource for a task that needs to be completed regularly, such as
cleaning up a database and notifying other parts of the system, or sending out
scheduled/aggregated messages to users.

## ServiceSource

A ServiceSource is implemented using a Knative Service. This lets an operator
easily create a Source that is addressable. This allows developers and operators
to work with 3rd party event sources that communicate via webhooks with ease.

## DeploymentSource

A DeploymentSource creates a Deployment that can send events. This is a common
and simple use case for just about any kind of no-frills long-lived source.
