package render

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/codec"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/model"
)

func inputs(t *testing.T) (string, []model.Arm, []model.Fixture) {
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
	return root, arms, fixtures
}

func TestRenderedPromptsDoNotLeakScorerLabelsOrGold(t *testing.T) {
	root, arms, fixtures := inputs(t)
	for _, arm := range arms {
		for _, fixture := range fixtures {
			sender, err := Sender(root, arm, fixture)
			if err != nil {
				t.Fatal(err)
			}
			receiver, err := Receiver(root, arm, fixture, "handoff")
			if err != nil {
				t.Fatal(err)
			}
			for stage, prompt := range map[string]string{"sender": sender, "receiver": receiver} {
				for _, leaked := range []string{`"gold"`, `"class"`, `"title"`, fixture.Class, fixture.Title} {
					if strings.Contains(prompt, leaked) {
						t.Fatalf("%s prompt leaked %q for %s/%s", stage, leaked, fixture.ID, arm.ID)
					}
				}
			}
		}
	}
}

func TestReceiverCannotReadSourceContents(t *testing.T) {
	root, arms, fixtures := inputs(t)
	for _, arm := range arms {
		for _, fixture := range fixtures {
			prompt, err := Receiver(root, arm, fixture, "handoff")
			if err != nil {
				t.Fatal(err)
			}
			start := strings.Index(prompt, "<receiver_task_json>")
			end := strings.Index(prompt, "</receiver_task_json>")
			if start < 0 || end <= start {
				t.Fatalf("receiver task shell missing for %s/%s", fixture.ID, arm.ID)
			}
			shell := prompt[start:end]
			for _, source := range fixture.Sources {
				if strings.Contains(shell, source.Content) {
					t.Fatalf("receiver leaked source content %s for %s/%s", source.ID, fixture.ID, arm.ID)
				}
				if !strings.Contains(shell, `"`+source.ID+`"`) {
					t.Fatalf("receiver lacks source ID %s for %s/%s", source.ID, fixture.ID, arm.ID)
				}
			}
			jsonText := strings.TrimPrefix(shell, "<receiver_task_json>\n")
			var task map[string]any
			if err := json.Unmarshal([]byte(jsonText), &task); err != nil {
				t.Fatal(err)
			}
			for _, forbiddenKey := range []string{"user_request", "sources", "decision_options", "assertion_options", "action_options", "concept_options", "human_question_options"} {
				if _, ok := task[forbiddenKey]; ok {
					t.Fatalf("receiver task shell leaked semantic field %s for %s/%s", forbiddenKey, fixture.ID, arm.ID)
				}
			}
		}
	}
}

func TestReceiverIntentAuthorityExcludesUntrustedProvenance(t *testing.T) {
	root, arms, fixtures := inputs(t)
	var fixture *model.Fixture
	for i := range fixtures {
		if fixtures[i].IntentEnvelope != nil {
			fixture = &fixtures[i]
			break
		}
	}
	if fixture == nil {
		t.Fatal("intent fixture missing")
	}
	prompt, err := Receiver(root, arms[0], *fixture, "handoff")
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(prompt, "<receiver_task_json>")
	end := strings.Index(prompt, "</receiver_task_json>")
	if start < 0 || end <= start {
		t.Fatal("receiver task shell missing")
	}
	shell := prompt[start:end]
	for _, required := range []string{`"intent_authority"`, `"user_authorized"`, `"side_effects"`, `"non_goals"`} {
		if !strings.Contains(shell, required) {
			t.Fatalf("receiver authority projection lacks %s", required)
		}
	}
	for _, forbidden := range []string{`"intent_envelope"`, `"assistant_inference"`, `"external_evidence"`} {
		if strings.Contains(shell, forbidden) {
			t.Fatalf("receiver authority projection leaked %s", forbidden)
		}
	}
}

func TestCarrierAndPolicyAxesRemainOrthogonal(t *testing.T) {
	root, arms, fixtures := inputs(t)
	fixture := fixtures[0]
	prompts := map[string]string{}
	for _, arm := range arms {
		prompt, err := Sender(root, arm, fixture)
		if err != nil {
			t.Fatal(err)
		}
		prompts[arm.ID] = prompt
	}
	if !strings.Contains(prompts["B"], "Control policy: BFV") || strings.Contains(prompts["A"], "Control policy: BFV") {
		t.Fatal("native BFV pair does not differ on policy as expected")
	}
	if !strings.Contains(prompts["C"], "genshijin-like normal mode") || !strings.Contains(prompts["E"], "json-carrier-v1") {
		t.Fatal("carrier instructions missing")
	}
	if strings.Contains(prompts["E"], "R1|target") || strings.Contains(prompts["E"], "Q1|target") {
		t.Fatal("neutral JSON carrier leaked BFV-coupled R1/Q1/D1 grammar")
	}
}

func TestReceiverPromptUsesCommonOutcomeSchema(t *testing.T) {
	root, arms, fixtures := inputs(t)
	for _, arm := range arms {
		prompt, err := Receiver(root, arm, fixtures[0], "handoff")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(prompt, `"version":"phase3-outcome-v1"`) {
			t.Fatalf("arm %s lacks common outcome schema", arm.ID)
		}
	}
}
