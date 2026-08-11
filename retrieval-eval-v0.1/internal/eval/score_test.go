package eval

import "testing"

func TestScoreSeparatesRecallStalenessAndDownstream(t *testing.T) {
	ok := true
	input := int64(10)
	fixtures := []Fixture{{ID: "q1", Query: "where", Answerable: true, GoldPaths: []string{"a.go", "b.go"}, K: 2}}
	observations := []Observation{{
		CaseID: "q1", System: "zoekt", Available: true,
		Candidates: []Candidate{{Path: "x.go"}, {Path: "b.go"}},
		LatencyMS:  5, CandidateBytes: 100, ToolCalls: 1, InputTokens: &input, DownstreamCorrect: &ok,
	}}
	report := Score(fixtures, observations)
	if !report.Coverage.Complete || len(report.Trials) != 1 {
		t.Fatalf("unexpected report: %#v", report)
	}
	trial := report.Trials[0]
	if trial.RecallAtK != 0.5 || trial.ReciprocalRank != 0.5 || !trial.DownstreamCorrect {
		t.Fatalf("unexpected trial: %#v", trial)
	}
}

func TestUnsupportedAbsenceAndStaleAreIndependentFailures(t *testing.T) {
	fixtures := []Fixture{{ID: "q1", Answerable: true, GoldPaths: []string{"a.go"}}}
	observations := []Observation{{CaseID: "q1", System: "gateway", Available: true, Stale: true, ClaimedAbsent: true}}
	trial := Score(fixtures, observations).Trials[0]
	if !trial.StaleFailure || !trial.UnsupportedAbsence {
		t.Fatalf("expected both failures: %#v", trial)
	}
}

func TestCoverageDetectsMissingCells(t *testing.T) {
	fixtures := []Fixture{{ID: "q1"}, {ID: "q2"}}
	observations := []Observation{{CaseID: "q1", System: "a", Available: true}}
	report := Score(fixtures, observations)
	if report.Coverage.Complete || len(report.Coverage.Missing) != 1 {
		t.Fatalf("expected missing coverage: %#v", report.Coverage)
	}
}

func TestExplicitSystemsDetectEntirelyMissingSystem(t *testing.T) {
	fixtures := []Fixture{{ID: "q1"}}
	observations := []Observation{{CaseID: "q1", System: "filesystem", Available: true}}
	report := ScoreWithSystems(fixtures, observations, []string{"filesystem", "zoekt"})
	if report.Coverage.Complete || len(report.Coverage.Missing) != 1 {
		t.Fatalf("expected missing system coverage: %#v", report.Coverage)
	}
}
