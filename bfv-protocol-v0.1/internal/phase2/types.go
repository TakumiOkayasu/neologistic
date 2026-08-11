package phase2

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

const (
	CandidateDirect = "direct-v1"
	CandidateNL     = "nl-v1"
	CandidatePipe   = "pipe-v1"
	CandidateJSON   = "json-v1"
)

var Candidates = []string{CandidateDirect, CandidateNL, CandidatePipe, CandidateJSON}

var requiredCoverage = []string{
	"serialization", "semantic-derivation", "downstream-consumption", "downstream-task",
	"epistemic", "evidence-association", "Q1-FP", "Q1-FN", "D1-transmission",
	"D1-authority", "D1-invented-rejection", "punctuation-tolerance", "copy-fields",
}

var safeFixtureID = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Record is the candidate-independent semantic record used by fixtures and the consumer receipt.
type Record struct {
	Kind               string  `json:"kind"`
	Target             string  `json:"target"`
	Epistemic          string  `json:"epistemic,omitempty"`
	Content            string  `json:"content"`
	Recommendation     string  `json:"recommendation,omitempty"`
	Evidence           *string `json:"evidence"`
	AuthorityReference string  `json:"authority_reference,omitempty"`
	Bounds             *string `json:"bounds"`
}

// Fixture keeps the model-visible source/task separate from the hidden oracle.
type Fixture struct {
	ID                      string         `json:"id"`
	Cohort                  string         `json:"cohort"`
	Coverage                []string       `json:"coverage"`
	Source                  string         `json:"source"`
	DownstreamTask          string         `json:"downstream_task"`
	AuthoritativeReferences []string       `json:"authoritative_references"`
	ExpectedRecords         []Record       `json:"expected_records"`
	TaskOracle              map[string]any `json:"task_oracle"`
}

// Attempt records every provider invocation, including failed calls later retried.
type Attempt struct {
	Attempt      int        `json:"attempt"`
	CallExecuted bool       `json:"call_executed"`
	RuntimeOK    bool       `json:"runtime_ok"`
	RetryReason  string     `json:"retry_reason,omitempty"`
	Prompt       string     `json:"prompt"`
	RawEvents    string     `json:"raw_events"`
	Response     string     `json:"response"`
	Usage        CodexUsage `json:"usage"`
	ExitStatus   int        `json:"exit_status"`
	Stderr       string     `json:"stderr"`
}

// StageObservation explicitly represents applicable, blocked, and not-applicable stages.
type StageObservation struct {
	Mode          string    `json:"mode"`
	Attempts      []Attempt `json:"attempts"`
	SkippedReason string    `json:"skipped_reason,omitempty"`
}

// DefectAttribution is evidence-backed diagnosis, independent of failure kind.
type DefectAttribution struct {
	Stage    string `json:"stage"`
	Category string `json:"category"`
	Evidence string `json:"evidence"`
}

// Observation is one candidate x fixture cell in the first coverage pass.
type Observation struct {
	CaseID       string              `json:"case_id"`
	Candidate    string              `json:"candidate"`
	CoveragePass int                 `json:"coverage_pass"`
	Handoff      string              `json:"handoff"`
	Producer     StageObservation    `json:"producer"`
	Consumer     StageObservation    `json:"consumer"`
	Defects      []DefectAttribution `json:"defects"`
}

type ConsumerOutput struct {
	Records []Record       `json:"records"`
	Task    map[string]any `json:"task"`
}

