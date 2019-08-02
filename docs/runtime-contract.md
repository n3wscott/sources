similar but way simpler to this: https://github.com/knative/serving/blob/master/docs/runtime-contract.md

We can define the general shared runtime contract for Source the interface:

- `K_SINK` <-- URI to send to.
- `K_OUTPUT_FORMAT` <-- "binary" or "structured"
TODO: extra Sources stuff.
  
- Source 


DataPlane contract :

- HTTP POST of CloudEvents to `K_SINK`. (any version, please in `K_OUTPUT_FORMAT` if possible)

TODO: fill in details, add examples.


Each source will talk about how they expect to run ? JobSource is Source contract + will run as a k8s job to completion.  
