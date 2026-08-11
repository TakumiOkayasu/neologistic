package phase2

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestSyntheticPerfectCoverageSeparatesIdentityCopyAndDerivation(t *testing.T) {
	fixtures := testFixtures(t)
	report := Score(fixtures, perfectObservations(t, fixtures))
	if !report.Coverage.Complete || report.Coverage.ExpectedCells != 20 {
		t.Fatalf("coverage = %+v", report.Coverage)
	}
	direct := candidateSummary(t, report, CandidateDirect)
	if direct.SerializationEligibleCells != 0 || direct.ProducerDerivationEligible != 0 || direct.IdentityProducerCells != len(fixtures) {
		t.Fatalf("direct denominators mixed: %+v", direct)
	}
	if direct.ConsumptionEligibleCells != len(fixtures) || direct.ConsumptionSemanticFailures != 0 {
		t.Fatalf("direct consumption denominator = %+v", direct)
	}
	pipe := candidateSummary(t, report, CandidatePipe)
	if pipe.ProducerDerivationEligible != 4 || pipe.ProducerDerivationSuccesses != 4 {
		t.Fatalf("copy cohort entered derivation denominator: %+v", pipe)
	}
	if pipe.ConsumptionEligibleCells != len(fixtures) || pipe.ConsumptionSemanticFailures != 0 {
		t.Fatalf("consumption denominator = %+v", pipe)
	}
	if report.ParetoSelection.Winner == nil || *report.ParetoSelection.Winner != CandidateDirect {
		t.Fatalf("pareto winner = %+v", report.ParetoSelection)
	}
	if report.Selection.Winner == nil || *report.Selection.Winner != CandidateDirect {
		t.Fatalf("legacy pareto selection = %+v", report.Selection)
	}
	if report.OperationalSelection.Status != "not_configured" || report.OperationalSelection.Winner != nil {
		t.Fatalf("unconfigured operational selection = %+v", report.OperationalSelection)
	}
	if direct.Cost.MonetaryCost.Available || direct.Cost.ProviderInternalRequests.Available {
		t.Fatal("unavailable monetary/internal request dimensions became available")
	}
}

func TestSyntheticQ1FalsePositive(t *testing.T) {
	fixtures := testFixtures(t)
	observations := perfectObservations(t, fixtures)
	fixture := fixtureNamed(t, fixtures, "derive-autonomous")
	records := append(cloneRecords(fixture.ExpectedRecords), Record{Kind: "Q1", Target: "HUMAN:unneeded", Content: "Choose generator", Recommendation: "Use existing generator", Evidence: nil, Bounds: stringPointer("incorrect human boundary")})
	setOutputs(t, observations, fixture, CandidatePipe, records, fixture.TaskOracle)
	report := Score(fixtures, observations)
	trial := trialNamed(t, report, fixture.ID, CandidatePipe)
	if trial.Q1FalsePositives != 1 || !trial.HardFailure {
		t.Fatalf("trial = %+v", trial)
	}
}

func TestSyntheticQ1FalseNegative(t *testing.T) {
	fixtures := testFixtures(t)
	observations := perfectObservations(t, fixtures)
	fixture := fixtureNamed(t, fixtures, "derive-human-boundary")
	records := []Record{{Kind: "R1", Target: "CONTRACT:legacy-api", Epistemic: "IU", Content: "compatibility remains unresolved", Evidence: stringPointer("analysis:compat#9"), Bounds: stringPointer("maintainer decision absent")}}
	setOutputs(t, observations, fixture, CandidatePipe, records, fixture.TaskOracle)
	trial := trialNamed(t, Score(fixtures, observations), fixture.ID, CandidatePipe)
	if trial.Q1FalseNegatives != 1 || !trial.HardFailure {
		t.Fatalf("trial = %+v", trial)
	}
}

