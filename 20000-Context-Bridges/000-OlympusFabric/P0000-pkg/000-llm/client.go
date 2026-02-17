// Filename: client.go
// Copyright: © 2026 VelociKey LLC. All Rights Reserved.
// Version: 1.0.0
// Author: Joseph A. White, III
// Status: Approved
// Approver: Joseph A. White, III
// Timestamp: 2026-02-13T14:41:00Z
// License: Proprietary

package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles communication with the Ollama API
type Client struct {
	BaseURL string
	Model   string
	Timeout time.Duration
}

// Request represents an Ollama generation request
type Request struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// Response represents an Ollama generation response
type Response struct {
	Response string    `json:"response"`
	Created  time.Time `json:"created_at"`
	Done     bool      `json:"done"`
}

func NewClient(baseUrl, model string) *Client {
	if baseUrl == "" {
		baseUrl = "http://localhost:11434"
	}
	if model == "" {
		model = "gemma3:27b"
	}
	return &Client{
		BaseURL: baseUrl,
		Model:   model,
		Timeout: 60 * time.Second,
	}
}

func (c *Client) Generate(prompt string) (string, error) {
	reqBody := Request{
		Model:  c.Model,
		Prompt: prompt,
		Stream: false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	httpClient := &http.Client{Timeout: c.Timeout}
	resp, err := httpClient.Post(c.BaseURL+"/api/generate", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned error (%d): %s", resp.StatusCode, string(body))
	}

	var ollamaResp Response
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", err
	}

	return ollamaResp.Response, nil
}
