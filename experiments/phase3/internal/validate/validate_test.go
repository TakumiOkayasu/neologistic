package validate

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/codec"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/model"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/runplan"
)

func phase3Root(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func loadInputs(t *testing.T) ([]model.Arm, []model.Fixture) {
	t.Helper()
	root := phase3Root(t)
	arms, err := codec.LoadJSON[[]model.Arm](filepath.Join(root, "fixtures", "arms.json"))
	if err != nil {
		t.Fatal(err)
	}
	fixtures, err := codec.LoadJSONL[model.Fixture](filepath.Join(root, "fixtures", "cases.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	return arms, fixtures
}

func TestCanonicalInputsValidate(t *testing.T) {
	arms, fixtures := loadInputs(t)
	if err := ValidateArms(phase3Root(t), arms); err != nil {
		t.Fatal(err)
	}
	if err := ValidateFixtures(fixtures); err != nil {
		t.Fatal(err)
	}
}

func TestSideEffectGrantRequiresUserAuthorization(t *testing.T) {
	_, fixtures := loadInputs(t)
	envelope := *fixtures[len(fixtures)-1].IntentEnvelope
	envelope.SideEffects.Allowed[0].SourceIntentIDs = []string{"I1"}
	if err := ValidateIntentEnvelope(envelope); err == nil {
		t.Fatal("expected assistant inference to be rejected as side-effect authority")
	}
}

func TestAcceptanceCriterionRequiresUserAuthorization(t *testing.T) {
	_, fixtures := loadInputs(t)
	envelope := *fixtures[len(fixtures)-1].IntentEnvelope
	envelope.Contract.AcceptanceCriteria[0].SourceIntentIDs = []string{"E1"}
	if err := ValidateIntentEnvelope(envelope); err == nil {
		t.Fatal("expected external evidence to be rejected as contract authority")
	}
}

func TestAcceptedReceiptRejectsUnauthorizedSideEffect(t *testing.T) {
	_, fixtures := loadInputs(t)
	envelope := *fixtures[len(fixtures)-1].IntentEnvelope
	receipt := model.IntentReceipt{
		Version: "intent-v1", Status: "accepted", UnderstoodContract: "read-only evaluation",
		PlannedSideEffects: []string{"create-repository"}, AddedAssumptions: []string{}, Conflicts: []string{},
	}
	if err := ValidateIntentReceipt(envelope, receipt); err == nil {
		t.Fatal("expected unauthorized accepted receipt to fail")
	}
}

func TestConflictReceiptAcceptsObservedConflict(t *testing.T) {
	_, fixtures := loadInputs(t)
	envelope := *fixtures[len(fixtures)-1].IntentEnvelope
	receipt := model.IntentReceipt{
		Version: "intent-v1", Status: "conflict", UnderstoodContract: "read-only evaluation",
		PlannedSideEffects: []string{"create-repository"}, AddedAssumptions: []string{}, Conflicts: []string{"create-repository is forbidden"},
	}
	if err := ValidateIntentReceipt(envelope, receipt); err != nil {
		t.Fatal(err)
	}
}

func TestFixtureRequiresDeclaredAssertionAndQuestionIDs(t *testing.T) {
	_, fixtures := loadInputs(t)
	fixture := fixtures[0]
	fixture.Gold.Assertions[0].Key = "hidden-gold-only-key"
	if err := ValidateFixture(fixture); err == nil {
		t.Fatal("expected undeclared assertion rejection")
	}
	fixture = fixtures[2]
	fixture.Gold.HumanRequest.QuestionID = "hidden-question"
	if err := ValidateFixture(fixture); err == nil {
		t.Fatal("expected undeclared human question rejection")
	}
}

func TestConflictReceiptRejectsAddedAssumptions(t *testing.T) {
	_, fixtures := loadInputs(t)
	envelope := *fixtures[len(fixtures)-1].IntentEnvelope
	receipt := model.IntentReceipt{
		Version: "intent-v1", Status: "conflict", UnderstoodContract: "read-only evaluation",
		PlannedSideEffects: []string{"create-repository"}, AddedAssumptions: []string{"repository creation is implied"}, Conflicts: []string{"create-repository is forbidden"},
	}
	if err := ValidateIntentReceipt(envelope, receipt); err == nil {
		t.Fatal("expected added assumption rejection")
	}
}

func TestFixtureIDsAndClassesAreFrozenAndOpaque(t *testing.T) {
	_, fixtures := loadInputs(t)
	fixture := fixtures[0]
	fixture.DecisionOptions[0].ID = "opt-semantic-leak"
	if err := ValidateFixture(fixture); err == nil {
		t.Fatal("expected semantic option ID rejection")
	}
	fixtures = append([]model.Fixture(nil), fixtures...)
	fixtures[0].Class = "different_class"
	if err := ValidateFixtures(fixtures); err == nil {
		t.Fatal("expected fixture lineage mismatch rejection")
	}
}

func TestRunManifestRequiresCompleteUniqueMatrix(t *testing.T) {
	arms, fixtures := loadInputs(t)
	manifest := validManifest(arms, fixtures)
	if err := ValidateRunManifest(manifest, fixtures, arms); err != nil {
		t.Fatal(err)
	}
	manifest.ExecutionOrder = manifest.ExecutionOrder[:len(manifest.ExecutionOrder)-1]
	if err := ValidateRunManifest(manifest, fixtures, arms); err == nil {
		t.Fatal("expected incomplete execution order rejection")
	}
}

func TestObservationRejectsExtraCallsAndTools(t *testing.T) {
	arms, fixtures := loadInputs(t)
	fixtureIDs := map[string]struct{}{fixtures[0].ID: {}}
	armIDs := map[string]struct{}{arms[0].ID: {}}
	usage := model.Usage{}
	observation := model.Observation{
		Version: model.ObservationVersion, RunID: "run-1", CaseID: fixtures[0].ID, ArmID: arms[0].ID,
		Sender:   model.Attempt{Response: "sender", ExitCode: 0, Calls: 2, Retries: 0, ToolCalls: 0, Usage: usage},
		Handoff:  "sender",
		Receiver: model.Attempt{Response: "receiver", ExitCode: 0, Calls: 1, Retries: 0, ToolCalls: 0, Usage: usage},
		Defects:  []model.DefectAttribution{},
	}
	if err := ValidateObservation(observation, fixtureIDs, armIDs); err == nil {
		t.Fatal("expected extra call rejection")
	}
	observation.Sender.Calls = 1
	observation.Sender.ToolCalls = 1
	if err := ValidateObservation(observation, fixtureIDs, armIDs); err == nil {
		t.Fatal("expected tool-call rejection")
	}
}

func TestPricingPolicyMustCoverBothManifestModels(t *testing.T) {
	arms, fixtures := loadInputs(t)
	manifest := validManifest(arms, fixtures)
	policy := model.PricingPolicy{
		Version:  model.PricingVersion,
		Currency: "USD",
		Entries: []model.PricingEntry{{
			Provider: manifest.Sender.Provider, Model: manifest.Sender.Model,
			ReasoningIncludedInOutput: true, Source: "https://pricing.invalid/sender", VerifiedAt: "2026-08-12T00:00:00Z",
		}},
	}
	if err := ValidatePricingForManifest(policy, manifest); err == nil {
		t.Fatal("expected missing receiver pricing entry rejection")
	}
	policy.Entries = append(policy.Entries, model.PricingEntry{
		Provider: manifest.Receiver.Provider, Model: manifest.Receiver.Model,
		ReasoningIncludedInOutput: true, Source: "https://pricing.invalid/receiver", VerifiedAt: "2026-08-12T00:00:00Z",
	})
	if err := ValidatePricingForManifest(policy, manifest); err != nil {
		t.Fatal(err)
	}
}

func TestRunManifestRejectsRepositoryAccess(t *testing.T) {
	arms, fixtures := loadInputs(t)
	manifest := validManifest(arms, fixtures)
	manifest.Isolation.RepositoryAccess = true
	if err := ValidateRunManifest(manifest, fixtures, arms); err == nil {
		t.Fatal("expected repository access rejection")
	}
}

func TestRunManifestRejectsNonCanonicalOrder(t *testing.T) {
	arms, fixtures := loadInputs(t)
	manifest := validManifest(arms, fixtures)
	manifest.ExecutionOrder[0], manifest.ExecutionOrder[1] = manifest.ExecutionOrder[1], manifest.ExecutionOrder[0]
	if err := ValidateRunManifest(manifest, fixtures, arms); err == nil {
		t.Fatal("expected non-canonical execution order rejection")
	}
}

func validManifest(arms []model.Arm, fixtures []model.Fixture) model.RunManifest {
	manifest := model.RunManifest{
		Version: model.RunManifestVersion, Scope: model.ExperimentScope, RunID: "run-1", FreezeLockSHA256: strings.Repeat("a", 64),
		Sender:    model.ModelConfig{Provider: "test", Model: "sender", ReasoningEffort: "fixed"},
		Receiver:  model.ModelConfig{Provider: "test", Model: "receiver", ReasoningEffort: "fixed"},
		Transport: "test-harness", Isolation: model.IsolationConfig{Workspace: "empty-temporary", Network: "provider-only", Tools: []string{}}, EliminationPolicy: model.EliminationPolicy, RuntimeRetriesFixed: 0,
	}
	order, err := runplan.ExecutionOrder(fixtures, arms)
	if err != nil {
		panic(err)
	}
	manifest.ExecutionOrder = order
	return manifest
}
