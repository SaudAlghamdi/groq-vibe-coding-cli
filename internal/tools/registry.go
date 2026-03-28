package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/api"
)

// Tool defines the interface all tools must implement.
type Tool interface {
	Name() string
	Description() string
	Parameters() map[string]any
	Execute(ctx context.Context, params map[string]any) (string, error)
}

// Registry holds all registered tools.
type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	r := &Registry{tools: make(map[string]Tool)}
	// Register all built-in tools
	r.Register(&ReadFileTool{})
	r.Register(&WriteFileTool{})
	r.Register(&EditFileTool{})
	r.Register(&RunCommandTool{})
	r.Register(&SearchFilesTool{})
	r.Register(&ListDirectoryTool{})
	r.Register(&GetFileInfoTool{})
	return r
}

func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
}

func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// Definitions returns tool definitions for the API request.
func (r *Registry) Definitions() []api.ToolDef {
	var defs []api.ToolDef
	for _, t := range r.tools {
		params, _ := json.Marshal(t.Parameters())
		defs = append(defs, api.ToolDef{
			Type: "function",
			Function: api.FunctionDef{
				Name:        t.Name(),
				Description: t.Description(),
				Parameters:  params,
			},
		})
	}
	return defs
}

// ParseParams parses a JSON arguments string into a parameter map.
func ParseParams(args string) (map[string]any, error) {
	var params map[string]any
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return nil, fmt.Errorf("parsing tool arguments: %w", err)
	}
	return params, nil
}

// GetString extracts a string parameter.
func GetString(params map[string]any, key string) string {
	v, ok := params[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}
