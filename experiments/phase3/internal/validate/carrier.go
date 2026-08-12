package validate

import (
	"fmt"
	"strings"

	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/codec"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/model"
)

type JSONCarrierFact struct {
	Key         string   `json:"key"`
	Origin      string   `json:"origin"`
	State       string   `json:"state"`
	EvidenceIDs []string `json:"evidence_ids"`
}

type JSONCarrierAction struct {
	ID          string `json:"id"`
	Disposition string `json:"disposition"`
}

type JSONCarrier struct {
	Summary           string               `json:"summary"`
	Facts             []JSONCarrierFact    `json:"facts"`
	RecommendedOption string               `json:"recommended_option"`
	Actions           []JSONCarrierAction  `json:"actions"`
	Concepts          []string             `json:"concepts"`
	HumanBoundary     *model.HumanRequest  `json:"human_boundary"`
	Constraints       []string             `json:"constraints"`
	IntentReceipt     *model.IntentReceipt `json:"intent_receipt"`
}

func ValidateCarrier(arm model.Arm, fixture model.Fixture, response string) error {
	if strings.TrimSpace(response) == "" {
		return fmt.Errorf("empty sender response")
	}
	if strings.Contains(response, "```") {
		return fmt.Errorf("sender response contains Markdown fence")
	}
	switch arm.Carrier {
	case "native", "genshijin-normal":
		return nil
	case "json-carrier-v1":
		value, err := codec.DecodeStrict[JSONCarrier]([]byte(response))
		if err != nil {
			return fmt.Errorf("json carrier parse: %w", err)
		}
		return validateJSONCarrier(fixture, value)
	default:
		return fmt.Errorf("unsupported carrier %q", arm.Carrier)
	}
}

func validateJSONCarrier(f model.Fixture, value JSONCarrier) error {
	if strings.TrimSpace(value.Summary) == "" {
		return fmt.Errorf("summary is required")
	}
	decisions := map[string]struct{}{}
	for _, option := range f.DecisionOptions {
		decisions[option.ID] = struct{}{}
	}
	assertions := map[string]struct{}{}
	for _, option := range f.AssertionOptions {
		assertions[option.ID] = struct{}{}
	}
	if _, ok := decisions[value.RecommendedOption]; !ok {
		return fmt.Errorf("recommended_option %q is not declared", value.RecommendedOption)
	}
	sources := map[string]struct{}{}
	for _, source := range f.Sources {
		sources[source.ID] = struct{}{}
	}
	factKeys := map[string]struct{}{}
	for _, fact := range value.Facts {
		if fact.Key == "" {
			return fmt.Errorf("fact key is required")
		}
		if _, ok := assertions[fact.Key]; !ok {
			return fmt.Errorf("fact key %q is not declared", fact.Key)
		}
		if _, ok := factKeys[fact.Key]; ok {
			return fmt.Errorf("duplicate fact key %q", fact.Key)
		}
		factKeys[fact.Key] = struct{}{}
		if _, ok := allowedOrigins[fact.Origin]; !ok {
			return fmt.Errorf("fact %s has invalid origin %q", fact.Key, fact.Origin)
		}
		if _, ok := allowedStates[fact.State]; !ok {
			return fmt.Errorf("fact %s has invalid state %q", fact.Key, fact.State)
		}
		if (fact.State == "verified" || fact.State == "refuted") && len(fact.EvidenceIDs) == 0 {
			return fmt.Errorf("fact %s requires evidence", fact.Key)
		}
		if err := validateUniqueStrings("fact evidence", fact.EvidenceIDs, false); err != nil {
			return fmt.Errorf("fact %s: %w", fact.Key, err)
		}
		for _, evidenceID := range fact.EvidenceIDs {
			if _, ok := sources[evidenceID]; !ok {
				return fmt.Errorf("fact %s references unknown evidence %q", fact.Key, evidenceID)
			}
		}
	}
	if len(factKeys) != len(assertions) {
		return fmt.Errorf("json carrier must contain exactly one fact for every declared assertion")
	}
	actions := map[string]struct{}{}
	for _, option := range f.ActionOptions {
		actions[option.ID] = struct{}{}
	}
	seenActions := map[string]struct{}{}
	for _, action := range value.Actions {
		if _, ok := actions[action.ID]; !ok {
			return fmt.Errorf("action %q is not declared", action.ID)
		}
		if _, ok := seenActions[action.ID]; ok {
			return fmt.Errorf("duplicate action %q", action.ID)
		}
		seenActions[action.ID] = struct{}{}
		if action.Disposition != "include" && action.Disposition != "exclude" {
			return fmt.Errorf("action %s has invalid disposition %q", action.ID, action.Disposition)
		}
	}
	if len(seenActions) != len(actions) {
		return fmt.Errorf("json carrier must classify every declared action as include or exclude")
	}
	if err := validateUniqueStrings("constraints", value.Constraints, false); err != nil {
		return err
	}
	concepts := map[string]struct{}{}
	for _, option := range f.ConceptOptions {
		concepts[option.ID] = struct{}{}
	}
	seenConcepts := map[string]struct{}{}
	for _, concept := range value.Concepts {
		if _, ok := concepts[concept]; !ok {
			return fmt.Errorf("concept %q is not declared", concept)
		}
		if _, ok := seenConcepts[concept]; ok {
			return fmt.Errorf("duplicate concept %q", concept)
		}
		seenConcepts[concept] = struct{}{}
	}
	questions := map[string]struct{}{}
	for _, option := range f.HumanQuestionOptions {
		questions[option.ID] = struct{}{}
	}
	if value.HumanBoundary != nil {
		if value.HumanBoundary.QuestionID == "" || value.HumanBoundary.RecommendationOption == "" || value.HumanBoundary.BlockingReason == "" {
			return fmt.Errorf("human_boundary is incomplete")
		}
		if _, ok := questions[value.HumanBoundary.QuestionID]; !ok {
			return fmt.Errorf("human_boundary question %q is not declared", value.HumanBoundary.QuestionID)
		}
		if _, ok := decisions[value.HumanBoundary.RecommendationOption]; !ok {
			return fmt.Errorf("human_boundary recommendation %q is not declared", value.HumanBoundary.RecommendationOption)
		}
	}
	if f.IntentEnvelope == nil {
		if value.IntentReceipt != nil {
			return fmt.Errorf("intent_receipt present without intent_envelope")
		}
	} else {
		if value.IntentReceipt == nil {
			return fmt.Errorf("intent_receipt is required when intent_envelope is present")
		}
		if err := ValidateIntentReceipt(*f.IntentEnvelope, *value.IntentReceipt); err != nil {
			return err
		}
	}
	return nil
}