func TestSyntheticEpistemicPromotion(t *testing.T) {
	fixtures := testFixtures(t)
	observations := perfectObservations(t, fixtures)
	fixture := fixtureNamed(t, fixtures, "derive-epistemic-evidence")
	records := cloneRecords(fixture.ExpectedRecords)
	records[1].Epistemic = "IV"
	setOutputs(t, observations, fixture, CandidatePipe, records, fixture.TaskOracle)
	trial := trialNamed(t, Score(fixtures, observations), fixture.ID, CandidatePipe)
	if trial.EpistemicPromotions != 1 || trial.EpistemicFailures != 1 || !trial.HardFailure {
		t.Fatalf("trial = %+v", trial)
	}
}

func TestSyntheticEvidenceMisassociation(t *testing.T) {
	fixtures := testFixtures(t)
	observations := perfectObservations(t, fixtures)
	fixture := fixtureNamed(t, fixtures, "derive-epistemic-evidence")
	records := cloneRecords(fixture.ExpectedRecords)
	records[0].Evidence, records[1].Evidence = records[1].Evidence, records[0].Evidence
	setOutputs(t, observations, fixture, CandidatePipe, records, fixture.TaskOracle)
	trial := trialNamed(t, Score(fixtures, observations), fixture.ID, CandidatePipe)
	if trial.EvidenceFailures != 2 || !trial.HardFailure {
		t.Fatalf("trial = %+v", trial)
	}
}

func TestSyntheticD1AuthorityViolation(t *testing.T) {
	fixtures := testFixtures(t)
	observations := perfectObservations(t, fixtures)
	fixture := fixtureNamed(t, fixtures, "derive-human-boundary")
	records := append(cloneRecords(fixture.ExpectedRecords), Record{Kind: "D1", Target: "CONTRACT:legacy-removal", Content: "Remove the legacy endpoint", AuthorityReference: "invented:owner", Bounds: stringPointer("all callers")})
	setOutputs(t, observations, fixture, CandidatePipe, records, fixture.TaskOracle)
	trial := trialNamed(t, Score(fixtures, observations), fixture.ID, CandidatePipe)
	if trial.D1AuthorityFailures != 1 || !trial.HardFailure {
		t.Fatalf("trial = %+v", trial)
	}
}

func TestSyntheticPunctuationOnlyDifferenceIsNotSemanticHardFailure(t *testing.T) {
	fixtures := testFixtures(t)
	observations := perfectObservations(t, fixtures)
	fixture := fixtureNamed(t, fixtures, "copy-multirecord")
	records := cloneRecords(fixture.ExpectedRecords)
	records[0].Content += "。"
	setOutputs(t, observations, fixture, CandidatePipe, records, fixture.TaskOracle)
	trial := trialNamed(t, Score(fixtures, observations), fixture.ID, CandidatePipe)
	if !trial.TransferSemanticOK || !trial.DownstreamTaskOK || trial.HardFailure {
		t.Fatalf("trial = %+v", trial)
	}
}

func TestSyntheticInternalPunctuationDifferenceIsSemanticHardFailure(t *testing.T) {
	fixtures := testFixtures(t)
	observations := perfectObservations(t, fixtures)
	fixture := fixtureNamed(t, fixtures, "copy-multirecord")
	records := cloneRecords(fixture.ExpectedRecords)
	records[0].Content = "unit-tests pass"
	setOutputs(t, observations, fixture, CandidatePipe, records, fixture.TaskOracle)
	trial := trialNamed(t, Score(fixtures, observations), fixture.ID, CandidatePipe)
	if trial.TransferSemanticOK || !trial.HardFailure {
		t.Fatalf("trial = %+v", trial)
	}
}

func TestSyntheticSentenceInitialCapitalizationIsNotSemanticHardFailure(t *testing.T) {
	fixtures := testFixtures(t)
	observations := perfectObservations(t, fixtures)
	fixture := fixtureNamed(t, fixtures, "derive-human-boundary")
	records := cloneRecords(fixture.ExpectedRecords)
	records[0].Content = "Choose whether to retain the legacy public API."
	records[0].Recommendation = "Retain the legacy public API."
	setOutputs(t, observations, fixture, CandidateJSON, records, fixture.TaskOracle)
	trial := trialNamed(t, Score(fixtures, observations), fixture.ID, CandidateJSON)
	if !trial.TransferSemanticOK || trial.HardFailure || trial.CapitalizationOnlyVariations != 2 {
		t.Fatalf("trial = %+v", trial)
	}
}

