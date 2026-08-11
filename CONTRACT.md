# workflow-core/v1 contract

## Scope and version

This document defines `workflow-core/v1`, the shared public-library contract
used to compare the Python `agent-framework` implementation with the Rust
`agent-framework-rs` implementation.

The contract covers only capabilities exposed by both public libraries:

- graph construction and validation;
- superstep execution and terminal state;
- fan-out and fan-in routing;
- switch-case routing;
- checkpoint creation and resume;
- human-in-loop request/response handling;
- fresh-run isolation after success or failure.

Python-only chat clients, MCP adapters, middleware, content types, and
observability integrations are outside this cross-language contract.

## Participant boundary

An E2E participant must invoke the installed library through its public API.
Suite-owned drivers may translate a language-neutral case into native builder
calls and normalize native events, errors, generated IDs, and checkpoint IDs.
They must not implement routing, checkpointing, request tracking, iteration
accounting, or other workflow behavior on behalf of the participant.

## Normalized observation

Each case produces one JSON object with these fields:

```json
{
  "contract": "workflow-core/v1",
  "case_id": "chain_happy",
  "participant": "python",
  "outcome": "ok",
  "terminal_state": "idle",
  "outputs": ["Hello, World!"],
  "requests": [],
  "checkpoints": [],
  "error": null,
  "evidence": {}
}
```

Rules:

- `outcome` is `ok` or `error`.
- `terminal_state` is `idle`, `idle_with_pending_requests`, `failed`, or
  `null` when construction fails before execution.
- `outputs` preserves workflow output order. A case may sort explicitly when
  the public contract declares branch completion order irrelevant.
- `requests` contains stable case-local labels, never generated request IDs.
- `checkpoints` contains stable observations such as iteration counts, never
  generated checkpoint IDs or timestamps.
- `error.category` is one of `validation`, `max_iterations`,
  `unknown_request`, `graph_signature_mismatch`, or `handler_failure`.
- `evidence` may contain deterministic case-specific state needed by an
  oracle. It must not contain paths, timings, random IDs, or timestamps.

## Required conformance cases

| Case ID | Public behavior | Required observation |
|---|---|---|
| `chain_happy` | A two-node chain transforms and yields one value. | `ok`, `idle`, one matching output. |
| `build_missing_start` | Building without a start executor is invalid. | `error.category=validation`. |
| `max_iterations` | A live cycle exceeds its configured superstep budget. | `error.category=max_iterations`; a later fresh run starts with a fresh budget. |
| `fan_out_fan_in` | One value is sent to two workers and aggregated once. | `ok`, `idle`, one aggregate containing both worker values. |
| `fresh_run_isolation` | Reusing a workflow does not reuse messages, fan-in buffers, iteration state, or prior outputs. | Two independent observations with no cross-run aggregation. |
| `switch_default_position` | Non-default predicates are evaluated before the single default regardless of where default was declared. | Matching input reaches the matching case, not default. |
| `checkpoint_resume` | Resume restores the complete run state and continues through the normal superstep boundary. | Resumed result equals uninterrupted result and observed iterations do not gain a hidden step. |
| `checkpoint_graph_mismatch` | A checkpoint cannot resume on a semantically different graph. | `error.category=graph_signature_mismatch`. |
| `request_partial_response` | Answering one of two requests leaves the other pending. | `idle_with_pending_requests`, one remaining request label. |
| `request_batch_atomicity` | A batch containing an unknown request ID changes no pending request and delivers no response. | `error.category=unknown_request`; both original requests remain pending. |

## Equality and parity

For every required case, compare the normalized records after removing only
the `participant` field. A parity failure is any difference in outcome,
terminal state, outputs, request labels, checkpoint observations, error
category, or deterministic evidence.

Capability absence is not a pass. Both participants are mandatory for
`workflow-core/v1`; a missing, skipped, or non-runnable participant fails the
suite.

## Suite entry points

The suite consuming this contract must expose:

- `doctor`: readiness and participant discovery;
- `list`: stable case IDs, tags, required participants, and timeout policy;
- `run`: one case or all cases, with console and JSON report modes;
- a harness-only test command that never invokes participants;
- a real E2E command that runs every required case against both participants
  and enforces parity.

The JSON report must include contract version, suite version, participant
versions, case results, parity results, and reproducible rerun commands.

The Python repository also carries this contract at
`spec/06-cross-language-e2e-contract.md`; changes to either copy must keep the
same `workflow-core/v1` semantics.
