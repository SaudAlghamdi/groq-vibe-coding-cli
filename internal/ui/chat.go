package ui

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"

	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/agent"
	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/api"
	ctxpkg "github.com/saudalghamdi/groq-vibe-coding-cli/internal/context"
	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/safety"
	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/tools"
)

type state int

const (
	stateInput state = iota
	stateThinking
)

// Msgs for bubbletea
type agentEventMsg agent.Event
type agentDoneMsg struct{}
type agentErrorMsg struct{ err error }

// waitForEvent returns a cmd that waits for the next event on the channel.
func waitForEvent(ch <-chan agent.Event) tea.Cmd {
	return func() tea.Msg {
		evt, ok := <-ch
		if !ok {
			return agentDoneMsg{}
		}
		return agentEventMsg(evt)
	}
}

type Model struct {
	// Dependencies
	client   *api.Client
	registry *tools.Registry
	conv     *ctxpkg.Conversation
	safety   *safety.Checker
	model    string
	maxIter  int

	// UI state
	state       state
	input       string
	history     []historyEntry // conversation display history
	output      strings.Builder
	toolSteps   []string
	spinner     Spinner
	width       int
	height      int
	usage       *api.Usage
	err         error
	quitting    bool
	mdRenderer  *glamour.TermRenderer
	eventCh     chan agent.Event
	cancelAgent context.CancelFunc
}

type historyEntry struct {
	role    string // "user" or "assistant"
	content string
	tools   []string
}

func NewModel(client *api.Client, registry *tools.Registry, conv *ctxpkg.Conversation, checker *safety.Checker, model string, maxIter int) Model {
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	return Model{
		client:     client,
		registry:   registry,
		conv:       conv,
		safety:     checker,
		model:      model,
		maxIter:    maxIter,
		state:      stateInput,
		spinner:    NewSpinner("Thinking..."),
		mdRenderer: renderer,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case SpinnerTickMsg:
		if m.state == stateThinking {
			m.spinner.Update()
			return m, m.spinner.Tick()
		}
		return m, nil

	case agentEventMsg:
		evt := agent.Event(msg)
		switch evt.Type {
		case agent.EventText:
			m.output.WriteString(evt.Content)

		case agent.EventToolCall:
			step := FormatToolCall(evt.ToolName, evt.ToolParams)
			m.toolSteps = append(m.toolSteps, step)
			m.spinner = NewSpinner(fmt.Sprintf("Running %s...", evt.ToolName))

		case agent.EventUsage:
			m.usage = evt.Usage

		case agent.EventDone:
			m.finishResponse()
			return m, nil

		case agent.EventError:
			m.finishResponse()
			m.err = fmt.Errorf("%s", evt.Content)
			return m, nil
		}
		// Keep listening for more events
		return m, waitForEvent(m.eventCh)

	case agentDoneMsg:
		m.finishResponse()
		return m, nil

	case agentErrorMsg:
		m.finishResponse()
		m.err = msg.err
		return m, nil
	}

	return m, nil
}

