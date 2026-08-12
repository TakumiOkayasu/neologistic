package validate

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/model"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/pathguard"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/runplan"
)

var expectedArms = map[string]struct {
	carrier string
	bfv     bool
}{
	"A": {carrier: "native", bfv: false},
	"B": {carrier: "native", bfv: true},
	"C": {carrier: "genshijin-normal", bfv: false},
	"D": {carrier: "genshijin-normal", bfv: true},
	"E": {carrier: "json-carrier-v1", bfv: false},
	"F": {carrier: "json-carrier-v1", bfv: true},
}

var expectedFixtures = map[string]string{
	"p3-001": "atomic_evidence_transfer",
	"p3-002": "epistemic_origin_preservation",
	"p3-003": "human_authority_boundary",
	"p3-004": "false_human_escalation",
	"p3-005": "scope_expansion_control",
	"p3-006": "reopen_after_refutation",
	"p3-007": "abstraction_preservation",
	"p3-008": "intent_provenance_gate",
}

var allowedOrigins = set("observed", "inferred", "synthetic")
var allowedStates = set("open", "verified", "refuted")
var allowedStatuses = set("resolved", "needs_human", "reopened", "blocked")
var allowedCompletion = set("ready_for_acceptance", "not_ready")
var allowedHumanModes = set("required", "forbidden")
var allowedDefectCategories = set("fixture", "carrier", "policy", "model", "runtime", "scorer", "human_contract", "unclassified")

func set(values ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(values))
	for _, value := range values {
		m[value] = struct{}{}
	}
	return m
}

func ValidateArms(root string, arms []model.Arm) error {
	if len(arms) != len(expectedArms) {
		return fmt.Errorf("expected %d arms, got %d", len(expectedArms), len(arms))
	}
	seen := map[string]struct{}{}
	for _, arm := range arms {
		if arm.Version != model.ArmVersion {
			return fmt.Errorf("arm %q: unsupported version %q", arm.ID, arm.Version)
		}
		expected, ok := expectedArms[arm.ID]
		if !ok {
			return fmt.Errorf("unexpected arm %q", arm.ID)
		}
		if _, duplicate := seen[arm.ID]; duplicate {
			return fmt.Errorf("duplicate arm %q", arm.ID)
		}
		seen[arm.ID] = struct{}{}
		if arm.Carrier != expected.carrier || arm.BFV != expected.bfv {
			return fmt.Errorf("arm %s: expected carrier=%s bfv=%v, got carrier=%s bfv=%v", arm.ID, expected.carrier, expected.bfv, arm.Carrier, arm.BFV)
		}
		if err := validateMechanisms(arm); err != nil {
			return err
		}
		if _, err := pathguard.ResolveRegular(root, arm.CarrierPrompt); err != nil {
			return fmt.Errorf("arm %s carrier prompt: %w", arm.ID, err)
		}
		if arm.BFV {
			if arm.PolicyPrompt == nil || strings.TrimSpace(*arm.PolicyPrompt) == "" {
				return fmt.Errorf("arm %s: BFV arm requires policy_prompt", arm.ID)
			}
			if _, err := pathguard.ResolveRegular(root, *arm.PolicyPrompt); err != nil {
				return fmt.Errorf("arm %s policy prompt: %w", arm.ID, err)
			}
		} else if arm.PolicyPrompt != nil {
			return fmt.Errorf("arm %s: non-BFV arm must not set policy_prompt", arm.ID)
		}
	}
	return nil
}

func validateMechanisms(arm model.Arm) error {
	expected := []string{}
	switch arm.Carrier {
	case "native":
	case "genshijin-normal":
		expected = append(expected, "genshijin-normal")
	case "json-carrier-v1":
		expected = append(expected, "json-carrier-v1")
	default:
		return fmt.Errorf("arm %s: unsupported carrier %q", arm.ID, arm.Carrier)
	}
	if arm.BFV {
		expected = append(expected, "bfv")
	}
	actual := append([]string(nil), arm.CustomMechanisms...)
	sort.Strings(actual)
	sort.Strings(expected)
	if strings.Join(actual, "\x00") != strings.Join(expected, "\x00") {
		return fmt.Errorf("arm %s: custom_mechanisms=%v, expected %v", arm.ID, arm.CustomMechanisms, expected)
	}
	return nil
}

