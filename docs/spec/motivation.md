# Motivation

The goal of the Knative Sources is to provide a common toolkit and API
framework for integrating with event producers and provide those events into 
the cluster.

By extending basic primitives like Job, Deployment, and custom resources like
Knative Sering Service, implementors can implement an event producer or importer
without requiring the implementor to introduce an additional custom controller.
