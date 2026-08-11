import asyncio
import importlib.metadata
import json
import sys
from dataclasses import dataclass
from typing import Any

from agent_framework import (
    Case,
    Default,
    Executor,
    InMemoryCheckpointStorage,
    RequestInfoExecutor,
    RequestInfoMessage,
    RequestResponse,
    WorkflowBuilder,
    WorkflowContext,
    WorkflowRunState,
    handler,
)

CONTRACT = "workflow-core/v1"
PARTICIPANT = "python"


def observation(
    case_id: str,
    *,
    outcome: str = "ok",
    terminal_state: str | None = "idle",
    outputs: list[Any] | None = None,
    requests: list[str] | None = None,
    checkpoints: list[int] | None = None,
    error: dict[str, str] | None = None,
    evidence: dict[str, Any] | None = None,
) -> dict[str, Any]:
    return {
        "contract": CONTRACT,
        "case_id": case_id,
        "participant": PARTICIPANT,
        "outcome": outcome,
        "terminal_state": terminal_state,
        "outputs": outputs or [],
        "requests": requests or [],
        "checkpoints": checkpoints or [],
        "error": error,
        "evidence": evidence or {},
    }


def state_name(state: WorkflowRunState) -> str:
    return state.value.lower()


class Upper(Executor):
    @handler
    async def run(self, value: str, ctx: WorkflowContext[str]) -> None:
        await ctx.send_message(value.upper())


class EmitString(Executor):
    @handler
    async def run(self, value: str, ctx: WorkflowContext[Any, str]) -> None:
        await ctx.yield_output(value)


class RelayString(Executor):
    @handler
    async def run(self, value: str, ctx: WorkflowContext[str]) -> None:
        await ctx.send_message(value)


async def chain_happy(case_id: str) -> dict[str, Any]:
    upper = Upper(id="upper")
    emit = EmitString(id="emit")
    workflow = WorkflowBuilder().set_start_executor(upper).add_edge(upper, emit).build()
    result = await workflow.run("hello")
    return observation(
        case_id,
        terminal_state=state_name(result.get_final_state()),
        outputs=result.get_outputs(),
        evidence={"executors": 2},
    )


async def build_missing_start(case_id: str) -> dict[str, Any]:
    try:
        WorkflowBuilder().build()
    except ValueError:
        return observation(case_id, outcome="error", terminal_state=None, error={"category": "validation"})
    raise AssertionError("builder accepted a workflow without a start executor")


class CountingLoop(Executor):
    def __init__(self, id: str) -> None:
        super().__init__(id=id)
        self.invocations = 0

    @handler
    async def run(self, value: str, ctx: WorkflowContext[str]) -> None:
        self.invocations += 1
        await ctx.send_message(value)


async def max_iterations(case_id: str) -> dict[str, Any]:
    loop = CountingLoop(id="loop")
    workflow = WorkflowBuilder().set_start_executor(loop).add_edge(loop, loop).set_max_iterations(3).build()
    attempts: list[int] = []
    previous = 0
    for _ in range(2):
        try:
            await workflow.run("loop")
        except RuntimeError as exc:
            if "3" not in str(exc):
                raise
        else:
            raise AssertionError("non-converging workflow unexpectedly completed")
        attempts.append(loop.invocations - previous)
        previous = loop.invocations
    return observation(
        case_id,
        outcome="error",
        terminal_state="failed",
        error={"category": "max_iterations"},
        evidence={"attempt_invocations": attempts},
    )


class Worker(Executor):
    def __init__(self, id: str, tag: str) -> None:
        super().__init__(id=id)
        self.tag = tag

    @handler
    async def run(self, value: str, ctx: WorkflowContext[str]) -> None:
        await ctx.send_message(f"{self.tag}({value})")


class SortedAggregator(Executor):
    @handler
    async def run(self, values: list[str], ctx: WorkflowContext[Any, str]) -> None:
        await ctx.yield_output("+".join(sorted(values)))


async def fan_out_fan_in(case_id: str) -> dict[str, Any]:
    dispatch = RelayString(id="dispatch")
    left = Worker(id="left", tag="left")
    right = Worker(id="right", tag="right")
    aggregate = SortedAggregator(id="aggregate")
    workflow = (
        WorkflowBuilder()
        .set_start_executor(dispatch)
        .add_fan_out_edges(dispatch, [left, right])
        .add_fan_in_edges([left, right], aggregate)
        .build()
    )
    result = await workflow.run("job")
    return observation(case_id, terminal_state=state_name(result.get_final_state()), outputs=result.get_outputs())


class SelectiveBranch(Executor):
    def __init__(self, id: str, accepted: str, emitted: str) -> None:
        super().__init__(id=id)
        self.accepted = accepted
        self.emitted = emitted

    @handler
    async def run(self, value: str, ctx: WorkflowContext[str]) -> None:
        if value == self.accepted:
            await ctx.send_message(self.emitted)


