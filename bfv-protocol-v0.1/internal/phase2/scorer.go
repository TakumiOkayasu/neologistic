package phase2

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

type TrialScore struct {
	CaseID                       string              `json:"case_id"`
	Cohort                       string              `json:"cohort"`
	Candidate                    string              `json:"candidate"`
	ProducerSerializationStatus  string              `json:"producer_serialization_status"`
	ProducerCanonicalExact       *bool               `json:"producer_canonical_exact"`
	ProducerDerivationStatus     string              `json:"producer_derivation_status"`
	ConsumptionSemanticStatus    string              `json:"consumption_semantic_status"`
	ConsumerParseOK              bool                `json:"consumer_parse_ok"`
	TransferSemanticOK           bool                `json:"transfer_semantic_ok"`
	DownstreamTaskOK             bool                `json:"downstream_task_ok"`
	Q1FalsePositives             int                 `json:"q1_false_positives"`
	Q1FalseNegatives             int                 `json:"q1_false_negatives"`
	EpistemicFailures            int                 `json:"epistemic_failures"`
	EpistemicMissing             int                 `json:"epistemic_missing"`
	EpistemicPromotions          int                 `json:"epistemic_promotions"`
	EvidenceFailures             int                 `json:"evidence_failures"`
	EvidenceMissing              int                 `json:"evidence_missing"`
	EvidenceMisassociated        int                 `json:"evidence_misassociated"`
	VerifiedWithoutEvidence      int                 `json:"verified_without_evidence"`
	BoundsFailures               int                 `json:"bounds_failures"`
	D1TransmissionFailures       int                 `json:"d1_transmission_failures"`
	D1AuthorityFailures          int                 `json:"d1_authority_failures"`
	D1Invented                   int                 `json:"d1_invented"`
	PunctuationOnlyVariations    int                 `json:"punctuation_only_variations"`
	CapitalizationOnlyVariations int                 `json:"capitalization_only_variations"`
	TaskSupportOK                bool                `json:"task_support_ok"`
	RuntimeHardFailureAttempts   int                 `json:"runtime_hard_failure_attempts"`
	CLIInvocations               int                 `json:"cli_invocations"`
	RuntimeRetries               int                 `json:"runtime_retries"`
	HardFailure                  bool                `json:"hard_failure"`
	Errors                       []string            `json:"errors,omitempty"`
	Defects                      []DefectAttribution `json:"defects"`
	cost                         usageAccumulator
}

type Metric struct {
	Available       bool   `json:"available"`
	Value           *int64 `json:"value"`
	ObservedTotal   int64  `json:"observed_total"`
	ReportedCalls   int    `json:"reported_calls"`
	UnreportedCalls int    `json:"unreported_calls"`
	Reason          string `json:"reason,omitempty"`
}

type CostVector struct {
	CLIInvocations           int    `json:"cli_invocations"`
	RuntimeRetries           int    `json:"runtime_retries"`
	InputTokens              Metric `json:"input_tokens"`
	MaxInputTokensPerCall    Metric `json:"max_input_tokens_per_call"`
	UncachedInputTokens      Metric `json:"uncached_input_tokens"`
	CachedInputTokens        Metric `json:"cached_input_tokens"`
	CacheWriteInputTokens    Metric `json:"cache_write_input_tokens"`
	OutputTokens             Metric `json:"output_tokens"`
	ReasoningOutputTokens    Metric `json:"reasoning_output_tokens"`
	InstructionTokens        Metric `json:"instruction_tokens"`
	TaskInputTokens          Metric `json:"task_input_tokens"`
	ProviderInternalRequests Metric `json:"provider_internal_requests"`
	MonetaryCost             Metric `json:"monetary_cost"`
}

type CandidateSummary struct {
	Candidate                    string         `json:"candidate"`
	ExpectedCells                int            `json:"expected_cells"`
	ObservedCells                int            `json:"observed_cells"`
	CopyCells                    int            `json:"copy_cells"`
	SourceDerivedCells           int            `json:"derivation_cells"`
	IdentityProducerCells        int            `json:"identity_producer_cells"`
	SerializationEligibleCells   int            `json:"serialization_eligible_cells"`
	SerializationFailures        int            `json:"serialization_failures"`
	ProducerDerivationEligible   int            `json:"producer_derivation_eligible_cells"`
	ProducerDerivationSuccesses  int            `json:"producer_derivation_successes"`
	ConsumptionEligibleCells     int            `json:"consumption_eligible_cells"`
	ConsumptionSemanticSuccesses int            `json:"consumption_semantic_successes"`
	ConsumptionSemanticFailures  int            `json:"consumption_semantic_failures"`
	ConsumerParseFailures        int            `json:"consumer_parse_failures"`
	TransferSemanticSuccesses    int            `json:"transfer_semantic_successes"`
	TransferSemanticFailures     int            `json:"transfer_semantic_failures"`
	DownstreamTaskSuccesses      int            `json:"downstream_task_successes"`
	DownstreamTaskFailures       int            `json:"downstream_task_failures"`
	Q1FalsePositives             int            `json:"q1_false_positives"`
	Q1FalseNegatives             int            `json:"q1_false_negatives"`
	RoutingFailures              int            `json:"routing_failures"`
	EpistemicFailures            int            `json:"epistemic_failures"`
	EpistemicMissing             int            `json:"epistemic_missing"`
	EpistemicPromotions          int            `json:"epistemic_promotions"`
	EvidenceFailures             int            `json:"evidence_failures"`
	EvidenceMissing              int            `json:"evidence_missing"`
	EvidenceMisassociated        int            `json:"evidence_misassociated"`
	VerifiedWithoutEvidence      int            `json:"verified_without_evidence"`
	BoundsFailures               int            `json:"bounds_failures"`
	D1TransmissionFailures       int            `json:"d1_transmission_failures"`
	D1AuthorityFailures          int            `json:"d1_authority_failures"`
	D1Invented                   int            `json:"d1_invented"`
	PunctuationOnlyVariations    int            `json:"punctuation_only_variations"`
	CapitalizationOnlyVariations int            `json:"capitalization_only_variations"`
	RuntimeHardFailureAttempts   int            `json:"runtime_hard_failure_attempts"`
	HardFailures                 int            `json:"hard_failures"`
	Cost                         CostVector     `json:"cost"`
	Defects                      map[string]int `json:"defects"`
	cost                         usageAccumulator
}

type CoverageReport struct {
	ExpectedCells  int      `json:"expected_cells"`
	ObservedCells  int      `json:"observed_cells"`
	Complete       bool     `json:"complete"`
	HardError      bool     `json:"hard_error"`
	MissingCells   []string `json:"missing_cells"`
	DuplicateCells []string `json:"duplicate_cells"`
}

type NextDiscriminatingTest struct {
	CaseID     string   `json:"case_id"`
	Candidates []string `json:"candidates"`
	Trigger    string   `json:"trigger"`
	Purpose    string   `json:"purpose"`
}

// Selection is the observation-only Pareto-frontier diagnostic. Report.Selection
// and selection.json retain this historical contract.
type Selection struct {
	Status                 string                  `json:"status"`
	Winner                 *string                 `json:"winner"`
	Rule                   string                  `json:"rule"`
	NextDiscriminatingTest *NextDiscriminatingTest `json:"next_discriminating_test"`
}

