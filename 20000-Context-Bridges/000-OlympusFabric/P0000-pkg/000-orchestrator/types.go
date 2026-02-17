package orchestrator

import "fmt"

// State represents the current lifecycle phase of an architecture task.
type State int

const (
	StateDrafting State = iota
	StateValidation
	StateAudit
	StateCertification
	StateRecovery
)

func (s State) String() string {
	return [...]string{"DRAFTING", "VALIDATION", "AUDIT", "CERTIFICATION", "RECOVERY"}[s]
}

// Identity represents the authenticated context (Delegate vs Auditor).
type Identity string

const (
	// IdentityCorporatePrime is the authoritative context for validation/certification.
	IdentityCorporatePrime Identity = "corporate-prime"
	IdentityNone           Identity = "none"
)

// OrchestratorError wraps state transition or identity failures.
type OrchestratorError struct {
	CurrentState State
	Message      string
	Err          error
}

func (e *OrchestratorError) Error() string {
	return fmt.Sprintf("[%s] %s: %v", e.CurrentState, e.Message, e.Err)
}