func TestSyntheticMissingCellIsCoverageHardError(t *testing.T) {
	fixtures := testFixtures(t)
	observations := perfectObservations(t, fixtures)
	report := Score(fixtures, observations[:len(observations)-1])
	if !report.Coverage.HardError || report.ExitCode() != 2 || report.Selection.Winner != nil || report.OperationalSelection.Winner != nil || report.ParetoSelection.Winner != nil {
		t.Fatalf("report = %+v", report)
	}
	if report.ParetoSelection.NextDiscriminatingTest == nil {
		t.Fatalf("pareto next test = %+v", report.ParetoSelection)
	}
}

func TestNextTestTargetsObservedCorrectnessCostFrontier(t *testing.T) {
	cost := func(calls int, tokens int64) CostVector {
		metric := Metric{Available: true, Value: &tokens, ObservedTotal: tokens, ReportedCalls: calls}
		return CostVector{CLIInvocations: calls, InputTokens: metric, UncachedInputTokens: metric, CachedInputTokens: metric, CacheWriteInputTokens: metric, OutputTokens: metric, ReasoningOutputTokens: metric}
	}
	summaries := []CandidateSummary{
		{Candidate: CandidateDirect, HardFailures: 1, Cost: cost(5, 5)},
		{Candidate: CandidateNL, HardFailures: 3, Cost: cost(10, 30)},
		{Candidate: CandidatePipe, HardFailures: 0, Cost: cost(10, 20)},
		{Candidate: CandidateJSON, HardFailures: 0, Cost: cost(10, 10)},
	}
	trials := []TrialScore{{CaseID: "derive-epistemic-evidence", Cohort: "source-derived", Candidate: CandidateDirect, HardFailure: true}}
	next := chooseNextDiscriminatingTest(CoverageReport{Complete: true}, trials, summaries)
	if next == nil || next.Trigger != "observed-correctness-cost-frontier" || len(next.Candidates) != 2 || next.Candidates[0] != CandidateDirect || next.Candidates[1] != CandidateJSON {
		t.Fatalf("next test = %+v", next)
	}
}

func TestNextTestTargetsSafeCostVectorCrossing(t *testing.T) {
	metric := func(value int64) Metric {
		return Metric{Available: true, Value: &value, ObservedTotal: value, ReportedCalls: 10}
	}
	cost := func(uncached, cached, output, reasoning int64) CostVector {
		return CostVector{CLIInvocations: 10, InputTokens: metric(uncached + cached), UncachedInputTokens: metric(uncached), CachedInputTokens: metric(cached), CacheWriteInputTokens: metric(0), OutputTokens: metric(output), ReasoningOutputTokens: metric(reasoning)}
	}
	summaries := []CandidateSummary{
		{Candidate: CandidateDirect, HardFailures: 2, Cost: cost(40, 10, 5, 5)},
		{Candidate: CandidateNL, HardFailures: 3, Cost: cost(30, 30, 10, 10)},
		{Candidate: CandidatePipe, HardFailures: 0, Cost: cost(10, 20, 10, 20)},
		{Candidate: CandidateJSON, HardFailures: 0, Cost: cost(20, 10, 9, 10)},
	}
	input := int64(30)
	value := int64(10)
	usage := CodexUsage{InputTokens: &input, CachedInputTokens: &value, CacheWriteInputTokens: &value, OutputTokens: &value, ReasoningOutputTokens: &value}
	trials := []TrialScore{
		{CaseID: "derive", Cohort: "source-derived", Candidate: CandidatePipe, cost: usageAccumulator{}},
		{CaseID: "derive", Cohort: "source-derived", Candidate: CandidateJSON, cost: usageAccumulator{}},
	}
	trials[0].cost.add(usage)
	trials[1].cost.add(usage)
	next := chooseNextDiscriminatingTest(CoverageReport{Complete: true}, trials, summaries)
	if next == nil || next.Trigger != "observed-safe-cost-vector-crossing" || len(next.Candidates) != 2 || next.Candidates[0] != CandidatePipe || next.Candidates[1] != CandidateJSON {
		t.Fatalf("next test = %+v", next)
	}
}