type OperationalPricing struct {
	PolicyVersion                     string `json:"policy_version"`
	Model                             string `json:"model"`
	VerifiedOn                        string `json:"verified_on"`
	Source                            string `json:"source"`
	UncachedInputNanoUSDPerToken      int64  `json:"uncached_input_nano_usd_per_token"`
	CachedInputNanoUSDPerToken        int64  `json:"cached_input_nano_usd_per_token"`
	CacheWriteInputNanoUSDPerToken    int64  `json:"cache_write_input_nano_usd_per_token"`
	OutputNanoUSDPerToken             int64  `json:"output_nano_usd_per_token"`
	StandardRateMaxInputTokensPerCall int64  `json:"standard_rate_max_input_tokens_per_call"`
	ReasoningOutputCostTreatment      string `json:"reasoning_output_cost_treatment"`
}

type OperationalEvaluation struct {
	Candidate                     string   `json:"candidate"`
	CoverageComplete              bool     `json:"coverage_complete"`
	HardFailures                  int      `json:"hard_failures"`
	ConsumptionSemanticSuccesses  int      `json:"consumption_semantic_successes"`
	ConsumptionSemanticFailures   int      `json:"consumption_semantic_failures"`
	TransferSemanticSuccesses     int      `json:"transfer_semantic_successes"`
	TransferSemanticFailures      int      `json:"transfer_semantic_failures"`
	DownstreamTaskSuccesses       int      `json:"downstream_task_successes"`
	DownstreamTaskFailures        int      `json:"downstream_task_failures"`
	EstimatedPaidTokenCostNanoUSD *int64   `json:"estimated_paid_token_cost_nano_usd"`
	EstimatedPaidTokenCostUSD     *float64 `json:"estimated_paid_token_cost_usd"`
	InputOutputTokens             *int64   `json:"input_output_tokens"`
	Eligible                      bool     `json:"eligible"`
	Disposition                   string   `json:"disposition"`
	Reason                        string   `json:"reason,omitempty"`
}

type OperationalSelection struct {
	Status      string                  `json:"status"`
	Winner      *string                 `json:"winner"`
	Challenger  *string                 `json:"challenger"`
	Rejected    []string                `json:"rejected"`
	Rule        string                  `json:"rule"`
	Reason      string                  `json:"reason,omitempty"`
	Pricing     *OperationalPricing     `json:"pricing,omitempty"`
	Evaluations []OperationalEvaluation `json:"evaluations,omitempty"`
}

type CohortSummary struct {
	Candidate    string `json:"candidate"`
	Cohort       string `json:"cohort"`
	Observations int    `json:"observations"`
	HardFailures int    `json:"hard_failures"`
	TaskFailures int    `json:"task_failures"`
}

type Report struct {
	Coverage             CoverageReport       `json:"coverage"`
	Trials               []TrialScore         `json:"trials"`
	Candidates           []CandidateSummary   `json:"candidates"`
	Cohorts              []CohortSummary      `json:"cohorts"`
	Selection            Selection            `json:"selection"`
	ParetoSelection      Selection            `json:"pareto_selection"`
	OperationalSelection OperationalSelection `json:"operational_selection"`
	HardErrors           []string             `json:"hard_errors,omitempty"`
}

type recordChecks struct {
	Semantic                bool
	Q1FP, Q1FN              int
	Epistemic               int
	Promotions              int
	EpistemicMissing        int
	Evidence                int
	EvidenceMissing         int
	EvidenceMisassociated   int
	VerifiedWithoutEvidence int
	Bounds                  int
	D1Transmission          int
	D1Authority             int
	D1Invented              int
	PunctuationOnly         int
	CapitalizationOnly      int
}

type usageAccumulator struct {
	Calls, Retries int
	values         [5]int64
	maxInputTokens int64
	reported       [5]int
	unreported     [5]int
	invalidUsage   bool
	invalidReason  string
}

func Score(fixtures []Fixture, observations []Observation) Report {
	return ScoreWithOperationalPolicy(fixtures, observations, nil)
}

func ScoreWithOperationalPolicy(fixtures []Fixture, observations []Observation, policy *OperationalPricing) Report {
	report := Report{}
	fixtureByID := map[string]Fixture{}
	for _, fixture := range fixtures {
		fixtureByID[fixture.ID] = fixture
	}
	expected := map[string]bool{}
	for _, fixture := range fixtures {
		for _, candidate := range Candidates {
			expected[cellKey(fixture.ID, candidate)] = true
		}
	}
	seen := map[string]int{}
	validObservations := make([]Observation, 0, len(observations))
	for _, observation := range observations {
		key := cellKey(observation.CaseID, observation.Candidate)
		if observation.CoveragePass != 1 {
			report.HardErrors = append(report.HardErrors, "coverage_pass must be 1 for "+key)
			continue
		}
		if err := validateDefects(observation.Defects); err != nil {
			report.HardErrors = append(report.HardErrors, key+": "+err.Error())
			continue
		}
		seen[key]++
		if !expected[key] {
			report.HardErrors = append(report.HardErrors, "unexpected coverage cell "+key)
			continue
		}
		if seen[key] == 1 {
			validObservations = append(validObservations, observation)
		}
	}
	for key := range expected {
		if seen[key] == 0 {
			report.Coverage.MissingCells = append(report.Coverage.MissingCells, key)
		}
		if seen[key] > 1 {
			report.Coverage.DuplicateCells = append(report.Coverage.DuplicateCells, key)
		}
	}
	sort.Strings(report.Coverage.MissingCells)
	sort.Strings(report.Coverage.DuplicateCells)
	report.Coverage.ExpectedCells = len(expected)
	report.Coverage.ObservedCells = len(expected) - len(report.Coverage.MissingCells)
	report.Coverage.Complete = len(report.Coverage.MissingCells) == 0 && len(report.Coverage.DuplicateCells) == 0 && len(report.HardErrors) == 0
	report.Coverage.HardError = !report.Coverage.Complete
	if report.Coverage.HardError {
		report.HardErrors = append(report.HardErrors, "coverage matrix is incomplete or duplicated")
	}

	summaryByCandidate := map[string]*CandidateSummary{}
	cohortByKey := map[string]*CohortSummary{}
	for _, candidate := range Candidates {
		summaryByCandidate[candidate] = &CandidateSummary{
			Candidate: candidate, ExpectedCells: len(fixtures),
			Defects: map[string]int{"fixture": 0, "protocol": 0, "model": 0, "scorer": 0, "runtime": 0, "unclassified": 0},
		}
		for _, cohort := range []string{"copy", "source-derived"} {
			cohortByKey[candidate+"\x00"+cohort] = &CohortSummary{Candidate: candidate, Cohort: cohort}
		}
	}
	for _, observation := range validObservations {
		fixture := fixtureByID[observation.CaseID]
		trial := scoreObservation(fixture, observation)
		report.Trials = append(report.Trials, trial)
		addTrial(summaryByCandidate[observation.Candidate], trial)
		cohort := cohortByKey[observation.Candidate+"\x00"+fixture.Cohort]
		cohort.Observations++
		if trial.HardFailure {
			cohort.HardFailures++
		}
		if !trial.DownstreamTaskOK {
			cohort.TaskFailures++
		}
	}
	sort.Slice(report.Trials, func(i, j int) bool {
		if report.Trials[i].CaseID == report.Trials[j].CaseID {
			return report.Trials[i].Candidate < report.Trials[j].Candidate
		}
		return report.Trials[i].CaseID < report.Trials[j].CaseID
	})
	for _, candidate := range Candidates {
		summary := summaryByCandidate[candidate]
		summary.Cost = summary.cost.report()
		report.Candidates = append(report.Candidates, *summary)
		for _, cohort := range []string{"copy", "source-derived"} {
			report.Cohorts = append(report.Cohorts, *cohortByKey[candidate+"\x00"+cohort])
		}
	}
	report.Selection = selectWinner(report.Candidates, report.Coverage, report.Trials)
	report.ParetoSelection = report.Selection
	report.OperationalSelection = selectOperationalWinner(report.Candidates, report.Coverage, policy)
	return report
}

