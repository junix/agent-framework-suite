package suite

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

type ParticipantClient interface {
	ParticipantName() string
	Run(context.Context, string) (Observation, error)
}

type RunOptions struct {
	Timeout time.Duration
	Stream  func(CaseResult)
}

func Run(cases []Case, clients []ParticipantClient, versions map[string]ParticipantVersion, options RunOptions) (Report, bool) {
	report := Report{
		SuiteVersion: SuiteVersion,
		Contract:     ContractVersion,
		Participants: versions,
		Cases:        make([]CaseResult, 0, len(cases)),
	}
	infrastructureOK := true
	for _, item := range cases {
		result := runCase(item, clients, options.Timeout)
		report.Cases = append(report.Cases, result)
		if result.Status == "passed" {
			report.Summary.Passed++
		} else {
			report.Summary.Failed++
		}
		for _, participant := range result.Participants {
			if participant.Status == "infrastructure_error" {
				infrastructureOK = false
			}
		}
		if options.Stream != nil {
			options.Stream(result)
		}
	}
	report.Summary.Total = len(report.Cases)
	return report, infrastructureOK
}

func runCase(item Case, clients []ParticipantClient, override time.Duration) CaseResult {
	result := CaseResult{
		ID:           item.ID,
		Name:         item.Name,
		Status:       "passed",
		Parity:       "passed",
		Rerun:        "agent-framework-suite run " + item.ID,
		Participants: make([]ParticipantResult, 0, len(clients)),
	}
	deadline := item.Timeout
	if override > 0 && override < deadline {
		deadline = override
	}
	for _, client := range clients {
		participantName := client.ParticipantName()
		ctx, cancel := context.WithTimeout(context.Background(), deadline)
		observed, err := client.Run(ctx, item.ID)
		cancel()
		participant := ParticipantResult{Participant: participantName, Status: "passed"}
		if err != nil {
			participant.Status = "infrastructure_error"
			participant.Proof = err.Error()
			result.Status = "failed"
			result.Proof = append(result.Proof, err.Error())
		} else {
			participant.Observation = &observed
			if proof := ValidateObservation(item, observed, participantName); proof != "" {
				participant.Status = "failed"
				participant.Proof = proof
				result.Status = "failed"
				result.Proof = append(result.Proof, proof)
			}
		}
		result.Participants = append(result.Participants, participant)
	}
	if len(result.Participants) != 2 {
		result.Status = "failed"
		result.Parity = "failed"
		result.Proof = append(result.Proof, fmt.Sprintf("mandatory participant count: got %d, want 2", len(result.Participants)))
		return result
	}
	left := result.Participants[0].Observation
	right := result.Participants[1].Observation
	if left == nil || right == nil || !Equivalent(*left, *right) {
		result.Status = "failed"
		result.Parity = "failed"
		if left != nil && right != nil {
			leftJSON, _ := json.Marshal(normalized(*left))
			rightJSON, _ := json.Marshal(normalized(*right))
			result.Proof = append(result.Proof, fmt.Sprintf("parity mismatch: %s != %s", leftJSON, rightJSON))
		}
	}
	return result
}

func ValidateObservation(item Case, observed Observation, participant string) string {
	if observed.Participant != participant {
		return fmt.Sprintf("%s %s identity mismatch: observation participant=%q", participant, item.ID, observed.Participant)
	}
	expected := item.Expected
	expected.Participant = ""
	actual := normalized(observed)
	if !reflect.DeepEqual(expected, actual) {
		expectedJSON, _ := json.Marshal(expected)
		actualJSON, _ := json.Marshal(actual)
		return fmt.Sprintf("%s %s oracle mismatch: got %s want %s", participant, item.ID, actualJSON, expectedJSON)
	}
	return ""
}

func Equivalent(left, right Observation) bool {
	return reflect.DeepEqual(normalized(left), normalized(right))
}

func normalized(value Observation) Observation {
	value.Participant = ""
	return value
}
