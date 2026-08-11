package suite

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

// fakeClient is a configurable ParticipantClient that records its interactions.
type fakeClient struct {
	participant string
	result      Observation
	err         error
	runCalls    int
	caseIDs     []string
	lastCtx     context.Context
}

func (client *fakeClient) ParticipantName() string { return client.participant }

func (client *fakeClient) Run(ctx context.Context, caseID string) (Observation, error) {
	client.runCalls++
	client.caseIDs = append(client.caseIDs, caseID)
	client.lastCtx = ctx
	if client.err != nil {
		return Observation{}, client.err
	}
	return client.result, nil
}

// multiCaseClient returns a per-case observation, so runs over several distinct
// cases can be exercised without reusing a single fixed observation.
type multiCaseClient struct {
	participant string
	results     map[string]Observation
}

func (client *multiCaseClient) ParticipantName() string { return client.participant }

func (client *multiCaseClient) Run(_ context.Context, caseID string) (Observation, error) {
	return client.results[caseID], nil
}

// passingObservation builds the observation a well-behaved participant would
// return for a case: the catalog oracle with the participant identity set.
func passingObservation(item Case, participant string) Observation {
	obs := item.Expected
	obs.Participant = participant
	return obs
}

// ---- Catalog ----

func TestCatalogHasStableUniqueOrderedIDs(t *testing.T) {
	cases := Catalog()
	if len(cases) != 10 {
		t.Fatalf("got %d cases, want 10", len(cases))
	}
	seen := map[string]bool{}
	for index, item := range cases {
		if seen[item.ID] {
			t.Fatalf("duplicate case id %s", item.ID)
		}
		seen[item.ID] = true
		wantID := fmt.Sprintf("WF-%03d", index+1)
		if item.ID != wantID {
			t.Fatalf("case %d id=%q want %q", index, item.ID, wantID)
		}
		if item.Name == "" {
			t.Fatalf("case %s has empty name", item.ID)
		}
		if item.Timeout != 30*time.Second {
			t.Fatalf("case %s timeout=%v want 30s", item.ID, item.Timeout)
		}
		if !contains(item.Tags, "parity") {
			t.Fatalf("case %s missing parity tag: %v", item.ID, item.Tags)
		}
		// Contract self-consistency: the oracle echoes the case identity.
		if item.Expected.Contract != ContractVersion {
			t.Fatalf("case %s expected contract=%q want %q", item.ID, item.Expected.Contract, ContractVersion)
		}
		if item.Expected.CaseID != item.ID {
			t.Fatalf("case %s expected CaseID=%q want %q", item.ID, item.Expected.CaseID, item.ID)
		}
	}
}

func TestCatalogOracleFieldsArePopulatedAndConsistent(t *testing.T) {
	for _, item := range Catalog() {
		// baseExpected initializes these to non-nil empties; every oracle must keep them populated.
		if item.Expected.Outputs == nil {
			t.Fatalf("case %s Outputs nil", item.ID)
		}
		if item.Expected.Requests == nil {
			t.Fatalf("case %s Requests nil", item.ID)
		}
		if item.Expected.Checkpoints == nil {
			t.Fatalf("case %s Checkpoints nil", item.ID)
		}
		if item.Expected.Evidence == nil {
			t.Fatalf("case %s Evidence nil", item.ID)
		}
		// Outcome is binary; an error outcome must carry a categorized error.
		switch item.Expected.Outcome {
		case "ok":
			if item.Expected.Error != nil {
				t.Fatalf("case %s ok outcome must not carry error: %+v", item.ID, item.Expected.Error)
			}
		case "error":
			if item.Expected.Error == nil || item.Expected.Error.Category == "" {
				t.Fatalf("case %s error outcome missing error category", item.ID)
			}
		default:
			t.Fatalf("case %s unexpected outcome %q", item.ID, item.Expected.Outcome)
		}
	}
}

