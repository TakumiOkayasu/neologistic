package eval

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"bfvprotocol/internal/protocol"
)

func ReadCases(r io.Reader) (map[string]Case, error) {
	cases := make(map[string]Case)
	decoder := json.NewDecoder(r)
	index := 0
	for {
		var c Case
		if err := decoder.Decode(&c); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("case %d: %w", index+1, err)
		}
		index++
		if c.ID == "" || c.Expected == "" {
			return nil, fmt.Errorf("case %d: id and expected are required", index)
		}
		if _, exists := cases[c.ID]; exists {
			return nil, fmt.Errorf("case %d: duplicate id %q", index, c.ID)
		}
		if _, err := protocol.ParseMessage(c.Expected); err != nil {
			return nil, fmt.Errorf("case %d: invalid expected message: %w", index, err)
		}
		cases[c.ID] = c
	}
	return cases, nil
}

func ReadResponses(r io.Reader) ([]Response, error) {
	var responses []Response
	decoder := json.NewDecoder(r)
	index := 0
	for {
		var response Response
		if err := decoder.Decode(&response); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("response %d: %w", index+1, err)
		}
		index++
		if response.CaseID == "" || response.Candidate == "" || response.Raw == "" {
			return nil, fmt.Errorf("response %d: case_id, candidate, and raw are required", index)
		}
		responses = append(responses, response)
	}
	return responses, nil
}

func Score(cases map[string]Case, responses []Response) ([]TrialResult, []CandidateSummary, error) {
	results := make([]TrialResult, 0, len(responses))
	summaries := make(map[string]*CandidateSummary)

	for _, response := range responses {
		c, ok := cases[response.CaseID]
		if !ok {
			return nil, nil, fmt.Errorf("unknown case_id %q", response.CaseID)
		}
		result := scoreOne(c, response)
		results = append(results, result)

		summary := summaries[response.Candidate]
		if summary == nil {
			summary = &CandidateSummary{Candidate: response.Candidate}
			summaries[response.Candidate] = summary
		}
		summary.Trials++
		if !result.ParseOK {
			summary.ParseFailures++
		}
		if result.CanonicalExact {
			summary.CanonicalExact++
		}
		if result.SemanticExact {
			summary.SemanticExact++
		}
		if !result.RoutingOK {
			summary.RoutingFailures++
		}
		if !result.EpistemicOK {
			summary.EpistemicFailures++
		}
		if !result.EvidenceOK {
			summary.EvidenceFailures++
		}
		if result.HardFailure {
			summary.HardFailures++
		}
		summary.InputTokens += response.Usage.InputTokens
		summary.CachedTokens += response.Usage.CachedTokens
		summary.OutputTokens += response.Usage.OutputTokens
		summary.ReasoningTokens += response.Usage.ReasoningTokens
	}

	ordered := make([]CandidateSummary, 0, len(summaries))
	for _, summary := range summaries {
		ordered = append(ordered, *summary)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Candidate < ordered[j].Candidate })
	return results, ordered, nil
}

func scoreOne(c Case, response Response) TrialResult {
	result := TrialResult{
		CaseID:    response.CaseID,
		Candidate: response.Candidate,
		Trial:     response.Trial,
		Usage:     response.Usage,
	}

	expected, err := protocol.ParseMessage(c.Expected)
	if err != nil {
		result.Error = "invalid expected fixture: " + err.Error()
		result.HardFailure = true
		return result
	}
	actual, err := protocol.ParseMessage(response.Raw)
	if err != nil {
		result.Error = err.Error()
		result.HardFailure = true
		return result
	}
	result.ParseOK = true

	canonical, err := protocol.EncodeMessage(actual)
	if err == nil {
		result.CanonicalExact = canonical == c.Expected && response.Raw == canonical
	}
	result.SemanticExact = equalRecordMultiset(expected, actual)
	result.RoutingOK = equalRouting(expected, actual)
	result.EpistemicOK = epistemicSafe(expected, actual)
	result.EvidenceOK = evidenceSafe(expected, actual)
	result.HardFailure = !result.SemanticExact || !result.RoutingOK || !result.EpistemicOK || !result.EvidenceOK
	return result
}

func equalRecordMultiset(a, b []protocol.Record) bool {
	if len(a) != len(b) {
		return false
	}
	counts := make(map[string]int, len(a))
	for _, record := range a {
		line, _ := protocol.EncodeLine(record)
		counts[line]++
	}
	for _, record := range b {
		line, _ := protocol.EncodeLine(record)
		counts[line]--
	}
	for _, count := range counts {
		if count != 0 {
			return false
		}
	}
	return true
}

func equalRouting(expected, actual []protocol.Record) bool {
	return kindTargets(expected, protocol.KindQuestion) == kindTargets(actual, protocol.KindQuestion)
}

func kindTargets(records []protocol.Record, kind protocol.Kind) string {
	var values []string
	for _, record := range records {
		if record.Kind == kind {
			values = append(values, record.Target)
		}
	}
	sort.Strings(values)
	encoded, _ := json.Marshal(values)
	return string(encoded)
}

func epistemicSafe(expected, actual []protocol.Record) bool {
	expectedByKey := make(map[string]protocol.Record)
	for _, record := range expected {
		if record.Kind == protocol.KindResult {
			expectedByKey[resultKey(record)] = record
		}
	}
	for _, record := range actual {
		if record.Kind != protocol.KindResult {
			continue
		}
		want, ok := expectedByKey[resultKey(record)]
		if !ok {
			continue
		}
		if record.Epistemic != want.Epistemic {
			return false
		}
	}
	return true
}

func evidenceSafe(expected, actual []protocol.Record) bool {
	expectedByKey := make(map[string]protocol.Record)
	for _, record := range expected {
		if record.Kind == protocol.KindResult {
			expectedByKey[resultKey(record)] = record
		}
	}
	for _, record := range actual {
		if record.Kind != protocol.KindResult {
			continue
		}
		want, ok := expectedByKey[resultKey(record)]
		if !ok {
			continue
		}
		if want.Evidence != "" && record.Evidence != want.Evidence {
			return false
		}
	}
	return true
}

func resultKey(record protocol.Record) string {
	return string(record.Kind) + "\x00" + record.Target + "\x00" + record.Content
}
