package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const baseURL = "https://api.groq.com/openai/v1"

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ChatCompletion sends a non-streaming chat completion request.
func (c *Client) ChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	req.Stream = false

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	var resp ChatCompletionResponse
	if err := c.doWithRetry(ctx, "POST", "/chat/completions", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// StreamCallback receives streaming chunks. Content is the text delta, done indicates the final chunk.
type StreamCallback func(content string, toolCalls []ToolCall, done bool, usage *Usage)

// ChatCompletionStream sends a streaming chat completion request.
// If the stream completes successfully but produces no content and no tool calls,
// it falls back to a non-streaming request.
func (c *Client) ChatCompletionStream(ctx context.Context, req ChatCompletionRequest, cb StreamCallback) error {
	gotContent, gotToolCalls, err := c.chatCompletionStreamOnce(ctx, req, cb)
	if err != nil {
		return err
	}

	// Fallback: stream returned empty (no content, no tool calls) — some models/providers
	// occasionally produce valid HTTP 200 SSE streams with no payload. Retrying with the
	// non-streaming endpoint is more reliable in that case.
	if !gotContent && !gotToolCalls {
		resp, err := c.ChatCompletion(ctx, req)
		if err != nil {
			return err
		}
		if len(resp.Choices) > 0 {
			msg := resp.Choices[0].Message
			if msg.Content != "" {
				cb(msg.Content, nil, false, nil)
			}
			if len(msg.ToolCalls) > 0 {
				cb("", msg.ToolCalls, false, nil)
			}
			if resp.Usage != nil {
				cb("", nil, false, resp.Usage)
			}
		}
		cb("", nil, true, nil)
	}

	return nil
}

// chatCompletionStreamOnce performs a single streaming request and returns whether
// any content or tool calls were received.
func (c *Client) chatCompletionStreamOnce(ctx context.Context, req ChatCompletionRequest, cb StreamCallback) (gotContent bool, gotToolCalls bool, err error) {
	streamReq := req
	streamReq.Stream = true
	streamReq.StreamOptions = &StreamOptions{IncludeUsage: true}

	body, err := json.Marshal(streamReq)
	if err != nil {
		return false, false, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return false, false, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return false, false, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return false, false, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	// Track accumulated tool calls across chunks
	accumulatedToolCalls := make(map[int]*ToolCall)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			cb("", nil, true, nil)
			return gotContent, gotToolCalls, nil
		}

		var chunk StreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return gotContent, gotToolCalls, fmt.Errorf("parsing stream chunk: %w", err)
		}

		if chunk.Usage != nil {
			cb("", nil, false, chunk.Usage)
			continue
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		delta := chunk.Choices[0].Delta
		if delta == nil {
			if chunk.Choices[0].FinishReason == "tool_calls" {
				// Collect all accumulated tool calls
				var toolCalls []ToolCall
				for _, tc := range accumulatedToolCalls {
					toolCalls = append(toolCalls, *tc)
				}
				cb("", toolCalls, false, nil)
				gotToolCalls = true
			}
			continue
		}

		if delta.Content != "" {
			cb(delta.Content, nil, false, nil)
			gotContent = true
		}

		// Accumulate tool call deltas
		for _, tc := range delta.ToolCalls {
			idx := tc.Index
			existing, ok := accumulatedToolCalls[idx]
			if !ok {
				newTC := ToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: FuncCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
				accumulatedToolCalls[idx] = &newTC
			} else {
				if tc.Function.Name != "" {
					existing.Function.Name += tc.Function.Name
				}
				existing.Function.Arguments += tc.Function.Arguments
			}
		}

		if chunk.Choices[0].FinishReason == "tool_calls" {
			var toolCalls []ToolCall
			for _, tc := range accumulatedToolCalls {
				toolCalls = append(toolCalls, *tc)
			}
			cb("", toolCalls, false, nil)
			gotToolCalls = true
			accumulatedToolCalls = make(map[int]*ToolCall)
		}
	}

	return gotContent, gotToolCalls, scanner.Err()
}

// ListModels fetches available models from the Groq API.
func (c *Client) ListModels(ctx context.Context) ([]ModelInfo, error) {
	var resp ModelsResponse
	if err := c.doWithRetry(ctx, "GET", "/models", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) doWithRetry(ctx context.Context, method, path string, body []byte, out any) error {
	maxRetries := 3
	backoff := time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		var bodyReader io.Reader
		if body != nil {
			bodyReader = bytes.NewReader(body)
		}

		req, err := http.NewRequestWithContext(ctx, method, baseURL+path, bodyReader)
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt == maxRetries {
				return fmt.Errorf("request failed after retries: %w", err)
			}
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("reading response: %w", err)
		}

		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			if attempt == maxRetries {
				return fmt.Errorf("API error (status %d) after retries: %s", resp.StatusCode, string(respBody))
			}
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
		}

		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
		return nil
	}
	return fmt.Errorf("request failed after %d retries", maxRetries)
}
