# Case matrix

| ID | Name | Tags | Observable invariant |
|---|---|---|---|
| WF-001 | `chain_happy` | `core,parity` | two public executors produce `HELLO` and idle |
| WF-002 | `build_missing_start` | `validation,parity` | missing start is a validation error |
| WF-003 | `max_iterations` | `runtime,parity` | both attempts consume the same fresh budget (initial delivery plus three supersteps) |
| WF-004 | `fan_out_fan_in` | `routing,parity` | two branches aggregate once |
| WF-005 | `fresh_run_isolation` | `state,parity` | partial fan-in never crosses fresh runs |
| WF-006 | `switch_default_position` | `routing,parity` | default is fallback even when declared first |
| WF-007 | `checkpoint_resume` | `checkpoint,parity` | fan-in state restores and iterations are `[1,2]` |
| WF-008 | `checkpoint_graph_mismatch` | `checkpoint,validation,parity` | changed graph rejects old checkpoint |
| WF-009 | `request_partial_response` | `request,parity` | one remaining request keeps pending terminal state |
| WF-010 | `request_batch_atomicity` | `request,validation,parity` | invalid batch rolls back before a valid batch succeeds |

All cases are hermetic, mandatory for both participants, and bounded by the
selected `--timeout` (30 seconds by default).
