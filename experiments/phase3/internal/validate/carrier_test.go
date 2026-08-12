package validate

import (
	"encoding/json"
	"testing"

	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/model"
)

func jsonArm(t *testing.T) (model.Arm, []model.Fixture) {
	t.Helper()
	arms, fixtures := loadInputs(t)
	for _, arm := range arms {
		if arm.Carrier == "json-carrier-v1" && !arm.BFV {
			return arm, fixtures
		}
	}
	t.Fatal("json arm missing")
	return model.Arm{}, nil
}

func canonicalJSONCarrier(f model.Fixture) JSONCarrier {
	required := map[string]struct{}{}
	for _, id := range f.Gold.RequiredActions {
		required[id] = struct{}{}
	}
	value := JSONCarrier{
		Summary:           "decision-relevant transfer",
		RecommendedOption: f.Gold.SelectedOption,
		Concepts:          append([]string(nil), f.Gold.RequiredConcepts...),
		Constraints:       []string{"preserve declared scope"},
	}
	for _, assertion := range f.Gold.Assertions {
		value.Facts = append(value.Facts, JSONCarrierFact(assertion))
	}
	for _, option := range f.ActionOptions {
		disposition := "exclude"
		if _, ok := required[option.ID]; ok {
			disposition = "include"
		}
		value.Actions = append(value.Actions, JSONCarrierAction{ID: option.ID, Disposition: disposition})
	}
	if f.Gold.HumanRequest.Mode == "required" {
		value.HumanBoundary = &model.HumanRequest{
			QuestionID: f.Gold.HumanRequest.QuestionID, RecommendationOption: f.Gold.HumanRequest.RecommendationOption,
			BlockingReason: "human-reserved authority",
		}
	}
	if f.IntentEnvelope != nil {
		value.IntentReceipt = &model.IntentReceipt{
			Version: "intent-v1", Status: "conflict", UnderstoodContract: "read-only evaluation only",
			AddedAssumptions: []string{}, PlannedSideEffects: append([]string(nil), f.Gold.IntentReceipt.PlannedSideEffects...),
			Conflicts: []string{"create-repository, install-plugin, and push-release are not authorized"},
		}
	}
	return value
}

func encodeCarrier(t *testing.T, value JSONCarrier) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestJSONCarrierRequiresCompleteFactAndActionCatalogs(t *testing.T) {
	arm, fixtures := jsonArm(t)
	value := canonicalJSONCarrier(fixtures[1])
	value.Facts = value.Facts[:1]
	if err := ValidateCarrier(arm, fixtures[1], encodeCarrier(t, value)); err == nil {
		t.Fatal("expected missing fact rejection")
	}
	value = canonicalJSONCarrier(fixtures[0])
	value.Actions = value.Actions[:len(value.Actions)-1]
	if err := ValidateCarrier(arm, fixtures[0], encodeCarrier(t, value)); err == nil {
		t.Fatal("expected missing action classification rejection")
	}
}

func TestJSONCarrierRequiresIntentReceiptOnlyForIntentFixture(t *testing.T) {
	arm, fixtures := jsonArm(t)
	intentFixture := fixtures[len(fixtures)-1]
	value := canonicalJSONCarrier(intentFixture)
	value.IntentReceipt = nil
	if err := ValidateCarrier(arm, intentFixture, encodeCarrier(t, value)); err == nil {
		t.Fatal("expected missing intent receipt rejection")
	}

	ordinary := fixtures[0]
	value = canonicalJSONCarrier(ordinary)
	value.IntentReceipt = &model.IntentReceipt{
		Version: "intent-v1", Status: "accepted", UnderstoodContract: "invented", AddedAssumptions: []string{},
		PlannedSideEffects: []string{}, Conflicts: []string{},
	}
	if err := ValidateCarrier(arm, ordinary, encodeCarrier(t, value)); err == nil {
		t.Fatal("expected unexpected intent receipt rejection")
	}
}
