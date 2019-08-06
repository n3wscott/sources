Something like https://github.com/knative/serving/blob/master/docs/spec/overview.md

We need to define:

- Sources have:

spec.sink <-- ObjectRef
status.sinkUri <-- URI of resolved sink ref object.

More for Scott to worry about:
TODO: Extra Source runtime contract data,
- Source Id (given by external thingy) <-- maybe cloudevent format extensions. ()
Maybe:
  spec.outputExtentions <-- Map[string]string then this goes into the outbound events.
- CRD spec event types.