func scoreObservation(fixture Fixture, observation Observation) TrialScore {
	trial := TrialScore{
		CaseID: fixture.ID, Cohort: fixture.Cohort, Candidate: observation.Candidate,
		ProducerSerializationStatus: "not-run", ProducerDerivationStatus: "not-run", ConsumptionSemanticStatus: "not-run",
		Defects: append([]DefectAttribution(nil), observation.Defects...),
	}
	var producer *Attempt
	var producerOK bool
	if observation.Candidate != CandidateDirect {
		producer, producerOK = inspectStage(observation.Producer, &trial)
	}
	consumer, consumerOK := inspectStage(observation.Consumer, &trial)
	var producerRecords []Record
	producerParsed := false

	if observation.Candidate == CandidateDirect {
		trial.ProducerSerializationStatus = "identity-not-applicable"
		trial.ProducerDerivationStatus = "identity-not-applicable"
		if observation.Producer.Mode != "not_applicable" || len(observation.Producer.Attempts) != 0 {
			trial.Errors = append(trial.Errors, "direct producer must be explicitly not_applicable with zero attempts")
		}
		if observation.Handoff != fixture.Source {
			trial.Errors = append(trial.Errors, "direct handoff did not preserve raw source bytes")
		}
	} else if producerOK && producer != nil {
		if observation.Handoff != producer.Response {
			trial.Errors = append(trial.Errors, "handoff differs from verbatim producer response")
		}
		actual, canonical, err := ParseCandidate(observation.Candidate, producer.Response)
		if err != nil {
			trial.ProducerSerializationStatus = "failed"
			if fixture.Cohort == "copy" {
				trial.ProducerDerivationStatus = "not-applicable-copy"
			} else {
				trial.ProducerDerivationStatus = "failed"
			}
			trial.Errors = append(trial.Errors, "producer serialization: "+err.Error())
		} else {
			trial.ProducerSerializationStatus = "passed"
			producerRecords = actual
			producerParsed = true
			exact := producer.Response == canonical
			trial.ProducerCanonicalExact = &exact
			if fixture.Cohort == "copy" {
				trial.ProducerDerivationStatus = "not-applicable-copy"
			} else {
				checks := compareRecords(fixture.ExpectedRecords, actual)
				if checks.Semantic {
					trial.ProducerDerivationStatus = "passed"
				} else {
					trial.ProducerDerivationStatus = "failed"
				}
				if !checks.Semantic {
					trial.Errors = append(trial.Errors, "producer semantic derivation differs from fixture")
				}
			}
		}
	} else {
		trial.ProducerSerializationStatus = "runtime-failed"
		if fixture.Cohort == "copy" {
			trial.ProducerDerivationStatus = "not-applicable-copy"
		} else {
			trial.ProducerDerivationStatus = "runtime-failed"
		}
	}

	if consumerOK && consumer != nil {
		output, err := ParseConsumer(consumer.Response)
		if err != nil {
			trial.Errors = append(trial.Errors, "consumer parse: "+err.Error())
		} else {
			trial.ConsumerParseOK = true
			if observation.Candidate != CandidateDirect {
				if producerParsed && compareRecords(producerRecords, output.Records).Semantic {
					trial.ConsumptionSemanticStatus = "passed"
				} else {
					trial.ConsumptionSemanticStatus = "failed"
					trial.Errors = append(trial.Errors, "consumer receipt differs from producer handoff")
				}
			}
			checks := compareRecords(fixture.ExpectedRecords, output.Records)
			trial.TransferSemanticOK = checks.Semantic
			if observation.Candidate == CandidateDirect {
				if checks.Semantic {
					trial.ConsumptionSemanticStatus = "passed"
				} else {
					trial.ConsumptionSemanticStatus = "failed"
				}
			}
			trial.Q1FalsePositives = checks.Q1FP
			trial.Q1FalseNegatives = checks.Q1FN
			trial.EpistemicFailures = checks.Epistemic
			trial.EpistemicMissing = checks.EpistemicMissing
			trial.EpistemicPromotions = checks.Promotions
			trial.EvidenceFailures = checks.Evidence
			trial.EvidenceMissing = checks.EvidenceMissing
			trial.EvidenceMisassociated = checks.EvidenceMisassociated
			trial.VerifiedWithoutEvidence = checks.VerifiedWithoutEvidence
			trial.BoundsFailures = checks.Bounds
			trial.D1TransmissionFailures = checks.D1Transmission
			trial.D1AuthorityFailures = checks.D1Authority
			trial.D1Invented = checks.D1Invented
			if violations := authorityViolations(output.Records, output.Task, fixture.AuthoritativeReferences); violations > trial.D1AuthorityFailures {
				trial.D1AuthorityFailures = violations
			}
			trial.PunctuationOnlyVariations = checks.PunctuationOnly
			trial.CapitalizationOnlyVariations = checks.CapitalizationOnly
			trial.DownstreamTaskOK = semanticJSONEqual(fixture.TaskOracle, output.Task, "")
			trial.TaskSupportOK = taskSupportMatches(output.Task, output.Records)
			if !checks.Semantic {
				trial.Errors = append(trial.Errors, "consumer semantic receipt differs from fixture")
			}
			if !trial.DownstreamTaskOK {
				trial.Errors = append(trial.Errors, "downstream task oracle mismatch")
			}
			if !trial.TaskSupportOK {
				trial.Errors = append(trial.Errors, "downstream task support is missing or misassociated")
			}
		}
	} else {
		trial.Errors = append(trial.Errors, "consumer runtime did not produce a scoreable response")
	}

	producerHard := observation.Candidate != CandidateDirect && trial.ProducerSerializationStatus != "passed"
	if observation.Candidate != CandidateDirect && fixture.Cohort == "source-derived" && trial.ProducerDerivationStatus != "passed" {
		producerHard = true
	}
	consumptionHard := trial.ConsumptionSemanticStatus != "passed"
	identityHard := observation.Candidate == CandidateDirect && (observation.Producer.Mode != "not_applicable" || len(observation.Producer.Attempts) != 0 || observation.Handoff != fixture.Source)
	handoffHard := observation.Candidate != CandidateDirect && producer != nil && observation.Handoff != producer.Response
	trial.HardFailure = trial.RuntimeHardFailureAttempts > 0 || producerHard || consumptionHard || identityHard || handoffHard || !trial.ConsumerParseOK || !trial.TransferSemanticOK || !trial.DownstreamTaskOK || !trial.TaskSupportOK
	return trial
}

func inspectStage(stage StageObservation, trial *TrialScore) (*Attempt, bool) {
	if len(stage.Attempts) == 0 {
		return nil, false
	}
	var selected *Attempt
	for index := range stage.Attempts {
		attempt := &stage.Attempts[index]
		if attempt.CallExecuted {
			trial.CLIInvocations++
			trial.cost.add(attempt.Usage)
		}
		if index > 0 {
			trial.RuntimeRetries++
			trial.cost.Retries++
			if stage.Attempts[index-1].RuntimeOK || attempt.RetryReason != "runtime_hard_failure" {
				trial.Errors = append(trial.Errors, "retry was not caused only by the preceding runtime hard failure")
				trial.RuntimeHardFailureAttempts++
			}
		}
		if !attempt.RuntimeOK && attempt.CallExecuted {
			trial.RuntimeHardFailureAttempts++
		}
		if attempt.RuntimeOK {
			selected = attempt
		}
	}
	trial.cost.Calls += countCalls(stage.Attempts)
	return selected, selected != nil
}

func countCalls(attempts []Attempt) int {
	count := 0
	for _, attempt := range attempts {
		if attempt.CallExecuted {
			count++
		}
	}
	return count
}

