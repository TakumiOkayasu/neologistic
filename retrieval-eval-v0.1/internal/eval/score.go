package eval

import (
	"fmt"
	"sort"
)

func Score(fixtures []Fixture, observations []Observation) Report {
	systems := map[string]struct{}{}
	for _, observation := range observations {
		systems[observation.System] = struct{}{}
	}
	var names []string
	for system := range systems {
		names = append(names, system)
	}
	return ScoreWithSystems(fixtures, observations, names)
}

func ScoreWithSystems(fixtures []Fixture, observations []Observation, expectedSystems []string) Report {
	report := Report{}
	fixtureByID := make(map[string]Fixture, len(fixtures))
	for _, fixture := range fixtures {
		if fixture.K <= 0 {
			fixture.K = 5
		}
		fixtureByID[fixture.ID] = fixture
	}
	systems := map[string]struct{}{}
	for _, system := range expectedSystems {
		if system != "" {
			systems[system] = struct{}{}
		}
	}
	expected := map[string]struct{}{}
	for system := range systems {
		for _, fixture := range fixtures {
			expected[key(fixture.ID, system)] = struct{}{}
		}
	}
	seen := map[string]int{}
	valid := make([]Observation, 0, len(observations))
	for _, observation := range observations {
		cell := key(observation.CaseID, observation.System)
		seen[cell]++
		if _, ok := expected[cell]; !ok {
			report.Coverage.Unexpected = append(report.Coverage.Unexpected, cell)
			continue
		}
		if seen[cell] == 1 {
			valid = append(valid, observation)
		}
	}
	for cell := range expected {
		if seen[cell] == 0 {
			report.Coverage.Missing = append(report.Coverage.Missing, cell)
		}
		if seen[cell] > 1 {
			report.Coverage.Duplicates = append(report.Coverage.Duplicates, cell)
		}
	}
	sort.Strings(report.Coverage.Missing)
	sort.Strings(report.Coverage.Duplicates)
	sort.Strings(report.Coverage.Unexpected)
	report.Coverage.Expected = len(expected)
	report.Coverage.Observed = len(expected) - len(report.Coverage.Missing)
	report.Coverage.Complete = len(report.Coverage.Missing) == 0 && len(report.Coverage.Duplicates) == 0 && len(report.Coverage.Unexpected) == 0

	summaryBySystem := map[string]*Summary{}
	for system := range systems {
		summaryBySystem[system] = &Summary{System: system, Expected: len(fixtures)}
	}
	for _, observation := range valid {
		fixture := fixtureByID[observation.CaseID]
		trial := scoreTrial(fixture, observation)
		report.Trials = append(report.Trials, trial)
		addTrial(summaryBySystem[observation.System], trial)
	}
	sort.Slice(report.Trials, func(i, j int) bool {
		if report.Trials[i].CaseID == report.Trials[j].CaseID {
			return report.Trials[i].System < report.Trials[j].System
		}
		return report.Trials[i].CaseID < report.Trials[j].CaseID
	})
	var names []string
	for name := range summaryBySystem {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		summary := summaryBySystem[name]
		if summary.Observed > 0 {
			summary.MeanRecallAtK /= float64(summary.Observed)
			summary.MeanReciprocalRank /= float64(summary.Observed)
		}
		summary.InputTokensTotal = summary.inputAccumulator.value()
		summary.OutputTokensTotal = summary.outputAccumulator.value()
		summary.ReasoningTokensTotal = summary.reasoningAccumulator.value()
		report.Summaries = append(report.Summaries, *summary)
	}
	if !report.Coverage.Complete {
		report.Errors = append(report.Errors, "coverage is incomplete, duplicated, or unexpected")
	}
	return report
}

func scoreTrial(fixture Fixture, observation Observation) Trial {
	trial := Trial{
		CaseID: observation.CaseID, System: observation.System, Available: observation.Available,
		LatencyMS: observation.LatencyMS, CandidateBytes: observation.CandidateBytes,
		ToolCalls: observation.ToolCalls, InputTokens: observation.InputTokens,
		OutputTokens: observation.OutputTokens, ReasoningTokens: observation.ReasoningTokens,
	}
	if observation.DownstreamCorrect != nil {
		trial.DownstreamEvaluated = true
		trial.DownstreamCorrect = *observation.DownstreamCorrect
	}
	if !observation.Available {
		trial.Errors = append(trial.Errors, "retrieval system unavailable")
		return trial
	}
	trial.StaleFailure = observation.Stale
	for _, candidate := range observation.Candidates {
		trial.StaleFailure = trial.StaleFailure || candidate.Stale
	}
	k := fixture.K
	if k <= 0 {
		k = 5
	}
	candidates := observation.Candidates
	if len(candidates) > k {
		candidates = candidates[:k]
	}
	gold := make(map[string]struct{}, len(fixture.GoldPaths))
	for _, path := range fixture.GoldPaths {
		gold[path] = struct{}{}
	}
	found := map[string]struct{}{}
	firstRank := 0
	for index, candidate := range candidates {
		if _, ok := gold[candidate.Path]; ok {
			found[candidate.Path] = struct{}{}
			if firstRank == 0 {
				firstRank = index + 1
			}
		}
	}
	if len(gold) > 0 {
		trial.RecallAtK = float64(len(found)) / float64(len(gold))
	}
	if firstRank > 0 {
		trial.ReciprocalRank = 1 / float64(firstRank)
	}
	trial.UnsupportedAbsence = observation.ClaimedAbsent && fixture.Answerable
	if trial.StaleFailure {
		trial.Errors = append(trial.Errors, "stale retrieval result")
	}
	if trial.UnsupportedAbsence {
		trial.Errors = append(trial.Errors, "unsupported absence claim")
	}
	return trial
}

func addTrial(summary *Summary, trial Trial) {
	summary.Observed++
	if !trial.Available {
		summary.Unavailable++
	}
	if trial.StaleFailure {
		summary.StaleFailures++
	}
	if trial.UnsupportedAbsence {
		summary.UnsupportedAbsenceClaims++
	}
	summary.MeanRecallAtK += trial.RecallAtK
	summary.MeanReciprocalRank += trial.ReciprocalRank
	if trial.DownstreamEvaluated {
		summary.DownstreamEvaluated++
		if trial.DownstreamCorrect {
			summary.DownstreamCorrect++
		}
	}
	summary.LatencyMSTotal += trial.LatencyMS
	summary.CandidateBytesTotal += trial.CandidateBytes
	summary.ToolCallsTotal += trial.ToolCalls
	summary.inputAccumulator.add(trial.InputTokens)
	summary.outputAccumulator.add(trial.OutputTokens)
	summary.reasoningAccumulator.add(trial.ReasoningTokens)
}

func (m *metricAccumulator) add(value *int64) {
	m.Seen = true
	if value == nil {
		m.Unavailable = true
		return
	}
	m.Sum += *value
}

func (m metricAccumulator) value() *int64 {
	if !m.Seen || m.Unavailable {
		return nil
	}
	value := m.Sum
	return &value
}

func key(caseID, system string) string {
	return fmt.Sprintf("%s\x00%s", caseID, system)
}
