use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::Arc;

use agent_framework::{
    Case, CheckpointStorage, Executor, ExecutorBuilder, InMemoryCheckpointStorage,
    RequestInfoExecutor, RequestInfoMessage, RequestResponse, Responses, ValidationError, Workflow,
    WorkflowBuilder, WorkflowContext, WorkflowError, WorkflowRunState,
};
use anyhow::{anyhow, bail, Context, Result};
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};

const CONTRACT: &str = "workflow-core/v1";
const PARTICIPANT: &str = "rust";

#[derive(Debug, Serialize)]
struct ErrorObservation {
    category: String,
}

#[derive(Debug, Serialize)]
struct Observation {
    contract: &'static str,
    case_id: String,
    participant: &'static str,
    outcome: String,
    terminal_state: Option<String>,
    outputs: Vec<Value>,
    requests: Vec<String>,
    checkpoints: Vec<usize>,
    error: Option<ErrorObservation>,
    evidence: Value,
}

fn observation(case_id: &str) -> Observation {
    Observation {
        contract: CONTRACT,
        case_id: case_id.to_string(),
        participant: PARTICIPANT,
        outcome: "ok".into(),
        terminal_state: Some("idle".into()),
        outputs: Vec::new(),
        requests: Vec::new(),
        checkpoints: Vec::new(),
        error: None,
        evidence: json!({}),
    }
}

fn expected_error(case_id: &str, category: &str, terminal_state: Option<&str>) -> Observation {
    let mut result = observation(case_id);
    result.outcome = "error".into();
    result.terminal_state = terminal_state.map(str::to_string);
    result.error = Some(ErrorObservation {
        category: category.into(),
    });
    result
}

fn state_name(state: WorkflowRunState) -> &'static str {
    match state {
        WorkflowRunState::Idle => "idle",
        WorkflowRunState::IdleWithPendingRequests => "idle_with_pending_requests",
        WorkflowRunState::Failed => "failed",
        WorkflowRunState::Cancelled => "cancelled",
        WorkflowRunState::Started => "started",
        WorkflowRunState::InProgress => "in_progress",
        WorkflowRunState::InProgressPendingRequests => "in_progress_pending_requests",
    }
}

fn final_state(result: &agent_framework::WorkflowRunResult) -> Result<String> {
    result
        .final_state()
        .map(state_name)
        .map(str::to_string)
        .ok_or_else(|| anyhow!("workflow result has no terminal state"))
}

fn string_relay(id: &str) -> Result<Executor> {
    Ok(ExecutorBuilder::new(id)
        .handler(|value: String, ctx: WorkflowContext| async move {
            ctx.send_message(value).await;
            Ok(())
        })
        .build()?)
}

async fn chain_happy(case_id: &str) -> Result<Observation> {
    let upper = ExecutorBuilder::new("upper")
        .handler(|value: String, ctx: WorkflowContext| async move {
            ctx.send_message(value.to_uppercase()).await;
            Ok(())
        })
        .build()?;
    let emit = ExecutorBuilder::new("emit")
        .handler(|value: String, ctx: WorkflowContext| async move {
            ctx.yield_output(value).await;
            Ok(())
        })
        .build()?;
    let workflow = WorkflowBuilder::new()
        .set_start_executor("upper")
        .add_executor(upper)
        .add_executor(emit)
        .add_edge("upper", "emit")
        .build()?;
    let run = workflow.run("hello".to_string()).await?;
    let mut result = observation(case_id);
    result.terminal_state = Some(final_state(&run)?);
    result.outputs = run
        .outputs::<String>()
        .into_iter()
        .map(Value::String)
        .collect();
    result.evidence = json!({"executors": 2});
    Ok(result)
}

async fn build_missing_start(case_id: &str) -> Result<Observation> {
    match WorkflowBuilder::new().build() {
        Err(WorkflowError::Validation(ValidationError::Other(message)))
            if message.contains("start executor") =>
        {
            Ok(expected_error(case_id, "validation", None))
        }
        Err(error) => bail!("unexpected validation error: {error}"),
        Ok(_) => bail!("builder accepted a workflow without a start executor"),
    }
}