func compareRecords(expected, actual []Record) recordChecks {
	checks := recordChecks{Semantic: true}
	expectedByKey := map[string]Record{}
	actualByKey := map[string][]Record{}
	for _, record := range expected {
		expectedByKey[recordKey(record)] = record
	}
	for _, record := range actual {
		key := recordKey(record)
		actualByKey[key] = append(actualByKey[key], record)
	}
	if len(expectedByKey) != len(expected) || len(expected) != len(actual) {
		checks.Semantic = false
	}

	expectedQ, actualQ := map[string]int{}, map[string]int{}
	for _, record := range expected {
		if record.Kind == "Q1" {
			expectedQ[record.Target]++
		}
	}
	for _, record := range actual {
		if record.Kind == "Q1" {
			actualQ[record.Target]++
		}
	}
	for target, count := range actualQ {
		if count > expectedQ[target] {
			checks.Q1FP += count - expectedQ[target]
		}
	}
	for target, count := range expectedQ {
		if count > actualQ[target] {
			checks.Q1FN += count - actualQ[target]
		}
	}
	if checks.Q1FP+checks.Q1FN > 0 {
		checks.Semantic = false
	}

	for key, want := range expectedByKey {
		found := actualByKey[key]
		if len(found) == 0 {
			checks.Semantic = false
			if want.Kind == "R1" {
				checks.Epistemic++
				if want.Evidence != nil {
					checks.Evidence++
				}
				if want.Bounds != nil {
					checks.Bounds++
				}
			}
			if want.Kind == "D1" {
				checks.D1Transmission++
			}
			continue
		}
		if len(found) > 1 {
			checks.Semantic = false
		}
		got := found[0]
		if !proseEqual(want.Content, got.Content) {
			checks.Semantic = false
		} else if punctuationOnlyEqual(want.Content, got.Content) {
			checks.PunctuationOnly++
		} else if want.Content != got.Content {
			checks.CapitalizationOnly++
		}
		if !nullableProseEqual(want.Bounds, got.Bounds) {
			checks.Bounds++
			checks.Semantic = false
		} else if nullablePunctuationOnlyEqual(want.Bounds, got.Bounds) {
			checks.PunctuationOnly++
		} else if !nullableExact(want.Bounds, got.Bounds) {
			checks.CapitalizationOnly++
		}
		switch want.Kind {
		case "R1":
			if got.Epistemic == "" {
				checks.Epistemic++
				checks.EpistemicMissing++
				checks.Semantic = false
			} else if want.Epistemic != got.Epistemic {
				checks.Epistemic++
				checks.Semantic = false
				if len(want.Epistemic) == 2 && len(got.Epistemic) == 2 && want.Epistemic[1] == 'U' && (got.Epistemic[1] == 'V' || got.Epistemic[1] == 'R') {
					checks.Promotions++
				}
			}
			if !nullableExact(want.Evidence, got.Evidence) {
				checks.Evidence++
				checks.Semantic = false
				if want.Evidence != nil && got.Evidence == nil {
					checks.EvidenceMissing++
				} else if evidenceBelongsToAnotherRecord(got.Evidence, expectedByKey, key) {
					checks.EvidenceMisassociated++
				}
			}
		case "Q1":
			if !proseEqual(want.Recommendation, got.Recommendation) || !nullableExact(want.Evidence, got.Evidence) {
				checks.Semantic = false
				if punctuationOnlyEqual(want.Recommendation, got.Recommendation) {
					checks.PunctuationOnly++
				} else if proseEqual(want.Recommendation, got.Recommendation) && want.Recommendation != got.Recommendation {
					checks.CapitalizationOnly++
				}
				if !nullableExact(want.Evidence, got.Evidence) {
					checks.Evidence++
					if want.Evidence != nil && got.Evidence == nil {
						checks.EvidenceMissing++
					} else if evidenceBelongsToAnotherRecord(got.Evidence, expectedByKey, key) {
						checks.EvidenceMisassociated++
					}
				}
			} else if punctuationOnlyEqual(want.Recommendation, got.Recommendation) {
				checks.PunctuationOnly++
			} else if want.Recommendation != got.Recommendation {
				checks.CapitalizationOnly++
			}
		case "D1":
			if want.AuthorityReference != got.AuthorityReference {
				checks.D1Authority++
				checks.D1Transmission++
				checks.Semantic = false
			}
			if !proseEqual(want.Content, got.Content) || !nullableProseEqual(want.Bounds, got.Bounds) {
				checks.D1Transmission++
				checks.Semantic = false
			}
		}
	}
	for key, found := range actualByKey {
		if _, ok := expectedByKey[key]; ok {
			continue
		}
		checks.Semantic = false
		for _, got := range found {
			if got.Kind == "D1" {
				checks.D1Authority++
				checks.D1Invented++
			}
		}
	}
	for _, got := range actual {
		if got.Kind == "R1" && len(got.Epistemic) == 2 && (got.Epistemic[1] == 'V' || got.Epistemic[1] == 'R') && got.Evidence == nil {
			checks.VerifiedWithoutEvidence++
			want, matched := expectedByKey[recordKey(got)]
			if !matched || want.Evidence == nil {
				checks.Evidence++
				checks.EvidenceMissing++
			}
			checks.Semantic = false
		}
	}
	return checks
}

func evidenceBelongsToAnotherRecord(evidence *string, expected map[string]Record, currentKey string) bool {
	if evidence == nil {
		return false
	}
	for key, record := range expected {
		if key != currentKey && record.Evidence != nil && *record.Evidence == *evidence {
			return true
		}
	}
	return false
}

func proseEqual(a, b string) bool {
	left, right := normalizeProse(a), normalizeProse(b)
	return left == right || sentenceCaseEqual(left, right)
}

func normalizeProse(value string) string {
	return strings.TrimRightFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || strings.ContainsRune(".!?。！？．", r)
	})
}

func sentenceCaseEqual(a, b string) bool {
	left, leftSize := utf8.DecodeRuneInString(a)
	right, rightSize := utf8.DecodeRuneInString(b)
	if leftSize == 0 || rightSize == 0 || leftSize != rightSize {
		return false
	}
	return unicode.ToLower(left) == unicode.ToLower(right) && a[leftSize:] == b[rightSize:]
}

func punctuationOnlyEqual(a, b string) bool {
	return a != b && normalizeProse(a) == normalizeProse(b)
}

func nullablePunctuationOnlyEqual(a, b *string) bool {
	return a != nil && b != nil && punctuationOnlyEqual(*a, *b)
}

func nullableExact(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}
func nullableProseEqual(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return proseEqual(*a, *b)
}

func semanticJSONEqual(a, b any, key string) bool {
	switch left := a.(type) {
	case map[string]any:
		right, ok := b.(map[string]any)
		if !ok || len(left) != len(right) {
			return false
		}
		for child, value := range left {
			if !semanticJSONEqual(value, right[child], child) {
				return false
			}
		}
		return true
	case []any:
		right, ok := b.([]any)
		if !ok || len(left) != len(right) {
			return false
		}
		for i := range left {
			if !semanticJSONEqual(left[i], right[i], key) {
				return false
			}
		}
		return true
	case string:
		right, ok := b.(string)
		if !ok {
			return false
		}
		if proseTaskKey(key) {
			return proseEqual(left, right)
		}
		return left == right
	default:
		return reflect.DeepEqual(a, b)
	}
}

func proseTaskKey(key string) bool {
	switch key {
	case "content", "decision", "request", "recommendation", "next_step", "reason", "statement":
		return true
	default:
		return false
	}
}

