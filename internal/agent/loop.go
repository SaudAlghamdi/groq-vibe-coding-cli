package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/api"
	ctxpkg "github.com/saudalghamdi/groq-vibe-coding-cli/internal/context"
	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/safety"
	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/tools"
)

// EventType represents the type of agent event.
type EventType int

const (
	EventText EventType = iota
	EventToolCall
	EventToolResult
	EventDone
	EventError
	EventUsage
)

// Event represents something happening in the agent loop.
type Event struct {
	Type       EventType
	Content    string
	ToolName   string
	ToolCallID string
	ToolParams map[string]any
	Usage      *api.Usage
}

// EventHandler receives events from the agent loop.
type EventHandler func(Event)

// Options configures the agent loop.
type Options struct {
	Client        *api.Client
	Registry      *tools.Registry
	Conversation  *ctxpkg.Conversation
	Safety        *safety.Checker
	Model         string
	MaxIterations int
	OnEvent       EventHandler
}

// Run executes the agent loop: send message -> check tool calls -> execute -> repeat.
func Run(ctx context.Context, opts Options) error {
	if opts.MaxIterations == 0 {
		opts.MaxIterations = 25
	}

	for iteration := 0; iteration < opts.MaxIterations; iteration++ {
		req := api.ChatCompletionRequest{
			Model:    opts.Model,
			Messages: opts.Conversation.GetMessages(),
			Tools:    opts.Registry.Definitions(),
		}

		// Use streaming
		var fullContent string
		var allToolCalls []api.ToolCall
		var lastUsage *api.Usage

		err := opts.Client.ChatCompletionStream(ctx, req, func(content string, toolCalls []api.ToolCall, done bool, usage *api.Usage) {
			if content != "" {
				fullContent += content
				opts.OnEvent(Event{Type: EventText, Content: content})
			}
			if len(toolCalls) > 0 {
				allToolCalls = toolCalls
			}
			if usage != nil {
				lastUsage = usage
				opts.OnEvent(Event{Type: EventUsage, Usage: usage})
			}
		})

		if err != nil {
			opts.OnEvent(Event{Type: EventError, Content: err.Error()})
			return err
		}

		// Add assistant message to conversation
		opts.Conversation.AddAssistantMessage(fullContent, allToolCalls)

		if lastUsage != nil {
			opts.Conversation.TokensUsed = lastUsage.TotalTokens
		}

		// If no tool calls, we're done
		if len(allToolCalls) == 0 {
			opts.OnEvent(Event{Type: EventDone})
			return nil
		}

		// Execute tool calls (in parallel if multiple)
		if len(allToolCalls) == 1 {
			tc := allToolCalls[0]
			result := executeToolCall(ctx, opts, tc)
			opts.Conversation.AddToolResult(tc.ID, result)
		} else {
			// Parallel execution
			type toolResult struct {
				id     string
				result string
				index  int
			}
			results := make([]toolResult, len(allToolCalls))
			var wg sync.WaitGroup

			for i, tc := range allToolCalls {
				wg.Add(1)
				go func(idx int, call api.ToolCall) {
					defer wg.Done()
					res := executeToolCall(ctx, opts, call)
					results[idx] = toolResult{id: call.ID, result: res, index: idx}
				}(i, tc)
			}
			wg.Wait()

			for _, r := range results {
				opts.Conversation.AddToolResult(r.id, r.result)
			}
		}
	}

	opts.OnEvent(Event{Type: EventError, Content: "Maximum iterations reached"})
	return fmt.Errorf("maximum iterations (%d) reached", opts.MaxIterations)
}

func executeToolCall(ctx context.Context, opts Options, tc api.ToolCall) string {
	params, err := tools.ParseParams(tc.Function.Arguments)
	if err != nil {
		errMsg := fmt.Sprintf("Error parsing arguments: %v", err)
		opts.OnEvent(Event{Type: EventToolResult, ToolName: tc.Function.Name, ToolCallID: tc.ID, Content: errMsg})
		return errMsg
	}

	opts.OnEvent(Event{Type: EventToolCall, ToolName: tc.Function.Name, ToolCallID: tc.ID, ToolParams: params})

	// Check safety
	approved, err := opts.Safety.Check(tc.Function.Name, params)
	if err != nil {
		errMsg := fmt.Sprintf("Error checking approval: %v", err)
		opts.OnEvent(Event{Type: EventToolResult, ToolName: tc.Function.Name, ToolCallID: tc.ID, Content: errMsg})
		return errMsg
	}
	if !approved {
		result := "Tool call was denied by user."
		opts.OnEvent(Event{Type: EventToolResult, ToolName: tc.Function.Name, ToolCallID: tc.ID, Content: result})
		return result
	}

	tool, ok := opts.Registry.Get(tc.Function.Name)
	if !ok {
		errMsg := fmt.Sprintf("Unknown tool: %s", tc.Function.Name)
		opts.OnEvent(Event{Type: EventToolResult, ToolName: tc.Function.Name, ToolCallID: tc.ID, Content: errMsg})
		return errMsg
	}

	result, err := tool.Execute(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("Error: %v", err)
		opts.OnEvent(Event{Type: EventToolResult, ToolName: tc.Function.Name, ToolCallID: tc.ID, Content: errMsg})
		return errMsg
	}

	// Truncate very long results
	if len(result) > 10000 {
		result = result[:10000] + "\n... (truncated)"
	}

	opts.OnEvent(Event{Type: EventToolResult, ToolName: tc.Function.Name, ToolCallID: tc.ID, Content: result})
	return result
}
