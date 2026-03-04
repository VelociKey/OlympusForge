package parser

import (
	"fmt"
	"strings"
)

// AuditGate performs deep-DNA scanning of generated artifacts.
// It ensures that code produced in the DRAFTING state matches Sovereign identity standards.
type AuditGate struct {
	CorporateBaseline string
}

// ScanIdentityDNA verifies that a file's metadata matches the required Sovereign identity.
func (a *AuditGate) ScanIdentityDNA(content string, expectedIdentity string) error {
	// Pattern check for Sovereign Identity blocks
	// Example: // Identity: @personal

	if !strings.Contains(content, fmt.Sprintf("Identity: %s", expectedIdentity)) {
		return fmt.Errorf("Identity DNA mismatch: expected %q but finding un-attested logic", expectedIdentity)
	}

	// Security Check: Look for hardcoded secrets or PII leaks
	if strings.Contains(content, "PRIVATE KEY") || strings.Contains(content, "sk_live") {
		return fmt.Errorf("security breach: artifacts contain un-encrypted secrets")
	}

	return nil
}

// EnforceNumericPrefix ensures the WSG "numeric prefixing" rule is followed.
func (a *AuditGate) EnforceNumericPrefix(filename string) error {
	// Pattern: 00000-DirectoryName
	if len(filename) < 6 || filename[5] != '-' {
		return fmt.Errorf("WSG violation: filename %q must follow numeric prefixing (prefix '-' name)", filename)
	}
	return nil
}