func taskSupportMatches(task map[string]any, records []Record) bool {
	target, targetOK := task["support_target"].(string)
	epistemic, epistemicIsString := task["support_epistemic"].(string)
	evidence, evidenceIsString := task["support_evidence"].(string)
	authority, authorityIsString := task["authority_reference"].(string)
	if !targetOK || target == "" {
		return false
	}
	matchedSupport := false
	matchedAuthority := task["authority_reference"] == nil
	for _, record := range records {
		recordEvidence := ""
		if record.Evidence != nil {
			recordEvidence = *record.Evidence
		}
		epistemicMatches := (record.Kind == "R1" && epistemicIsString && record.Epistemic == epistemic) || (record.Kind != "R1" && task["support_epistemic"] == nil)
		evidenceMatches := (evidenceIsString && recordEvidence == evidence) || (task["support_evidence"] == nil && record.Evidence == nil)
		if record.Target == target && epistemicMatches && evidenceMatches {
			matchedSupport = true
		}
		if authorityIsString && record.Kind == "D1" && record.AuthorityReference == authority {
			matchedAuthority = true
		}
	}
	return matchedSupport && matchedAuthority
}

func authorityViolations(records []Record, task map[string]any, allowed []string) int {
	violations := 0
	for _, record := range records {
		if record.Kind == "D1" && !containsString(allowed, record.AuthorityReference) {
			violations++
		}
	}
	if authority, ok := task["authority_reference"].(string); ok && !containsString(allowed, authority) {
		violations++
	}
	return violations
}

func addTrial(summary *CandidateSummary, trial TrialScore) {
	summary.ObservedCells++
	if trial.Cohort == "copy" {
		summary.CopyCells++
	} else {
		summary.SourceDerivedCells++
	}
	if trial.ProducerSerializationStatus == "identity-not-applicable" {
		summary.IdentityProducerCells++
	} else {
		summary.SerializationEligibleCells++
		if trial.ProducerSerializationStatus != "passed" {
			summary.SerializationFailures++
		}
		if trial.ProducerDerivationStatus == "passed" || trial.ProducerDerivationStatus == "failed" || trial.ProducerDerivationStatus == "runtime-failed" {
			summary.ProducerDerivationEligible++
			if trial.ProducerDerivationStatus == "passed" {
				summary.ProducerDerivationSuccesses++
			}
		}
	}
	summary.ConsumptionEligibleCells++
	if trial.ConsumptionSemanticStatus == "passed" {
		summary.ConsumptionSemanticSuccesses++
	} else {
		summary.ConsumptionSemanticFailures++
	}
	if !trial.ConsumerParseOK {
		summary.ConsumerParseFailures++
	}
	if trial.TransferSemanticOK {
		summary.TransferSemanticSuccesses++
	} else {
		summary.TransferSemanticFailures++
	}
	if trial.DownstreamTaskOK {
		summary.DownstreamTaskSuccesses++
	} else {
		summary.DownstreamTaskFailures++
	}
	summary.Q1FalsePositives += trial.Q1FalsePositives
	summary.Q1FalseNegatives += trial.Q1FalseNegatives
	if trial.Q1FalsePositives+trial.Q1FalseNegatives > 0 {
		summary.RoutingFailures++
	}
	summary.EpistemicFailures += trial.EpistemicFailures
	summary.EpistemicMissing += trial.EpistemicMissing
	summary.EpistemicPromotions += trial.EpistemicPromotions
	summary.EvidenceFailures += trial.EvidenceFailures
	summary.EvidenceMissing += trial.EvidenceMissing
	summary.EvidenceMisassociated += trial.EvidenceMisassociated
	summary.VerifiedWithoutEvidence += trial.VerifiedWithoutEvidence
	summary.BoundsFailures += trial.BoundsFailures
	summary.D1TransmissionFailures += trial.D1TransmissionFailures
	summary.D1AuthorityFailures += trial.D1AuthorityFailures
	summary.D1Invented += trial.D1Invented
	summary.PunctuationOnlyVariations += trial.PunctuationOnlyVariations
	summary.CapitalizationOnlyVariations += trial.CapitalizationOnlyVariations
	summary.RuntimeHardFailureAttempts += trial.RuntimeHardFailureAttempts
	if trial.HardFailure {
		summary.HardFailures++
	}
	for _, defect := range trial.Defects {
		summary.Defects[defect.Category]++
	}
	summary.cost.merge(trial.cost)
}

func (usage *usageAccumulator) add(value CodexUsage) {
	fields := []*int64{value.InputTokens, value.CachedInputTokens, value.CacheWriteInputTokens, value.OutputTokens, value.ReasoningOutputTokens}
	for _, field := range fields {
		if field != nil && *field < 0 {
			usage.invalidate("a Codex CLI invocation reported a negative token counter")
		}
	}
	if value.InputTokens != nil && value.CachedInputTokens != nil && value.CacheWriteInputTokens != nil &&
		*value.InputTokens >= 0 && *value.CachedInputTokens >= 0 && *value.CacheWriteInputTokens >= 0 &&
		(*value.CachedInputTokens > *value.InputTokens || *value.CacheWriteInputTokens > *value.InputTokens-*value.CachedInputTokens) {
		usage.invalidate("a Codex CLI invocation reported cached_input_tokens + cache_write_input_tokens greater than input_tokens")
	}
	for index, field := range fields {
		if field == nil {
			usage.unreported[index]++
		} else {
			usage.reported[index]++
			if *field >= 0 {
				if usage.values[index] > maxOperationalCostValue-*field {
					usage.invalidate("token counters overflow int64 while aggregating Codex CLI invocations")
				} else {
					usage.values[index] += *field
				}
			}
			if index == 0 && (usage.reported[index] == 1 || *field > usage.maxInputTokens) {
				usage.maxInputTokens = *field
			}
		}
	}
}

func (usage *usageAccumulator) merge(other usageAccumulator) {
	hadReportedInput := usage.reported[0] > 0
	if other.invalidUsage {
		usage.invalidate(other.invalidReason)
	}
	usage.Calls += other.Calls
	usage.Retries += other.Retries
	for i := range usage.values {
		if other.values[i] < 0 {
			usage.invalidate("scored trial contains a negative aggregate token counter")
		} else if usage.values[i] > maxOperationalCostValue-other.values[i] {
			usage.invalidate("token counters overflow int64 while merging scored trials")
		} else {
			usage.values[i] += other.values[i]
		}
		usage.reported[i] += other.reported[i]
		usage.unreported[i] += other.unreported[i]
	}
	if other.reported[0] > 0 && (!hadReportedInput || other.maxInputTokens > usage.maxInputTokens) {
		usage.maxInputTokens = other.maxInputTokens
	}
}

func (usage *usageAccumulator) invalidate(reason string) {
	if usage.invalidUsage {
		return
	}
	usage.invalidUsage = true
	usage.invalidReason = reason
}