func LoadFixtures(r io.Reader) ([]Fixture, error) {
	var fixtures []Fixture
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	seen := map[string]bool{}
	coverage := map[string]bool{}
	cohorts := map[string]bool{}
	for line := 1; scanner.Scan(); line++ {
		var fixture Fixture
		if err := decodeStrictLine(scanner.Bytes(), &fixture); err != nil {
			return nil, fmt.Errorf("fixture line %d: %w", line, err)
		}
		if fixture.ID == "" || fixture.Source == "" || fixture.DownstreamTask == "" || len(fixture.ExpectedRecords) == 0 || len(fixture.TaskOracle) == 0 {
			return nil, fmt.Errorf("fixture line %d: id, source, downstream_task, expected_records, and task_oracle are required", line)
		}
		if !safeFixtureID.MatchString(fixture.ID) {
			return nil, fmt.Errorf("fixture line %d: unsafe id %q", line, fixture.ID)
		}
		if seen[fixture.ID] {
			return nil, fmt.Errorf("fixture line %d: duplicate id %q", line, fixture.ID)
		}
		if fixture.Cohort != "copy" && fixture.Cohort != "source-derived" {
			return nil, fmt.Errorf("fixture %q: cohort must be copy or source-derived", fixture.ID)
		}
		seen[fixture.ID] = true
		cohorts[fixture.Cohort] = true
		keys := map[string]bool{}
		for index, record := range fixture.ExpectedRecords {
			if err := ValidateRecord(record); err != nil {
				return nil, fmt.Errorf("fixture %q record %d: %w", fixture.ID, index+1, err)
			}
			key := recordKey(record)
			if keys[key] {
				return nil, fmt.Errorf("fixture %q: duplicate record key %q", fixture.ID, key)
			}
			keys[key] = true
			if record.Kind == "D1" && !containsString(fixture.AuthoritativeReferences, record.AuthorityReference) {
				return nil, fmt.Errorf("fixture %q: D1 authority reference %q is not allowlisted", fixture.ID, record.AuthorityReference)
			}
		}
		if fixture.Cohort == "copy" {
			for _, record := range fixture.ExpectedRecords {
				for _, value := range recordValues(record) {
					if value != "" && !strings.Contains(fixture.Source, value) {
						return nil, fmt.Errorf("fixture %q: copy source does not contain expected value %q", fixture.ID, value)
					}
				}
			}
		}
		if err := validateTaskObject(fixture.TaskOracle); err != nil {
			return nil, fmt.Errorf("fixture %q task_oracle: %w", fixture.ID, err)
		}
		if authority, ok := fixture.TaskOracle["authority_reference"].(string); ok && authority != "" && !containsString(fixture.AuthoritativeReferences, authority) {
			return nil, fmt.Errorf("fixture %q: task authority reference %q is not allowlisted", fixture.ID, authority)
		}
		for _, label := range fixture.Coverage {
			coverage[label] = true
		}
		fixtures = append(fixtures, fixture)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(fixtures) == 0 {
		return nil, fmt.Errorf("no fixtures")
	}
	if !cohorts["copy"] || !cohorts["source-derived"] {
		return nil, fmt.Errorf("both copy and source-derived cohorts are required")
	}
	var missing []string
	for _, label := range requiredCoverage {
		if !coverage[label] {
			missing = append(missing, label)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("fixture coverage missing: %v", missing)
	}
	return fixtures, nil
}

func LoadObservations(r io.Reader) ([]Observation, error) {
	var observations []Observation
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 32*1024*1024)
	for line := 1; scanner.Scan(); line++ {
		var observation Observation
		if err := decodeStrictLine(scanner.Bytes(), &observation); err != nil {
			return nil, fmt.Errorf("observation line %d: %w", line, err)
		}
		if observation.CaseID == "" || !containsString(Candidates, observation.Candidate) || observation.CoveragePass != 1 {
			return nil, fmt.Errorf("observation line %d: valid case_id, candidate, and coverage_pass=1 are required", line)
		}
		if err := validateDefects(observation.Defects); err != nil {
			return nil, fmt.Errorf("observation line %d: %w", line, err)
		}
		observations = append(observations, observation)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return observations, nil
}

func ValidateRecord(record Record) error {
	if record.Target == "" || record.Content == "" {
		return fmt.Errorf("kind, target, and content are required")
	}
	switch record.Kind {
	case "R1":
		if len(record.Epistemic) != 2 || !containsByte("OIS", record.Epistemic[0]) || !containsByte("UVR", record.Epistemic[1]) {
			return fmt.Errorf("R1 epistemic must be origin O/I/S plus state U/V/R")
		}
		if (record.Epistemic[1] == 'V' || record.Epistemic[1] == 'R') && record.Evidence == nil {
			return fmt.Errorf("verified or refuted R1 requires evidence")
		}
		if record.Recommendation != "" || record.AuthorityReference != "" {
			return fmt.Errorf("R1 contains fields reserved for another kind")
		}
	case "Q1":
		if record.Recommendation == "" || record.Bounds == nil {
			return fmt.Errorf("Q1 requires recommendation and blocking boundary")
		}
		if record.Epistemic != "" || record.AuthorityReference != "" {
			return fmt.Errorf("Q1 contains fields reserved for another kind")
		}
	case "D1":
		if record.AuthorityReference == "" {
			return fmt.Errorf("D1 requires authority_reference")
		}
		if record.Epistemic != "" || record.Recommendation != "" || record.Evidence != nil {
			return fmt.Errorf("D1 contains fields reserved for another kind")
		}
	default:
		return fmt.Errorf("unsupported record kind %q", record.Kind)
	}
	return nil
}

func decodeStrictLine(raw []byte, target any) error {
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("line must contain exactly one JSON object")
	}
	return nil
}

func validateDefects(defects []DefectAttribution) error {
	allowed := map[string]bool{"fixture": true, "protocol": true, "model": true, "scorer": true, "runtime": true, "unclassified": true}
	for index, defect := range defects {
		if defect.Stage == "" || !allowed[defect.Category] || strings.TrimSpace(defect.Evidence) == "" {
			return fmt.Errorf("defect %d requires stage, supported category, and evidence", index+1)
		}
	}
	return nil
}

func containsByte(values string, value byte) bool {
	for index := range values {
		if values[index] == value {
			return true
		}
	}
	return false
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func recordKey(record Record) string { return record.Kind + "\x00" + record.Target }

func recordValues(record Record) []string {
	values := []string{record.Target, record.Epistemic, record.Content, record.Recommendation, record.AuthorityReference}
	if record.Evidence != nil {
		values = append(values, *record.Evidence)
	}
	if record.Bounds != nil {
		values = append(values, *record.Bounds)
	}
	return values
}