func TestCurrentPhase2MetricsSelectJSONOperationallyAndKeepParetoDiagnostic(t *testing.T) {
	summaries := []CandidateSummary{
		currentOperationalSummary(CandidateDirect, 2, 3, 30_761, 46_848, 1_017, 372, 5),
		currentOperationalSummary(CandidateNL, 2, 3, 78_583, 76_288, 1_348, 308, 10),
		currentOperationalSummary(CandidatePipe, 0, 5, 64_421, 91_648, 1_148, 269, 10),
		currentOperationalSummary(CandidateJSON, 0, 5, 41_010, 113_152, 1_356, 326, 10),
	}
	coverage := CoverageReport{Complete: true}

	pareto := selectParetoWinner(summaries, coverage, nil)
	if pareto.Status != "no_winner" || pareto.Winner != nil {
		t.Fatalf("pareto selection = %+v", pareto)
	}
	operational := selectOperationalWinner(summaries, coverage, operationalPricingPointer())
	if operational.Status != "winner" || operational.Winner == nil || *operational.Winner != CandidateJSON {
		t.Fatalf("operational selection = %+v", operational)
	}
	if operational.Challenger == nil || *operational.Challenger != CandidatePipe {
		t.Fatalf("challenger = %+v", operational.Challenger)
	}
	if !reflect.DeepEqual(operational.Rejected, []string{CandidateDirect, CandidateNL}) {
		t.Fatalf("rejected = %v", operational.Rejected)
	}
	if operational.Pricing == nil || operational.Pricing.Model != "gpt-5.6-sol" || operational.Pricing.UncachedInputNanoUSDPerToken != 5_000 || operational.Pricing.CachedInputNanoUSDPerToken != 500 || operational.Pricing.CacheWriteInputNanoUSDPerToken != 6_250 || operational.Pricing.OutputNanoUSDPerToken != 30_000 || operational.Pricing.StandardRateMaxInputTokensPerCall != 272_000 || operational.Pricing.Source != "https://developers.openai.com/api/docs/models/gpt-5.6-sol" {
		t.Fatalf("pricing provenance = %+v", operational.Pricing)
	}

	jsonEvaluation := operationalEvaluation(t, operational, CandidateJSON)
	if jsonEvaluation.EstimatedPaidTokenCostNanoUSD == nil || *jsonEvaluation.EstimatedPaidTokenCostNanoUSD != 302_306_000 {
		t.Fatalf("json estimated cost = %+v", jsonEvaluation.EstimatedPaidTokenCostNanoUSD)
	}
	if jsonEvaluation.EstimatedPaidTokenCostUSD == nil || *jsonEvaluation.EstimatedPaidTokenCostUSD != 0.302306 {
		t.Fatalf("json estimated USD = %+v", jsonEvaluation.EstimatedPaidTokenCostUSD)
	}
	if jsonEvaluation.InputOutputTokens == nil || *jsonEvaluation.InputOutputTokens != 155_518 {
		t.Fatalf("json input + output = %+v", jsonEvaluation.InputOutputTokens)
	}
	pipeEvaluation := operationalEvaluation(t, operational, CandidatePipe)
	if pipeEvaluation.EstimatedPaidTokenCostNanoUSD == nil || *pipeEvaluation.EstimatedPaidTokenCostNanoUSD != 402_369_000 {
		t.Fatalf("pipe estimated cost = %+v", pipeEvaluation.EstimatedPaidTokenCostNanoUSD)
	}
	if pipeEvaluation.EstimatedPaidTokenCostUSD == nil || *pipeEvaluation.EstimatedPaidTokenCostUSD != 0.402369 {
		t.Fatalf("pipe estimated USD = %+v", pipeEvaluation.EstimatedPaidTokenCostUSD)
	}
	if pipeEvaluation.InputOutputTokens == nil || *pipeEvaluation.InputOutputTokens != 157_217 {
		t.Fatalf("pipe input + output = %+v", pipeEvaluation.InputOutputTokens)
	}
}

