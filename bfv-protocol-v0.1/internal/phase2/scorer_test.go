package phase2

import (
	"os"
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
	if report.Selection.Winner == nil || *report.Selection.Winner != CandidateDirect {
		t.Fatalf("winner = %+v", report.Selection)
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
	if !report.Coverage.HardError || report.ExitCode() != 2 || report.Selection.Winner != nil {
		t.Fatalf("report = %+v", report)
	}
	if report.Selection.NextDiscriminatingTest == nil {
		t.Fatalf("next test = %+v", report.Selection)
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
	value := int64(10)
	usage := CodexUsage{InputTokens: &value, CachedInputTokens: &value, CacheWriteInputTokens: &value, OutputTokens: &value, ReasoningOutputTokens: &value}
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
	value := int64(1)
	usage := CodexUsage{InputTokens: &value, CachedInputTokens: &value, CacheWriteInputTokens: &value, OutputTokens: &value, ReasoningOutputTokens: &value}
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
