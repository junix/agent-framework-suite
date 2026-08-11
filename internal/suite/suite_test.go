package suite

import (
	"context"
	"errors"
	"testing"
)

type fakeClient struct {
	participant string
	result      Observation
	err         error
}

func (client fakeClient) ParticipantName() string { return client.participant }

func (client fakeClient) Run(context.Context, string) (Observation, error) {
	return client.result, client.err
}

func TestCatalogHasStableUniqueIDs(t *testing.T) {
	seen := map[string]bool{}
	for _, item := range Catalog() {
		if seen[item.ID] {
			t.Fatalf("duplicate case id %s", item.ID)
		}
		seen[item.ID] = true
	}
	if len(seen) != 10 {
		t.Fatalf("got %d cases, want 10", len(seen))
	}
}

func TestSelectByPrefixNameAndTag(t *testing.T) {
	selected, err := Select(Catalog(), []string{"WF-001", "fan_out_fan_in"}, "parity")
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 2 || selected[0].ID != "WF-001" || selected[1].ID != "WF-004" {
		t.Fatalf("unexpected selection: %+v", selected)
	}
	if _, err := Select(Catalog(), []string{"missing"}, ""); err == nil {
		t.Fatal("unknown selector was accepted")
	}
}

func TestRunRequiresOracleAndParity(t *testing.T) {
	item := Catalog()[0]
	python := item.Expected
	python.Participant = "python"
	rust := item.Expected
	rust.Participant = "rust"
	report, infrastructureOK := Run(
		[]Case{item},
		[]ParticipantClient{fakeClient{participant: "python", result: python}, fakeClient{participant: "rust", result: rust}},
		map[string]ParticipantVersion{},
		RunOptions{},
	)
	if !infrastructureOK || report.Summary.Passed != 1 {
		t.Fatalf("expected pass, got %+v", report)
	}
	rust.Outputs = []any{"DRIFT"}
	report, _ = Run(
		[]Case{item},
		[]ParticipantClient{fakeClient{participant: "python", result: python}, fakeClient{participant: "rust", result: rust}},
		map[string]ParticipantVersion{},
		RunOptions{},
	)
	if report.Summary.Failed != 1 || report.Cases[0].Parity != "failed" {
		t.Fatalf("expected parity failure, got %+v", report)
	}
}

func TestRunClassifiesDriverFailureAsInfrastructure(t *testing.T) {
	item := Catalog()[0]
	result := item.Expected
	result.Participant = "python"
	report, infrastructureOK := Run(
		[]Case{item},
		[]ParticipantClient{fakeClient{participant: "python", result: result}, fakeClient{participant: "rust", err: errors.New("boom")}},
		map[string]ParticipantVersion{},
		RunOptions{},
	)
	if infrastructureOK || report.Summary.Failed != 1 {
		t.Fatalf("expected infrastructure failure, got %+v", report)
	}
}
