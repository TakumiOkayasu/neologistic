package eval

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/codec"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/model"
	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/validate"
)

type TrialCost struct {
	InputTokens           *int64 `json:"input_tokens"`
	UncachedInputTokens   *int64 `json:"uncached_input_tokens"`
	CachedInputTokens     *int64 `json:"cached_input_tokens"`
	CacheWriteInputTokens *int64 `json:"cache_write_input_tokens"`
	OutputTokens          *int64 `json:"output_tokens"`
	ReasoningOutputTokens *int64 `json:"reasoning_output_tokens"`
	PaidNanoUSD           *int64 `json:"paid_nano_usd"`
	Calls                 int    `json:"calls"`
	Retries               int    `json:"retries"`
	ToolCalls             int    `json:"tool_calls"`
	LatencyMS             int64  `json:"latency_ms"`
	HumanInterventions    int    `json:"human_interventions"`
	CorrectionTurns       int    `json:"correction_turns"`
	ConverterCalls        int    `json:"converter_calls"`
	ConverterLatencyMS    int64  `json:"converter_latency_ms"`
}

type FailureCounts struct {
	Runtime         int `json:"runtime"`
	CarrierParse    int `json:"carrier_parse"`
	Decision        int `json:"decision"`
	Semantic        int `json:"semantic"`
	Epistemic       int `json:"epistemic"`
	Evidence        int `json:"evidence"`
	Authority       int `json:"authority"`
	Scope           int `json:"scope"`
	Completion      int `json:"completion"`
	Reopen          int `json:"reopen"`
	IntentGate      int `json:"intent_gate"`
	HumanCorrection int `json:"human_correction"`
}

func (f FailureCounts) Total() int {
	return f.Runtime + f.CarrierParse + f.Decision + f.Semantic + f.Epistemic + f.Evidence + f.Authority + f.Scope + f.Completion + f.Reopen + f.IntentGate + f.HumanCorrection
}

type Trial struct {
	CaseID         string                    `json:"case_id"`
	ArmID          string                    `json:"arm_id"`
	HardFailure    bool                      `json:"hard_failure"`
	Failures       FailureCounts             `json:"failures"`
	Errors         []string                  `json:"errors"`
	Cost           TrialCost                 `json:"cost"`
	Defects        []model.DefectAttribution `json:"defects"`
	SelectedOption string                    `json:"selected_option,omitempty"`
	Status         string                    `json:"status,omitempty"`
	Completion     string                    `json:"completion,omitempty"`
}

type Metric struct {
	Available bool   `json:"available"`
	Value     *int64 `json:"value"`
	Reason    string `json:"reason,omitempty"`
}

type ArmCost struct {
	InputTokens           Metric `json:"input_tokens"`
	UncachedInputTokens   Metric `json:"uncached_input_tokens"`
	CachedInputTokens     Metric `json:"cached_input_tokens"`
	CacheWriteInputTokens Metric `json:"cache_write_input_tokens"`
	OutputTokens          Metric `json:"output_tokens"`
	ReasoningOutputTokens Metric `json:"reasoning_output_tokens"`
	PaidNanoUSD           Metric `json:"paid_nano_usd"`
	Calls                 int    `json:"calls"`
	Retries               int    `json:"retries"`
	ToolCalls             int    `json:"tool_calls"`
	LatencyMS             int64  `json:"latency_ms"`
	HumanInterventions    int    `json:"human_interventions"`
	CorrectionTurns       int    `json:"correction_turns"`
	ConverterCalls        int    `json:"converter_calls"`
	ConverterLatencyMS    int64  `json:"converter_latency_ms"`
}

type ArmSummary struct {
	ArmID            string         `json:"arm_id"`
	Carrier          string         `json:"carrier"`
	BFV              bool           `json:"bfv"`
	CustomMechanisms []string       `json:"custom_mechanisms"`
	ExpectedCells    int            `json:"expected_cells"`
	ObservedCells    int            `json:"observed_cells"`
	HardFailures     int            `json:"hard_failures"`
	Failures         FailureCounts  `json:"failures"`
	Cost             ArmCost        `json:"cost"`
	Defects          map[string]int `json:"defects"`
}

type Coverage struct {
	ExpectedCells   int      `json:"expected_cells"`
	ObservedCells   int      `json:"observed_cells"`
	SkippedCells    []string `json:"skipped_after_elimination"`
	Complete        bool     `json:"complete"`
	MissingCells    []string `json:"missing_cells"`
	DuplicateCells  []string `json:"duplicate_cells"`
	UnexpectedCells []string `json:"unexpected_cells"`
}

type DeletionDecision struct {
	Mechanism string `json:"mechanism"`
	Decision  string `json:"decision"`
	Reason    string `json:"reason"`
}

type Selection struct {
	Status             string             `json:"status"`
	Winner             *string            `json:"winner"`
	Reason             string             `json:"reason"`
	EligibleArms       []string           `json:"eligible_arms"`
	BaselineDominators []string           `json:"baseline_dominators"`
	DeletionDecisions  []DeletionDecision `json:"deletion_decisions"`
}

