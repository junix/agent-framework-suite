package suite

import "time"

const (
	ContractVersion = "workflow-core/v1"
	SuiteVersion    = "0.1.0"
)

type ObservedError struct {
	Category string `json:"category"`
}

type Observation struct {
	Contract      string         `json:"contract"`
	CaseID        string         `json:"case_id"`
	Participant   string         `json:"participant"`
	Outcome       string         `json:"outcome"`
	TerminalState *string        `json:"terminal_state"`
	Outputs       []any          `json:"outputs"`
	Requests      []string       `json:"requests"`
	Checkpoints   []int          `json:"checkpoints"`
	Error         *ObservedError `json:"error"`
	Evidence      map[string]any `json:"evidence"`
}

type Case struct {
	ID       string
	Name     string
	Tags     []string
	Timeout  time.Duration
	Expected Observation
}

type ParticipantVersion struct {
	Driver         string `json:"driver"`
	Participant    string `json:"participant"`
	SubjectVersion string `json:"subject_version"`
}

type ParticipantResult struct {
	Participant string       `json:"participant"`
	Status      string       `json:"status"`
	Observation *Observation `json:"observation,omitempty"`
	Proof       string       `json:"proof,omitempty"`
}

type CaseResult struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Status       string              `json:"status"`
	Participants []ParticipantResult `json:"participants"`
	Parity       string              `json:"parity"`
	Proof        []string            `json:"proof,omitempty"`
	Rerun        string              `json:"rerun"`
}

type Summary struct {
	Passed int `json:"passed"`
	Failed int `json:"failed"`
	Total  int `json:"total"`
}

type Report struct {
	SuiteVersion string                        `json:"suite_version"`
	Contract     string                        `json:"contract"`
	Participants map[string]ParticipantVersion `json:"participants"`
	Cases        []CaseResult                  `json:"cases"`
	Summary      Summary                       `json:"summary"`
}
