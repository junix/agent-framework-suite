# agent-framework-suite

Hermetic cross-language E2E conformance suite for the sibling Python
`agent-framework` and Rust `agent-framework-rs` workflow runtimes.

The suite drives both installed libraries through public APIs, normalizes only
unstable IDs and native error shapes, checks each result against a shared
oracle, and then requires Python/Rust parity. Drivers do not implement routing,
checkpointing, request tracking, or iteration behavior.

## Quick start

```sh
just doctor
just list
just run WF-001
just check
```

`just test` is harness-only and never invokes a framework. `just check` builds
the Rust driver, checks both mandatory participants, and runs every hermetic
case. No case uses a model, network, ambient credential, or external service.

## CLI

```text
agent-framework-suite doctor [--json]
agent-framework-suite list [--json] [--select SPEC] [--tag TAG]
agent-framework-suite run [--json] [--select SPEC] [--tag TAG]
                          [--timeout DURATION] [--stream] [--report PATH]
agent-framework-suite version [--json]
```

Selectors accept a complete case ID, a prefix such as `WF`, a case name, or
`all`. Both participants are mandatory; a missing driver is an infrastructure
error, not a skip.

Exit codes are stable:

| Code | Meaning |
|---:|---|
| 0 | command succeeded and every selected case passed |
| 1 | one or more conformance/parity cases failed |
| 2 | invalid command, selector, flag, or report path |
| 3 | mandatory driver is missing, invalid, or timed out |

See [CONTRACT.md](CONTRACT.md), [CASES.md](CASES.md), and
[DRIVER_PROTOCOL.md](DRIVER_PROTOCOL.md) for the frozen behavioral and
transport contracts.