type Report struct {
	Version           string       `json:"version"`
	Scope             string       `json:"scope"`
	RunID             string       `json:"run_id"`
	EliminationPolicy string       `json:"elimination_policy"`
	Coverage          Coverage     `json:"coverage"`
	Trials            []Trial      `json:"trials"`
	Arms              []ArmSummary `json:"arms"`
	Selection         Selection    `json:"selection"`
	HardErrors        []string     `json:"hard_errors,omitempty"`
}

type metricAccumulator struct {
	value     int64
	available bool
	reason    string
}

type armAccumulator struct {
	summary                                                      ArmSummary
	input, uncached, cached, cacheWrite, output, reasoning, paid metricAccumulator
}

func Score(fixtures []model.Fixture, arms []model.Arm, observations []model.Observation, pricing *model.PricingPolicy, manifest model.RunManifest) Report {
	report := Report{Version: model.ReportVersion, Scope: manifest.Scope, RunID: manifest.RunID, EliminationPolicy: manifest.EliminationPolicy}
	if err := validate.ValidateRunManifest(manifest, fixtures, arms); err != nil {
		report.HardErrors = append(report.HardErrors, "run manifest: "+err.Error())
	}
	if pricing != nil {
		if err := validate.ValidatePricingForManifest(*pricing, manifest); err != nil {
			report.HardErrors = append(report.HardErrors, "pricing policy: "+err.Error())
		}
	}
	fixtureByID := map[string]model.Fixture{}
	armByID := map[string]model.Arm{}
	for _, f := range fixtures {
		fixtureByID[f.ID] = f
	}
	for _, arm := range arms {
		armByID[arm.ID] = arm
	}

	expected := map[string]struct{}{}
	for _, f := range fixtures {
		for _, arm := range arms {
			expected[cellKey(f.ID, arm.ID)] = struct{}{}
		}
	}
	seen := map[string]int{}
	valid := make([]model.Observation, 0, len(observations))
	fixtureIDs := keys(fixtureByID)
	armIDs := keys(armByID)
	for _, observation := range observations {
		key := cellKey(observation.CaseID, observation.ArmID)
		if observation.RunID != manifest.RunID {
			report.HardErrors = append(report.HardErrors, key+": observation run_id does not match run manifest")
			continue
		}
		if observation.Sender.Retries != manifest.RuntimeRetriesFixed || observation.Receiver.Retries != manifest.RuntimeRetriesFixed {
			report.HardErrors = append(report.HardErrors, key+": runtime retries do not match frozen run policy")
			continue
		}
		if err := validate.ValidateObservation(observation, fixtureIDs, armIDs); err != nil {
			report.HardErrors = append(report.HardErrors, key+": "+err.Error())
			continue
		}
		seen[key]++
		if _, ok := expected[key]; !ok {
			report.Coverage.UnexpectedCells = append(report.Coverage.UnexpectedCells, key)
			continue
		}
		if seen[key] == 1 {
			valid = append(valid, observation)
		}
	}
	for key := range expected {
		if seen[key] > 1 {
			report.Coverage.DuplicateCells = append(report.Coverage.DuplicateCells, key)
		}
	}
	sort.Strings(report.Coverage.DuplicateCells)
	sort.Strings(report.Coverage.UnexpectedCells)
	report.Coverage.ExpectedCells = len(expected)
	report.Coverage.ObservedCells = len(valid)

	accumulators := map[string]*armAccumulator{}
	for _, arm := range arms {
		accumulators[arm.ID] = &armAccumulator{
			summary: ArmSummary{
				ArmID: arm.ID, Carrier: arm.Carrier, BFV: arm.BFV,
				CustomMechanisms: append([]string(nil), arm.CustomMechanisms...),
				ExpectedCells:    len(fixtures),
				Defects:          map[string]int{"fixture": 0, "carrier": 0, "policy": 0, "model": 0, "runtime": 0, "scorer": 0, "human_contract": 0, "unclassified": 0},
			},
			input: metricAccumulator{available: true}, uncached: metricAccumulator{available: true},
			cached: metricAccumulator{available: true}, cacheWrite: metricAccumulator{available: true},
			output: metricAccumulator{available: true}, reasoning: metricAccumulator{available: true},
			paid: metricAccumulator{available: pricing != nil},
		}
	}

	trialByCell := map[string]Trial{}
	for _, observation := range valid {
		fixture := fixtureByID[observation.CaseID]
		arm := armByID[observation.ArmID]
		trial := scoreObservation(fixture, arm, observation, pricing, manifest.Sender, manifest.Receiver)
		report.Trials = append(report.Trials, trial)
		trialByCell[cellKey(trial.CaseID, trial.ArmID)] = trial
		addTrial(accumulators[arm.ID], trial)
	}
	sort.Slice(report.Trials, func(i, j int) bool {
		if report.Trials[i].CaseID == report.Trials[j].CaseID {
			return report.Trials[i].ArmID < report.Trials[j].ArmID
		}
		return report.Trials[i].CaseID < report.Trials[j].CaseID
	})
	classifyCoverage(&report, expected, seen, trialByCell, manifest, valid)
	for _, arm := range arms {
		acc := accumulators[arm.ID]
		finalizeCost(acc)
		report.Arms = append(report.Arms, acc.summary)
	}
	report.Selection = selectArm(fixtures, arms, report.Arms, trialByCell, report.Coverage)
	return report
}