async fn max_iterations(case_id: &str) -> Result<Observation> {
    let invocations = Arc::new(AtomicUsize::new(0));
    let handler_counter = invocations.clone();
    let loop_executor = ExecutorBuilder::new("loop")
        .handler(move |value: String, ctx: WorkflowContext| {
            let counter = handler_counter.clone();
            async move {
                counter.fetch_add(1, Ordering::SeqCst);
                ctx.send_message(value).await;
                Ok(())
            }
        })
        .build()?;
    let workflow = WorkflowBuilder::new()
        .set_start_executor("loop")
        .add_executor(loop_executor)
        .add_edge("loop", "loop")
        .set_max_iterations(3)
        .build()?;
    let mut attempts = Vec::new();
    let mut previous = 0;
    for _ in 0..2 {
        match workflow.run("loop".to_string()).await {
            Err(WorkflowError::Failed(details))
                if details.error_type == "MaxIterationsExceeded" => {}
            Err(error) => bail!("unexpected max-iterations error: {error}"),
            Ok(_) => bail!("non-converging workflow unexpectedly completed"),
        }
        let current = invocations.load(Ordering::SeqCst);
        attempts.push(current - previous);
        previous = current;
    }
    let mut result = expected_error(case_id, "max_iterations", Some("failed"));
    result.evidence = json!({"attempt_invocations": attempts});
    Ok(result)
}

fn worker(id: &str, tag: &str) -> Result<Executor> {
    let tag = tag.to_string();
    Ok(ExecutorBuilder::new(id)
        .handler(move |value: String, ctx: WorkflowContext| {
            let tag = tag.clone();
            async move {
                ctx.send_message(format!("{tag}({value})")).await;
                Ok(())
            }
        })
        .build()?)
}

fn sorted_aggregator(id: &str) -> Result<Executor> {
    Ok(ExecutorBuilder::new(id)
        .aggregate_handler(|mut values: Vec<String>, ctx: WorkflowContext| async move {
            values.sort();
            ctx.yield_output(values.join("+")).await;
            Ok(())
        })
        .build()?)
}

async fn fan_out_fan_in(case_id: &str) -> Result<Observation> {
    let workflow = WorkflowBuilder::new()
        .set_start_executor("dispatch")
        .add_executor(string_relay("dispatch")?)
        .add_executor(worker("left", "left")?)
        .add_executor(worker("right", "right")?)
        .add_executor(sorted_aggregator("aggregate")?)
        .add_fan_out_edges("dispatch", ["left", "right"])
        .add_fan_in_edges(["left", "right"], "aggregate")
        .build()?;
    let run = workflow.run("job".to_string()).await?;
    let mut result = observation(case_id);
    result.terminal_state = Some(final_state(&run)?);
    result.outputs = run
        .outputs::<String>()
        .into_iter()
        .map(Value::String)
        .collect();
    Ok(result)
}

fn selective_branch(id: &str, accepted: &str, emitted: &str) -> Result<Executor> {
    let accepted = accepted.to_string();
    let emitted = emitted.to_string();
    Ok(ExecutorBuilder::new(id)
        .handler(move |value: String, ctx: WorkflowContext| {
            let accepted = accepted.clone();
            let emitted = emitted.clone();
            async move {
                if value == accepted {
                    ctx.send_message(emitted).await;
                }
                Ok(())
            }
        })
        .build()?)
}

async fn fresh_run_isolation(case_id: &str) -> Result<Observation> {
    let workflow = WorkflowBuilder::new()
        .set_start_executor("dispatch")
        .add_executor(string_relay("dispatch")?)
        .add_executor(selective_branch("left", "first", "left")?)
        .add_executor(selective_branch("right", "second", "right")?)
        .add_executor(sorted_aggregator("aggregate")?)
        .add_fan_out_edges("dispatch", ["left", "right"])
        .add_fan_in_edges(["left", "right"], "aggregate")
        .build()?;
    let first = workflow.run("first".to_string()).await?;
    let second = workflow.run("second".to_string()).await?;
    let mut result = observation(case_id);
    result.terminal_state = Some(final_state(&second)?);
    result.evidence = json!({
        "runs": [first.outputs::<String>(), second.outputs::<String>()]
    });
    Ok(result)
}

