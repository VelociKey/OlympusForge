// Filename: prompts_test.go
// Copyright: © 2026 VelociKey LLC. All Rights Reserved.
// Version: 1.0.0
// Author: Joseph A. White, III
// Status: Draft
// Approver: Joseph A. White, III
// Timestamp: 2026-02-13T14:38:00Z
// License: Proprietary

package llm

import (
	"strings"
	"testing"
)

func TestPromptFactoryCreate(t *testing.T) {
	factory := NewPromptFactory()

	t.Run("Grammar Generation", func(t *testing.T) {
		prompt := factory.Create(PromptGrammarGeneration, "Create a JSON grammar")
		if !strings.Contains(prompt, "Language Engineer") {
			t.Errorf("expected role context, got %q", prompt)
		}
		if !strings.Contains(prompt, "Create a JSON grammar") {
			t.Errorf("expected user input in prompt, got %q", prompt)
		}
	})

	t.Run("Program Generation", func(t *testing.T) {
		prompt := factory.Create(PromptProgramGeneration, "Define a variable x", "INT IDENTIFIER")
		if !strings.Contains(prompt, "compiler engineer") {
			t.Errorf("expected role context, got %q", prompt)
		}
		if !strings.Contains(prompt, "INT IDENTIFIER") {
			t.Errorf("expected context grammar in prompt, got %q", prompt)
		}
		if !strings.Contains(prompt, "Define a variable x") {
			t.Errorf("expected user requirement in prompt, got %q", prompt)
		}
	})

	t.Run("Explanation", func(t *testing.T) {
		prompt := factory.Create(PromptExplanation, "What does INT mean?", "INT = [0..9]+")
		if !strings.Contains(prompt, "Explain the following") {
			t.Errorf("expected explanation context, got %q", prompt)
		}
		if !strings.Contains(prompt, "INT = [0..9]+") {
			t.Errorf("expected content in prompt, got %q", prompt)
		}
		if !strings.Contains(prompt, "What does INT mean?") {
			t.Errorf("expected query in prompt, got %q", prompt)
		}
	})

	t.Run("Default fallback", func(t *testing.T) {
		prompt := factory.Create(PromptEvolution, "Make it faster")
		if prompt != "Make it faster" {
			t.Errorf("expected raw prompt fallback, got %q", prompt)
		}
	})
}