func classifyCoverage(report *Report, expected map[string]struct{}, seen map[string]int, trials map[string]Trial, manifest model.RunManifest, observations []model.Observation) {
	order := make(map[string]int, len(manifest.ExecutionOrder))
	for index, cell := range manifest.ExecutionOrder {
		order[cellKey(cell.CaseID, cell.ArmID)] = index
	}
	previous := -1
	for _, observation := range observations {
		key := cellKey(observation.CaseID, observation.ArmID)
		index, ok := order[key]
		if !ok {
			continue
		}
		if index <= previous {
			report.HardErrors = append(report.HardErrors, "observations are not in run-manifest execution order at "+key)
		}
		previous = index
	}

	firstFailure := map[string]int{}
	for key, trial := range trials {
		if !trial.HardFailure {
			continue
		}
		index := order[key]
		if current, ok := firstFailure[trial.ArmID]; !ok || index < current {
			firstFailure[trial.ArmID] = index
		}
	}
	for key, trial := range trials {
		if failureIndex, ok := firstFailure[trial.ArmID]; ok && order[key] > failureIndex {
			report.HardErrors = append(report.HardErrors, "arm "+trial.ArmID+" continued after hard-failure elimination at "+key)
		}
	}

	for key := range expected {
		if seen[key] != 0 {
			continue
		}
		armID := strings.SplitN(key, "\x00", 2)[1]
		if failureIndex, ok := firstFailure[armID]; ok && order[key] > failureIndex {
			report.Coverage.SkippedCells = append(report.Coverage.SkippedCells, key)
		} else {
			report.Coverage.MissingCells = append(report.Coverage.MissingCells, key)
		}
	}
	sort.Strings(report.Coverage.SkippedCells)
	sort.Strings(report.Coverage.MissingCells)
	report.Coverage.Complete = len(report.Coverage.MissingCells) == 0 && len(report.Coverage.DuplicateCells) == 0 && len(report.Coverage.UnexpectedCells) == 0 && len(report.HardErrors) == 0
}

func scoreObservation(f model.Fixture, arm model.Arm, o model.Observation, pricing *model.PricingPolicy, senderConfig, receiverConfig model.ModelConfig) Trial {
	trial := Trial{CaseID: f.ID, ArmID: arm.ID, Defects: append([]model.DefectAttribution(nil), o.Defects...)}
	trial.Cost = buildCost(o, pricing, senderConfig, receiverConfig, &trial)
	if o.Sender.ExitCode != 0 || strings.TrimSpace(o.Sender.Response) == "" {
		trial.Failures.Runtime++
		trial.Errors = append(trial.Errors, "sender runtime failed or returned no response")
	}
	if o.Receiver.ExitCode != 0 || strings.TrimSpace(o.Receiver.Response) == "" {
		trial.Failures.Runtime++
		trial.Errors = append(trial.Errors, "receiver runtime failed or returned no response")
	}
	if o.Handoff != o.Sender.Response {
		trial.Failures.Runtime++
		trial.Errors = append(trial.Errors, "handoff differs from exact sender response")
	}
	if err := validate.ValidateCarrier(arm, f, o.Sender.Response); err != nil {
		trial.Failures.CarrierParse++
		trial.Errors = append(trial.Errors, err.Error())
	}
	outcome, err := codec.DecodeStrict[model.Outcome]([]byte(o.Receiver.Response))
	if err != nil {
		trial.Failures.CarrierParse++
		trial.Errors = append(trial.Errors, "receiver outcome parse: "+err.Error())
		trial.HardFailure = true
		return trial
	}
	trial.SelectedOption = outcome.SelectedOption
	trial.Status = outcome.Status
	trial.Completion = outcome.Completion
	if err := validate.ValidateOutcome(f, outcome); err != nil {
		trial.Failures.Semantic++
		trial.Errors = append(trial.Errors, "receiver outcome validation: "+err.Error())
	}
	compareGold(f, outcome, &trial)
	if o.CorrectionTurns > 0 {
		trial.Failures.HumanCorrection += o.CorrectionTurns
		trial.Errors = append(trial.Errors, fmt.Sprintf("avoidable human correction turns: %d", o.CorrectionTurns))
	}
	trial.HardFailure = trial.Failures.Total() > 0
	return trial
}

