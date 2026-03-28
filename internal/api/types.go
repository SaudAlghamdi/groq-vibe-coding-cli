package api

import (
	"encoding/json"
	"fmt"
)

// Request types

type ChatCompletionRequest struct {
	Model         string         `json:"model"`
	Messages      []Message      `json:"messages"`
	Tools         []ToolDef      `json:"tools,omitempty"`
	Stream        bool           `json:"stream"`
	StreamOptions *StreamOptions `json:"stream_options,omitempty"`
	Temperature   float64        `json:"temperature,omitempty"`
	MaxTokens     int            `json:"max_tokens,omitempty"`
	TopP          float64        `json:"top_p,omitempty"`
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type Message struct {
	Role       string      `json:"role"`
	Content    string      `json:"content"`
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
	Name       string      `json:"name,omitempty"`
}

type ToolDef struct {
	Type     string       `json:"type"`
	Function FunctionDef  `json:"function"`
}

type FunctionDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type ToolCall struct {
	Index    int      `json:"index"`
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Function FuncCall `json:"function"`
}

type FuncCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Response types

type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Usage   *Usage   `json:"usage,omitempty"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	Delta        *Delta  `json:"delta,omitempty"`
	FinishReason string  `json:"finish_reason"`
}

type Delta struct {
	Role      string     `json:"role,omitempty"`
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Streaming chunk

type StreamChunk struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Usage   *Usage   `json:"usage,omitempty"`
}

// APIError represents a structured error response from the API.
type APIError struct {
	StatusCode int
	Message    string
	ErrorType  string
	Code       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

// apiErrorBody is used to decode the JSON error response body.
type apiErrorBody struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

type ModelsResponse struct {
	Data []ModelInfo `json:"data"`
}

type ModelInfo struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	OwnedBy string `json:"owned_by"`
}
