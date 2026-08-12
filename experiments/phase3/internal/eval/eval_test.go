package eval

import (
	"encoding/json"
	"math"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/codec"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/model"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/runplan"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/validate"
)

func loadCanonical(t *testing.T) (string, []model.Arm, []model.Fixture) {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	arms, err := codec.LoadJSON[[]model.Arm](filepath.Join(root, "fixtures", "arms.json"))
	if err != nil {
		t.Fatal(err)
	}
	fixtures, err := codec.LoadJSONL[model.Fixture](filepath.Join(root, "fixtures", "cases.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if err := validate.ValidateArms(root, arms); err != nil {
		t.Fatal(err)
	}
	if err := validate.ValidateFixtures(fixtures); err != nil {
		t.Fatal(err)
	}
	return root, arms, fixtures
}

func i64(value int64) *int64 { return &value }

func manifestFor(arms []model.Arm, fixtures []model.Fixture) model.RunManifest {
	manifest := model.RunManifest{
		Version: model.RunManifestVersion, Scope: model.ExperimentScope, RunID: "test-run", FreezeLockSHA256: strings.Repeat("a", 64),
		Sender:    model.ModelConfig{Provider: "test", Model: "sender", ReasoningEffort: "fixed"},
		Receiver:  model.ModelConfig{Provider: "test", Model: "receiver", ReasoningEffort: "fixed"},
		Transport: "test", Isolation: model.IsolationConfig{Workspace: "empty-temporary", Network: "provider-only", Tools: []string{}}, EliminationPolicy: model.EliminationPolicy, RuntimeRetriesFixed: 0,
	}
	order, err := runplan.ExecutionOrder(fixtures, arms)
	if err != nil {
		panic(err)
	}
	manifest.ExecutionOrder = order
	return manifest
}

func perfectOutcome(f model.Fixture) model.Outcome {
	outcome := model.Outcome{
		Version: model.OutcomeVersion, TaskID: f.ID,
		SelectedOption: f.Gold.SelectedOption, Status: f.Gold.Status,
		Assertions: append([]model.Assertion(nil), f.Gold.Assertions...),
		Actions:    append([]string(nil), f.Gold.RequiredActions...),
		Completion: f.Gold.Completion,
		Concepts:   append([]string(nil), f.Gold.RequiredConcepts...),
	}
	if f.Gold.HumanRequest.Mode == "required" {
		outcome.HumanRequest = &model.HumanRequest{
			QuestionID:           f.Gold.HumanRequest.QuestionID,
			RecommendationOption: f.Gold.HumanRequest.RecommendationOption,
			BlockingReason:       "human-reserved authority",
		}
	}
	if f.Gold.IntentReceipt.Required {
		conflicts := make([]string, 0, len(f.Gold.IntentReceipt.RequiredConflictTerms))
		for _, term := range f.Gold.IntentReceipt.RequiredConflictTerms {
			conflicts = append(conflicts, term+" is not authorized")
		}
		outcome.IntentReceipt = &model.IntentReceipt{
			Version: "intent-v1", Status: f.Gold.IntentReceipt.Status,
			UnderstoodContract: "read-only evaluation only",
			AddedAssumptions:   []string{},
			PlannedSideEffects: append([]string(nil), f.Gold.IntentReceipt.PlannedSideEffects...),
			Conflicts:          conflicts,
		}
	}
	return outcome
}

func senderResponse(t *testing.T, arm model.Arm, f model.Fixture) string {
	t.Helper()
	if arm.Carrier != "json-carrier-v1" {
		return "判断=" + f.Gold.SelectedOption + "。根拠ID維持。必要actionのみ。"
	}
	actions := make([]validate.JSONCarrierAction, 0, len(f.ActionOptions))
	required := map[string]struct{}{}
	for _, action := range f.Gold.RequiredActions {
		required[action] = struct{}{}
	}
	for _, option := range f.ActionOptions {
		disposition := "exclude"
		if _, ok := required[option.ID]; ok {
			disposition = "include"
		}
		actions = append(actions, validate.JSONCarrierAction{ID: option.ID, Disposition: disposition})
	}
	facts := make([]validate.JSONCarrierFact, 0, len(f.Gold.Assertions))
	for _, assertion := range f.Gold.Assertions {
		facts = append(facts, validate.JSONCarrierFact(assertion))
	}
	value := validate.JSONCarrier{
		Summary: "decision-relevant transfer", Facts: facts,
		RecommendedOption: f.Gold.SelectedOption, Actions: actions,
		Concepts: append([]string(nil), f.Gold.RequiredConcepts...),
	}
	if f.Gold.HumanRequest.Mode == "required" {
		value.HumanBoundary = &model.HumanRequest{
			QuestionID:           f.Gold.HumanRequest.QuestionID,
			RecommendationOption: f.Gold.HumanRequest.RecommendationOption,
			BlockingReason:       "human-reserved authority",
		}
	}
	if f.Gold.IntentReceipt.Required {
		value.IntentReceipt = perfectOutcome(f).IntentReceipt
	}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func perfectObservations(t *testing.T, arms []model.Arm, fixtures []model.Fixture) []model.Observation {
	t.Helper()
	armByID := make(map[string]model.Arm, len(arms))
	for _, arm := range arms {
		armByID[arm.ID] = arm
	}
	fixtureByID := make(map[string]model.Fixture, len(fixtures))
	for _, fixture := range fixtures {
		fixtureByID[fixture.ID] = fixture
	}
	order, err := runplan.ExecutionOrder(fixtures, arms)
	if err != nil {
		t.Fatal(err)
	}
	observations := make([]model.Observation, 0, len(order))
	for _, cell := range order {
		fixture := fixtureByID[cell.CaseID]
		arm := armByID[cell.ArmID]
		handoff := senderResponse(t, arm, fixture)
		outcomeData, err := json.Marshal(perfectOutcome(fixture))
		if err != nil {
			t.Fatal(err)
		}
		attemptUsage := model.Usage{
			InputTokens: i64(100), CachedInputTokens: i64(20), CacheWriteInputTokens: i64(0),
			OutputTokens: i64(20), ReasoningOutputTokens: i64(5),
		}
		observations = append(observations, model.Observation{
			Version: model.ObservationVersion, RunID: "test-run", CaseID: fixture.ID, ArmID: arm.ID,
			Sender:   model.Attempt{Response: handoff, ExitCode: 0, Calls: 1, Usage: attemptUsage, LatencyMS: 10},
			Handoff:  handoff,
			Receiver: model.Attempt{Response: string(outcomeData), ExitCode: 0, Calls: 1, Usage: attemptUsage, LatencyMS: 10},
			Defects:  []model.DefectAttribution{},
		})
	}
	return observations
}

func fixtureByIDForTest(fixtures []model.Fixture, id string) model.Fixture {
	for _, fixture := range fixtures {
		if fixture.ID == id {
			return fixture
		}
	}
	panic("fixture not found: " + id)
}

func scorePerfect(t *testing.T, arms []model.Arm, fixtures []model.Fixture, observations []model.Observation) Report {
	t.Helper()
	return Score(fixtures, arms, observations, nil, manifestFor(arms, fixtures))
}

func TestEqualPerfectArmsSelectNativeDeletionDefault(t *testing.T) {
	_, arms, fixtures := loadCanonical(t)
	report := scorePerfect(t, arms, fixtures, perfectObservations(t, arms, fixtures))
	if !report.Coverage.Complete {
		t.Fatalf("coverage incomplete: %+v", report.Coverage)
	}
	if report.Selection.Winner == nil || *report.Selection.Winner != "A" {
		t.Fatalf("expected A, got %+v", report.Selection)
	}
	for _, summary := range report.Arms {
		if summary.HardFailures != 0 {
			t.Fatalf("arm %s has failures: %+v", summary.ArmID, summary.Failures)
		}
	}
}

func TestCustomArmMustDominateEveryTask(t *testing.T) {
	_, arms, fixtures := loadCanonical(t)
	observations := perfectObservations(t, arms, fixtures)
	for i := range observations {
		if observations[i].ArmID != "E" {
			continue
		}
		observations[i].Sender.Usage.InputTokens = i64(80)
		observations[i].Receiver.Usage.InputTokens = i64(80)
		observations[i].Sender.Usage.OutputTokens = i64(15)
		observations[i].Receiver.Usage.OutputTokens = i64(15)
		observations[i].Sender.LatencyMS = 8
		observations[i].Receiver.LatencyMS = 8
	}
	report := scorePerfect(t, arms, fixtures, observations)
	if report.Selection.Winner == nil || *report.Selection.Winner != "E" {
		t.Fatalf("expected E to dominate A, got %+v", report.Selection)
	}
}

func TestNativeHardFailureSelectsLowestMechanismEligibleArm(t *testing.T) {
	_, arms, fixtures := loadCanonical(t)
	observations := perfectObservations(t, arms, fixtures)
	firstFixture := fixtureByIDForTest(fixtures, "p3-002")
	var filtered []model.Observation
	for i := range observations {
		observation := observations[i]
		if observation.ArmID == "A" && observation.CaseID == firstFixture.ID {
			outcome := perfectOutcome(firstFixture)
			outcome.SelectedOption = firstFixture.DecisionOptions[0].ID
			data, _ := json.Marshal(outcome)
			observation.Receiver.Response = string(data)
			filtered = append(filtered, observation)
			continue
		}
		if observation.ArmID == "A" {
			continue
		}
		filtered = append(filtered, observation)
	}
	report := scorePerfect(t, arms, fixtures, filtered)
	if !report.Coverage.Complete || len(report.Coverage.SkippedCells) != len(fixtures)-1 {
		t.Fatalf("expected valid elimination coverage, got %+v", report.Coverage)
	}
	if report.Selection.Winner == nil || *report.Selection.Winner != "B" {
		t.Fatalf("expected B after A failure, got %+v", report.Selection)
	}
}

func TestContinuingArmAfterHardFailureBlocksSelection(t *testing.T) {
	_, arms, fixtures := loadCanonical(t)
	observations := perfectObservations(t, arms, fixtures)
	firstFixture := fixtureByIDForTest(fixtures, "p3-002")
	for i := range observations {
		if observations[i].ArmID == "A" && observations[i].CaseID == firstFixture.ID {
			outcome := perfectOutcome(firstFixture)
			outcome.SelectedOption = firstFixture.DecisionOptions[0].ID
			data, _ := json.Marshal(outcome)
			observations[i].Receiver.Response = string(data)
			break
		}
	}
	report := scorePerfect(t, arms, fixtures, observations)
	if report.Coverage.Complete || report.Selection.Winner != nil || len(report.HardErrors) == 0 {
		t.Fatalf("expected post-failure continuation rejection, got %+v", report)
	}
}

func TestMissingCellBlocksSelection(t *testing.T) {
	_, arms, fixtures := loadCanonical(t)
	observations := perfectObservations(t, arms, fixtures)
	observations = observations[:len(observations)-1]
	report := scorePerfect(t, arms, fixtures, observations)
	if report.Coverage.Complete || report.Selection.Winner != nil {
		t.Fatalf("expected incomplete coverage, got %+v", report)
	}
}

func TestMixedRunIDsBlockSelection(t *testing.T) {
	_, arms, fixtures := loadCanonical(t)
	observations := perfectObservations(t, arms, fixtures)
	observations[0].RunID = "other-run"
	report := scorePerfect(t, arms, fixtures, observations)
	if report.Coverage.Complete || report.Selection.Winner != nil || len(report.HardErrors) == 0 {
		t.Fatalf("expected mixed run rejection, got %+v", report)
	}
}

func TestEpistemicLossIsHardFailure(t *testing.T) {
	_, arms, fixtures := loadCanonical(t)
	observations := perfectObservations(t, arms, fixtures)
	for i := range observations {
		if observations[i].ArmID == "C" && observations[i].CaseID == fixtures[1].ID {
			outcome := perfectOutcome(fixtures[1])
			outcome.Assertions[1].Origin = "observed"
			outcome.Assertions[1].State = "verified"
			data, _ := json.Marshal(outcome)
			observations[i].Receiver.Response = string(data)
			break
		}
	}
	report := scorePerfect(t, arms, fixtures, observations)
	for _, summary := range report.Arms {
		if summary.ArmID == "C" {
			if summary.Failures.Epistemic == 0 || summary.HardFailures == 0 {
				t.Fatalf("expected epistemic hard failure: %+v", summary)
			}
			if summary.Defects["unclassified"] == 0 {
				t.Fatalf("expected unattributed hard failure to remain visible: %+v", summary.Defects)
			}
			return
		}
	}
	t.Fatal("arm C summary missing")
}

func TestUnexpectedAssertionIsHardFailure(t *testing.T) {
	_, arms, fixtures := loadCanonical(t)
	observations := perfectObservations(t, arms, fixtures)
	for i := range observations {
		if observations[i].ArmID == "A" && observations[i].CaseID == fixtures[0].ID {
			outcome := perfectOutcome(fixtures[0])
			outcome.Assertions = append(outcome.Assertions, model.Assertion{Key: "claim-unknown", Origin: "inferred", State: "open"})
			data, _ := json.Marshal(outcome)
			observations[i].Receiver.Response = string(data)
			break
		}
	}
	report := scorePerfect(t, arms, fixtures, observations)
	if report.Selection.Winner != nil && *report.Selection.Winner == "A" {
		t.Fatalf("unexpected assertion should disqualify A: %+v", report.Selection)
	}
}

func TestTokenOverflowIsHardFailure(t *testing.T) {
	_, arms, fixtures := loadCanonical(t)
	observations := perfectObservations(t, arms, fixtures)
	target := observations[0]
	observations[0].Sender.Usage.InputTokens = i64(math.MaxInt64)
	observations[0].Receiver.Usage.InputTokens = i64(1)
	report := scorePerfect(t, arms, fixtures, observations)
	for _, trial := range report.Trials {
		if trial.CaseID == target.CaseID && trial.ArmID == target.ArmID {
			if trial.Failures.Runtime == 0 {
				t.Fatalf("expected overflow failure: %+v", trial)
			}
			return
		}
	}
	t.Fatalf("target trial missing: case=%s arm=%s", target.CaseID, target.ArmID)
}

func pricingForTest(manifest model.RunManifest) model.PricingPolicy {
	limit := int64(1000)
	return model.PricingPolicy{
		Version:  model.PricingVersion,
		Currency: "USD",
		Entries: []model.PricingEntry{
			{
				Provider: manifest.Sender.Provider, Model: manifest.Sender.Model,
				UncachedInputNanoPerToken: 10, CachedInputNanoPerToken: 1,
				CacheWriteNanoPerToken: 0, OutputNanoPerToken: 100,
				ReasoningIncludedInOutput: true, MaxInputTokensPerCall: &limit,
				Source: "https://pricing.invalid/sender", VerifiedAt: "2026-08-12T00:00:00Z",
			},
			{
				Provider: manifest.Receiver.Provider, Model: manifest.Receiver.Model,
				UncachedInputNanoPerToken: 20, CachedInputNanoPerToken: 2,
				CacheWriteNanoPerToken: 0, OutputNanoPerToken: 200,
				ReasoningIncludedInOutput: true, MaxInputTokensPerCall: &limit,
				Source: "https://pricing.invalid/receiver", VerifiedAt: "2026-08-12T00:00:00Z",
			},
		},
	}
}

func TestPricingUsesStageSpecificModelRates(t *testing.T) {
	_, arms, fixtures := loadCanonical(t)
	observations := perfectObservations(t, arms, fixtures)
	manifest := manifestFor(arms, fixtures)
	pricing := pricingForTest(manifest)
	target := observations[0]
	report := Score(fixtures, arms, observations, &pricing, manifest)
	for _, trial := range report.Trials {
		if trial.CaseID == target.CaseID && trial.ArmID == target.ArmID {
			if trial.Cost.PaidNanoUSD == nil || *trial.Cost.PaidNanoUSD != 8460 {
				t.Fatalf("unexpected stage-specific paid cost: %+v", trial.Cost)
			}
			return
		}
	}
	t.Fatalf("target trial missing: case=%s arm=%s", target.CaseID, target.ArmID)
}

func TestPricingApplicabilityLimitIsHardFailure(t *testing.T) {
	_, arms, fixtures := loadCanonical(t)
	observations := perfectObservations(t, arms, fixtures)
	manifest := manifestFor(arms, fixtures)
	pricing := pricingForTest(manifest)
	limit := int64(99)
	pricing.Entries[0].MaxInputTokensPerCall = &limit
	target := observations[0]
	report := Score(fixtures, arms, observations, &pricing, manifest)
	for _, trial := range report.Trials {
		if trial.CaseID == target.CaseID && trial.ArmID == target.ArmID {
			if trial.Failures.Runtime == 0 || trial.Cost.PaidNanoUSD != nil {
				t.Fatalf("expected pricing applicability failure: %+v", trial)
			}
			return
		}
	}
	t.Fatalf("target trial missing: case=%s arm=%s", target.CaseID, target.ArmID)
}

func TestScoreCellReportsEliminationSignal(t *testing.T) {
	_, arms, fixtures := loadCanonical(t)
	observations := perfectObservations(t, arms, fixtures)
	manifest := manifestFor(arms, fixtures)
	observation := observations[0]
	fixture := fixtureByIDForTest(fixtures, observation.CaseID)
	var arm model.Arm
	for _, candidate := range arms {
		if candidate.ID == observation.ArmID {
			arm = candidate
			break
		}
	}
	outcome := perfectOutcome(fixture)
	outcome.SelectedOption = fixture.DecisionOptions[0].ID
	data, err := json.Marshal(outcome)
	if err != nil {
		t.Fatal(err)
	}
	observation.Receiver.Response = string(data)
	trial, err := ScoreCell(fixture, arm, observation, nil, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if !trial.HardFailure || trial.Failures.Decision == 0 {
		t.Fatalf("expected cell-level hard failure: %+v", trial)
	}
}
