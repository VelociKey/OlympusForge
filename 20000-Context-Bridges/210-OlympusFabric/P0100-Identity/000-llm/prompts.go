// Filename: prompts.go
// Copyright: © 2026 VelociKey LLC. All Rights Reserved.
// Version: 1.0.0
// Author: Joseph A. White, III
// Status: Approved
// Approver: Joseph A. White, III
// Timestamp: 2026-02-13T14:41:30Z
// License: Proprietary

package llm

import (
	"fmt"
	"strings"
)

// PromptType defines the intent of the LLM request
type PromptType string

const (
	PromptGrammarGeneration PromptType = "grammar"
	PromptProgramGeneration PromptType = "program"
	PromptExplanation       PromptType = "explain"
	PromptEvolution         PromptType = "evolve"
)

// PromptFactory manages templates for different LLM interactions
type PromptFactory struct{}

func NewPromptFactory() *PromptFactory {
	return &PromptFactory{}
}

func (f *PromptFactory) Create(pType PromptType, userPrompt string, context ...string) string {
	ctx := strings.Join(context, "\n")

	switch pType {
	case PromptGrammarGeneration:
		return fmt.Sprintf(`You are a world-class Language Engineer. 
Your task is to design an eBNF grammar based on the following requirements:
---
%s
---
Rules for generation:
1. Use standard ISO 14977 eBNF.
2. Ensure the grammar is clear and unambiguous.
3. Use lowercase for rules and uppercase for terminals where appropriate.
4. Output ONLY the eBNF code block.

Current Context/Existing Rules (if any):
%s`, userPrompt, ctx)

	case PromptProgramGeneration:
		return fmt.Sprintf(`You are a compiler engineer. 
Your task is to write a compact program/specification that conforms to the provided eBNF/jeBNF grammar.
---
Grammar Specification:
%s
---
User Requirement:
%s
---
Output ONLY the program text, no explanations.`, ctx, userPrompt)

	case PromptExplanation:
		return fmt.Sprintf(`Explain the following grammar or DSL program in detail.
Identify the core rules, expected inputs, and any potential ambiguities.
---
Content:
%s
---
User Query:
%s`, ctx, userPrompt)

	default:
		return userPrompt
	}
}