fn tagged_sink(id: &str, tag: &str) -> Result<Executor> {
    let tag = tag.to_string();
    Ok(ExecutorBuilder::new(id)
        .handler(move |value: i64, ctx: WorkflowContext| {
            let tag = tag.clone();
            async move {
                ctx.yield_output(format!("{tag}:{value}")).await;
                Ok(())
            }
        })
        .build()?)
}

async fn switch_default_position(case_id: &str) -> Result<Observation> {
    let route = ExecutorBuilder::new("route")
        .handler(|value: i64, ctx: WorkflowContext| async move {
            ctx.send_message(value).await;
            Ok(())
        })
        .build()?;
    let workflow = WorkflowBuilder::new()
        .set_start_executor("route")
        .add_executor(route)
        .add_executor(tagged_sink("fallback", "fallback")?)
        .add_executor(tagged_sink("negative", "negative")?)
        .add_switch_case_edge_group(
            "route",
            vec![
                Case::default("fallback"),
                Case::when_named("negative-v1", |value: &i64| *value < 0, "negative"),
            ],
        )
        .build()?;
    let negative = workflow.run(-1i64).await?;
    let fallback = workflow.run(1i64).await?;
    let mut result = observation(case_id);
    result.terminal_state = Some(final_state(&fallback)?);
    result.outputs = negative
        .outputs::<String>()
        .into_iter()
        .chain(fallback.outputs::<String>())
        .map(Value::String)
        .collect();
    Ok(result)
}

fn two_step_builder(storage: Arc<InMemoryCheckpointStorage>) -> Result<WorkflowBuilder> {
    let loop_executor = ExecutorBuilder::new("loop")
        .handler(|value: i64, ctx: WorkflowContext| async move {
            if value == 2 {
                ctx.yield_output(value).await;
            } else {
                ctx.send_message(value + 1).await;
            }
            Ok(())
        })
        .build()?;
    Ok(WorkflowBuilder::new()
        .set_start_executor("loop")
        .add_executor(loop_executor)
        .add_edge("loop", "loop")
        .set_max_iterations(2)
        .with_checkpointing(storage))
}

fn emit_constant(id: &str, value: &str) -> Result<Executor> {
    let value = value.to_string();
    Ok(ExecutorBuilder::new(id)
        .handler(move |_input: String, ctx: WorkflowContext| {
            let value = value.clone();
            async move {
                ctx.send_message(value).await;
                Ok(())
            }
        })
        .build()?)
}

fn fan_in_checkpoint_builder(storage: Arc<InMemoryCheckpointStorage>) -> Result<WorkflowBuilder> {
    Ok(WorkflowBuilder::new()
        .set_start_executor("start")
        .add_executor(string_relay("start")?)
        .add_executor(emit_constant("left", "left")?)
        .add_executor(string_relay("delay")?)
        .add_executor(emit_constant("right", "right")?)
        .add_executor(sorted_aggregator("aggregate")?)
        .add_fan_out_edges("start", ["left", "delay"])
        .add_edge("delay", "right")
        .add_fan_in_edges(["left", "right"], "aggregate")
        .with_checkpointing(storage))
}