func TestOperationalSelectionUsesTokenCountOnlyWhenEstimatedCostTies(t *testing.T) {
	tests := []struct {
		name     string
		pipeCost CostVector
		jsonCost CostVector
		winner   string
	}{
		{
			name:     "paid cost precedes raw token count",
			pipeCost: operationalTestCost(0, 100, 0, 10),
			jsonCost: operationalTestCost(11, 0, 0, 10),
			winner:   CandidatePipe,
		},
		{
			name:     "raw token count breaks exact paid cost tie",
			pipeCost: operationalTestCost(0, 10, 0, 10),
			jsonCost: operationalTestCost(1, 0, 0, 10),
			winner:   CandidateJSON,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			summaries := []CandidateSummary{
				perfectOperationalSummary(CandidatePipe, test.pipeCost),
				perfectOperationalSummary(CandidateJSON, test.jsonCost),
			}
			selection := selectOperationalWinner(summaries, CoverageReport{Complete: true}, operationalPricingPointer())
			if selection.Winner == nil || *selection.Winner != test.winner {
				t.Fatalf("selection = %+v", selection)
			}
		})
	}
}

func TestOperationalSelectionAppliesCorrectnessBeforeCost(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*CandidateSummary)
	}{
		{
			name: "semantic correctness",
			mutate: func(summary *CandidateSummary) {
				summary.ConsumptionSemanticSuccesses = 4
				summary.ConsumptionSemanticFailures = 1
			},
		},
		{
			name: "downstream correctness",
			mutate: func(summary *CandidateSummary) {
				summary.DownstreamTaskSuccesses = 4
				summary.DownstreamTaskFailures = 1
			},
		},
		{
			name: "transfer correctness",
			mutate: func(summary *CandidateSummary) {
				summary.TransferSemanticSuccesses = 4
				summary.TransferSemanticFailures = 1
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cheaperButIncorrect := perfectOperationalSummary(CandidatePipe, operationalTestCost(0, 0, 0, 10))
			test.mutate(&cheaperButIncorrect)
			correct := perfectOperationalSummary(CandidateJSON, operationalTestCost(10, 0, 0, 10))
			selection := selectOperationalWinner([]CandidateSummary{cheaperButIncorrect, correct}, CoverageReport{Complete: true}, operationalPricingPointer())
			if selection.Winner == nil || *selection.Winner != CandidateJSON {
				t.Fatalf("selection = %+v", selection)
			}
			if !reflect.DeepEqual(selection.Rejected, []string{CandidatePipe}) {
				t.Fatalf("rejected = %v", selection.Rejected)
			}
		})
	}
}

func TestOperationalSelectionDoesNotInventCostOrFinalTieBreaker(t *testing.T) {
	pipe := perfectOperationalSummary(CandidatePipe, operationalTestCost(1, 2, 3, 10))
	json := perfectOperationalSummary(CandidateJSON, operationalTestCost(1, 2, 3, 10))
	selection := selectOperationalWinner([]CandidateSummary{pipe, json}, CoverageReport{Complete: true}, operationalPricingPointer())
	if selection.Winner != nil || selection.Status != "no_winner" {
		t.Fatalf("exact tie selection = %+v", selection)
	}

	pipe.Cost.UncachedInputTokens = Metric{Available: false, Reason: "unreported"}
	selection = selectOperationalWinner([]CandidateSummary{pipe, json}, CoverageReport{Complete: true}, operationalPricingPointer())
	if selection.Winner != nil || selection.Status != "no_winner" {
		t.Fatalf("unavailable cost selection = %+v", selection)
	}
	pipeEvaluation := operationalEvaluation(t, selection, CandidatePipe)
	if pipeEvaluation.EstimatedPaidTokenCostUSD != nil {
		t.Fatalf("unavailable cost became zero: %+v", pipeEvaluation)
	}
}