func compareGold(f model.Fixture, outcome model.Outcome, trial *Trial) {
	gold := f.Gold
	if outcome.SelectedOption != gold.SelectedOption {
		trial.Failures.Decision++
		trial.Errors = append(trial.Errors, fmt.Sprintf("selected option %q, expected %q", outcome.SelectedOption, gold.SelectedOption))
	}
	if outcome.Status != gold.Status {
		trial.Failures.Semantic++
		trial.Errors = append(trial.Errors, fmt.Sprintf("status %q, expected %q", outcome.Status, gold.Status))
	}
	if outcome.Completion != gold.Completion {
		trial.Failures.Completion++
		trial.Errors = append(trial.Errors, fmt.Sprintf("completion %q, expected %q", outcome.Completion, gold.Completion))
	}
	actualAssertions := map[string]model.Assertion{}
	for _, assertion := range outcome.Assertions {
		actualAssertions[assertion.Key] = assertion
	}
	for _, expected := range gold.Assertions {
		actual, ok := actualAssertions[expected.Key]
		if !ok {
			trial.Failures.Semantic++
			trial.Errors = append(trial.Errors, "missing assertion "+expected.Key)
			continue
		}
		if actual.Origin != expected.Origin || actual.State != expected.State {
			trial.Failures.Epistemic++
			trial.Errors = append(trial.Errors, fmt.Sprintf("assertion %s epistemic=%s/%s, expected %s/%s", expected.Key, actual.Origin, actual.State, expected.Origin, expected.State))
		}
		if !sameSet(actual.EvidenceIDs, expected.EvidenceIDs) {
			trial.Failures.Evidence++
			trial.Errors = append(trial.Errors, fmt.Sprintf("assertion %s evidence=%v, expected %v", expected.Key, actual.EvidenceIDs, expected.EvidenceIDs))
		}
	}
	expectedAssertionKeys := stringSetFromAssertions(gold.Assertions)
	for key := range actualAssertions {
		if _, ok := expectedAssertionKeys[key]; !ok {
			trial.Failures.Semantic++
			trial.Errors = append(trial.Errors, "unexpected assertion "+key)
		}
	}
	actions := stringSet(outcome.Actions)
	for _, required := range gold.RequiredActions {
		if _, ok := actions[required]; !ok {
			trial.Failures.Scope++
			trial.Errors = append(trial.Errors, "missing required action "+required)
		}
	}
	for _, forbidden := range gold.ForbiddenActions {
		if _, ok := actions[forbidden]; ok {
			trial.Failures.Scope++
			trial.Errors = append(trial.Errors, "forbidden action selected "+forbidden)
		}
	}
	if gold.HumanRequest.Mode == "required" {
		if outcome.HumanRequest == nil {
			trial.Failures.Authority++
			trial.Errors = append(trial.Errors, "required human request missing")
		} else {
			if outcome.HumanRequest.QuestionID != gold.HumanRequest.QuestionID || outcome.HumanRequest.RecommendationOption != gold.HumanRequest.RecommendationOption {
				trial.Failures.Authority++
				trial.Errors = append(trial.Errors, "human request question or recommendation differs from gold")
			}
		}
	} else if outcome.HumanRequest != nil {
		trial.Failures.Authority++
		trial.Errors = append(trial.Errors, "unnecessary human request")
	}
	concepts := stringSet(outcome.Concepts)
	for _, required := range gold.RequiredConcepts {
		if _, ok := concepts[required]; !ok {
			trial.Failures.Semantic++
			trial.Errors = append(trial.Errors, "missing required concept "+required)
		}
	}
	for _, forbidden := range gold.ForbiddenConcepts {
		if _, ok := concepts[forbidden]; ok {
			trial.Failures.Semantic++
			trial.Errors = append(trial.Errors, "forbidden concretization concept selected "+forbidden)
		}
	}
	if gold.IntentReceipt.Required {
		if outcome.IntentReceipt == nil {
			trial.Failures.IntentGate++
			trial.Errors = append(trial.Errors, "required intent receipt missing")
		} else {
			if outcome.IntentReceipt.Status != gold.IntentReceipt.Status {
				trial.Failures.IntentGate++
				trial.Errors = append(trial.Errors, fmt.Sprintf("intent receipt status %q, expected %q", outcome.IntentReceipt.Status, gold.IntentReceipt.Status))
			}
			if !sameSet(outcome.IntentReceipt.PlannedSideEffects, gold.IntentReceipt.PlannedSideEffects) {
				trial.Failures.IntentGate++
				trial.Errors = append(trial.Errors, "intent receipt planned side effects differ from gold")
			}
			joined := strings.Join(outcome.IntentReceipt.Conflicts, "\n")
			for _, term := range gold.IntentReceipt.RequiredConflictTerms {
				if !strings.Contains(joined, term) {
					trial.Failures.IntentGate++
					trial.Errors = append(trial.Errors, "intent receipt missing conflict term "+term)
				}
			}
		}
	} else if outcome.IntentReceipt != nil {
		trial.Failures.IntentGate++
		trial.Errors = append(trial.Errors, "unexpected intent receipt")
	}
	if f.Class == "reopen_after_refutation" && (outcome.Status != "reopened" || outcome.Completion != "not_ready") {
		trial.Failures.Reopen++
		trial.Errors = append(trial.Errors, "refutation did not reopen work")
	}
}