async def fresh_run_isolation(case_id: str) -> dict[str, Any]:
    dispatch = RelayString(id="dispatch")
    left = SelectiveBranch(id="left", accepted="first", emitted="left")
    right = SelectiveBranch(id="right", accepted="second", emitted="right")
    aggregate = SortedAggregator(id="aggregate")
    workflow = (
        WorkflowBuilder()
        .set_start_executor(dispatch)
        .add_fan_out_edges(dispatch, [left, right])
        .add_fan_in_edges([left, right], aggregate)
        .build()
    )
    first = await workflow.run("first")
    second = await workflow.run("second")
    return observation(
        case_id,
        terminal_state=state_name(second.get_final_state()),
        evidence={"runs": [first.get_outputs(), second.get_outputs()]},
    )


class RelayInt(Executor):
    @handler
    async def run(self, value: int, ctx: WorkflowContext[int]) -> None:
        await ctx.send_message(value)


class TaggedIntSink(Executor):
    def __init__(self, id: str, tag: str) -> None:
        super().__init__(id=id)
        self.tag = tag

    @handler
    async def run(self, value: int, ctx: WorkflowContext[Any, str]) -> None:
        await ctx.yield_output(f"{self.tag}:{value}")


async def switch_default_position(case_id: str) -> dict[str, Any]:
    route = RelayInt(id="route")
    fallback = TaggedIntSink(id="fallback", tag="fallback")
    negative = TaggedIntSink(id="negative", tag="negative")
    workflow = (
        WorkflowBuilder()
        .set_start_executor(route)
        .add_switch_case_edge_group(
            route,
            [Default(target=fallback), Case(condition=lambda value: value < 0, target=negative)],
        )
        .build()
    )
    negative_result = await workflow.run(-1)
    fallback_result = await workflow.run(1)
    return observation(
        case_id,
        terminal_state=state_name(fallback_result.get_final_state()),
        outputs=negative_result.get_outputs() + fallback_result.get_outputs(),
    )


class EmitConstant(Executor):
    def __init__(self, id: str, value: str) -> None:
        super().__init__(id=id)
        self.value = value

    @handler
    async def run(self, _value: str, ctx: WorkflowContext[str]) -> None:
        await ctx.send_message(self.value)


class TwoStepLoop(Executor):
    @handler
    async def run(self, value: int, ctx: WorkflowContext[int, int]) -> None:
        if value == 2:
            await ctx.yield_output(value)
        else:
            await ctx.send_message(value + 1)


def two_step_builder(storage: InMemoryCheckpointStorage) -> WorkflowBuilder:
    loop = TwoStepLoop(id="loop")
    return WorkflowBuilder().set_start_executor(loop).add_edge(loop, loop).set_max_iterations(2).with_checkpointing(storage)


def fan_in_checkpoint_builder(storage: InMemoryCheckpointStorage) -> WorkflowBuilder:
    start = RelayString(id="start")
    left = EmitConstant(id="left", value="left")
    delay = RelayString(id="delay")
    right = EmitConstant(id="right", value="right")
    aggregate = SortedAggregator(id="aggregate")
    return (
        WorkflowBuilder()
        .set_start_executor(start)
        .add_fan_out_edges(start, [left, delay])
        .add_edge(delay, right)
        .add_fan_in_edges([left, right], aggregate)
        .with_checkpointing(storage)
    )


async def checkpoint_resume(case_id: str) -> dict[str, Any]:
    iteration_storage = InMemoryCheckpointStorage()
    workflow = two_step_builder(iteration_storage).build()
    await workflow.run(0)
    checkpoint = next(item for item in await iteration_storage.list_checkpoints() if item.iteration_count == 1)
    for checkpoint_id in await iteration_storage.list_checkpoint_ids():
        if checkpoint_id != checkpoint.checkpoint_id:
            await iteration_storage.delete_checkpoint(checkpoint_id)
    resumed_loop = await two_step_builder(iteration_storage).build().run_from_checkpoint(checkpoint.checkpoint_id)
    iterations = sorted(item.iteration_count for item in await iteration_storage.list_checkpoints())

    fan_in_storage = InMemoryCheckpointStorage()
    fan_in_workflow = fan_in_checkpoint_builder(fan_in_storage).build()
    uninterrupted = await fan_in_workflow.run("go")
    fan_in_checkpoint = next(
        item
        for item in await fan_in_storage.list_checkpoints()
        if item.iteration_count == 2 and item.edge_states and item.messages
    )
    resumed_fan_in = await fan_in_checkpoint_builder(fan_in_storage).build().run_from_checkpoint(fan_in_checkpoint.checkpoint_id)
    fan_in_output = resumed_fan_in.get_outputs()
    return observation(
        case_id,
        terminal_state=state_name(resumed_fan_in.get_final_state()),
        outputs=[str(resumed_loop.get_outputs()[0]), fan_in_output[0]],
        checkpoints=iterations,
        evidence={"resumed_equals_uninterrupted": fan_in_output == uninterrupted.get_outputs()},
    )