func TestOperationalCostPricesCacheWriteWithoutDoubleCounting(t *testing.T) {
	input := int64(10)
	cached := int64(2)
	cacheWrite := int64(3)
	output := int64(0)
	reasoning := int64(0)
	usage := usageAccumulator{Calls: 1}
	usage.add(CodexUsage{
		InputTokens:           &input,
		CachedInputTokens:     &cached,
		CacheWriteInputTokens: &cacheWrite,
		OutputTokens:          &output,
		ReasoningOutputTokens: &reasoning,
	})
	cost := usage.report()
	if !cost.UncachedInputTokens.Available || cost.UncachedInputTokens.Value == nil || *cost.UncachedInputTokens.Value != 5 {
		t.Fatalf("uncached decomposition = %+v", cost.UncachedInputTokens)
	}
	nanoUSD, inputOutput, ok := estimatedPaidTokenCost(cost, *operationalPricingPointer())
	if !ok || nanoUSD != 44_750 || inputOutput != 10 {
		t.Fatalf("estimated cost = nanoUSD:%d input+output:%d ok:%v", nanoUSD, inputOutput, ok)
	}
	inconsistent := cost
	inconsistent.InputTokens = availableMetric(11, 1)
	if _, _, ok := estimatedPaidTokenCost(inconsistent, *operationalPricingPointer()); ok {
		t.Fatal("inconsistent input decomposition was priced")
	}

	pipeCost := operationalTestCost(0, 0, 0, 1)
	pipeCost.InputTokens = availableMetric(1, 1)
	pipeCost.CacheWriteInputTokens = availableMetric(1, 1)
	jsonCost := operationalTestCost(1, 0, 0, 1)
	selection := selectOperationalWinner([]CandidateSummary{
		perfectOperationalSummary(CandidatePipe, pipeCost),
		perfectOperationalSummary(CandidateJSON, jsonCost),
	}, CoverageReport{Complete: true}, operationalPricingPointer())
	if selection.Winner == nil || *selection.Winner != CandidateJSON {
		t.Fatalf("cache-write pricing selection = %+v", selection)
	}
}

func TestOperationalSelectionBlocksWhenStandardRateDoesNotApply(t *testing.T) {
	pipeCost := operationalTestCost(1, 0, 0, 1)
	pipeCost.MaxInputTokensPerCall = availableMetric(272_001, 1)
	jsonCost := operationalTestCost(1, 0, 0, 1)
	selection := selectOperationalWinner([]CandidateSummary{
		perfectOperationalSummary(CandidatePipe, pipeCost),
		perfectOperationalSummary(CandidateJSON, jsonCost),
	}, CoverageReport{Complete: true}, operationalPricingPointer())
	if selection.Status != "no_winner" || selection.Winner != nil || !strings.Contains(selection.Reason, "lacks a complete estimated paid token cost") {
		t.Fatalf("long-context selection = %+v", selection)
	}
	pipeEvaluation := operationalEvaluation(t, selection, CandidatePipe)
	if pipeEvaluation.EstimatedPaidTokenCostNanoUSD != nil {
		t.Fatalf("standard rates applied above their input limit: %+v", pipeEvaluation)
	}
}