func ValidateFixtures(fixtures []model.Fixture) error {
	if len(fixtures) != len(expectedFixtures) {
		return fmt.Errorf("expected %d fixture classes, got %d", len(expectedFixtures), len(fixtures))
	}
	seenID := map[string]struct{}{}
	seenClass := map[string]struct{}{}
	for _, fixture := range fixtures {
		expectedClass, ok := expectedFixtures[fixture.ID]
		if !ok {
			return fmt.Errorf("unexpected fixture id %q", fixture.ID)
		}
		if fixture.Class != expectedClass {
			return fmt.Errorf("fixture %s has class %q, expected %q", fixture.ID, fixture.Class, expectedClass)
		}
		if err := ValidateFixture(fixture); err != nil {
			return fmt.Errorf("fixture %q: %w", fixture.ID, err)
		}
		if _, duplicate := seenID[fixture.ID]; duplicate {
			return fmt.Errorf("duplicate fixture id %q", fixture.ID)
		}
		seenID[fixture.ID] = struct{}{}
		if _, duplicate := seenClass[fixture.Class]; duplicate {
			return fmt.Errorf("duplicate fixture class %q", fixture.Class)
		}
		seenClass[fixture.Class] = struct{}{}
	}
	return nil
}

func ValidateFixture(f model.Fixture) error {
	if f.Version != model.FixtureVersion {
		return fmt.Errorf("unsupported version %q", f.Version)
	}
	if strings.TrimSpace(f.ID) == "" || strings.TrimSpace(f.Class) == "" || strings.TrimSpace(f.Title) == "" || strings.TrimSpace(f.UserRequest) == "" {
		return fmt.Errorf("id, class, title, and user_request are required")
	}
	sourceIDs, err := uniqueOptions(f.Sources, func(v model.Source) string { return v.ID }, "source")
	if err != nil {
		return err
	}
	if len(sourceIDs) == 0 {
		return fmt.Errorf("at least one source is required")
	}
	if err := validateOpaqueSequence("source", sourceIDs, "s"); err != nil {
		return err
	}
	for _, source := range f.Sources {
		if strings.TrimSpace(source.Kind) == "" || strings.TrimSpace(source.Content) == "" {
			return fmt.Errorf("source %q requires kind and content", source.ID)
		}
	}
	decisionIDs, err := validateOptions(f.DecisionOptions, "decision option", true)
	if err != nil {
		return err
	}
	if err := validateOpaqueSequence("decision option", decisionIDs, "d"); err != nil {
		return err
	}
	assertionIDs, err := validateOptions(f.AssertionOptions, "assertion option", true)
	if err != nil {
		return err
	}
	if err := validateOpaqueSequence("assertion option", assertionIDs, "k"); err != nil {
		return err
	}
	actionIDs, err := validateOptions(f.ActionOptions, "action option", true)
	if err != nil {
		return err
	}
	if err := validateOpaqueSequence("action option", actionIDs, "a"); err != nil {
		return err
	}
	conceptIDs, err := validateOptions(f.ConceptOptions, "concept option", false)
	if err != nil {
		return err
	}
	if err := validateOpaqueSequence("concept option", conceptIDs, "c"); err != nil {
		return err
	}
	questionIDs, err := validateOptions(f.HumanQuestionOptions, "human question option", false)
	if err != nil {
		return err
	}
	if err := validateOpaqueSequence("human question option", questionIDs, "h"); err != nil {
		return err
	}
	if _, ok := decisionIDs[f.Gold.SelectedOption]; !ok {
		return fmt.Errorf("gold selected_option %q is not declared", f.Gold.SelectedOption)
	}
	if _, ok := allowedStatuses[f.Gold.Status]; !ok {
		return fmt.Errorf("invalid gold status %q", f.Gold.Status)
	}
	if _, ok := allowedCompletion[f.Gold.Completion]; !ok {
		return fmt.Errorf("invalid gold completion %q", f.Gold.Completion)
	}
	if _, ok := allowedHumanModes[f.Gold.HumanRequest.Mode]; !ok {
		return fmt.Errorf("invalid human_request mode %q", f.Gold.HumanRequest.Mode)
	}
	if f.Gold.HumanRequest.Mode == "required" {
		if f.Gold.HumanRequest.QuestionID == "" || f.Gold.HumanRequest.RecommendationOption == "" {
			return fmt.Errorf("required human request needs question_id and recommendation_option")
		}
		if _, ok := questionIDs[f.Gold.HumanRequest.QuestionID]; !ok {
			return fmt.Errorf("human question %q is not declared", f.Gold.HumanRequest.QuestionID)
		}
		if _, ok := decisionIDs[f.Gold.HumanRequest.RecommendationOption]; !ok {
			return fmt.Errorf("human recommendation %q is not a decision option", f.Gold.HumanRequest.RecommendationOption)
		}
	} else if f.Gold.HumanRequest.QuestionID != "" || f.Gold.HumanRequest.RecommendationOption != "" {
		return fmt.Errorf("forbidden human request must not declare question or recommendation")
	}
	assertionKeys := map[string]struct{}{}
	for _, assertion := range f.Gold.Assertions {
		if assertion.Key == "" {
			return fmt.Errorf("gold assertion key is required")
		}
		if _, ok := assertionIDs[assertion.Key]; !ok {
			return fmt.Errorf("gold assertion key %q is not declared", assertion.Key)
		}
		if _, ok := assertionKeys[assertion.Key]; ok {
			return fmt.Errorf("duplicate gold assertion key %q", assertion.Key)
		}
		assertionKeys[assertion.Key] = struct{}{}
		if _, ok := allowedOrigins[assertion.Origin]; !ok {
			return fmt.Errorf("assertion %s: invalid origin %q", assertion.Key, assertion.Origin)
		}
		if _, ok := allowedStates[assertion.State]; !ok {
			return fmt.Errorf("assertion %s: invalid state %q", assertion.Key, assertion.State)
		}
		if (assertion.State == "verified" || assertion.State == "refuted") && len(assertion.EvidenceIDs) == 0 {
			return fmt.Errorf("assertion %s: state %s requires evidence", assertion.Key, assertion.State)
		}
		if err := validateUniqueStrings("assertion evidence", assertion.EvidenceIDs, false); err != nil {
			return fmt.Errorf("assertion %s: %w", assertion.Key, err)
		}
		for _, evidenceID := range assertion.EvidenceIDs {
			if _, ok := sourceIDs[evidenceID]; !ok {
				return fmt.Errorf("assertion %s references unknown evidence %q", assertion.Key, evidenceID)
			}
		}
	}
	if len(assertionKeys) != len(assertionIDs) {
		return fmt.Errorf("every assertion option must have exactly one gold assertion")
	}
	if err := validateMembership("required action", f.Gold.RequiredActions, actionIDs); err != nil {
		return err
	}
	if err := validateMembership("forbidden action", f.Gold.ForbiddenActions, actionIDs); err != nil {
		return err
	}
	if overlap(f.Gold.RequiredActions, f.Gold.ForbiddenActions) {
		return fmt.Errorf("required_actions and forbidden_actions overlap")
	}
	if len(f.Gold.RequiredActions)+len(f.Gold.ForbiddenActions) != len(actionIDs) {
		return fmt.Errorf("every action option must be classified as required or forbidden")
	}
	if err := validateMembership("required concept", f.Gold.RequiredConcepts, conceptIDs); err != nil {
		return err
	}
	if err := validateMembership("forbidden concept", f.Gold.ForbiddenConcepts, conceptIDs); err != nil {
		return err
	}
	if overlap(f.Gold.RequiredConcepts, f.Gold.ForbiddenConcepts) {
		return fmt.Errorf("required_concepts and forbidden_concepts overlap")
	}
	if len(f.Gold.RequiredConcepts)+len(f.Gold.ForbiddenConcepts) != len(conceptIDs) {
		return fmt.Errorf("every concept option must be classified as required or forbidden")
	}
	if f.IntentEnvelope == nil && f.Gold.IntentReceipt.Required {
		return fmt.Errorf("intent receipt cannot be required without intent_envelope")
	}
	if f.IntentEnvelope != nil {
		if err := ValidateIntentEnvelope(*f.IntentEnvelope); err != nil {
			return fmt.Errorf("intent_envelope: %w", err)
		}
	}
	if err := validateIntentExpectation(f.Gold.IntentReceipt); err != nil {
		return err
	}
	return nil
}