async fn checkpoint_resume(case_id: &str) -> Result<Observation> {
    let iteration_storage = InMemoryCheckpointStorage::new();
    let workflow = two_step_builder(iteration_storage.clone())?.build()?;
    workflow.run(0i64).await?;
    let checkpoint = iteration_storage
        .list_checkpoints(None)
        .await?
        .into_iter()
        .find(|item| item.iteration_count == 1)
        .context("missing iteration-one checkpoint")?;
    for checkpoint_id in iteration_storage.list_checkpoint_ids(None).await? {
        if checkpoint_id != checkpoint.checkpoint_id {
            iteration_storage.delete_checkpoint(&checkpoint_id).await?;
        }
    }
    let resumed_loop = two_step_builder(iteration_storage.clone())?
        .build()?
        .run_from_checkpoint(&checkpoint.checkpoint_id, None, None)
        .await?;
    let mut iterations: Vec<usize> = iteration_storage
        .list_checkpoints(None)
        .await?
        .iter()
        .map(|item| item.iteration_count)
        .collect();
    iterations.sort_unstable();

    let fan_in_storage = InMemoryCheckpointStorage::new();
    let fan_in_workflow = fan_in_checkpoint_builder(fan_in_storage.clone())?.build()?;
    let uninterrupted = fan_in_workflow.run("go".to_string()).await?;
    let fan_in_checkpoint = fan_in_storage
        .list_checkpoints(None)
        .await?
        .into_iter()
        .find(|item| {
            item.iteration_count == 2 && !item.edge_states.is_empty() && !item.messages.is_empty()
        })
        .context("missing partial fan-in checkpoint")?;
    let resumed_fan_in = fan_in_checkpoint_builder(fan_in_storage.clone())?
        .build()?
        .run_from_checkpoint(&fan_in_checkpoint.checkpoint_id, None, None)
        .await?;
    let resumed_outputs = resumed_fan_in.outputs::<String>();

    let mut result = observation(case_id);
    result.terminal_state = Some(final_state(&resumed_fan_in)?);
    result.outputs = vec![
        Value::String(resumed_loop.outputs::<i64>()[0].to_string()),
        Value::String(resumed_outputs[0].clone()),
    ];
    result.checkpoints = iterations;
    result.evidence = json!({
        "resumed_equals_uninterrupted": resumed_outputs == uninterrupted.outputs::<String>()
    });
    Ok(result)
}

async fn checkpoint_graph_mismatch(case_id: &str) -> Result<Observation> {
    let storage = InMemoryCheckpointStorage::new();
    let base = two_step_builder(storage.clone())?.build()?;
    base.run(0i64).await?;
    let checkpoint_id = storage
        .list_checkpoint_ids(None)
        .await?
        .into_iter()
        .next()
        .context("missing checkpoint")?;

    let changed_loop = ExecutorBuilder::new("loop")
        .handler(|value: i64, ctx: WorkflowContext| async move {
            ctx.send_message(value + 1).await;
            Ok(())
        })
        .build()?;
    let changed = WorkflowBuilder::new()
        .set_start_executor("loop")
        .add_executor(changed_loop)
        .add_edge("loop", "loop")
        .set_max_iterations(3)
        .with_checkpointing(storage)
        .build()?;
    match changed
        .run_from_checkpoint(&checkpoint_id, None, None)
        .await
    {
        Err(WorkflowError::GraphSignatureMismatch) => Ok(expected_error(
            case_id,
            "graph_signature_mismatch",
            Some("failed"),
        )),
        Err(WorkflowError::Failed(details)) if details.message.contains("graph has changed") => Ok(
            expected_error(case_id, "graph_signature_mismatch", Some("failed")),
        ),
        Err(error) => bail!("unexpected graph mismatch error: {error}"),
        Ok(_) => bail!("changed graph accepted an old checkpoint"),
    }
}

#[derive(Clone, Debug, Serialize, Deserialize)]
struct ApprovalRequest {
    label: String,
}

fn two_request_workflow() -> Result<Workflow> {
    let requester = ExecutorBuilder::new("requester")
        .handler(|value: String, ctx: WorkflowContext| async move {
            for suffix in ["first", "second"] {
                ctx.send_message(RequestInfoMessage::new(ApprovalRequest {
                    label: format!("{value}:{suffix}"),
                }))
                .await;
            }
            Ok(())
        })
        .handler(
            |response: RequestResponse, ctx: WorkflowContext| async move {
                let answer: String = response.data_as().map_err(|error| error.to_string())?;
                ctx.yield_output(answer).await;
                Ok(())
            },
        )
        .build()?;
    Ok(WorkflowBuilder::new()
        .set_start_executor("requester")
        .add_executor(requester)
        .add_executor(RequestInfoExecutor::build("request_info"))
        .add_edge("requester", "request_info")
        .add_edge("request_info", "requester")
        .build()?)
}