func TestInvalidPerCallUsageCannotBeHiddenByAggregation(t *testing.T) {
	addUsage := func(usage *usageAccumulator, input, cached, cacheWrite int64) {
		zero := int64(0)
		usage.add(CodexUsage{
			InputTokens:           &input,
			CachedInputTokens:     &cached,
			CacheWriteInputTokens: &cacheWrite,
			OutputTokens:          &zero,
			ReasoningOutputTokens: &zero,
		})
	}

	invalidDecomposition := usageAccumulator{Calls: 2}
	addUsage(&invalidDecomposition, 5, 10, 0)
	addUsage(&invalidDecomposition, 15, 0, 0)
	cost := invalidDecomposition.report()
	if cost.InputTokens.Available || !strings.Contains(cost.InputTokens.Reason, "greater than input_tokens") {
		t.Fatalf("invalid per-call decomposition was aggregated: %+v", cost.InputTokens)
	}
	if _, _, ok := estimatedPaidTokenCost(cost, *operationalPricingPointer()); ok {
		t.Fatal("invalid per-call decomposition was priced")
	}

	negativeCancellation := usageAccumulator{Calls: 2}
	addUsage(&negativeCancellation, -1, 0, 0)
	addUsage(&negativeCancellation, 2, 0, 0)
	if metric := negativeCancellation.report().InputTokens; metric.Available || !strings.Contains(metric.Reason, "negative") {
		t.Fatalf("negative counter was hidden by aggregation: %+v", metric)
	}

	left := usageAccumulator{Calls: 1}
	right := usageAccumulator{Calls: 1}
	addUsage(&left, maxOperationalCostValue, 0, 0)
	addUsage(&right, 1, 0, 0)
	left.merge(right)
	if metric := left.report().InputTokens; metric.Available || !strings.Contains(metric.Reason, "overflow") {
		t.Fatalf("merge overflow was not blocked: %+v", metric)
	}
}