func (usage usageAccumulator) report() CostVector {
	metrics := make([]Metric, 5)
	for i := range metrics {
		if usage.reported[i] > 0 {
			value := usage.values[i]
			metrics[i].Value = &value
		}
		metrics[i].ObservedTotal = usage.values[i]
		metrics[i].ReportedCalls = usage.reported[i]
		metrics[i].UnreportedCalls = usage.unreported[i]
		metrics[i].Available = usage.unreported[i] == 0 && usage.reported[i] == usage.Calls
		if !metrics[i].Available {
			metrics[i].Reason = "not reported for every executed Codex CLI invocation"
		}
		if usage.invalidUsage {
			metrics[i].Available = false
			metrics[i].Reason = usage.invalidReason
		}
	}
	unavailable := func(reason string) Metric {
		return Metric{Available: false, UnreportedCalls: usage.Calls, Reason: reason}
	}
	uncached := Metric{Available: false, Reason: "requires complete nonnegative input_tokens, cached_input_tokens, and cache_write_input_tokens whose sum does not exceed total input", UnreportedCalls: usage.Calls}
	if metrics[0].Available && metrics[1].Available && metrics[2].Available &&
		metrics[0].Value != nil && metrics[1].Value != nil && metrics[2].Value != nil &&
		*metrics[0].Value >= 0 && *metrics[1].Value >= 0 && *metrics[2].Value >= 0 &&
		*metrics[1].Value <= *metrics[0].Value && *metrics[2].Value <= *metrics[0].Value-*metrics[1].Value {
		value := *metrics[0].Value - *metrics[1].Value - *metrics[2].Value
		uncached = Metric{Available: true, Value: &value, ObservedTotal: value, ReportedCalls: usage.Calls}
	}
	maxInput := Metric{
		Available:       metrics[0].Available,
		ObservedTotal:   usage.values[0],
		ReportedCalls:   usage.reported[0],
		UnreportedCalls: usage.unreported[0],
	}
	if usage.reported[0] > 0 {
		value := usage.maxInputTokens
		maxInput.Value = &value
	}
	if !maxInput.Available {
		maxInput.Reason = metrics[0].Reason
	}
	return CostVector{
		CLIInvocations: usage.Calls, RuntimeRetries: usage.Retries,
		InputTokens: metrics[0], MaxInputTokensPerCall: maxInput, UncachedInputTokens: uncached, CachedInputTokens: metrics[1], CacheWriteInputTokens: metrics[2], OutputTokens: metrics[3], ReasoningOutputTokens: metrics[4],
		InstructionTokens:        unavailable("Codex turn usage does not separate instruction tokens from case input tokens"),
		TaskInputTokens:          unavailable("Codex turn usage does not separate case input tokens from instruction tokens"),
		ProviderInternalRequests: unavailable("Codex CLI invocation count is observed; internal provider request count is not exposed"),
		MonetaryCost:             unavailable("no billing amount is reported by the Codex event stream"),
	}
}

func selectParetoWinner(summaries []CandidateSummary, coverage CoverageReport, trials []TrialScore) Selection {
	selection := Selection{
		Status: "no_winner",
		Rule:   "winner requires complete coverage, no hard failures, and one candidate that weakly dominates every other candidate on correctness and every observed cost dimension with at least one strict improvement",
	}
	if coverage.Complete {
		var winners []string
		for i := range summaries {
			candidate := summaries[i]
			if candidate.HardFailures != 0 {
				continue
			}
			dominatesAll := true
			strict := false
			for j := range summaries {
				if i == j {
					continue
				}
				correctOK, correctStrict := correctnessDominance(candidate, summaries[j])
				costOK, costStrict := costDominance(candidate.Cost, summaries[j].Cost)
				if !correctOK || !costOK {
					dominatesAll = false
					break
				}
				strict = strict || correctStrict || costStrict
			}
			if dominatesAll && strict {
				winners = append(winners, candidate.Candidate)
			}
		}
		if len(winners) == 1 {
			winner := winners[0]
			selection.Winner = &winner
			selection.Status = "winner"
			return selection
		}
	}
	selection.NextDiscriminatingTest = chooseNextDiscriminatingTest(coverage, trials, summaries)
	return selection
}

// selectWinner preserves the former internal entry point for callers that
// still treat the dominance result as a diagnostic.
func selectWinner(summaries []CandidateSummary, coverage CoverageReport, trials []TrialScore) Selection {
	return selectParetoWinner(summaries, coverage, trials)
}

const (
	nanoUSDPerUSD            int64 = 1_000_000_000
	maxOperationalCostValue  int64 = 1<<63 - 1
	operationalSelectionRule       = "complete coverage; hard failures == 0; complete semantic consumption, transfer, and downstream correctness; lowest estimated paid-token cost; only on an exact estimated-cost tie, lowest input + output token count; an exact remaining tie leaves operational selection unresolved; unavailable cost remains unavailable and blocks selection"
)

type operationalRank struct {
	candidate         string
	costUnits         int64
	inputOutputTokens int64
}

func validateOperationalPricing(pricing OperationalPricing, expectedModel string) error {
	if strings.TrimSpace(pricing.PolicyVersion) == "" || strings.TrimSpace(pricing.Model) == "" ||
		strings.TrimSpace(pricing.VerifiedOn) == "" || strings.TrimSpace(pricing.Source) == "" ||
		strings.TrimSpace(pricing.ReasoningOutputCostTreatment) == "" {
		return fmt.Errorf("policy_version, model, verified_on, source, and reasoning_output_cost_treatment are required")
	}
	if expectedModel != "" && pricing.Model != expectedModel {
		return fmt.Errorf("model %q does not match run model %q", pricing.Model, expectedModel)
	}
	if pricing.UncachedInputNanoUSDPerToken <= 0 || pricing.CachedInputNanoUSDPerToken <= 0 ||
		pricing.CacheWriteInputNanoUSDPerToken <= 0 || pricing.OutputNanoUSDPerToken <= 0 ||
		pricing.StandardRateMaxInputTokensPerCall <= 0 {
		return fmt.Errorf("all nanoUSD-per-token rates and standard_rate_max_input_tokens_per_call must be positive")
	}
	return nil
}