func TestCatalogSpecificCaseContent(t *testing.T) {
	byID := map[string]Case{}
	for _, item := range Catalog() {
		byID[item.ID] = item
	}
	if got := byID["WF-001"].Expected.Outputs; len(got) != 1 || got[0] != "HELLO" {
		t.Fatalf("WF-001 outputs=%v want [HELLO]", got)
	}
	if byID["WF-002"].Expected.TerminalState != nil {
		t.Fatalf("WF-002 terminal state should be nil, got %q", *byID["WF-002"].Expected.TerminalState)
	}
	if got := byID["WF-002"].Expected.Error; got == nil || got.Category != "validation" {
		t.Fatalf("WF-002 error=%v want validation", got)
	}
	if got := byID["WF-003"].Expected.Error; got == nil || got.Category != "max_iterations" {
		t.Fatalf("WF-003 error=%v want max_iterations", got)
	}
	if got := byID["WF-007"].Expected.Checkpoints; len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("WF-007 checkpoints=%v want [1 2]", got)
	}
	if got := byID["WF-008"].Expected.Error; got == nil || got.Category != "graph_signature_mismatch" {
		t.Fatalf("WF-008 error=%v want graph_signature_mismatch", got)
	}
	if got := byID["WF-009"].Expected.Requests; len(got) != 1 || got[0] != "deploy:second" {
		t.Fatalf("WF-009 requests=%v want [deploy:second]", got)
	}
	wf009 := byID["WF-009"].Expected
	if wf009.Outcome != "ok" || wf009.TerminalState == nil || *wf009.TerminalState != "idle_with_pending_requests" {
		t.Fatalf("WF-009 outcome/terminal drift: %+v", wf009)
	}
	if got := byID["WF-010"].Expected.Error; got == nil || got.Category != "unknown_request" {
		t.Fatalf("WF-010 error=%v want unknown_request", got)
	}
}

// ---- Select ----