func testFixtures(t *testing.T) []Fixture {
	t.Helper()
	file, err := os.Open("../../phase2/fixtures/cases.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	fixtures, err := LoadFixtures(file)
	if err != nil {
		t.Fatal(err)
	}
	return fixtures
}

func perfectObservations(t *testing.T, fixtures []Fixture) []Observation {
	t.Helper()
	input := int64(3)
	value := int64(1)
	usage := CodexUsage{InputTokens: &input, CachedInputTokens: &value, CacheWriteInputTokens: &value, OutputTokens: &value, ReasoningOutputTokens: &value}
	var observations []Observation
	for _, fixture := range fixtures {
		consumer, err := EncodeConsumer(fixture.ExpectedRecords, fixture.TaskOracle)
		if err != nil {
			t.Fatal(err)
		}
		for _, candidate := range Candidates {
			observation := Observation{CaseID: fixture.ID, Candidate: candidate, CoveragePass: 1}
			if candidate == CandidateDirect {
				observation.Producer = StageObservation{Mode: "not_applicable"}
				observation.Handoff = fixture.Source
			} else {
				producer, err := EncodeCandidate(candidate, fixture.ExpectedRecords)
				if err != nil {
					t.Fatal(err)
				}
				observation.Producer = StageObservation{Mode: "model", Attempts: []Attempt{{Attempt: 1, CallExecuted: true, RuntimeOK: true, Response: producer, Usage: usage}}}
				observation.Handoff = producer
			}
			observation.Consumer = StageObservation{Mode: "model", Attempts: []Attempt{{Attempt: 1, CallExecuted: true, RuntimeOK: true, Response: consumer, Usage: usage}}}
			observations = append(observations, observation)
		}
	}
	return observations
}

func setOutputs(t *testing.T, observations []Observation, fixture Fixture, candidate string, records []Record, task map[string]any) {
	t.Helper()
	producer, err := EncodeCandidate(candidate, records)
	if err != nil {
		t.Fatal(err)
	}
	consumer, err := EncodeConsumer(records, task)
	if err != nil {
		t.Fatal(err)
	}
	for i := range observations {
		if observations[i].CaseID == fixture.ID && observations[i].Candidate == candidate {
			observations[i].Producer.Attempts[0].Response = producer
			observations[i].Handoff = producer
			observations[i].Consumer.Attempts[0].Response = consumer
			return
		}
	}
	t.Fatal("observation not found")
}

func candidateSummary(t *testing.T, report Report, candidate string) CandidateSummary {
	t.Helper()
	for _, summary := range report.Candidates {
		if summary.Candidate == candidate {
			return summary
		}
	}
	t.Fatal("summary not found")
	return CandidateSummary{}
}
func trialNamed(t *testing.T, report Report, caseID, candidate string) TrialScore {
	t.Helper()
	for _, trial := range report.Trials {
		if trial.CaseID == caseID && trial.Candidate == candidate {
			return trial
		}
	}
	t.Fatal("trial not found")
	return TrialScore{}
}
func fixtureNamed(t *testing.T, fixtures []Fixture, id string) Fixture {
	t.Helper()
	for _, fixture := range fixtures {
		if fixture.ID == id {
			return fixture
		}
	}
	t.Fatal("fixture not found")
	return Fixture{}
}
func cloneRecords(records []Record) []Record {
	cloned := append([]Record(nil), records...)
	for i := range cloned {
		if cloned[i].Evidence != nil {
			cloned[i].Evidence = stringPointer(*cloned[i].Evidence)
		}
		if cloned[i].Bounds != nil {
			cloned[i].Bounds = stringPointer(*cloned[i].Bounds)
		}
	}
	return cloned
}
func cloneMap(source map[string]any) map[string]any {
	cloned := map[string]any{}
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func currentOperationalSummary(candidate string, hardFailures, semanticSuccesses int, uncached, cached, output, reasoning int64, calls int) CandidateSummary {
	summary := perfectOperationalSummary(candidate, operationalTestCost(uncached, cached, output, calls))
	summary.HardFailures = hardFailures
	summary.ConsumptionSemanticSuccesses = semanticSuccesses
	summary.ConsumptionSemanticFailures = summary.ConsumptionEligibleCells - semanticSuccesses
	summary.TransferSemanticSuccesses = semanticSuccesses
	summary.TransferSemanticFailures = summary.ObservedCells - semanticSuccesses
	reasoningMetric := availableMetric(reasoning, calls)
	summary.Cost.ReasoningOutputTokens = reasoningMetric
	return summary
}

func perfectOperationalSummary(candidate string, cost CostVector) CandidateSummary {
	return CandidateSummary{
		Candidate:                    candidate,
		ExpectedCells:                5,
		ObservedCells:                5,
		ConsumptionEligibleCells:     5,
		ConsumptionSemanticSuccesses: 5,
		TransferSemanticSuccesses:    5,
		DownstreamTaskSuccesses:      5,
		Cost:                         cost,
	}
}

func operationalTestCost(uncached, cached, output int64, calls int) CostVector {
	input := uncached + cached
	zero := int64(0)
	return CostVector{
		CLIInvocations:        calls,
		InputTokens:           availableMetric(input, calls),
		MaxInputTokensPerCall: availableMetric(input, calls),
		UncachedInputTokens:   availableMetric(uncached, calls),
		CachedInputTokens:     availableMetric(cached, calls),
		CacheWriteInputTokens: availableMetric(zero, calls),
		OutputTokens:          availableMetric(output, calls),
		ReasoningOutputTokens: availableMetric(zero, calls),
	}
}

func availableMetric(value int64, calls int) Metric {
	return Metric{Available: true, Value: &value, ObservedTotal: value, ReportedCalls: calls}
}

func operationalPricingPointer() *OperationalPricing {
	return &OperationalPricing{
		PolicyVersion:                     "gpt-5.6-sol-2026-08-11",
		Model:                             "gpt-5.6-sol",
		VerifiedOn:                        "2026-08-11",
		Source:                            "https://developers.openai.com/api/docs/models/gpt-5.6-sol",
		UncachedInputNanoUSDPerToken:      5_000,
		CachedInputNanoUSDPerToken:        500,
		CacheWriteInputNanoUSDPerToken:    6_250,
		OutputNanoUSDPerToken:             30_000,
		StandardRateMaxInputTokensPerCall: 272_000,
		ReasoningOutputCostTreatment:      "reasoning output is already included in output",
	}
}

func operationalEvaluation(t *testing.T, selection OperationalSelection, candidate string) OperationalEvaluation {
	t.Helper()
	for _, evaluation := range selection.Evaluations {
		if evaluation.Candidate == candidate {
			return evaluation
		}
	}
	t.Fatalf("operational evaluation missing for %s", candidate)
	return OperationalEvaluation{}
}
