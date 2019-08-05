Lets do https://github.com/knative/serving/blob/master/docs/spec/spec.md style.

# Knative Source API Spec

## Resource YAML Definitions

### JobSource

```yaml
apiVersion: sources.knative.dev/v1alpha1
kind: JobSource
metadata:
  name: my-jobsource
  labels:
    controller-tools.k8s.io: "1.0"
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
  sink: ...

  # is either the string "structured" or "binary".
  outputFormat: ...
```

### CronJobSource

TODO

### ServiceSource

TODO

### DeploymentSource

TODO
