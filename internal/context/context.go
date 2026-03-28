package context

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/api"
)

const maxContextChars = 120000 // ~30k tokens

// Conversation manages the message history and system prompt.
type Conversation struct {
	Messages     []api.Message
	SystemPrompt string
	TokensUsed   int
}

func NewConversation(customPrompt string) *Conversation {
	system := buildSystemPrompt(customPrompt)
	return &Conversation{
		SystemPrompt: system,
		Messages: []api.Message{
			{Role: "system", Content: system},
		},
	}
}

func (c *Conversation) AddUserMessage(content string) {
	c.Messages = append(c.Messages, api.Message{
		Role:    "user",
		Content: content,
	})
	c.truncateIfNeeded()
}

func (c *Conversation) AddAssistantMessage(content string, toolCalls []api.ToolCall) {
	c.Messages = append(c.Messages, api.Message{
		Role:      "assistant",
		Content:   content,
		ToolCalls: toolCalls,
	})
}

func (c *Conversation) AddToolResult(toolCallID, content string) {
	c.Messages = append(c.Messages, api.Message{
		Role:       "tool",
		Content:    content,
		ToolCallID: toolCallID,
	})
	c.truncateIfNeeded()
}

func (c *Conversation) GetMessages() []api.Message {
	return c.Messages
}

func (c *Conversation) Clear() {
	c.Messages = []api.Message{
		{Role: "system", Content: c.SystemPrompt},
	}
}

func (c *Conversation) truncateIfNeeded() {
	totalChars := 0
	for _, m := range c.Messages {
		totalChars += len(m.Content)
		for _, tc := range m.ToolCalls {
			totalChars += len(tc.Function.Arguments)
		}
	}

	if totalChars <= maxContextChars {
		return
	}

	// Keep system message and last N messages
	system := c.Messages[0]
	rest := c.Messages[1:]

	for totalChars > maxContextChars && len(rest) > 4 {
		removed := rest[0]
		totalChars -= len(removed.Content)
		for _, tc := range removed.ToolCalls {
			totalChars -= len(tc.Function.Arguments)
		}
		rest = rest[1:]
	}

	c.Messages = append([]api.Message{system}, rest...)
}

func buildSystemPrompt(customPrompt string) string {
	cwd, _ := os.Getwd()

	var sb strings.Builder
	sb.WriteString("You are groqcode, an AI-powered coding assistant running in the terminal.\n")
	sb.WriteString(fmt.Sprintf("Working directory: %s\n\n", cwd))

	// Detect project type
	projectInfo := detectProject(cwd)
	if projectInfo != "" {
		sb.WriteString("Project context:\n")
		sb.WriteString(projectInfo)
		sb.WriteString("\n")
	}

	sb.WriteString(`You have access to the following tools to help users with coding tasks:
- ReadFile: Read file contents with line numbers
- WriteFile: Write content to a file (creates or overwrites)
- EditFile: Make targeted edits using exact string replacement
- RunCommand: Execute shell commands with a 30s timeout
- SearchFiles: Search for patterns in files recursively
- ListDirectory: List files in a directory
- GetFileInfo: Get file metadata (size, modified time, language)

Guidelines:
- Read files before editing them to understand the existing code
- Use EditFile for targeted changes, WriteFile for new files or complete rewrites
- Show your reasoning when making decisions
- Be concise and practical in your responses
- When you need to explore the codebase, use ListDirectory and SearchFiles
`)

	if customPrompt != "" {
		sb.WriteString("\nCustom instructions:\n")
		sb.WriteString(customPrompt)
		sb.WriteString("\n")
	}

	return sb.String()
}

func detectProject(dir string) string {
	var parts []string

	checks := []struct {
		file string
		lang string
	}{
		{"go.mod", "Go"},
		{"package.json", "Node.js"},
		{"Cargo.toml", "Rust"},
		{"pyproject.toml", "Python"},
		{"requirements.txt", "Python"},
		{"pom.xml", "Java (Maven)"},
		{"build.gradle", "Java (Gradle)"},
		{"Gemfile", "Ruby"},
		{"composer.json", "PHP"},
	}

	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(dir, c.file)); err == nil {
			parts = append(parts, fmt.Sprintf("  - %s project (%s detected)", c.lang, c.file))
		}
	}

	// Check for git
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		parts = append(parts, "  - Git repository")
	}

	return strings.Join(parts, "\n")
}
