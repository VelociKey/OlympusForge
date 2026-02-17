// Filename: client_test.go
// Copyright: © 2026 VelociKey LLC. All Rights Reserved.
// Version: 1.0.0
// Author: Joseph A. White, III
// Status: Draft
// Approver: Joseph A. White, III
// Timestamp: 2026-02-13T14:36:00Z
// License: Proprietary

package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	c := NewClient("", "")
	if c.BaseURL != "http://localhost:11434" {
		t.Errorf("expected default base URL, got %s", c.BaseURL)
	}
	if c.Model != "gemma3:27b" {
		t.Errorf("expected default model, got %s", c.Model)
	}

	c2 := NewClient("http://other:11434", "gemma3:1b")
	if c2.BaseURL != "http://other:11434" {
		t.Errorf("expected custom base URL, got %s", c2.BaseURL)
	}
	if c2.Model != "gemma3:1b" {
		t.Errorf("expected custom model, got %s", c2.Model)
	}
}

func TestGenerate(t *testing.T) {
	// Mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Errorf("expected path /api/generate, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if req.Prompt == "" {
			t.Error("expected non-empty prompt")
		}

		resp := Response{
			Response: "This is a mock response from Gemma 3.",
			Created:  time.Now(),
			Done:     true,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "gemma3:1b")
	res, err := client.Generate("Tell me about jeBNF.")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expected := "This is a mock response from Gemma 3."
	if res != expected {
		t.Errorf("expected %q, got %q", expected, res)
	}
}

func TestGenerateError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error occurred"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "gemma3:1b")
	_, err := client.Generate("test prompt")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
