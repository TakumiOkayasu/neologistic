package model

const (
	FixtureVersion     = "phase3-fixture-v2"
	ObservationVersion = "phase3-observation-v1"
	OutcomeVersion     = "phase3-outcome-v1"
	ArmVersion         = "phase3-arm-v1"
	FreezeVersion      = "phase3-freeze-v2"
	ReportVersion      = "phase3-report-v2"
	RunManifestVersion = "phase3-run-manifest-v2"
	PricingVersion     = "phase3-pricing-v2"
	ExperimentScope    = "semantic-transfer-pilot-v1"
	EliminationPolicy  = "after-first-hard-failure"
)

type Arm struct {
	Version          string   `json:"version"`
	ID               string   `json:"id"`
	Carrier          string   `json:"carrier"`
	BFV              bool     `json:"bfv"`
	CustomMechanisms []string `json:"custom_mechanisms"`
	CarrierPrompt    string   `json:"carrier_prompt"`
	PolicyPrompt     *string  `json:"policy_prompt"`
}

type Source struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Content string `json:"content"`
}

type Option struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type Assertion struct {
	Key         string   `json:"key"`
	Origin      string   `json:"origin"`
	State       string   `json:"state"`
	EvidenceIDs []string `json:"evidence_ids"`
}

type HumanRequest struct {
	QuestionID           string `json:"question_id"`
	RecommendationOption string `json:"recommendation_option"`
	BlockingReason       string `json:"blocking_reason"`
}

type IntentItem struct {
	ID        string  `json:"id"`
	Text      string  `json:"text"`
	Reference *string `json:"reference,omitempty"`
}

type AcceptanceCriterion struct {
	ID              string   `json:"id"`
	Criterion       string   `json:"criterion"`
	SourceIntentIDs []string `json:"source_intent_ids"`
}

type IntentContract struct {
	Outcome            string                `json:"outcome"`
	AcceptanceCriteria []AcceptanceCriterion `json:"acceptance_criteria"`
}

type IntentProvenance struct {
	UserAuthorized     []IntentItem `json:"user_authorized"`
	AssistantInference []IntentItem `json:"assistant_inference"`
	ExternalEvidence   []IntentItem `json:"external_evidence"`
}

type SideEffectGrant struct {
	ID              string   `json:"id"`
	Effect          string   `json:"effect"`
	SourceIntentIDs []string `json:"source_intent_ids"`
}

type SideEffects struct {
	Allowed   []SideEffectGrant `json:"allowed"`
	Forbidden []string          `json:"forbidden"`
}

type IntentEnvelope struct {
	Version     string           `json:"version"`
	UserRequest string           `json:"user_request"`
	Contract    IntentContract   `json:"contract"`
	Provenance  IntentProvenance `json:"provenance"`
	SideEffects SideEffects      `json:"side_effects"`
	NonGoals    []string         `json:"non_goals"`
}

type IntentReceipt struct {
	Version            string   `json:"version"`
	Status             string   `json:"status"`
	UnderstoodContract string   `json:"understood_contract"`
	AddedAssumptions   []string `json:"added_assumptions"`
	PlannedSideEffects []string `json:"planned_side_effects"`
	Conflicts          []string `json:"conflicts"`
}

type ReceiverIntentAuthority struct {
	Version        string         `json:"version"`
	UserRequest    string         `json:"user_request"`
	Contract       IntentContract `json:"contract"`
	UserAuthorized []IntentItem   `json:"user_authorized"`
	SideEffects    SideEffects    `json:"side_effects"`
	NonGoals       []string       `json:"non_goals"`
}

type HumanExpectation struct {
	Mode                 string `json:"mode"`
	QuestionID           string `json:"question_id,omitempty"`
	RecommendationOption string `json:"recommendation_option,omitempty"`
}

type IntentExpectation struct {
	Required              bool     `json:"required"`
	Status                string   `json:"status,omitempty"`
	PlannedSideEffects    []string `json:"planned_side_effects,omitempty"`
	RequiredConflictTerms []string `json:"required_conflict_terms,omitempty"`
}

type Gold struct {
	SelectedOption    string            `json:"selected_option"`
	Status            string            `json:"status"`
	Assertions        []Assertion       `json:"assertions"`
	RequiredActions   []string          `json:"required_actions"`
	ForbiddenActions  []string          `json:"forbidden_actions"`
	HumanRequest      HumanExpectation  `json:"human_request"`
	Completion        string            `json:"completion"`
	RequiredConcepts  []string          `json:"required_concepts"`
	ForbiddenConcepts []string          `json:"forbidden_concepts"`
	IntentReceipt     IntentExpectation `json:"intent_receipt"`
}