func selectOperationalWinner(summaries []CandidateSummary, coverage CoverageReport, policy *OperationalPricing) OperationalSelection {
	selection := OperationalSelection{Status: "not_configured", Rule: operationalSelectionRule}
	if policy == nil {
		selection.Reason = "operational selection requires an explicit pricing policy"
		return selection
	}
	if err := validateOperationalPricing(*policy, ""); err != nil {
		selection.Status = "invalid_policy"
		selection.Reason = err.Error()
		return selection
	}
	pricing := *policy
	selection.Status = "no_winner"
	selection.Pricing = &pricing
	ranked := make([]operationalRank, 0, len(summaries))
	evaluationIndex := make(map[string]int, len(summaries))
	costUnavailable := false

	for _, summary := range summaries {
		evaluation := OperationalEvaluation{
			Candidate:                    summary.Candidate,
			CoverageComplete:             coverage.Complete && summary.ExpectedCells > 0 && summary.ObservedCells == summary.ExpectedCells,
			HardFailures:                 summary.HardFailures,
			ConsumptionSemanticSuccesses: summary.ConsumptionSemanticSuccesses,
			ConsumptionSemanticFailures:  summary.ConsumptionSemanticFailures,
			TransferSemanticSuccesses:    summary.TransferSemanticSuccesses,
			TransferSemanticFailures:     summary.TransferSemanticFailures,
			DownstreamTaskSuccesses:      summary.DownstreamTaskSuccesses,
			DownstreamTaskFailures:       summary.DownstreamTaskFailures,
			Disposition:                  "blocked",
		}
		if nanoUSD, inputOutput, ok := estimatedPaidTokenCost(summary.Cost, pricing); ok {
			costUSD := float64(nanoUSD) / float64(nanoUSDPerUSD)
			evaluation.EstimatedPaidTokenCostUSD = &costUSD
			evaluation.EstimatedPaidTokenCostNanoUSD = &nanoUSD
			evaluation.InputOutputTokens = &inputOutput
		}

		switch {
		case !evaluation.CoverageComplete:
			evaluation.Reason = "complete candidate coverage is required"
		case summary.HardFailures != 0:
			evaluation.Disposition = "rejected"
			evaluation.Reason = "hard failures must equal zero"
			selection.Rejected = append(selection.Rejected, summary.Candidate)
		case !completeOperationalCorrectness(summary):
			evaluation.Disposition = "rejected"
			evaluation.Reason = "complete semantic consumption and downstream correctness are required"
			selection.Rejected = append(selection.Rejected, summary.Candidate)
		default:
			evaluation.Eligible = true
			evaluation.Disposition = "eligible"
			if evaluation.EstimatedPaidTokenCostNanoUSD == nil || evaluation.InputOutputTokens == nil {
				evaluation.Reason = "estimated paid-token cost requires complete, consistent, nonnegative input/output metrics and every call within the pricing policy's standard-rate input limit"
				costUnavailable = true
			} else {
				ranked = append(ranked, operationalRank{
					candidate:         summary.Candidate,
					costUnits:         *evaluation.EstimatedPaidTokenCostNanoUSD,
					inputOutputTokens: *evaluation.InputOutputTokens,
				})
			}
		}
		evaluationIndex[summary.Candidate] = len(selection.Evaluations)
		selection.Evaluations = append(selection.Evaluations, evaluation)
	}

	sort.Strings(selection.Rejected)
	if !coverage.Complete {
		selection.Reason = "operational selection is blocked because coverage is incomplete"
		return selection
	}
	if costUnavailable {
		selection.Reason = "operational selection is blocked because an eligible candidate lacks a complete estimated paid token cost"
		return selection
	}
	if len(ranked) == 0 {
		selection.Reason = "no candidate passed the hard-failure and semantic/downstream correctness gates"
		return selection
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].costUnits != ranked[j].costUnits {
			return ranked[i].costUnits < ranked[j].costUnits
		}
		return ranked[i].inputOutputTokens < ranked[j].inputOutputTokens
	})
	if len(ranked) > 1 && sameOperationalRank(ranked[0], ranked[1]) {
		selection.Reason = "the leading eligible candidates tie on estimated paid token cost and input + output token count"
		return selection
	}

	winner := ranked[0].candidate
	selection.Status = "winner"
	selection.Winner = &winner
	selection.Reason = winner + " has the lowest estimated paid token cost after the coverage, hard-failure, and semantic/downstream correctness gates"
	selection.Evaluations[evaluationIndex[winner]].Disposition = "winner"
	selection.Evaluations[evaluationIndex[winner]].Reason = "selected by the operational order"

	if len(ranked) > 1 && (len(ranked) == 2 || !sameOperationalRank(ranked[1], ranked[2])) {
		challenger := ranked[1].candidate
		selection.Challenger = &challenger
		selection.Evaluations[evaluationIndex[challenger]].Disposition = "challenger"
		selection.Evaluations[evaluationIndex[challenger]].Reason = "next eligible candidate under the operational order"
	}
	return selection
}

func completeOperationalCorrectness(summary CandidateSummary) bool {
	return summary.ConsumptionEligibleCells == summary.ObservedCells &&
		summary.ConsumptionSemanticFailures == 0 &&
		summary.ConsumptionSemanticSuccesses == summary.ConsumptionEligibleCells &&
		summary.TransferSemanticFailures == 0 &&
		summary.TransferSemanticSuccesses == summary.ObservedCells &&
		summary.DownstreamTaskFailures == 0 &&
		summary.DownstreamTaskSuccesses == summary.ObservedCells
}

func estimatedPaidTokenCost(cost CostVector, pricing OperationalPricing) (nanoUSD, inputOutputTokens int64, ok bool) {
	if !cost.MaxInputTokensPerCall.Available || cost.MaxInputTokensPerCall.Value == nil ||
		*cost.MaxInputTokensPerCall.Value < 0 || *cost.MaxInputTokensPerCall.Value > pricing.StandardRateMaxInputTokensPerCall {
		return 0, 0, false
	}
	metrics := []struct {
		metric Metric
		rate   int64
	}{
		{metric: cost.UncachedInputTokens, rate: pricing.UncachedInputNanoUSDPerToken},
		{metric: cost.CachedInputTokens, rate: pricing.CachedInputNanoUSDPerToken},
		{metric: cost.CacheWriteInputTokens, rate: pricing.CacheWriteInputNanoUSDPerToken},
		{metric: cost.OutputTokens, rate: pricing.OutputNanoUSDPerToken},
	}
	for _, item := range metrics {
		if item.rate <= 0 || !item.metric.Available || item.metric.Value == nil || *item.metric.Value < 0 {
			return 0, 0, false
		}
		value := *item.metric.Value
		if value > maxOperationalCostValue/item.rate {
			return 0, 0, false
		}
		term := value * item.rate
		if nanoUSD > maxOperationalCostValue-term {
			return 0, 0, false
		}
		nanoUSD += term
	}
	if !cost.InputTokens.Available || cost.InputTokens.Value == nil || *cost.InputTokens.Value < 0 ||
		!cost.OutputTokens.Available || cost.OutputTokens.Value == nil || *cost.OutputTokens.Value < 0 ||
		*cost.InputTokens.Value > maxOperationalCostValue-*cost.OutputTokens.Value {
		return 0, 0, false
	}
	uncached := *cost.UncachedInputTokens.Value
	cached := *cost.CachedInputTokens.Value
	cacheWrite := *cost.CacheWriteInputTokens.Value
	if uncached > maxOperationalCostValue-cached {
		return 0, 0, false
	}
	accountedInput := uncached + cached
	if accountedInput > maxOperationalCostValue-cacheWrite || accountedInput+cacheWrite != *cost.InputTokens.Value {
		return 0, 0, false
	}
	inputOutputTokens = *cost.InputTokens.Value + *cost.OutputTokens.Value
	return nanoUSD, inputOutputTokens, true
}

func sameOperationalRank(left, right operationalRank) bool {
	return left.costUnits == right.costUnits && left.inputOutputTokens == right.inputOutputTokens
}

func chooseNextDiscriminatingTest(coverage CoverageReport, trials []TrialScore, summaries []CandidateSummary) *NextDiscriminatingTest {
	if len(coverage.MissingCells) > 0 {
		parts := strings.SplitN(coverage.MissingCells[0], "/", 2)
		candidate := Candidates
		caseID := parts[0]
		if len(parts) == 2 && containsString(Candidates, parts[1]) {
			candidate = []string{parts[1]}
		}
		return &NextDiscriminatingTest{
			CaseID: caseID, Candidates: append([]string(nil), candidate...), Trigger: "missing-coverage-cell",
			Purpose: "complete the one missing preregistered coverage cell, then reapply the same decision rule",
		}
	}
	if safe, unsafe, ok := correctnessCostFrontierPair(summaries); ok {
		for _, trial := range trials {
			if trial.Candidate == unsafe && trial.Cohort == "source-derived" && trial.HardFailure {
				return &NextDiscriminatingTest{
					CaseID: trial.CaseID, Candidates: []string{unsafe, safe}, Trigger: "observed-correctness-cost-frontier",
					Purpose: "run one matched second observation to distinguish stable semantic loss in the cheaper candidate from one-shot model variance; the initial hard failure remains part of the evidence",
				}
			}
		}
	}
	if left, right, ok := safeCostCrossingPair(summaries); ok {
		return &NextDiscriminatingTest{
			CaseID: largestUncachedSpreadCase(trials, left, right), Candidates: []string{left, right}, Trigger: "observed-safe-cost-vector-crossing",
			Purpose: "run one matched billing-enabled observation for the safe frontier pair so provider-reported total cost can resolve the crossing uncached-input, cached-input, output, and reasoning vector",
		}
	}
	for _, trial := range trials {
		if trial.HardFailure {
			return &NextDiscriminatingTest{
				CaseID: trial.CaseID, Candidates: append([]string(nil), Candidates...), Trigger: "observed-hard-failure",
				Purpose: "after evidence-backed defect classification or repair, rerun this one matched case once across all four candidates",
			}
		}
	}
	// With complete correctness but crossing or incomplete cost vectors, repeat
	// the derivation case with the largest observed input-token spread. This is
	// driven by the observed decision uncertainty, not a preset trial count.
	type span struct {
		caseID string
		min    int64
		max    int64
		set    bool
	}
	spans := map[string]*span{}
	for _, trial := range trials {
		value := trial.cost.values[0]
		entry := spans[trial.CaseID]
		if entry == nil {
			entry = &span{caseID: trial.CaseID, min: value, max: value, set: true}
			spans[trial.CaseID] = entry
		} else {
			if value < entry.min {
				entry.min = value
			}
			if value > entry.max {
				entry.max = value
			}
		}
	}
	var selected *span
	for _, entry := range spans {
		if selected == nil || entry.max-entry.min > selected.max-selected.min || (entry.max-entry.min == selected.max-selected.min && entry.caseID < selected.caseID) {
			selected = entry
		}
	}
	caseID := "unavailable"
	if selected != nil {
		caseID = selected.caseID
	}
	return &NextDiscriminatingTest{
		CaseID: caseID, Candidates: append([]string(nil), Candidates...), Trigger: "observed-cost-decision-uncertainty",
		Purpose: "repeat the single case with the largest observed input-token spread once to test whether the crossing cost vector is stable",
	}
}

