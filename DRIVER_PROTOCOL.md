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

Each captured stream has a 64 KiB (65,536-byte) budget, including whitespace.
Stdout exceeding that budget is an infrastructure error before JSON decoding;
the harness never validates a truncated response. Stderr has a separate budget:
its retained prefix is followed by `[stderr truncated after 65536 bytes]` when
oversized, and stderr truncation alone does not invalidate a successful response.
The capture writers keep consuming both pipes after their budgets are exhausted,
so truncation does not stop pipe draining. Existing subprocess cancellation is
unchanged, and overflow does not replace an existing process exit error.

Generated workflow IDs, checkpoint IDs, request IDs, timestamps, native enum
spellings, and native exception variants are normalized. All workflow behavior
must still come from the participant's public library API.
