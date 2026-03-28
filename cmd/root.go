package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/api"
	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/config"
	ctxpkg "github.com/saudalghamdi/groq-vibe-coding-cli/internal/context"
	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/safety"
	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/tools"
	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/ui"
)

var (
	cfgModel string
	yoloMode bool
)

var rootCmd = &cobra.Command{
	Use:   "groqcode",
	Short: "AI-powered coding agent in your terminal",
	Long:  "groqcode is a CLI tool that acts as an AI-powered coding agent using Groq's fast LLM inference.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if cfg.APIKey == "" {
			fmt.Fprintln(os.Stderr, "No API key found. Set GROQ_API_KEY or run 'groqcode init'.")
			os.Exit(1)
		}

		model := cfg.Model
		if cfgModel != "" {
			model = cfgModel
		}

		client := api.NewClient(cfg.APIKey)
		registry := tools.NewRegistry()
		conv := ctxpkg.NewConversation(cfg.CustomSystemPrompt)
		checker := safety.NewChecker(yoloMode)

		m := ui.NewModel(client, registry, conv, checker, model, cfg.MaxIterations)
		p := tea.NewProgram(m, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			return fmt.Errorf("running TUI: %w", err)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgModel, "model", "", "Model to use (default: llama-3.3-70b-versatile)")
	rootCmd.PersistentFlags().BoolVar(&yoloMode, "yolo", false, "Auto-approve all tool operations (use with caution!)")
}