func buildCost(o model.Observation, pricing *model.PricingPolicy, senderConfig, receiverConfig model.ModelConfig, trial *Trial) TrialCost {
	cost := TrialCost{
		Calls:              o.Sender.Calls + o.Receiver.Calls,
		Retries:            o.Sender.Retries + o.Receiver.Retries,
		ToolCalls:          o.Sender.ToolCalls + o.Receiver.ToolCalls,
		LatencyMS:          o.Sender.LatencyMS + o.Receiver.LatencyMS,
		HumanInterventions: o.HumanInterventions,
		CorrectionTurns:    o.CorrectionTurns,
		ConverterCalls:     o.Converter.Calls,
		ConverterLatencyMS: o.Converter.LatencyMS,
	}
	var err error
	if cost.InputTokens, err = sumOptional(o.Sender.Usage.InputTokens, o.Receiver.Usage.InputTokens); err != nil {
		trial.Failures.Runtime++
		trial.Errors = append(trial.Errors, "input tokens: "+err.Error())
	}
	if cost.CachedInputTokens, err = sumOptional(o.Sender.Usage.CachedInputTokens, o.Receiver.Usage.CachedInputTokens); err != nil {
		trial.Failures.Runtime++
		trial.Errors = append(trial.Errors, "cached input tokens: "+err.Error())
	}
	if cost.CacheWriteInputTokens, err = sumOptional(o.Sender.Usage.CacheWriteInputTokens, o.Receiver.Usage.CacheWriteInputTokens); err != nil {
		trial.Failures.Runtime++
		trial.Errors = append(trial.Errors, "cache-write input tokens: "+err.Error())
	}
	if cost.OutputTokens, err = sumOptional(o.Sender.Usage.OutputTokens, o.Receiver.Usage.OutputTokens); err != nil {
		trial.Failures.Runtime++
		trial.Errors = append(trial.Errors, "output tokens: "+err.Error())
	}
	if cost.ReasoningOutputTokens, err = sumOptional(o.Sender.Usage.ReasoningOutputTokens, o.Receiver.Usage.ReasoningOutputTokens); err != nil {
		trial.Failures.Runtime++
		trial.Errors = append(trial.Errors, "reasoning output tokens: "+err.Error())
	}
	if cost.InputTokens != nil && cost.CachedInputTokens != nil && cost.CacheWriteInputTokens != nil {
		value := *cost.InputTokens - *cost.CachedInputTokens - *cost.CacheWriteInputTokens
		if value < 0 {
			trial.Failures.Runtime++
			trial.Errors = append(trial.Errors, "invalid token decomposition: input < cached + cache-write")
		} else {
			cost.UncachedInputTokens = &value
		}
	}
	if pricing != nil {
		value, err := paidObservationCost(o, *pricing, senderConfig, receiverConfig)
		if err != nil {
			trial.Failures.Runtime++
			trial.Errors = append(trial.Errors, "pricing: "+err.Error())
		} else {
			cost.PaidNanoUSD = &value
		}
	}
	return cost
}

func paidObservationCost(o model.Observation, policy model.PricingPolicy, senderConfig, receiverConfig model.ModelConfig) (int64, error) {
	senderEntry, err := pricingEntry(policy, senderConfig)
	if err != nil {
		return 0, fmt.Errorf("sender: %w", err)
	}
	receiverEntry, err := pricingEntry(policy, receiverConfig)
	if err != nil {
		return 0, fmt.Errorf("receiver: %w", err)
	}
	senderCost, err := paidAttemptCost(o.Sender, senderEntry)
	if err != nil {
		return 0, fmt.Errorf("sender: %w", err)
	}
	receiverCost, err := paidAttemptCost(o.Receiver, receiverEntry)
	if err != nil {
		return 0, fmt.Errorf("receiver: %w", err)
	}
	if senderCost > math.MaxInt64-receiverCost {
		return 0, fmt.Errorf("cost overflow")
	}
	return senderCost + receiverCost, nil
}

func pricingEntry(policy model.PricingPolicy, config model.ModelConfig) (model.PricingEntry, error) {
	for _, entry := range policy.Entries {
		if entry.Provider == config.Provider && entry.Model == config.Model {
			return entry, nil
		}
	}
	return model.PricingEntry{}, fmt.Errorf("no pricing entry for %s/%s", config.Provider, config.Model)
}

