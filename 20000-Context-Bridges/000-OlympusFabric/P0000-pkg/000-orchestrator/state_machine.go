package orchestrator

import (
	"context"
	"fmt"
	"sync"

	"github.com/VelociKey/Olympus2/pkg/whisper"
)

// StateMachine manages the Sovereign Node advancement lifecycle.
type StateMachine struct {
	mu            sync.RWMutex
	currentState  State
	identity      Identity
	personalPool  []Identity // Pool of personal alter egos for delegate labor
	identityIndex int        // Index for rotating through the personal pool
	sc            *whisper.WhisperLog
}

// NewStateMachine initializes a new orchestrator with a pool of personal alter egos.
func NewStateMachine(personalEgos []string, sc *whisper.WhisperLog) *StateMachine {
	if sc == nil {
		sc = whisper.New("Fabric", "fabric.lpsv")
	}

	pool := make([]Identity, len(personalEgos))
	for i, ego := range personalEgos {
		pool[i] = Identity(ego)
	}

	return &StateMachine{
		currentState: StateDrafting,
		personalPool: pool,
		identity:     IdentityNone,
		sc:           sc,
	}
}

// GetStatus returns the current machine state and identity context.
func (sm *StateMachine) GetStatus() (State, Identity) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentState, sm.identity
}

// Transition advances the state machine if valid.
func (sm *StateMachine) Transition(ctx context.Context, next State) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.sc.Log("StateTransition", "Attempt", sm.currentState.String(), next.String())

	if !sm.isValidTransition(sm.currentState, next) {
		return &OrchestratorError{
			CurrentState: sm.currentState,
			Message:      fmt.Sprintf("Invalid transition to %s", next),
		}
	}

	sm.currentState = next
	return nil
}

func (sm *StateMachine) isValidTransition(current, next State) bool {
	if next == StateRecovery {
		return true // Recovery is always accessible from any error state
	}

	switch current {
	case StateDrafting:
		return next == StateValidation
	case StateValidation:
		return next == StateAudit
	case StateAudit:
		return next == StateCertification
	case StateCertification:
		return false // Terminal state
	case StateRecovery:
		return next == StateDrafting
	default:
		return false
	}
}