func validateOptions(values []model.Option, kind string, required bool) (map[string]struct{}, error) {
	if required && len(values) == 0 {
		return nil, fmt.Errorf("at least one %s is required", kind)
	}
	ids, err := uniqueOptions(values, func(v model.Option) string { return v.ID }, kind)
	if err != nil {
		return nil, err
	}
	for _, value := range values {
		if strings.TrimSpace(value.Text) == "" {
			return nil, fmt.Errorf("%s %q requires text", kind, value.ID)
		}
	}
	return ids, nil
}

func uniqueOptions[T any](values []T, id func(T) string, kind string) (map[string]struct{}, error) {
	out := map[string]struct{}{}
	for _, value := range values {
		key := strings.TrimSpace(id(value))
		if key == "" {
			return nil, fmt.Errorf("%s id is required", kind)
		}
		if _, ok := out[key]; ok {
			return nil, fmt.Errorf("duplicate %s id %q", kind, key)
		}
		out[key] = struct{}{}
	}
	return out, nil
}

func validateOpaqueSequence(kind string, values map[string]struct{}, prefix string) error {
	for i := 1; i <= len(values); i++ {
		expected := fmt.Sprintf("%s%d", prefix, i)
		if _, ok := values[expected]; !ok {
			return fmt.Errorf("%s IDs must be opaque sequential values %s1..%s%d; missing %q", kind, prefix, prefix, len(values), expected)
		}
	}
	return nil
}