func paidAttemptCost(attempt model.Attempt, entry model.PricingEntry) (int64, error) {
	input := attempt.Usage.InputTokens
	cached := attempt.Usage.CachedInputTokens
	cacheWrite := attempt.Usage.CacheWriteInputTokens
	output := attempt.Usage.OutputTokens
	for _, metric := range []*int64{input, cached, cacheWrite, output} {
		if metric == nil {
			return 0, fmt.Errorf("required token counter is unavailable")
		}
	}
	if entry.MaxInputTokensPerCall != nil && *input > *entry.MaxInputTokensPerCall {
		return 0, fmt.Errorf("input tokens %d exceed pricing applicability limit %d", *input, *entry.MaxInputTokensPerCall)
	}
	uncached := *input - *cached - *cacheWrite
	if uncached < 0 {
		return 0, fmt.Errorf("invalid token decomposition: input < cached + cache-write")
	}
	metrics := []int64{uncached, *cached, *cacheWrite, *output}
	rates := []int64{entry.UncachedInputNanoPerToken, entry.CachedInputNanoPerToken, entry.CacheWriteNanoPerToken, entry.OutputNanoPerToken}
	var total int64
	for i, metric := range metrics {
		if metric < 0 || rates[i] < 0 {
			return 0, fmt.Errorf("negative token or rate")
		}
		if rates[i] != 0 && metric > math.MaxInt64/rates[i] {
			return 0, fmt.Errorf("cost overflow")
		}
		part := metric * rates[i]
		if total > math.MaxInt64-part {
			return 0, fmt.Errorf("cost overflow")
		}
		total += part
	}
	return total, nil
}

func sumOptional(a, b *int64) (*int64, error) {
	if a == nil || b == nil {
		return nil, nil
	}
	if *a > math.MaxInt64-*b {
		return nil, fmt.Errorf("counter overflow")
	}
	value := *a + *b
	return &value, nil
}

func addTrial(acc *armAccumulator, trial Trial) {
	acc.summary.ObservedCells++
	if trial.HardFailure {
		acc.summary.HardFailures++
	}
	addFailures(&acc.summary.Failures, trial.Failures)
	for _, defect := range trial.Defects {
		acc.summary.Defects[defect.Category]++
	}
	if trial.HardFailure && len(trial.Defects) == 0 {
		acc.summary.Defects["unclassified"]++
	}
	addMetric(&acc.input, trial.Cost.InputTokens, "one or more input token counters are missing")
	addMetric(&acc.uncached, trial.Cost.UncachedInputTokens, "one or more uncached input counters are unavailable")
	addMetric(&acc.cached, trial.Cost.CachedInputTokens, "one or more cached input counters are missing")
	addMetric(&acc.cacheWrite, trial.Cost.CacheWriteInputTokens, "one or more cache-write counters are missing")
	addMetric(&acc.output, trial.Cost.OutputTokens, "one or more output token counters are missing")
	addMetric(&acc.reasoning, trial.Cost.ReasoningOutputTokens, "one or more reasoning token counters are missing")
	addMetric(&acc.paid, trial.Cost.PaidNanoUSD, "pricing or one or more billed token counters are unavailable")
	acc.summary.Cost.Calls += trial.Cost.Calls
	acc.summary.Cost.Retries += trial.Cost.Retries
	acc.summary.Cost.ToolCalls += trial.Cost.ToolCalls
	acc.summary.Cost.LatencyMS += trial.Cost.LatencyMS
	acc.summary.Cost.HumanInterventions += trial.Cost.HumanInterventions
	acc.summary.Cost.CorrectionTurns += trial.Cost.CorrectionTurns
	acc.summary.Cost.ConverterCalls += trial.Cost.ConverterCalls
	acc.summary.Cost.ConverterLatencyMS += trial.Cost.ConverterLatencyMS
}

func addFailures(dst *FailureCounts, src FailureCounts) {
	dst.Runtime += src.Runtime
	dst.CarrierParse += src.CarrierParse
	dst.Decision += src.Decision
	dst.Semantic += src.Semantic
	dst.Epistemic += src.Epistemic
	dst.Evidence += src.Evidence
	dst.Authority += src.Authority
	dst.Scope += src.Scope
	dst.Completion += src.Completion
	dst.Reopen += src.Reopen
	dst.IntentGate += src.IntentGate
	dst.HumanCorrection += src.HumanCorrection
}

func addMetric(acc *metricAccumulator, value *int64, reason string) {
	if value == nil {
		acc.available = false
		acc.reason = reason
		return
	}
	if acc.available {
		acc.value += *value
	}
}

func finalizeCost(acc *armAccumulator) {
	acc.summary.Cost.InputTokens = metric(acc.input)
	acc.summary.Cost.UncachedInputTokens = metric(acc.uncached)
	acc.summary.Cost.CachedInputTokens = metric(acc.cached)
	acc.summary.Cost.CacheWriteInputTokens = metric(acc.cacheWrite)
	acc.summary.Cost.OutputTokens = metric(acc.output)
	acc.summary.Cost.ReasoningOutputTokens = metric(acc.reasoning)
	acc.summary.Cost.PaidNanoUSD = metric(acc.paid)
}

func metric(acc metricAccumulator) Metric {
	if !acc.available {
		return Metric{Available: false, Reason: acc.reason}
	}
	value := acc.value
	return Metric{Available: true, Value: &value}
}