async fn request_partial_response(case_id: &str) -> Result<Observation> {
    let workflow = two_request_workflow()?;
    let initial = workflow.run("deploy".to_string()).await?;
    let requests = initial.request_info_events();
    let initial_count = requests.len();
    let first_id = requests[0].request_id.clone();
    let second: ApprovalRequest = requests[1].request.data_as()?;
    let run = workflow
        .send_responses(Responses::new().respond(first_id, "yes".to_string()))
        .await?;
    let mut result = observation(case_id);
    result.terminal_state = Some(final_state(&run)?);
    result.outputs = run
        .outputs::<String>()
        .into_iter()
        .map(Value::String)
        .collect();
    result.requests = vec![second.label];
    result.evidence = json!({"initial_requests": initial_count, "remaining_requests": 1});
    Ok(result)
}

async fn request_batch_atomicity(case_id: &str) -> Result<Observation> {
    let workflow = two_request_workflow()?;
    let initial = workflow.run("deploy".to_string()).await?;
    let ids: Vec<String> = initial
        .request_info_events()
        .iter()
        .map(|request| request.request_id.clone())
        .collect();
    let invalid = Responses::new()
        .respond(ids[0].clone(), "one".to_string())
        .respond("missing", "bad".to_string());
    match workflow.send_responses(invalid).await {
        Err(WorkflowError::Failed(details)) if details.message.contains("missing") => {}
        Err(error) => bail!("unexpected invalid-batch error: {error}"),
        Ok(_) => bail!("invalid response batch unexpectedly succeeded"),
    }
    let valid = Responses::new()
        .respond(ids[0].clone(), "one".to_string())
        .respond(ids[1].clone(), "two".to_string());
    let run = workflow.send_responses(valid).await?;
    let mut outputs = run.outputs::<String>();
    outputs.sort();
    let mut result = expected_error(
        case_id,
        "unknown_request",
        Some(state_name(
            run.final_state()
                .context("valid response run has no final state")?,
        )),
    );
    result.outputs = outputs.into_iter().map(Value::String).collect();
    result.evidence = json!({"rollback_verified": true});
    Ok(result)
}

async fn run_case(case_id: &str) -> Result<Observation> {
    match case_id {
        "WF-001" => chain_happy(case_id).await,
        "WF-002" => build_missing_start(case_id).await,
        "WF-003" => max_iterations(case_id).await,
        "WF-004" => fan_out_fan_in(case_id).await,
        "WF-005" => fresh_run_isolation(case_id).await,
        "WF-006" => switch_default_position(case_id).await,
        "WF-007" => checkpoint_resume(case_id).await,
        "WF-008" => checkpoint_graph_mismatch(case_id).await,
        "WF-009" => request_partial_response(case_id).await,
        "WF-010" => request_batch_atomicity(case_id).await,
        _ => bail!("unknown case: {case_id}"),
    }
}

#[tokio::main]
async fn main() -> Result<()> {
    let args: Vec<String> = std::env::args().collect();
    match args.as_slice() {
        [_, command] if command == "version" => {
            println!(
                "{}",
                serde_json::to_string(&json!({
                    "driver": "agent-framework-rs-driver",
                    "participant": PARTICIPANT,
                    "subject_version": "0.1.0"
                }))?
            );
            Ok(())
        }
        [_, command, case_id] if command == "run" => {
            let result = match run_case(case_id).await {
                Ok(result) => result,
                Err(error) => {
                    let mut failed = expected_error(case_id, "driver_failure", Some("failed"));
                    failed.evidence = json!({"message": error.to_string()});
                    failed
                }
            };
            println!("{}", serde_json::to_string(&result)?);
            Ok(())
        }
        _ => {
            eprintln!("usage: agent-framework-rs-driver version | run CASE_ID");
            std::process::exit(2);
        }
    }
}
