package protocol

import "fmt"

// Kind identifies the wire record and protocol version.
type Kind string

const (
	KindResult   Kind = "R1"
	KindQuestion Kind = "Q1"
	KindDecision Kind = "D1"
)

// Origin identifies how a result was produced.
type Origin byte

const (
	OriginObserved  Origin = 'O'
	OriginInferred  Origin = 'I'
	OriginSynthetic Origin = 'S'
)

// State identifies the current epistemic status of a result.
type State byte

const (
	StateUnverified State = 'U'
	StateVerified   State = 'V'
	StateRefuted    State = 'R'
)

// Epistemic keeps provenance and verification state separate.
type Epistemic struct {
	Origin Origin
	State  State
}

func (e Epistemic) String() string {
	return string([]byte{byte(e.Origin), byte(e.State)})
}

func ParseEpistemic(raw string) (Epistemic, error) {
	if len(raw) != 2 {
		return Epistemic{}, fmt.Errorf("epistemic value must contain exactly two ASCII characters")
	}

	e := Epistemic{Origin: Origin(raw[0]), State: State(raw[1])}
	if !e.Origin.Valid() {
		return Epistemic{}, fmt.Errorf("unsupported epistemic origin %q", raw[0])
	}
	if !e.State.Valid() {
		return Epistemic{}, fmt.Errorf("unsupported epistemic state %q", raw[1])
	}
	return e, nil
}

func (o Origin) Valid() bool {
	switch o {
	case OriginObserved, OriginInferred, OriginSynthetic:
		return true
	default:
		return false
	}
}

func (s State) Valid() bool {
	switch s {
	case StateUnverified, StateVerified, StateRefuted:
		return true
	default:
		return false
	}
}

// Record is the normalized representation of R1, Q1, or D1.
// Fields unused by a record kind remain empty.
type Record struct {
	Kind           Kind      `json:"kind"`
	Target         string    `json:"target"`
	Epistemic      Epistemic `json:"epistemic,omitempty"`
	Content        string    `json:"content"`
	Recommendation string    `json:"recommendation,omitempty"`
	Evidence       string    `json:"evidence,omitempty"`
	Reference      string    `json:"reference,omitempty"`
	Bounds         string    `json:"bounds,omitempty"`
}

func (r Record) Validate() error {
	if r.Target == "" {
		return fmt.Errorf("target is required")
	}
	if r.Content == "" {
		return fmt.Errorf("content is required")
	}

	switch r.Kind {
	case KindResult:
		if !r.Epistemic.Origin.Valid() || !r.Epistemic.State.Valid() {
			return fmt.Errorf("valid epistemic origin and state are required for R1")
		}
		if (r.Epistemic.State == StateVerified || r.Epistemic.State == StateRefuted) && r.Evidence == "" {
			return fmt.Errorf("verified or refuted R1 requires evidence")
		}
		if r.Recommendation != "" || r.Reference != "" {
			return fmt.Errorf("R1 contains fields reserved for another record kind")
		}
	case KindQuestion:
		if r.Recommendation == "" {
			return fmt.Errorf("Q1 requires a best-supported recommendation")
		}
		if r.Bounds == "" {
			return fmt.Errorf("Q1 requires a blocking condition or human-only decision boundary")
		}
		if r.Epistemic.Origin != 0 || r.Epistemic.State != 0 || r.Reference != "" {
			return fmt.Errorf("Q1 contains fields reserved for another record kind")
		}
	case KindDecision:
		if r.Reference == "" {
			return fmt.Errorf("D1 requires an authoritative decision reference")
		}
		if r.Epistemic.Origin != 0 || r.Epistemic.State != 0 || r.Recommendation != "" || r.Evidence != "" {
			return fmt.Errorf("D1 contains fields reserved for another record kind")
		}
	default:
		return fmt.Errorf("unsupported record kind %q", r.Kind)
	}

	return nil
}