func selectArm(fixtures []model.Fixture, arms []model.Arm, summaries []ArmSummary, trials map[string]Trial, coverage Coverage) Selection {
	selection := Selection{Status: "undetermined"}
	if !coverage.Complete {
		selection.Reason = "coverage incomplete; no operational decision"
		selection.DeletionDecisions = holdAll("coverage incomplete")
		return selection
	}
	summaryByID := map[string]ArmSummary{}
	armByID := map[string]model.Arm{}
	for _, arm := range arms {
		armByID[arm.ID] = arm
	}
	for _, summary := range summaries {
		summaryByID[summary.ArmID] = summary
		if summary.HardFailures == 0 && summary.ObservedCells == summary.ExpectedCells {
			selection.EligibleArms = append(selection.EligibleArms, summary.ArmID)
		}
	}
	sort.Strings(selection.EligibleArms)
	if len(selection.EligibleArms) == 0 {
		selection.Reason = "all arms failed hard correctness gates"
		selection.DeletionDecisions = holdAll("no eligible arm")
		return selection
	}
	baseline, baselineEligible := summaryByID["A"]
	if baselineEligible && baseline.HardFailures == 0 {
		for _, armID := range selection.EligibleArms {
			if armID == "A" {
				continue
			}
			if dominatesBaseline(fixtures, trials, armID) {
				selection.BaselineDominators = append(selection.BaselineDominators, armID)
			}
		}
		if len(selection.BaselineDominators) == 0 {
			winner := "A"
			selection.Status = "selected"
			selection.Winner = &winner
			selection.Reason = "no custom arm dominated native communication on every matched task; native wins by deletion default"
			selection.DeletionDecisions = decisionsForWinner("A")
			return selection
		}
		winner := bestAmong(selection.BaselineDominators, summaryByID, armByID)
		selection.Status = "selected"
		selection.Winner = &winner
		selection.Reason = "custom arm dominated native communication on every matched task"
		selection.DeletionDecisions = decisionsForWinner(winner)
		return selection
	}
	winner := bestAmong(selection.EligibleArms, summaryByID, armByID)
	selection.Status = "selected"
	selection.Winner = &winner
	selection.Reason = "native communication failed a hard gate; selected the lowest-cost eligible arm with the fewest custom mechanisms"
	selection.DeletionDecisions = decisionsForWinner(winner)
	return selection
}

func dominatesBaseline(fixtures []model.Fixture, trials map[string]Trial, candidate string) bool {
	strict := false
	for _, fixture := range fixtures {
		base, okBase := trials[cellKey(fixture.ID, "A")]
		cand, okCandidate := trials[cellKey(fixture.ID, candidate)]
		if !okBase || !okCandidate || base.HardFailure || cand.HardFailure {
			return false
		}
		cmp, comparable := compareCost(cand.Cost, base.Cost)
		if !comparable || cmp > 0 {
			return false
		}
		if cmp < 0 {
			strict = true
		}
	}
	return strict
}

// compareCost returns -1 when a is strictly better, 0 when equal, and 1 when a is worse or crosses b.
// Paid cost is used when both are complete. Otherwise mutually exclusive token counters and operational counters must not worsen.
func compareCost(a, b TrialCost) (int, bool) {
	if a.PaidNanoUSD != nil && b.PaidNanoUSD != nil {
		if operationalWorse(a, b) {
			return 1, true
		}
		switch {
		case *a.PaidNanoUSD < *b.PaidNanoUSD:
			return -1, true
		case *a.PaidNanoUSD == *b.PaidNanoUSD:
			if operationalStrictlyBetter(a, b) {
				return -1, true
			}
			return 0, true
		default:
			return 1, true
		}
	}
	metricsA := []*int64{a.UncachedInputTokens, a.CachedInputTokens, a.CacheWriteInputTokens, a.OutputTokens}
	metricsB := []*int64{b.UncachedInputTokens, b.CachedInputTokens, b.CacheWriteInputTokens, b.OutputTokens}
	for i := range metricsA {
		if metricsA[i] == nil || metricsB[i] == nil {
			return 0, false
		}
	}
	worse := operationalWorse(a, b)
	strict := operationalStrictlyBetter(a, b)
	for i := range metricsA {
		if *metricsA[i] > *metricsB[i] {
			worse = true
		}
		if *metricsA[i] < *metricsB[i] {
			strict = true
		}
	}
	if worse {
		return 1, true
	}
	if strict {
		return -1, true
	}
	return 0, true
}

func operationalWorse(a, b TrialCost) bool {
	return a.Calls > b.Calls || a.Retries > b.Retries || a.ToolCalls > b.ToolCalls || a.HumanInterventions > b.HumanInterventions || a.CorrectionTurns > b.CorrectionTurns || a.ConverterCalls > b.ConverterCalls
}

func operationalStrictlyBetter(a, b TrialCost) bool {
	return a.Calls < b.Calls || a.Retries < b.Retries || a.ToolCalls < b.ToolCalls || a.HumanInterventions < b.HumanInterventions || a.CorrectionTurns < b.CorrectionTurns || a.ConverterCalls < b.ConverterCalls
}

