package suite

import "time"

func text(value string) *string { return &value }

func baseExpected(id, outcome string, terminal *string) Observation {
	return Observation{
		Contract:      ContractVersion,
		CaseID:        id,
		Outcome:       outcome,
		TerminalState: terminal,
		Outputs:       []any{},
		Requests:      []string{},
		Checkpoints:   []int{},
		Evidence:      map[string]any{},
	}
}

func Catalog() []Case {
	const timeout = 30 * time.Second

	wf001 := baseExpected("WF-001", "ok", text("idle"))
	wf001.Outputs = []any{"HELLO"}
	wf001.Evidence = map[string]any{"executors": float64(2)}

	wf002 := baseExpected("WF-002", "error", nil)
	wf002.Error = &ObservedError{Category: "validation"}

	wf003 := baseExpected("WF-003", "error", text("failed"))
	wf003.Error = &ObservedError{Category: "max_iterations"}
	wf003.Evidence = map[string]any{"attempt_invocations": []any{float64(4), float64(4)}}

	wf004 := baseExpected("WF-004", "ok", text("idle"))
	wf004.Outputs = []any{"left(job)+right(job)"}

	wf005 := baseExpected("WF-005", "ok", text("idle"))
	wf005.Evidence = map[string]any{"runs": []any{[]any{}, []any{}}}

	wf006 := baseExpected("WF-006", "ok", text("idle"))
	wf006.Outputs = []any{"negative:-1", "fallback:1"}

	wf007 := baseExpected("WF-007", "ok", text("idle"))
	wf007.Outputs = []any{"2", "left+right"}
	wf007.Checkpoints = []int{1, 2}
	wf007.Evidence = map[string]any{"resumed_equals_uninterrupted": true}

	wf008 := baseExpected("WF-008", "error", text("failed"))
	wf008.Error = &ObservedError{Category: "graph_signature_mismatch"}

	wf009 := baseExpected("WF-009", "ok", text("idle_with_pending_requests"))
	wf009.Outputs = []any{"yes"}
	wf009.Requests = []string{"deploy:second"}
	wf009.Evidence = map[string]any{"initial_requests": float64(2), "remaining_requests": float64(1)}

	wf010 := baseExpected("WF-010", "error", text("idle"))
	wf010.Outputs = []any{"one", "two"}
	wf010.Error = &ObservedError{Category: "unknown_request"}
	wf010.Evidence = map[string]any{"rollback_verified": true}

	return []Case{
		{ID: "WF-001", Name: "chain_happy", Tags: []string{"core", "parity"}, Timeout: timeout, Expected: wf001},
		{ID: "WF-002", Name: "build_missing_start", Tags: []string{"validation", "parity"}, Timeout: timeout, Expected: wf002},
		{ID: "WF-003", Name: "max_iterations", Tags: []string{"runtime", "parity"}, Timeout: timeout, Expected: wf003},
		{ID: "WF-004", Name: "fan_out_fan_in", Tags: []string{"routing", "parity"}, Timeout: timeout, Expected: wf004},
		{ID: "WF-005", Name: "fresh_run_isolation", Tags: []string{"state", "parity"}, Timeout: timeout, Expected: wf005},
		{ID: "WF-006", Name: "switch_default_position", Tags: []string{"routing", "parity"}, Timeout: timeout, Expected: wf006},
		{ID: "WF-007", Name: "checkpoint_resume", Tags: []string{"checkpoint", "parity"}, Timeout: timeout, Expected: wf007},
		{ID: "WF-008", Name: "checkpoint_graph_mismatch", Tags: []string{"checkpoint", "validation", "parity"}, Timeout: timeout, Expected: wf008},
		{ID: "WF-009", Name: "request_partial_response", Tags: []string{"request", "parity"}, Timeout: timeout, Expected: wf009},
		{ID: "WF-010", Name: "request_batch_atomicity", Tags: []string{"request", "validation", "parity"}, Timeout: timeout, Expected: wf010},
	}
}
