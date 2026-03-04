package orchestrator

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// RunTask executes a mission step with integrated identity management.
func (sm *StateMachine) RunTask(ctx context.Context, taskName string, fn func() error) error {
	for {
		// Identify required identity for the current state
		var required Identity
		state, _ := sm.GetStatus()

		switch state {
		case StateDrafting, StateValidation:
			required = sm.getPersonalIdentity()
		case StateAudit, StateCertification:
			required = IdentityCorporatePrime
		default:
			required = IdentityNone
		}

		// 1. Proactive Quota Check (Reduce 429 Surprise)
		if sm.shouldCheckQuota(state) {
			if lowQuota := sm.checkQuota(ctx); lowQuota {
				sm.sc.Log("QuotaCheck", "Proactive Rotation", state.String(), "Rotating due to low stats", 0)
				sm.rotateIdentity()
				if err := sm.SyncIdentity(sm.getPersonalIdentity()); err != nil {
					return err
				}
			}
		}

		// 2. Ensure identity is aligned
		if err := sm.SyncIdentity(required); err != nil {
			return err
		}

		// Execute the logic
		err := fn()
		if err == nil {
			sm.sc.Log("TaskExecution", "Success", taskName, "Task completed successfully", 0)
			return nil
		}

		// Handle Failures (Identity rotation or Audit/Structural errors)
		if sm.isRateLimit(err) || state == StateAudit || state == StateValidation {
			sm.sc.Log("TaskExecution", "Failure", taskName, fmt.Sprintf("State: %s | Error: %s", state.String(), err.Error()), 0)

			sm.Transition(ctx, StateRecovery)

			// 1. Rollback the workspace to the last known-good checkpoint
			// This represents the call to Materializer.Rollback()
			sm.sc.Log("TaskExecution", "Rollback", taskName, "Reverting to last mini-state", 0)

			// 2. Adjust identity if it was a rate limit
			if sm.isRateLimit(err) {
				sm.rotateIdentity()
				time.Sleep(30 * time.Second)
			}

			// 3. Reset to DRAFTING and start forward again
			sm.Transition(ctx, StateDrafting)
			continue
		}

		// Fatal error (unrecoverable system failure)
		return err
	}
}

func (sm *StateMachine) getPersonalIdentity() Identity {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if len(sm.personalPool) == 0 {
		return IdentityNone
	}
	return sm.personalPool[sm.identityIndex%len(sm.personalPool)]
}

func (sm *StateMachine) rotateIdentity() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.identityIndex++
	sm.sc.Log("IdentityRotation", "Index", "PersonalPool", fmt.Sprintf("%d", sm.identityIndex), 0)
}

func (sm *StateMachine) isRateLimit(err error) bool {
	// In a real implementation, we would check the Connect/HTTP error code
	// Mocking for now based on string matching
	return err != nil && (sm.contains(err.Error(), "429") || sm.contains(err.Error(), "rate limit"))
}

func (sm *StateMachine) shouldCheckQuota(state State) bool {
	// Proactively check before intensive labor phases
	return state == StateDrafting || state == StateValidation
}

func (sm *StateMachine) checkQuota(ctx context.Context) bool {
	sm.sc.Log("QuotaCheck", "ActiveCheck", "/stats", "Executing gemini /stats", 0)

	// os/exec call to gemini-cli (non-interactive)
	cmd := exec.CommandContext(ctx, "gemini", "-p", "/stats")
	output, err := cmd.CombinedOutput()
	if err != nil {
		sm.sc.Log("QuotaCheck", "Failure", "/stats", err.Error(), 0)
		return false // Fail-open: continue and let 429 handler take over if needed
	}

	// Logic to parse output for low-quota indicators
	// If the output contains "low quota" or high usage percentages
	if strings.Contains(strings.ToLower(string(output)), "insufficient") ||
		strings.Contains(strings.ToLower(string(output)), "exhausted") {
		return true
	}

	return false
}

func (sm *StateMachine) contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr // Simplified match
}