func bestAmong(ids []string, summaries map[string]ArmSummary, arms map[string]model.Arm) string {
	copyIDs := append([]string(nil), ids...)
	sort.Slice(copyIDs, func(i, j int) bool {
		a, b := summaries[copyIDs[i]], summaries[copyIDs[j]]
		if a.Cost.PaidNanoUSD.Available && b.Cost.PaidNanoUSD.Available && *a.Cost.PaidNanoUSD.Value != *b.Cost.PaidNanoUSD.Value {
			return *a.Cost.PaidNanoUSD.Value < *b.Cost.PaidNanoUSD.Value
		}
		if len(arms[a.ArmID].CustomMechanisms) != len(arms[b.ArmID].CustomMechanisms) {
			return len(arms[a.ArmID].CustomMechanisms) < len(arms[b.ArmID].CustomMechanisms)
		}
		if aTokens, aOK := totalObservedTokens(a.Cost); aOK {
			if bTokens, bOK := totalObservedTokens(b.Cost); bOK && aTokens != bTokens {
				return aTokens < bTokens
			}
		}
		if a.Cost.Calls != b.Cost.Calls {
			return a.Cost.Calls < b.Cost.Calls
		}
		if a.Cost.Retries != b.Cost.Retries {
			return a.Cost.Retries < b.Cost.Retries
		}
		if a.Cost.HumanInterventions != b.Cost.HumanInterventions {
			return a.Cost.HumanInterventions < b.Cost.HumanInterventions
		}
		if a.Cost.CorrectionTurns != b.Cost.CorrectionTurns {
			return a.Cost.CorrectionTurns < b.Cost.CorrectionTurns
		}
		return a.ArmID < b.ArmID
	})
	return copyIDs[0]
}

func totalObservedTokens(cost ArmCost) (int64, bool) {
	if !cost.InputTokens.Available || !cost.OutputTokens.Available || cost.InputTokens.Value == nil || cost.OutputTokens.Value == nil {
		return 0, false
	}
	if *cost.InputTokens.Value > math.MaxInt64-*cost.OutputTokens.Value {
		return 0, false
	}
	return *cost.InputTokens.Value + *cost.OutputTokens.Value, true
}

func decisionsForWinner(winner string) []DeletionDecision {
	retain := map[string]bool{}
	switch winner {
	case "B":
		retain["bfv"] = true
	case "C":
		retain["genshijin-normal"] = true
	case "D":
		retain["genshijin-normal"] = true
		retain["bfv"] = true
	case "E":
		retain["json-carrier-v1"] = true
	case "F":
		retain["json-carrier-v1"] = true
		retain["bfv"] = true
	}
	var out []DeletionDecision
	for _, mechanism := range []string{"genshijin-normal", "json-carrier-v1", "bfv"} {
		if retain[mechanism] {
			out = append(out, DeletionDecision{Mechanism: mechanism, Decision: "advance_to_stateful_validation", Reason: "required by semantic-transfer winner " + winner + "; not yet approved for operational default"})
		} else {
			out = append(out, DeletionDecision{Mechanism: mechanism, Decision: "delete_from_default", Reason: "not required by selected arm " + winner})
		}
	}
	return out
}

func holdAll(reason string) []DeletionDecision {
	return []DeletionDecision{
		{Mechanism: "genshijin-normal", Decision: "hold", Reason: reason},
		{Mechanism: "json-carrier-v1", Decision: "hold", Reason: reason},
		{Mechanism: "bfv", Decision: "hold", Reason: reason},
	}
}

func cellKey(caseID, armID string) string { return caseID + "\x00" + armID }

func keys[T any](values map[string]T) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for key := range values {
		out[key] = struct{}{}
	}
	return out
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	left := append([]string(nil), a...)
	right := append([]string(nil), b...)
	sort.Strings(left)
	sort.Strings(right)
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func stringSetFromAssertions(values []model.Assertion) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		out[value.Key] = struct{}{}
	}
	return out
}

func stringSet(values []string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		out[value] = struct{}{}
	}
	return out
}

// ScoreCell deterministically scores one completed sender/receiver cell so a
// runner can enforce early elimination before spending on later cells.
func ScoreCell(fixture model.Fixture, arm model.Arm, observation model.Observation, pricing *model.PricingPolicy, manifest model.RunManifest) (Trial, error) {
	if observation.RunID != manifest.RunID {
		return Trial{}, fmt.Errorf("observation run_id does not match run manifest")
	}
	if observation.CaseID != fixture.ID || observation.ArmID != arm.ID {
		return Trial{}, fmt.Errorf("observation cell does not match selected fixture/arm")
	}
	if observation.Sender.Retries != manifest.RuntimeRetriesFixed || observation.Receiver.Retries != manifest.RuntimeRetriesFixed {
		return Trial{}, fmt.Errorf("runtime retries do not match frozen run policy")
	}
	if err := validate.ValidateObservation(observation, map[string]struct{}{fixture.ID: {}}, map[string]struct{}{arm.ID: {}}); err != nil {
		return Trial{}, err
	}
	if pricing != nil {
		if err := validate.ValidatePricingForManifest(*pricing, manifest); err != nil {
			return Trial{}, err
		}
	}
	return scoreObservation(fixture, arm, observation, pricing, manifest.Sender, manifest.Receiver), nil
}