func safeCostCrossingPair(summaries []CandidateSummary) (string, string, bool) {
	for i := range summaries {
		if summaries[i].HardFailures != 0 {
			continue
		}
		for j := i + 1; j < len(summaries); j++ {
			if summaries[j].HardFailures != 0 {
				continue
			}
			leftCostOK, _ := costDominance(summaries[i].Cost, summaries[j].Cost)
			rightCostOK, _ := costDominance(summaries[j].Cost, summaries[i].Cost)
			if !leftCostOK && !rightCostOK {
				return summaries[i].Candidate, summaries[j].Candidate, true
			}
		}
	}
	return "", "", false
}

func largestUncachedSpreadCase(trials []TrialScore, left, right string) string {
	byCase := map[string]map[string]int64{}
	for _, trial := range trials {
		if trial.Cohort != "source-derived" || (trial.Candidate != left && trial.Candidate != right) {
			continue
		}
		if byCase[trial.CaseID] == nil {
			byCase[trial.CaseID] = map[string]int64{}
		}
		byCase[trial.CaseID][trial.Candidate] = trial.cost.values[0] - trial.cost.values[1]
	}
	selected := "unavailable"
	var largest int64 = -1
	for caseID, values := range byCase {
		leftValue, leftOK := values[left]
		rightValue, rightOK := values[right]
		if !leftOK || !rightOK {
			continue
		}
		spread := leftValue - rightValue
		if spread < 0 {
			spread = -spread
		}
		if spread > largest || (spread == largest && caseID < selected) {
			selected, largest = caseID, spread
		}
	}
	return selected
}

func correctnessCostFrontierPair(summaries []CandidateSummary) (safe, unsafe string, ok bool) {
	var safeCandidates []CandidateSummary
	for _, summary := range summaries {
		if summary.HardFailures == 0 {
			safeCandidates = append(safeCandidates, summary)
		}
	}
	for _, candidate := range safeCandidates {
		dominatesSafe := true
		for _, other := range safeCandidates {
			if candidate.Candidate == other.Candidate {
				continue
			}
			correctOK, _ := correctnessDominance(candidate, other)
			costOK, _ := costDominance(candidate.Cost, other.Cost)
			if !correctOK || !costOK {
				dominatesSafe = false
				break
			}
		}
		if dominatesSafe {
			if safe != "" {
				return "", "", false
			}
			safe = candidate.Candidate
		}
	}
	if safe == "" {
		return "", "", false
	}
	var anchor CandidateSummary
	for _, summary := range summaries {
		if summary.Candidate == safe {
			anchor = summary
			break
		}
	}
	var selected *CandidateSummary
	for index := range summaries {
		candidate := &summaries[index]
		if candidate.HardFailures == 0 || candidate.Cost.CLIInvocations >= anchor.Cost.CLIInvocations {
			continue
		}
		if selected == nil || candidate.HardFailures < selected.HardFailures ||
			(candidate.HardFailures == selected.HardFailures && candidate.Cost.CLIInvocations < selected.Cost.CLIInvocations) ||
			(candidate.HardFailures == selected.HardFailures && candidate.Cost.CLIInvocations == selected.Cost.CLIInvocations && candidate.Candidate < selected.Candidate) {
			selected = candidate
		}
	}
	if selected == nil {
		return "", "", false
	}
	return safe, selected.Candidate, true
}

func correctnessDominance(a, b CandidateSummary) (bool, bool) {
	lowerA := []int{a.SerializationFailures, a.ConsumerParseFailures, a.TransferSemanticFailures, a.DownstreamTaskFailures, a.Q1FalsePositives, a.Q1FalseNegatives, a.EpistemicFailures, a.EvidenceFailures, a.BoundsFailures, a.D1TransmissionFailures, a.D1AuthorityFailures, a.RuntimeHardFailureAttempts, a.HardFailures}
	lowerB := []int{b.SerializationFailures, b.ConsumerParseFailures, b.TransferSemanticFailures, b.DownstreamTaskFailures, b.Q1FalsePositives, b.Q1FalseNegatives, b.EpistemicFailures, b.EvidenceFailures, b.BoundsFailures, b.D1TransmissionFailures, b.D1AuthorityFailures, b.RuntimeHardFailureAttempts, b.HardFailures}
	strict := false
	for i := range lowerA {
		if lowerA[i] > lowerB[i] {
			return false, false
		}
		if lowerA[i] < lowerB[i] {
			strict = true
		}
	}
	if a.DownstreamTaskSuccesses < b.DownstreamTaskSuccesses || a.TransferSemanticSuccesses < b.TransferSemanticSuccesses {
		return false, false
	}
	if a.DownstreamTaskSuccesses > b.DownstreamTaskSuccesses || a.TransferSemanticSuccesses > b.TransferSemanticSuccesses {
		strict = true
	}
	return true, strict
}

func costDominance(a, b CostVector) (bool, bool) {
	left := []int64{int64(a.CLIInvocations), int64(a.RuntimeRetries)}
	right := []int64{int64(b.CLIInvocations), int64(b.RuntimeRetries)}
	// input_tokens includes cached_input_tokens in the Codex event schema.
	// Compare mutually exclusive uncached and cached components to avoid
	// counting cached tokens twice or claiming false dominance.
	metricsA := []Metric{a.UncachedInputTokens, a.CachedInputTokens, a.CacheWriteInputTokens, a.OutputTokens, a.ReasoningOutputTokens}
	metricsB := []Metric{b.UncachedInputTokens, b.CachedInputTokens, b.CacheWriteInputTokens, b.OutputTokens, b.ReasoningOutputTokens}
	for i := range metricsA {
		if !metricsA[i].Available || !metricsB[i].Available {
			return false, false
		}
		left = append(left, *metricsA[i].Value)
		right = append(right, *metricsB[i].Value)
	}
	strict := false
	for i := range left {
		if left[i] > right[i] {
			return false, false
		}
		if left[i] < right[i] {
			strict = true
		}
	}
	return true, strict
}

func cellKey(caseID, candidate string) string { return caseID + "/" + candidate }

func (report Report) ExitCode() int {
	if report.Coverage.HardError {
		return 2
	}
	return 0
}

func ValidateObservations(fixtures []Fixture, observations []Observation) error {
	report := Score(fixtures, observations)
	if report.Coverage.HardError {
		return fmt.Errorf("coverage hard error: missing=%v duplicate=%v", report.Coverage.MissingCells, report.Coverage.DuplicateCells)
	}
	return nil
}
