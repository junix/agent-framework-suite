# Driver protocol

Every native driver exposes:

```text
<driver> version
<driver> run CASE_ID
```

`version` writes one JSON object containing `driver`, `participant`, and
`subject_version`. `run` writes exactly one `workflow-core/v1` observation to
stdout. Diagnostics go to stderr. Drivers receive a clean, bounded subprocess
environment from the harness and must not access the network.

Generated workflow IDs, checkpoint IDs, request IDs, timestamps, native enum
spellings, and native exception variants are normalized. All workflow behavior
must still come from the participant's public library API.
