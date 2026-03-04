package orchestrator

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const ConfigEnv = "CLOUDSDK_ACTIVE_CONFIG_NAME"

// SyncIdentity aligns the OS gcloud configuration with the required state.
func (sm *StateMachine) SyncIdentity(required Identity) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.sc.Log("IdentitySync", "Context", string(required), "Aligning OS config", 0)

	// Set OS environment variable
	if err := os.Setenv(ConfigEnv, string(required)); err != nil {
		return fmt.Errorf("failed to set identity env: %w", err)
	}

	// Verify via gcloud CLI (os/exec only, no shell)
	cmd := exec.Command("gcloud", "config", "configurations", "list", "--filter", "IS_ACTIVE=true", "--format", "value(name)")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to verify gcloud identity: %w", err)
	}

	active := strings.TrimSpace(string(output))
	if active != string(required) {
		return fmt.Errorf("identity mismatch: OS reports %q, but required %q", active, required)
	}

	sm.identity = required
	sm.sc.Log("IdentitySync", "Synchronized", active, "OS reports required identity", 0)
	return nil
}

// CurrentIdentity returns the identity from the environment.
func CurrentIdentity() Identity {
	val := os.Getenv(ConfigEnv)
	if val == "" {
		return IdentityNone
	}
	return Identity(val)
}
