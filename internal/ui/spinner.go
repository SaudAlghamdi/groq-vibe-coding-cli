package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type SpinnerTickMsg time.Time

type Spinner struct {
	frame int
	label string
}

func NewSpinner(label string) Spinner {
	return Spinner{label: label}
}

func (s Spinner) Tick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return SpinnerTickMsg(t)
	})
}

func (s *Spinner) Update() {
	s.frame = (s.frame + 1) % len(spinnerFrames)
}

func (s Spinner) View() string {
	frame := SpinnerStyle.Render(spinnerFrames[s.frame])
	return frame + " " + DimStyle.Render(s.label)
}