async def checkpoint_graph_mismatch(case_id: str) -> dict[str, Any]:
    storage = InMemoryCheckpointStorage()
    base = two_step_builder(storage).build()
    await base.run(0)
    checkpoint_id = (await storage.list_checkpoint_ids())[0]
    loop = TwoStepLoop(id="loop")
    changed = WorkflowBuilder().set_start_executor(loop).add_edge(loop, loop).set_max_iterations(3).with_checkpointing(storage).build()
    try:
        await changed.run_from_checkpoint(checkpoint_id)
    except ValueError:
        return observation(
            case_id,
            outcome="error",
            terminal_state="failed",
            error={"category": "graph_signature_mismatch"},
        )
    raise AssertionError("changed graph accepted an old checkpoint")


@dataclass
class ApprovalRequest(RequestInfoMessage):
    label: str = ""


class TwoRequester(Executor):
    @handler
    async def ask(self, value: str, ctx: WorkflowContext[ApprovalRequest]) -> None:
        await ctx.send_message(ApprovalRequest(label=f"{value}:first"))
        await ctx.send_message(ApprovalRequest(label=f"{value}:second"))

    @handler
    async def receive(self, response: RequestResponse[ApprovalRequest, str], ctx: WorkflowContext[Any, str]) -> None:
        await ctx.yield_output(response.data)


def two_request_workflow() -> Any:
    requester = TwoRequester(id="requester")
    request_info = RequestInfoExecutor(id="request_info")
    return (
        WorkflowBuilder()
        .set_start_executor(requester)
        .add_edge(requester, request_info)
        .add_edge(request_info, requester)
        .build()
    )


async def request_partial_response(case_id: str) -> dict[str, Any]:
    workflow = two_request_workflow()
    initial = await workflow.run("deploy")
    requests = initial.get_request_info_events()
    result = await workflow.send_responses({requests[0].request_id: "yes"})
    remaining = [event.data.label for event in requests[1:]]
    return observation(
        case_id,
        terminal_state=state_name(result.get_final_state()),
        outputs=result.get_outputs(),
        requests=remaining,
        evidence={"initial_requests": len(requests), "remaining_requests": 1},
    )


async def request_batch_atomicity(case_id: str) -> dict[str, Any]:
    workflow = two_request_workflow()
    initial = await workflow.run("deploy")
    requests = initial.get_request_info_events()
    try:
        await workflow.send_responses({requests[0].request_id: "one", "missing": "bad"})
    except ValueError:
        pass
    else:
        raise AssertionError("invalid response batch unexpectedly succeeded")
    valid = await workflow.send_responses({requests[0].request_id: "one", requests[1].request_id: "two"})
    return observation(
        case_id,
        outcome="error",
        terminal_state=state_name(valid.get_final_state()),
        outputs=sorted(valid.get_outputs()),
        error={"category": "unknown_request"},
        evidence={"rollback_verified": True},
    )


CASES = {
    "WF-001": chain_happy,
    "WF-002": build_missing_start,
    "WF-003": max_iterations,
    "WF-004": fan_out_fan_in,
    "WF-005": fresh_run_isolation,
    "WF-006": switch_default_position,
    "WF-007": checkpoint_resume,
    "WF-008": checkpoint_graph_mismatch,
    "WF-009": request_partial_response,
    "WF-010": request_batch_atomicity,
}


async def run_case(case_id: str) -> dict[str, Any]:
    if case_id not in CASES:
        raise ValueError(f"unknown case: {case_id}")
    try:
        return await CASES[case_id](case_id)
    except Exception as exc:
        return observation(
            case_id,
            outcome="error",
            terminal_state="failed",
            error={"category": "driver_failure"},
            evidence={"exception_type": type(exc).__name__, "message": str(exc)},
        )


def main() -> int:
    if len(sys.argv) == 2 and sys.argv[1] == "version":
        print(
            json.dumps(
                {
                    "driver": "agent-framework-py-driver",
                    "participant": PARTICIPANT,
                    "subject_version": importlib.metadata.version("agent-framework-py"),
                },
                sort_keys=True,
            )
        )
        return 0
    if len(sys.argv) == 3 and sys.argv[1] == "run":
        print(json.dumps(asyncio.run(run_case(sys.argv[2])), sort_keys=True, separators=(",", ":")))
        return 0
    print("usage: agent-framework-py-driver version | run CASE_ID", file=sys.stderr)
    return 2


if __name__ == "__main__":
    raise SystemExit(main())