func TestSelectEmptySpecsReturnsAllInOrder(t *testing.T) {
	selected, err := Select(Catalog(), nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 10 {
		t.Fatalf("got %d, want 10", len(selected))
	}
	for index, item := range selected {
		if item.ID != Catalog()[index].ID {
			t.Fatalf("selection order changed at %d: %s", index, item.ID)
		}
	}
}

func TestSelectAllKeywordReturnsAll(t *testing.T) {
	selected, err := Select(Catalog(), []string{"all"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 10 {
		t.Fatalf("all matched %d, want 10", len(selected))
	}
}

func TestSelectIDPrefixMatchesEveryCase(t *testing.T) {
	// "WF" is the prefix before "-" in every "WF-NNN" id, so it selects all.
	selected, err := Select(Catalog(), []string{"WF"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 10 {
		t.Fatalf("prefix WF matched %d, want 10", len(selected))
	}
}

func TestSelectIsCaseInsensitive(t *testing.T) {
	selected, err := Select(Catalog(), []string{"wf-001"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 1 || selected[0].ID != "WF-001" {
		t.Fatalf("case-insensitive match failed: %+v", selected)
	}
}

func TestSelectByName(t *testing.T) {
	selected, err := Select(Catalog(), []string{"fan_out_fan_in"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 1 || selected[0].ID != "WF-004" {
		t.Fatalf("name match failed: %+v", selected)
	}
}

func TestSelectCommaSeparatedDedupsAndSkipsBlanks(t *testing.T) {
	selected, err := Select(Catalog(), []string{"WF-001, WF-002 ,, WF-001"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 2 || selected[0].ID != "WF-001" || selected[1].ID != "WF-002" {
		t.Fatalf("comma selection=%+v", selected)
	}
}

func TestSelectPreservesCatalogOrderRegardlessOfSpecOrder(t *testing.T) {
	selected, err := Select(Catalog(), []string{"WF-004", "WF-001"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 2 || selected[0].ID != "WF-001" || selected[1].ID != "WF-004" {
		t.Fatalf("order not preserved: %+v", selected)
	}
}

func TestSelectTagFiltersResult(t *testing.T) {
	selected, err := Select(Catalog(), []string{"WF-001", "WF-002"}, "core")
	if err != nil {
		t.Fatal(err)
	}
	// WF-001 is "core", WF-002 is "validation"; only core survives.
	if len(selected) != 1 || selected[0].ID != "WF-001" {
		t.Fatalf("tag filter result=%+v", selected)
	}
}

func TestSelectTagEmptiesResult(t *testing.T) {
	_, err := Select(Catalog(), []string{"WF-001"}, "nonexistent-tag")
	if err == nil || !strings.Contains(err.Error(), "selection is empty") {
		t.Fatalf("expected selection is empty error, got %v", err)
	}
}

func TestSelectUnknownSpecErrors(t *testing.T) {
	_, err := Select(Catalog(), []string{"missing"}, "")
	if err == nil || !strings.Contains(err.Error(), `selector "missing" matched no cases`) {
		t.Fatalf("expected matched-no-cases error, got %v", err)
	}
}

// ---- ValidateObservation / Equivalent / normalized ----

func TestValidateObservationIdentityMismatch(t *testing.T) {
	item := Catalog()[0]
	obs := passingObservation(item, "python")
	obs.Participant = "rust" // wrong identity
	proof := ValidateObservation(item, obs, "python")
	if proof == "" || !strings.Contains(proof, "identity mismatch") {
		t.Fatalf("expected identity mismatch proof, got %q", proof)
	}
}

func TestValidateObservationOracleMismatchOnOutcome(t *testing.T) {
	item := Catalog()[0]
	obs := passingObservation(item, "python")
	obs.Outcome = "error"
	proof := ValidateObservation(item, obs, "python")
	if proof == "" || !strings.Contains(proof, "oracle mismatch") {
		t.Fatalf("expected oracle mismatch proof, got %q", proof)
	}
}

func TestValidateObservationOracleMismatchOnOutputs(t *testing.T) {
	item := Catalog()[0]
	obs := passingObservation(item, "python")
	obs.Outputs = []any{"WRONG"}
	proof := ValidateObservation(item, obs, "python")
	if proof == "" || !strings.Contains(proof, "oracle mismatch") {
		t.Fatalf("expected oracle mismatch proof for output drift, got %q", proof)
	}
}

func TestValidateObservationMatchReturnsEmpty(t *testing.T) {
	item := Catalog()[0]
	obs := passingObservation(item, "python")
	if proof := ValidateObservation(item, obs, "python"); proof != "" {
		t.Fatalf("expected empty proof, got %q", proof)
	}
}

func TestNormalizedClearsParticipant(t *testing.T) {
	obs := Observation{Participant: "python", Outcome: "ok"}
	got := normalized(obs)
	if got.Participant != "" {
		t.Fatalf("normalized kept participant %q", got.Participant)
	}
}

func TestEquivalentIgnoresParticipantOnly(t *testing.T) {
	left := Observation{Participant: "python", Outcome: "ok", Outputs: []any{}, Requests: []string{}, Checkpoints: []int{}, Evidence: map[string]any{}}
	right := Observation{Participant: "rust", Outcome: "ok", Outputs: []any{}, Requests: []string{}, Checkpoints: []int{}, Evidence: map[string]any{}}
	if !Equivalent(left, right) {
		t.Fatal("observations differing only in participant should be equivalent")
	}
}

func TestEquivalentDetectsOutputDrift(t *testing.T) {
	left := Observation{Participant: "python", Outputs: []any{"A"}}
	right := Observation{Participant: "rust", Outputs: []any{"B"}}
	if Equivalent(left, right) {
		t.Fatal("different outputs should not be equivalent")
	}
}

// ---- Run / runCase ----

func TestRunEmptyCasesReportsEmptyAndInfrastructureOK(t *testing.T) {
	report, infrastructureOK := Run(nil, nil, map[string]ParticipantVersion{}, RunOptions{})
	if !infrastructureOK {
		t.Fatal("empty run should not report infrastructure failure")
	}
	if report.Summary.Total != 0 || report.Summary.Passed != 0 || report.Summary.Failed != 0 {
		t.Fatalf("empty summary=%+v", report.Summary)
	}
	if len(report.Cases) != 0 {
		t.Fatalf("empty run produced cases: %+v", report.Cases)
	}
	if report.SuiteVersion != SuiteVersion || report.Contract != ContractVersion {
		t.Fatalf("report header wrong: %+v", report)
	}
}

func TestRunHappyPathPassesOracleAndParity(t *testing.T) {
	item := Catalog()[0]
	python := &fakeClient{participant: "python", result: passingObservation(item, "python")}
	rust := &fakeClient{participant: "rust", result: passingObservation(item, "rust")}
	report, infrastructureOK := Run(
		[]Case{item},
		[]ParticipantClient{python, rust},
		map[string]ParticipantVersion{},
		RunOptions{},
	)
	if !infrastructureOK {
		t.Fatal("expected infrastructure OK")
	}
	if report.Summary.Passed != 1 || report.Summary.Failed != 0 || report.Summary.Total != 1 {
		t.Fatalf("summary=%+v", report.Summary)
	}
	if python.runCalls != 1 || rust.runCalls != 1 {
		t.Fatalf("call counts python=%d rust=%d", python.runCalls, rust.runCalls)
	}
	caseResult := report.Cases[0]
	if len(caseResult.Participants) != 2 {
		t.Fatalf("participants=%d", len(caseResult.Participants))
	}
	for _, p := range caseResult.Participants {
		if p.Status != "passed" {
			t.Fatalf("participant %s status=%s", p.Participant, p.Status)
		}
		if p.Observation == nil {
			t.Fatalf("participant %s observation nil on pass", p.Participant)
		}
	}
	if caseResult.Status != "passed" || caseResult.Parity != "passed" {
		t.Fatalf("case status/parity=%s/%s", caseResult.Status, caseResult.Parity)
	}
	if caseResult.Rerun != "agent-framework-suite run "+item.ID {
		t.Fatalf("rerun hint=%q", caseResult.Rerun)
	}
	if len(caseResult.Proof) != 0 {
		t.Fatalf("passing case should have no proof, got %+v", caseResult.Proof)
	}
}

func TestRunParityDriftFailsButIsNotInfrastructure(t *testing.T) {
	item := Catalog()[0]
	python := &fakeClient{participant: "python", result: passingObservation(item, "python")}
	rustObs := passingObservation(item, "rust")
	rustObs.Outputs = []any{"DRIFT"}
	rust := &fakeClient{participant: "rust", result: rustObs}
	report, infrastructureOK := Run(
		[]Case{item},
		[]ParticipantClient{python, rust},
		map[string]ParticipantVersion{},
		RunOptions{},
	)
	// Parity drift is a semantic failure, not an infrastructure one.
	if !infrastructureOK {
		t.Fatal("parity drift must not be classified as infrastructure error")
	}
	if report.Summary.Failed != 1 || report.Cases[0].Status != "failed" || report.Cases[0].Parity != "failed" {
		t.Fatalf("expected parity failure, got %+v", report.Cases[0])
	}
	if !strings.Contains(strings.Join(report.Cases[0].Proof, " "), "parity mismatch") {
		t.Fatalf("expected parity mismatch proof, got %+v", report.Cases[0].Proof)
	}
}

func TestRunOracleMismatchFailsStatusButParityHoldsWhenBothAgree(t *testing.T) {
	item := Catalog()[0]
	// Both participants return the SAME wrong output: oracle fails for each, but they agree on parity.
	pythonObs := passingObservation(item, "python")
	pythonObs.Outputs = []any{"SAME-DRIFT"}
	rustObs := passingObservation(item, "rust")
	rustObs.Outputs = []any{"SAME-DRIFT"}
	python := &fakeClient{participant: "python", result: pythonObs}
	rust := &fakeClient{participant: "rust", result: rustObs}
	report, infrastructureOK := Run(
		[]Case{item},
		[]ParticipantClient{python, rust},
		map[string]ParticipantVersion{},
		RunOptions{},
	)
	if !infrastructureOK {
		t.Fatal("oracle mismatch is not infrastructure")
	}
	if report.Summary.Failed != 1 || report.Cases[0].Status != "failed" {
		t.Fatalf("expected failed status, got %+v", report.Cases[0])
	}
	// Both failed validation identically, so parity is still "passed".
	if report.Cases[0].Parity != "passed" {
		t.Fatalf("parity should hold when both agree, got %s", report.Cases[0].Parity)
	}
	for _, p := range report.Cases[0].Participants {
		if !strings.Contains(p.Proof, "oracle mismatch") {
			t.Fatalf("participant %s proof=%q", p.Participant, p.Proof)
		}
	}
}

func TestRunDriverFailureIsInfrastructure(t *testing.T) {
	item := Catalog()[0]
	python := &fakeClient{participant: "python", result: passingObservation(item, "python")}
	rust := &fakeClient{participant: "rust", err: errors.New("boom")}
	report, infrastructureOK := Run(
		[]Case{item},
		[]ParticipantClient{python, rust},
		map[string]ParticipantVersion{},
		RunOptions{},
	)
	if infrastructureOK {
		t.Fatal("driver failure must flip infrastructure flag")
	}
	if report.Summary.Failed != 1 {
		t.Fatalf("summary=%+v", report.Summary)
	}
	// The failing participant is marked infrastructure_error, carries the error as proof,
	// and has no observation. The case proof also echoes the error.
	var failing *ParticipantResult
	for index := range report.Cases[0].Participants {
		if report.Cases[0].Participants[index].Participant == "rust" {
			failing = &report.Cases[0].Participants[index]
		}
	}
	if failing == nil {
		t.Fatal("rust participant missing")
	}
	if failing.Status != "infrastructure_error" || failing.Proof != "boom" || failing.Observation != nil {
		t.Fatalf("rust participant=%+v", failing)
	}
	if !strings.Contains(strings.Join(report.Cases[0].Proof, " "), "boom") {
		t.Fatalf("case proof should echo driver error: %+v", report.Cases[0].Proof)
	}
}

func TestRunRequiresExactlyTwoParticipants(t *testing.T) {
	item := Catalog()[0]
	// Only one participant: must fail with mandatory-count proof before any parity check.
	python := &fakeClient{participant: "python", result: passingObservation(item, "python")}
	report, _ := Run(
		[]Case{item},
		[]ParticipantClient{python},
		map[string]ParticipantVersion{},
		RunOptions{},
	)
	caseResult := report.Cases[0]
	if caseResult.Status != "failed" || caseResult.Parity != "failed" {
		t.Fatalf("expected failed/failed, got %s/%s", caseResult.Status, caseResult.Parity)
	}
	if !strings.Contains(strings.Join(caseResult.Proof, " "), "mandatory participant count: got 1, want 2") {
		t.Fatalf("expected mandatory count proof, got %+v", caseResult.Proof)
	}
}

func TestRunThreeParticipantsAlsoFailsMandatoryCount(t *testing.T) {
	item := Catalog()[0]
	a := &fakeClient{participant: "a", result: passingObservation(item, "a")}
	b := &fakeClient{participant: "b", result: passingObservation(item, "b")}
	c := &fakeClient{participant: "c", result: passingObservation(item, "c")}
	report, _ := Run(
		[]Case{item},
		[]ParticipantClient{a, b, c},
		map[string]ParticipantVersion{},
		RunOptions{},
	)
	if !strings.Contains(strings.Join(report.Cases[0].Proof, " "), "got 3, want 2") {
		t.Fatalf("expected got 3 proof, got %+v", report.Cases[0].Proof)
	}
}

func TestRunStreamsEachCaseResult(t *testing.T) {
	items := []Case{Catalog()[0], Catalog()[3]} // WF-001 and WF-004
	py := &multiCaseClient{participant: "python", results: map[string]Observation{
		items[0].ID: passingObservation(items[0], "python"),
		items[1].ID: passingObservation(items[1], "python"),
	}}
	rs := &multiCaseClient{participant: "rust", results: map[string]Observation{
		items[0].ID: passingObservation(items[0], "rust"),
		items[1].ID: passingObservation(items[1], "rust"),
	}}
	var streamed []CaseResult
	report, _ := Run(items, []ParticipantClient{py, rs}, map[string]ParticipantVersion{}, RunOptions{
		Stream: func(r CaseResult) { streamed = append(streamed, r) },
	})
	if len(streamed) != len(items) {
		t.Fatalf("streamed %d results, want %d", len(streamed), len(items))
	}
	if len(report.Cases) != len(streamed) {
		t.Fatalf("report cases %d != streamed %d", len(report.Cases), len(streamed))
	}
	for index, got := range streamed {
		if got.ID != report.Cases[index].ID {
			t.Fatalf("streamed[%d].ID=%s != report.ID=%s", index, got.ID, report.Cases[index].ID)
		}
	}
}

func TestRunTimeoutOverrideShrinksDeadline(t *testing.T) {
	item := Catalog()[0] // 30s timeout
	py := &fakeClient{participant: "python", result: passingObservation(item, "python")}
	rs := &fakeClient{participant: "rust", result: passingObservation(item, "rust")}
	Run([]Case{item}, []ParticipantClient{py, rs}, map[string]ParticipantVersion{}, RunOptions{Timeout: 1 * time.Second})
	deadline, ok := py.lastCtx.Deadline()
	if !ok {
		t.Fatal("python ctx had no deadline")
	}
	remaining := time.Until(deadline)
	if remaining <= 0 {
		t.Fatalf("override deadline already expired: %v", remaining)
	}
	// If the override were ignored, remaining would be ~30s; with a 1s override it must be well under.
	if remaining > 2*time.Second {
		t.Fatalf("override not applied: remaining=%v want <2s", remaining)
	}
}

func TestRunTimeoutZeroKeepsCaseTimeout(t *testing.T) {
	item := Catalog()[0] // 30s timeout
	py := &fakeClient{participant: "python", result: passingObservation(item, "python")}
	rs := &fakeClient{participant: "rust", result: passingObservation(item, "rust")}
	Run([]Case{item}, []ParticipantClient{py, rs}, map[string]ParticipantVersion{}, RunOptions{Timeout: 0})
	deadline, ok := py.lastCtx.Deadline()
	if !ok {
		t.Fatal("python ctx had no deadline")
	}
	remaining := time.Until(deadline)
	// A zero override must be ignored so the 30s case timeout is used (remaining ~30s, not expired).
	if remaining <= 20*time.Second {
		t.Fatalf("zero override wrongly shrank deadline: remaining=%v want >20s", remaining)
	}
}

func TestRunPassesCaseIDToClients(t *testing.T) {
	item := Catalog()[0]
	py := &fakeClient{participant: "python", result: passingObservation(item, "python")}
	rs := &fakeClient{participant: "rust", result: passingObservation(item, "rust")}
	Run([]Case{item}, []ParticipantClient{py, rs}, map[string]ParticipantVersion{}, RunOptions{})
	if len(py.caseIDs) != 1 || py.caseIDs[0] != item.ID {
		t.Fatalf("python saw caseIDs=%v, want [%s]", py.caseIDs, item.ID)
	}
	if len(rs.caseIDs) != 1 || rs.caseIDs[0] != item.ID {
		t.Fatalf("rust saw caseIDs=%v, want [%s]", rs.caseIDs, item.ID)
	}
}

func TestRunReportEchoesVersions(t *testing.T) {
	item := Catalog()[0]
	py := &fakeClient{participant: "python", result: passingObservation(item, "python")}
	rs := &fakeClient{participant: "rust", result: passingObservation(item, "rust")}
	versions := map[string]ParticipantVersion{
		"python": {Driver: "py", Participant: "python", SubjectVersion: "1.0"},
		"rust":   {Driver: "rs", Participant: "rust", SubjectVersion: "2.0"},
	}
	report, _ := Run([]Case{item}, []ParticipantClient{py, rs}, versions, RunOptions{})
	if !reflect.DeepEqual(report.Participants, versions) {
		t.Fatalf("report participants=%+v want %+v", report.Participants, versions)
	}
}