func validateMembership(kind string, values []string, allowed map[string]struct{}) error {
	seen := map[string]struct{}{}
	for _, value := range values {
		if _, ok := allowed[value]; !ok {
			return fmt.Errorf("%s %q is not declared", kind, value)
		}
		if _, ok := seen[value]; ok {
			return fmt.Errorf("duplicate %s %q", kind, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validateUniqueStrings(kind string, values []string, rejectEmptyList bool) error {
	if rejectEmptyList && len(values) == 0 {
		return fmt.Errorf("%s must not be empty", kind)
	}
	seen := map[string]struct{}{}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s contains an empty value", kind)
		}
		if _, ok := seen[value]; ok {
			return fmt.Errorf("%s contains duplicate %q", kind, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func overlap(a, b []string) bool {
	seen := map[string]struct{}{}
	for _, value := range a {
		seen[value] = struct{}{}
	}
	for _, value := range b {
		if _, ok := seen[value]; ok {
			return true
		}
	}
	return false
}

func validateIntentExpectation(expectation model.IntentExpectation) error {
	if !expectation.Required {
		if expectation.Status != "" || len(expectation.PlannedSideEffects) != 0 || len(expectation.RequiredConflictTerms) != 0 {
			return fmt.Errorf("non-required intent receipt must not contain expectations")
		}
		return nil
	}
	if expectation.Status != "accepted" && expectation.Status != "conflict" {
		return fmt.Errorf("required intent receipt has invalid status %q", expectation.Status)
	}
	if err := validateUniqueStrings("planned_side_effects", expectation.PlannedSideEffects, false); err != nil {
		return err
	}
	if err := validateUniqueStrings("required_conflict_terms", expectation.RequiredConflictTerms, expectation.Status == "conflict"); err != nil {
		return err
	}
	return nil
}

func ValidateIntentEnvelope(envelope model.IntentEnvelope) error {
	if envelope.Version != "intent-v1" {
		return fmt.Errorf("unsupported version %q", envelope.Version)
	}
	if strings.TrimSpace(envelope.UserRequest) == "" || strings.TrimSpace(envelope.Contract.Outcome) == "" {
		return fmt.Errorf("user_request and contract.outcome are required")
	}
	if len(envelope.Contract.AcceptanceCriteria) == 0 {
		return fmt.Errorf("at least one acceptance criterion is required")
	}
	all := map[string]string{}
	userAuthorized := map[string]struct{}{}
	addItems := func(kind string, items []model.IntentItem) error {
		for _, item := range items {
			if item.ID == "" || strings.TrimSpace(item.Text) == "" {
				return fmt.Errorf("%s intent item requires id and text", kind)
			}
			if previous, ok := all[item.ID]; ok {
				return fmt.Errorf("intent id %q appears in both %s and %s", item.ID, previous, kind)
			}
			all[item.ID] = kind
			if kind == "user_authorized" {
				userAuthorized[item.ID] = struct{}{}
			}
		}
		return nil
	}
	if err := addItems("user_authorized", envelope.Provenance.UserAuthorized); err != nil {
		return err
	}
	if err := addItems("assistant_inference", envelope.Provenance.AssistantInference); err != nil {
		return err
	}
	if err := addItems("external_evidence", envelope.Provenance.ExternalEvidence); err != nil {
		return err
	}
	if len(userAuthorized) == 0 {
		return fmt.Errorf("at least one user_authorized intent is required")
	}
	criterionIDs := map[string]struct{}{}
	for _, criterion := range envelope.Contract.AcceptanceCriteria {
		if criterion.ID == "" || strings.TrimSpace(criterion.Criterion) == "" || len(criterion.SourceIntentIDs) == 0 {
			return fmt.Errorf("acceptance criterion requires id, criterion, and source_intent_ids")
		}
		if _, ok := criterionIDs[criterion.ID]; ok {
			return fmt.Errorf("duplicate acceptance criterion id %q", criterion.ID)
		}
		criterionIDs[criterion.ID] = struct{}{}
		if err := validateUniqueStrings("criterion source_intent_ids", criterion.SourceIntentIDs, true); err != nil {
			return fmt.Errorf("criterion %s: %w", criterion.ID, err)
		}
		for _, sourceID := range criterion.SourceIntentIDs {
			if _, ok := userAuthorized[sourceID]; !ok {
				return fmt.Errorf("criterion %s is not authorized by USER_AUTHORIZED intent: %q", criterion.ID, sourceID)
			}
		}
	}
	grantIDs := map[string]struct{}{}
	effectValues := map[string]struct{}{}
	for _, grant := range envelope.SideEffects.Allowed {
		if grant.ID == "" || strings.TrimSpace(grant.Effect) == "" || len(grant.SourceIntentIDs) == 0 {
			return fmt.Errorf("allowed side effect requires id, effect, and source_intent_ids")
		}
		if _, ok := grantIDs[grant.ID]; ok {
			return fmt.Errorf("duplicate side-effect grant id %q", grant.ID)
		}
		grantIDs[grant.ID] = struct{}{}
		if _, ok := effectValues[grant.Effect]; ok {
			return fmt.Errorf("duplicate allowed side effect %q", grant.Effect)
		}
		effectValues[grant.Effect] = struct{}{}
		if err := validateUniqueStrings("side-effect source_intent_ids", grant.SourceIntentIDs, true); err != nil {
			return fmt.Errorf("side effect %s: %w", grant.ID, err)
		}
		for _, sourceID := range grant.SourceIntentIDs {
			if _, ok := userAuthorized[sourceID]; !ok {
				return fmt.Errorf("side effect %s is not authorized by USER_AUTHORIZED intent: %q", grant.ID, sourceID)
			}
		}
	}
	if err := validateUniqueStrings("forbidden side effects", envelope.SideEffects.Forbidden, false); err != nil {
		return err
	}
	for _, forbidden := range envelope.SideEffects.Forbidden {
		if _, ok := effectValues[forbidden]; ok {
			return fmt.Errorf("side effect %q is both allowed and forbidden", forbidden)
		}
	}
	if err := validateUniqueStrings("non_goals", envelope.NonGoals, false); err != nil {
		return err
	}
	return nil
}

func ValidateIntentReceipt(envelope model.IntentEnvelope, receipt model.IntentReceipt) error {
	if receipt.Version != "intent-v1" {
		return fmt.Errorf("unsupported receipt version %q", receipt.Version)
	}
	if receipt.Status != "accepted" && receipt.Status != "conflict" {
		return fmt.Errorf("invalid receipt status %q", receipt.Status)
	}
	if strings.TrimSpace(receipt.UnderstoodContract) == "" {
		return fmt.Errorf("understood_contract is required")
	}
	if err := validateUniqueStrings("added_assumptions", receipt.AddedAssumptions, false); err != nil {
		return err
	}
	if err := validateUniqueStrings("planned_side_effects", receipt.PlannedSideEffects, false); err != nil {
		return err
	}
	if err := validateUniqueStrings("conflicts", receipt.Conflicts, false); err != nil {
		return err
	}
	allowed := map[string]struct{}{}
	for _, grant := range envelope.SideEffects.Allowed {
		allowed[grant.Effect] = struct{}{}
	}
	forbidden := map[string]struct{}{}
	for _, effect := range envelope.SideEffects.Forbidden {
		forbidden[effect] = struct{}{}
	}
	hasConflict := false
	for _, effect := range receipt.PlannedSideEffects {
		if _, ok := allowed[effect]; !ok {
			hasConflict = true
		}
		if _, ok := forbidden[effect]; ok {
			hasConflict = true
		}
	}
	if len(receipt.AddedAssumptions) != 0 {
		return fmt.Errorf("intent receipt must not add assumptions; report discrepancies as conflicts")
	}
	if receipt.Status == "accepted" {
		if hasConflict || len(receipt.Conflicts) != 0 {
			return fmt.Errorf("accepted receipt contains unauthorized effects or conflicts")
		}
	} else if !hasConflict && len(receipt.Conflicts) == 0 {
		return fmt.Errorf("conflict receipt has no observable conflict")
	}
	return nil
}

func ValidateOutcome(f model.Fixture, outcome model.Outcome) error {
	if outcome.Version != model.OutcomeVersion {
		return fmt.Errorf("unsupported outcome version %q", outcome.Version)
	}
	if outcome.TaskID != f.ID {
		return fmt.Errorf("task_id %q does not match fixture %q", outcome.TaskID, f.ID)
	}
	if _, ok := allowedStatuses[outcome.Status]; !ok {
		return fmt.Errorf("invalid status %q", outcome.Status)
	}
	if _, ok := allowedCompletion[outcome.Completion]; !ok {
		return fmt.Errorf("invalid completion %q", outcome.Completion)
	}
	decisionIDs, _ := uniqueOptions(f.DecisionOptions, func(v model.Option) string { return v.ID }, "decision option")
	if _, ok := decisionIDs[outcome.SelectedOption]; !ok {
		return fmt.Errorf("selected_option %q is not declared", outcome.SelectedOption)
	}
	sourceIDs, _ := uniqueOptions(f.Sources, func(v model.Source) string { return v.ID }, "source")
	assertionIDs, _ := uniqueOptions(f.AssertionOptions, func(v model.Option) string { return v.ID }, "assertion option")
	assertionKeys := map[string]struct{}{}
	for _, assertion := range outcome.Assertions {
		if assertion.Key == "" {
			return fmt.Errorf("assertion key is required")
		}
		if _, ok := assertionIDs[assertion.Key]; !ok {
			return fmt.Errorf("assertion key %q is not declared", assertion.Key)
		}
		if _, ok := assertionKeys[assertion.Key]; ok {
			return fmt.Errorf("duplicate assertion key %q", assertion.Key)
		}
		assertionKeys[assertion.Key] = struct{}{}
		if _, ok := allowedOrigins[assertion.Origin]; !ok {
			return fmt.Errorf("assertion %s has invalid origin %q", assertion.Key, assertion.Origin)
		}
		if _, ok := allowedStates[assertion.State]; !ok {
			return fmt.Errorf("assertion %s has invalid state %q", assertion.Key, assertion.State)
		}
		if (assertion.State == "verified" || assertion.State == "refuted") && len(assertion.EvidenceIDs) == 0 {
			return fmt.Errorf("assertion %s requires evidence", assertion.Key)
		}
		if err := validateUniqueStrings("assertion evidence", assertion.EvidenceIDs, false); err != nil {
			return fmt.Errorf("assertion %s: %w", assertion.Key, err)
		}
		for _, evidenceID := range assertion.EvidenceIDs {
			if _, ok := sourceIDs[evidenceID]; !ok {
				return fmt.Errorf("assertion %s references unknown evidence %q", assertion.Key, evidenceID)
			}
		}
	}
	actionIDs, _ := uniqueOptions(f.ActionOptions, func(v model.Option) string { return v.ID }, "action option")
	if err := validateMembership("action", outcome.Actions, actionIDs); err != nil {
		return err
	}
	conceptIDs, _ := uniqueOptions(f.ConceptOptions, func(v model.Option) string { return v.ID }, "concept option")
	if err := validateMembership("concept", outcome.Concepts, conceptIDs); err != nil {
		return err
	}
	questionIDs, _ := uniqueOptions(f.HumanQuestionOptions, func(v model.Option) string { return v.ID }, "human question option")
	if outcome.HumanRequest != nil {
		if outcome.HumanRequest.QuestionID == "" || outcome.HumanRequest.RecommendationOption == "" || outcome.HumanRequest.BlockingReason == "" {
			return fmt.Errorf("human_request requires question_id, recommendation_option, and blocking_reason")
		}
		if _, ok := questionIDs[outcome.HumanRequest.QuestionID]; !ok {
			return fmt.Errorf("human question %q is not declared", outcome.HumanRequest.QuestionID)
		}
		if _, ok := decisionIDs[outcome.HumanRequest.RecommendationOption]; !ok {
			return fmt.Errorf("human recommendation %q is not a decision option", outcome.HumanRequest.RecommendationOption)
		}
	}
	if outcome.IntentReceipt != nil {
		if f.IntentEnvelope == nil {
			return fmt.Errorf("intent_receipt present without intent_envelope")
		}
		if err := ValidateIntentReceipt(*f.IntentEnvelope, *outcome.IntentReceipt); err != nil {
			return fmt.Errorf("intent_receipt: %w", err)
		}
	}
	return nil
}

func ValidateObservation(o model.Observation, fixtureIDs, armIDs map[string]struct{}) error {
	if o.Version != model.ObservationVersion {
		return fmt.Errorf("unsupported observation version %q", o.Version)
	}
	if o.RunID == "" {
		return fmt.Errorf("run_id is required")
	}
	if _, ok := fixtureIDs[o.CaseID]; !ok {
		return fmt.Errorf("unknown case_id %q", o.CaseID)
	}
	if _, ok := armIDs[o.ArmID]; !ok {
		return fmt.Errorf("unknown arm_id %q", o.ArmID)
	}
	if err := validateAttempt("sender", o.Sender); err != nil {
		return err
	}
	if err := validateAttempt("receiver", o.Receiver); err != nil {
		return err
	}
	if o.HumanInterventions < 0 || o.CorrectionTurns < 0 || o.Converter.Calls < 0 || o.Converter.LatencyMS < 0 {
		return fmt.Errorf("human and converter counters must be non-negative")
	}
	for _, defect := range o.Defects {
		if _, ok := allowedDefectCategories[defect.Category]; !ok {
			return fmt.Errorf("invalid defect category %q", defect.Category)
		}
		if defect.Stage == "" || strings.TrimSpace(defect.Reason) == "" {
			return fmt.Errorf("defect requires stage and reason")
		}
	}
	return nil
}

func validateAttempt(name string, attempt model.Attempt) error {
	if attempt.Calls < 1 {
		return fmt.Errorf("%s calls must be at least 1", name)
	}
	if attempt.Retries < 0 || attempt.Calls != attempt.Retries+1 {
		return fmt.Errorf("%s calls must equal retries+1", name)
	}
	if attempt.ToolCalls != 0 {
		return fmt.Errorf("%s tool_calls must be zero in the isolated semantic-transfer pilot", name)
	}
	if attempt.LatencyMS < 0 {
		return fmt.Errorf("%s latency counter must be non-negative", name)
	}
	for label, value := range map[string]*int64{
		"input_tokens": attempt.Usage.InputTokens, "cached_input_tokens": attempt.Usage.CachedInputTokens,
		"cache_write_input_tokens": attempt.Usage.CacheWriteInputTokens, "output_tokens": attempt.Usage.OutputTokens,
		"reasoning_output_tokens": attempt.Usage.ReasoningOutputTokens,
	} {
		if value != nil && *value < 0 {
			return fmt.Errorf("%s %s must be non-negative", name, label)
		}
	}
	return nil
}

func ValidatePricingPolicy(policy model.PricingPolicy) error {
	if policy.Version != model.PricingVersion {
		return fmt.Errorf("unsupported pricing version %q", policy.Version)
	}
	if policy.Currency != "USD" {
		return fmt.Errorf("unsupported pricing currency %q", policy.Currency)
	}
	if len(policy.Entries) == 0 {
		return fmt.Errorf("pricing policy requires at least one entry")
	}
	seen := map[string]struct{}{}
	for index, entry := range policy.Entries {
		if strings.TrimSpace(entry.Provider) == "" || strings.TrimSpace(entry.Model) == "" || strings.TrimSpace(entry.Source) == "" || strings.TrimSpace(entry.VerifiedAt) == "" {
			return fmt.Errorf("pricing entry %d requires provider, model, source, and verified_at", index)
		}
		key := entry.Provider + "\x00" + entry.Model
		if _, duplicate := seen[key]; duplicate {
			return fmt.Errorf("duplicate pricing entry for %s/%s", entry.Provider, entry.Model)
		}
		seen[key] = struct{}{}
		if _, err := time.Parse(time.RFC3339, entry.VerifiedAt); err != nil {
			return fmt.Errorf("pricing entry %s/%s verified_at must be RFC3339: %w", entry.Provider, entry.Model, err)
		}
		if entry.UncachedInputNanoPerToken < 0 || entry.CachedInputNanoPerToken < 0 || entry.CacheWriteNanoPerToken < 0 || entry.OutputNanoPerToken < 0 {
			return fmt.Errorf("pricing entry %s/%s rates must be non-negative", entry.Provider, entry.Model)
		}
		if !entry.ReasoningIncludedInOutput {
			return fmt.Errorf("pricing entry %s/%s must state that reasoning output is included in output tokens", entry.Provider, entry.Model)
		}
		if entry.MaxInputTokensPerCall != nil && *entry.MaxInputTokensPerCall < 1 {
			return fmt.Errorf("pricing entry %s/%s max_input_tokens_per_call must be positive", entry.Provider, entry.Model)
		}
	}
	return nil
}

func ValidatePricingForManifest(policy model.PricingPolicy, manifest model.RunManifest) error {
	if err := ValidatePricingPolicy(policy); err != nil {
		return err
	}
	for label, config := range map[string]model.ModelConfig{"sender": manifest.Sender, "receiver": manifest.Receiver} {
		found := false
		for _, entry := range policy.Entries {
			if entry.Provider == config.Provider && entry.Model == config.Model {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("pricing policy has no entry for %s model %s/%s", label, config.Provider, config.Model)
		}
	}
	return nil
}

func validateIsolation(isolation model.IsolationConfig) error {
	if isolation.Workspace != "empty-temporary" || isolation.RepositoryAccess || isolation.PluginsEnabled || isolation.MemoryEnabled || isolation.Network != "provider-only" || len(isolation.Tools) != 0 {
		return fmt.Errorf("run isolation must use an empty temporary workspace with no repository, tools, plugins, memory, or non-provider network access")
	}
	return nil
}

func ValidateRunManifest(manifest model.RunManifest, fixtures []model.Fixture, arms []model.Arm) error {
	if manifest.Version != model.RunManifestVersion {
		return fmt.Errorf("unsupported run manifest version %q", manifest.Version)
	}
	if manifest.Scope != model.ExperimentScope {
		return fmt.Errorf("unsupported experiment scope %q", manifest.Scope)
	}
	if manifest.EliminationPolicy != model.EliminationPolicy {
		return fmt.Errorf("unsupported elimination policy %q", manifest.EliminationPolicy)
	}
	if err := validateIsolation(manifest.Isolation); err != nil {
		return err
	}
	if strings.TrimSpace(manifest.RunID) == "" || strings.TrimSpace(manifest.Transport) == "" {
		return fmt.Errorf("run_id and transport are required")
	}
	if len(manifest.FreezeLockSHA256) != 64 {
		return fmt.Errorf("freeze_lock_sha256 must be 64 hexadecimal characters")
	}
	if _, err := hex.DecodeString(manifest.FreezeLockSHA256); err != nil {
		return fmt.Errorf("freeze_lock_sha256 is invalid: %w", err)
	}
	for label, config := range map[string]model.ModelConfig{"sender": manifest.Sender, "receiver": manifest.Receiver} {
		if strings.TrimSpace(config.Provider) == "" || strings.TrimSpace(config.Model) == "" || strings.TrimSpace(config.ReasoningEffort) == "" {
			return fmt.Errorf("%s provider, model, and reasoning_effort are required", label)
		}
		if config.Temperature != nil && (*config.Temperature < 0 || *config.Temperature > 2) {
			return fmt.Errorf("%s temperature must be between 0 and 2 when reported", label)
		}
		if config.MaxOutputTokens != nil && *config.MaxOutputTokens < 1 {
			return fmt.Errorf("%s max_output_tokens must be positive when reported", label)
		}
	}
	if manifest.RuntimeRetriesFixed < 0 {
		return fmt.Errorf("runtime_retries_fixed must be non-negative")
	}
	expectedOrder, err := runplan.ExecutionOrder(fixtures, arms)
	if err != nil {
		return err
	}
	if len(manifest.ExecutionOrder) != len(expectedOrder) {
		return fmt.Errorf("execution_order has %d cells, expected %d", len(manifest.ExecutionOrder), len(expectedOrder))
	}
	for index, expected := range expectedOrder {
		actual := manifest.ExecutionOrder[index]
		if actual != expected {
			return fmt.Errorf("execution_order[%d]=%s/%s, expected %s/%s", index, actual.CaseID, actual.ArmID, expected.CaseID, expected.ArmID)
		}
	}
	return nil
}