type Fixture struct {
	Version              string          `json:"version"`
	ID                   string          `json:"id"`
	Class                string          `json:"class"`
	Title                string          `json:"title"`
	UserRequest          string          `json:"user_request"`
	Sources              []Source        `json:"sources"`
	DecisionOptions      []Option        `json:"decision_options"`
	AssertionOptions     []Option        `json:"assertion_options"`
	ActionOptions        []Option        `json:"action_options"`
	ConceptOptions       []Option        `json:"concept_options"`
	HumanQuestionOptions []Option        `json:"human_question_options"`
	IntentEnvelope       *IntentEnvelope `json:"intent_envelope,omitempty"`
	Gold                 Gold            `json:"gold"`
}

// SenderFixture is the complete public evidence package provided only to the sender.
// Class, title, and hidden gold remain scorer-only labels.
type SenderFixture struct {
	Version              string          `json:"version"`
	ID                   string          `json:"id"`
	UserRequest          string          `json:"user_request"`
	Sources              []Source        `json:"sources"`
	DecisionOptions      []Option        `json:"decision_options"`
	AssertionOptions     []Option        `json:"assertion_options"`
	ActionOptions        []Option        `json:"action_options"`
	ConceptOptions       []Option        `json:"concept_options"`
	HumanQuestionOptions []Option        `json:"human_question_options"`
	IntentEnvelope       *IntentEnvelope `json:"intent_envelope,omitempty"`
}

// ReceiverFixture deliberately omits source contents. The receiver must rely on
// the sender handoff for derived facts while retaining immutable task authority.
type ReceiverFixture struct {
	Version                string                   `json:"version"`
	ID                     string                   `json:"id"`
	SourceIDs              []string                 `json:"source_ids"`
	DecisionOptionIDs      []string                 `json:"decision_option_ids"`
	AssertionOptionIDs     []string                 `json:"assertion_option_ids"`
	ActionOptionIDs        []string                 `json:"action_option_ids"`
	ConceptOptionIDs       []string                 `json:"concept_option_ids"`
	HumanQuestionOptionIDs []string                 `json:"human_question_option_ids"`
	IntentAuthority        *ReceiverIntentAuthority `json:"intent_authority,omitempty"`
}

func (f Fixture) SenderPublic() SenderFixture {
	return SenderFixture{
		Version: f.Version, ID: f.ID, UserRequest: f.UserRequest,
		Sources: f.Sources, DecisionOptions: f.DecisionOptions,
		AssertionOptions: f.AssertionOptions, ActionOptions: f.ActionOptions,
		ConceptOptions: f.ConceptOptions, HumanQuestionOptions: f.HumanQuestionOptions,
		IntentEnvelope: f.IntentEnvelope,
	}
}

func (f Fixture) ReceiverPublic() ReceiverFixture {
	var authority *ReceiverIntentAuthority
	if f.IntentEnvelope != nil {
		authority = &ReceiverIntentAuthority{
			Version:        f.IntentEnvelope.Version,
			UserRequest:    f.IntentEnvelope.UserRequest,
			Contract:       f.IntentEnvelope.Contract,
			UserAuthorized: f.IntentEnvelope.Provenance.UserAuthorized,
			SideEffects:    f.IntentEnvelope.SideEffects,
			NonGoals:       f.IntentEnvelope.NonGoals,
		}
	}

	return ReceiverFixture{
		Version:                f.Version,
		ID:                     f.ID,
		SourceIDs:              optionIDsFromSources(f.Sources),
		DecisionOptionIDs:      optionIDs(f.DecisionOptions),
		AssertionOptionIDs:     optionIDs(f.AssertionOptions),
		ActionOptionIDs:        optionIDs(f.ActionOptions),
		ConceptOptionIDs:       optionIDs(f.ConceptOptions),
		HumanQuestionOptionIDs: optionIDs(f.HumanQuestionOptions),
		IntentAuthority:        authority,
	}
}

func optionIDs(values []Option) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, value.ID)
	}
	return out
}

func optionIDsFromSources(values []Source) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, value.ID)
	}
	return out
}

