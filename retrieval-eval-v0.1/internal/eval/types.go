package eval

type Fixture struct {
	ID         string   `json:"id"`
	Query      string   `json:"query"`
	Answerable bool     `json:"answerable"`
	GoldPaths  []string `json:"gold_paths"`
	K          int      `json:"k,omitempty"`
}

type Candidate struct {
	Path     string  `json:"path"`
	Start    int     `json:"start_line,omitempty"`
	End      int     `json:"end_line,omitempty"`
	Score    float64 `json:"score,omitempty"`
	Revision string  `json:"revision,omitempty"`
	Stale    bool    `json:"stale,omitempty"`
}

type Observation struct {
	CaseID            string      `json:"case_id"`
	System            string      `json:"system"`
	Available         bool        `json:"available"`
	Revision          string      `json:"revision,omitempty"`
	Stale             bool        `json:"stale,omitempty"`
	Candidates        []Candidate `json:"candidates"`
	ClaimedAbsent     bool        `json:"claimed_absent,omitempty"`
	LatencyMS         int64       `json:"latency_ms,omitempty"`
	CandidateBytes    int64       `json:"candidate_bytes,omitempty"`
	ToolCalls         int64       `json:"tool_calls,omitempty"`
	InputTokens       *int64      `json:"input_tokens,omitempty"`
	OutputTokens      *int64      `json:"output_tokens,omitempty"`
	ReasoningTokens   *int64      `json:"reasoning_tokens,omitempty"`
	DownstreamCorrect *bool       `json:"downstream_correct,omitempty"`
}

type Trial struct {
	CaseID              string   `json:"case_id"`
	System              string   `json:"system"`
	Available           bool     `json:"available"`
	RecallAtK           float64  `json:"recall_at_k"`
	ReciprocalRank      float64  `json:"reciprocal_rank"`
	UnsupportedAbsence  bool     `json:"unsupported_absence"`
	StaleFailure        bool     `json:"stale_failure"`
	DownstreamEvaluated bool     `json:"downstream_evaluated"`
	DownstreamCorrect   bool     `json:"downstream_correct"`
	LatencyMS           int64    `json:"latency_ms"`
	CandidateBytes      int64    `json:"candidate_bytes"`
	ToolCalls           int64    `json:"tool_calls"`
	InputTokens         *int64   `json:"input_tokens,omitempty"`
	OutputTokens        *int64   `json:"output_tokens,omitempty"`
	ReasoningTokens     *int64   `json:"reasoning_tokens,omitempty"`
	Errors              []string `json:"errors,omitempty"`
}

type Summary struct {
	System                   string  `json:"system"`
	Expected                 int     `json:"expected"`
	Observed                 int     `json:"observed"`
	Unavailable              int     `json:"unavailable"`
	StaleFailures            int     `json:"stale_failures"`
	UnsupportedAbsenceClaims int     `json:"unsupported_absence_claims"`
	MeanRecallAtK            float64 `json:"mean_recall_at_k"`
	MeanReciprocalRank       float64 `json:"mean_reciprocal_rank"`
	DownstreamEvaluated      int     `json:"downstream_evaluated"`
	DownstreamCorrect        int     `json:"downstream_correct"`
	LatencyMSTotal           int64   `json:"latency_ms_total"`
	CandidateBytesTotal      int64   `json:"candidate_bytes_total"`
	ToolCallsTotal           int64   `json:"tool_calls_total"`
	InputTokensTotal         *int64  `json:"input_tokens_total,omitempty"`
	OutputTokensTotal        *int64  `json:"output_tokens_total,omitempty"`
	ReasoningTokensTotal     *int64  `json:"reasoning_tokens_total,omitempty"`
	inputAccumulator         metricAccumulator
	outputAccumulator        metricAccumulator
	reasoningAccumulator     metricAccumulator
}

type metricAccumulator struct {
	Sum         int64
	Seen        bool
	Unavailable bool
}

type Coverage struct {
	Expected   int      `json:"expected"`
	Observed   int      `json:"observed"`
	Complete   bool     `json:"complete"`
	Missing    []string `json:"missing,omitempty"`
	Duplicates []string `json:"duplicates,omitempty"`
	Unexpected []string `json:"unexpected,omitempty"`
}

type Report struct {
	Coverage  Coverage  `json:"coverage"`
	Trials    []Trial   `json:"trials"`
	Summaries []Summary `json:"summaries"`
	Errors    []string  `json:"errors,omitempty"`
}
