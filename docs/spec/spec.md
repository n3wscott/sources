# Knative Source API Spec

## Resource YAML Definitions

### JobSource

```yaml
apiVersion: sources.knative.dev/v1alpha1
kind: JobSource
metadata:
  name: my-jobsource
spec:
  # A template is required.
  template:
    spec:
      # is a singleton []corev1.Container.
      containers: ...

  # Any spec fields valid for a Kubernetes Job are also valid for a JobSource.
  # The semantics are unchanged as these fields will be passed directly to the
  # underlying Job.
  backoffLimit: 2

  # is a corev1.ObjectReference.
  sink:
    apiVersion: v1
    kind: Service
    name: my-service

  # is either the string "structured" or "binary".
  outputFormat: "binary"
```

### CronJobSource

```yaml
apiVersion: sources.knative.dev/v1alpha1
kind: CronJobSource
metadata:
  name: my-cronjobsource
spec:
  sink:
    # Change the kind and name for the sink as desired.
    apiVersion: eventing.knative.dev/v1alpha1
    kind: Broker
    name: default
  # schedule is a required crontab schedule.
  schedule: "*/20 * * * *"
  # A jobTemplate is required.
  jobTemplate:
    spec:
      template:
        spec:
          # containers is a singleton []corev1.Container.
          containers:
          - name: my-container
            image: example.com/container
```

### ServiceSource

Also see [Knative Serving Service spec](https://github.com/knative/serving/blob/master/docs/spec/spec.md#service).

```yaml
apiVersion: sources.knative.dev/v1alpha1
kind: ServiceSource
metadata:
  name: my-servicesource
spec:
  sink:
    apiVersion: eventing.knative.dev/v1alpha1
    kind: Broker
    name: default
  template:
    spec:
      # containers is a list of containers to run.
      # As with Knative Services, you can run any number of containers
      # and distribute traffic between them.
      containers:
      - name: my-container
        image: example.com/container
  traffic:
```