type Outcome struct {
	Version        string         `json:"version"`
	TaskID         string         `json:"task_id"`
	SelectedOption string         `json:"selected_option"`
	Status         string         `json:"status"`
	Assertions     []Assertion    `json:"assertions"`
	Actions        []string       `json:"actions"`
	HumanRequest   *HumanRequest  `json:"human_request"`
	Completion     string         `json:"completion"`
	Concepts       []string       `json:"concepts"`
	IntentReceipt  *IntentReceipt `json:"intent_receipt"`
}

type Usage struct {
	InputTokens           *int64 `json:"input_tokens"`
	CachedInputTokens     *int64 `json:"cached_input_tokens"`
	CacheWriteInputTokens *int64 `json:"cache_write_input_tokens"`
	OutputTokens          *int64 `json:"output_tokens"`
	ReasoningOutputTokens *int64 `json:"reasoning_output_tokens"`
}

type Attempt struct {
	Response  string `json:"response"`
	ExitCode  int    `json:"exit_code"`
	Calls     int    `json:"calls"`
	Retries   int    `json:"retries"`
	ToolCalls int    `json:"tool_calls"`
	LatencyMS int64  `json:"latency_ms"`
	Usage     Usage  `json:"usage"`
}

type ConverterUsage struct {
	Calls     int   `json:"calls"`
	LatencyMS int64 `json:"latency_ms"`
}

type DefectAttribution struct {
	Category string `json:"category"`
	Stage    string `json:"stage"`
	Reason   string `json:"reason"`
}

type Observation struct {
	Version            string              `json:"version"`
	RunID              string              `json:"run_id"`
	CaseID             string              `json:"case_id"`
	ArmID              string              `json:"arm_id"`
	Sender             Attempt             `json:"sender"`
	Handoff            string              `json:"handoff"`
	Receiver           Attempt             `json:"receiver"`
	HumanInterventions int                 `json:"human_interventions"`
	CorrectionTurns    int                 `json:"correction_turns"`
	Converter          ConverterUsage      `json:"converter"`
	Defects            []DefectAttribution `json:"defects"`
}

type PricingEntry struct {
	Provider                  string `json:"provider"`
	Model                     string `json:"model"`
	UncachedInputNanoPerToken int64  `json:"uncached_input_nano_per_token"`
	CachedInputNanoPerToken   int64  `json:"cached_input_nano_per_token"`
	CacheWriteNanoPerToken    int64  `json:"cache_write_nano_per_token"`
	OutputNanoPerToken        int64  `json:"output_nano_per_token"`
	ReasoningIncludedInOutput bool   `json:"reasoning_included_in_output"`
	MaxInputTokensPerCall     *int64 `json:"max_input_tokens_per_call"`
	Source                    string `json:"source"`
	VerifiedAt                string `json:"verified_at"`
}

type PricingPolicy struct {
	Version  string         `json:"version"`
	Currency string         `json:"currency"`
	Entries  []PricingEntry `json:"entries"`
}

type ModelConfig struct {
	Provider        string   `json:"provider"`
	Model           string   `json:"model"`
	ReasoningEffort string   `json:"reasoning_effort"`
	Temperature     *float64 `json:"temperature"`
	MaxOutputTokens *int64   `json:"max_output_tokens"`
}

type CellRef struct {
	CaseID string `json:"case_id"`
	ArmID  string `json:"arm_id"`
}

type IsolationConfig struct {
	Workspace        string   `json:"workspace"`
	RepositoryAccess bool     `json:"repository_access"`
	Tools            []string `json:"tools"`
	PluginsEnabled   bool     `json:"plugins_enabled"`
	MemoryEnabled    bool     `json:"memory_enabled"`
	Network          string   `json:"network"`
}

type RunManifest struct {
	Version             string          `json:"version"`
	Scope               string          `json:"scope"`
	RunID               string          `json:"run_id"`
	FreezeLockSHA256    string          `json:"freeze_lock_sha256"`
	Sender              ModelConfig     `json:"sender"`
	Receiver            ModelConfig     `json:"receiver"`
	Transport           string          `json:"transport"`
	Isolation           IsolationConfig `json:"isolation"`
	EliminationPolicy   string          `json:"elimination_policy"`
	ExecutionOrder      []CellRef       `json:"execution_order"`
	RuntimeRetriesFixed int             `json:"runtime_retries_fixed"`
}
