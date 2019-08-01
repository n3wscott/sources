# Event Replay

This is a demo of an idiomatic use of a JobSource. An operator may want to
"replay" events that happened in the past for additional processing. A JobSource
provides a resource that runs some job until completion, publishing events to a
sink. This demo JobSource publishes all documents in a Firestore collection to
its sink.

## Prerequisites

You must have a Kubernetes cluster with Knative running. You need `kubectl` and
`ko` ([github.com/google/ko](https://github.com/google/ko)). Firestore will be
used, so you'll need a GCP account as well.

You must install the CRDs from this repository. From the root of this repo:
```bash
ko apply -f config/
```

## Set up Firestore

If you already have Firestore set up with a working service account, skip this.

1. Enable the Cloud Firestore API.
1. Create a service account for your database. For better security, set the role to **Datastore >
   Cloud Datastore User**. Any name is fine.
1. Create a key for your service account and download the JSON.

## Creating mock data

Make sure you have documents in Firestore that can be retrieved.
The project is configured to query a collection named "purchases" by default.

There is a program `gen-fake-events` in this directory that can automatically
populate your database.

## Running the event replay JobSource

Upload your database credentials to Kubernetes as a secret so that they are
available to your containers:
```bash
kubectl create secret generic db-creds --from-file=./path/to/credentials/db-svc-acct.json
```
Be sure to change the path to your actual credential file.

Open `event-replay.yaml` and make sure all the fields are correct:
1. The environment variable `FROM_COLLECTION` should be the collection you would
   like to query. You do not need to change this if you generated data with
   `gen-fake-events`.
1. The environment variable `GOOGLE_APPLICATION_CREDS_JSON` should be configured
   to match the credentials you uploaded in the previous step.
1. Set the `sink` to the Kubernetes object you would like to receive the events.
   The default is [sockeye](https://github.com/n3wscott/sockeye), which delivers
   events it receives to the browser via websocket. If you would also like to
   use sockeye, simply run
   ```bash
   kubectl apply -n default -f https://github.com/n3wscott/sockeye/releases/download/0.1.0/sockeye.yaml
   ```

Finally, create the event replay JobSource.
```bash
ko create -f event-replay
```

Monitor its status with `kubectl get jobsources`.