func (m *Model) finishResponse() {
	if m.output.Len() > 0 {
		m.history = append(m.history, historyEntry{
			role:    "assistant",
			content: m.output.String(),
			tools:   m.toolSteps,
		})
	}
	m.output.Reset()
	m.toolSteps = nil
	m.state = stateInput
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.state == stateThinking {
		if msg.Type == tea.KeyCtrlC {
			if m.cancelAgent != nil {
				m.cancelAgent()
			}
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil
	}

	switch msg.Type {
	case tea.KeyCtrlC:
		m.quitting = true
		return m, tea.Quit

	case tea.KeyEnter:
		input := strings.TrimSpace(m.input)
		if input == "" {
			return m, nil
		}

		// Handle slash commands
		if handled := m.handleSlashCommand(input); handled {
			m.input = ""
			return m, nil
		}

		// Add to history
		m.history = append(m.history, historyEntry{role: "user", content: input})
		m.output.Reset()
		m.toolSteps = nil
		m.err = nil

		m.conv.AddUserMessage(input)
		m.state = stateThinking
		m.spinner = NewSpinner("Thinking...")
		m.input = ""

		// Create channel for events
		m.eventCh = make(chan agent.Event, 100)

		return m, tea.Batch(m.spinner.Tick(), m.runAgent(), waitForEvent(m.eventCh))

	case tea.KeyBackspace:
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
		return m, nil

	case tea.KeyRunes:
		m.input += string(msg.Runes)
		return m, nil

	case tea.KeySpace:
		m.input += " "
		return m, nil
	}

	return m, nil
}

func (m *Model) handleSlashCommand(input string) bool {
	switch input {
	case "/exit", "/quit":
		m.quitting = true
		return true
	case "/clear":
		m.conv.Clear()
		m.history = nil
		m.output.Reset()
		m.toolSteps = nil
		m.err = nil
		return true
	case "/help":
		m.history = append(m.history, historyEntry{role: "assistant", content: helpText()})
		return true
	case "/model":
		m.history = append(m.history, historyEntry{role: "assistant", content: "Current model: " + m.model})
		return true
	}
	if strings.HasPrefix(input, "/model ") {
		m.model = strings.TrimPrefix(input, "/model ")
		m.history = append(m.history, historyEntry{role: "assistant", content: "Model changed to: " + m.model})
		return true
	}
	return false
}

func (m Model) runAgent() tea.Cmd {
	ch := m.eventCh
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		_ = cancel // stored on model via closure isn't easy, just let it run

		err := agent.Run(ctx, agent.Options{
			Client:        m.client,
			Registry:      m.registry,
			Conversation:  m.conv,
			Safety:        m.safety,
			Model:         m.model,
			MaxIterations: m.maxIter,
			OnEvent: func(evt agent.Event) {
				ch <- evt
			},
		})
		close(ch)
		if err != nil {
			return agentErrorMsg{err: err}
		}
		return agentDoneMsg{}
	}
}

func (m Model) View() string {
	if m.quitting {
		return DimStyle.Render("Goodbye!") + "\n"
	}

	var sb strings.Builder

	// Header
	sb.WriteString(HeaderStyle.Render("groqcode") + " ")
	sb.WriteString(DimStyle.Render(m.model))
	sb.WriteString("\n\n")

	// History
	for _, entry := range m.history {
		switch entry.role {
		case "user":
			sb.WriteString(UserPromptStyle.Render("> "+entry.content) + "\n\n")
		case "assistant":
			for _, step := range entry.tools {
				sb.WriteString(ToolCallStyle.Render(step) + "\n")
			}
			content := entry.content
			if m.mdRenderer != nil {
				if rendered, err := m.mdRenderer.Render(content); err == nil {
					content = rendered
				}
			}
			sb.WriteString(content + "\n")
		}
	}

	// Current streaming output
	if m.output.Len() > 0 {
		for _, step := range m.toolSteps {
			sb.WriteString(ToolCallStyle.Render(step) + "\n")
		}
		sb.WriteString(m.output.String())
		sb.WriteString("\n")
	}

	// Error
	if m.err != nil {
		sb.WriteString(ErrorStyle.Render("Error: "+m.err.Error()) + "\n")
	}

	// Spinner or input
	if m.state == stateThinking {
		sb.WriteString(m.spinner.View() + "\n")
	} else {
		sb.WriteString(UserPromptStyle.Render("> ") + m.input)
		sb.WriteString(DimStyle.Render("█") + "\n")
	}

	// Status bar
	cwd, _ := os.Getwd()
	status := fmt.Sprintf(" %s | %s", m.model, cwd)
	if m.usage != nil {
		status += fmt.Sprintf(" | tokens: %d", m.usage.TotalTokens)
	}
	sb.WriteString("\n" + StatusBarStyle.Render(status) + "\n")

	return sb.String()
}

func helpText() string {
	return `Commands:
  /clear        — Clear conversation history
  /model        — Show current model
  /model <name> — Switch model
  /help         — Show this help
  /exit         — Exit groqcode

Shortcuts:
  Enter   — Send message
  Ctrl+C  — Exit`
}
