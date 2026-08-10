package eval

// Case describes one model-facing evaluation prompt and its expected wire message.
type Case struct {
	ID       string `json:"id"`
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

// Usage stores provider-reported usage when available.
type Usage struct {
	InputTokens     int `json:"input_tokens,omitempty"`
	CachedTokens    int `json:"cached_tokens,omitempty"`
	OutputTokens    int `json:"output_tokens,omitempty"`
	ReasoningTokens int `json:"reasoning_tokens,omitempty"`
}

// Response stores one model trial.
type Response struct {
	CaseID    string `json:"case_id"`
	Candidate string `json:"candidate"`
	Trial     int    `json:"trial"`
	Raw       string `json:"raw"`
	Usage     Usage  `json:"usage,omitempty"`
}

// TrialResult contains protocol and semantic checks for one response.
type TrialResult struct {
	CaseID         string `json:"case_id"`
	Candidate      string `json:"candidate"`
	Trial          int    `json:"trial"`
	ParseOK        bool   `json:"parse_ok"`
	CanonicalExact bool   `json:"canonical_exact"`
	SemanticExact  bool   `json:"semantic_exact"`
	RoutingOK      bool   `json:"routing_ok"`
	EpistemicOK    bool   `json:"epistemic_ok"`
	EvidenceOK     bool   `json:"evidence_ok"`
	HardFailure    bool   `json:"hard_failure"`
	Error          string `json:"error,omitempty"`
	Usage          Usage  `json:"usage,omitempty"`
}

// CandidateSummary aggregates trials for one candidate format or prompt arm.
type CandidateSummary struct {
	Candidate         string `json:"candidate"`
	Trials            int    `json:"trials"`
	ParseFailures     int    `json:"parse_failures"`
	CanonicalExact    int    `json:"canonical_exact"`
	SemanticExact     int    `json:"semantic_exact"`
	RoutingFailures   int    `json:"routing_failures"`
	EpistemicFailures int    `json:"epistemic_failures"`
	EvidenceFailures  int    `json:"evidence_failures"`
	HardFailures      int    `json:"hard_failures"`
	InputTokens       int    `json:"input_tokens"`
	CachedTokens      int    `json:"cached_tokens"`
	OutputTokens      int    `json:"output_tokens"`
	ReasoningTokens   int    `json:"reasoning_tokens"`
}
