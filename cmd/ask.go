package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/agent"
	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/api"
	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/config"
	ctxpkg "github.com/saudalghamdi/groq-vibe-coding-cli/internal/context"
	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/safety"
	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/tools"
	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/ui"
)

var askCmd = &cobra.Command{
	Use:   "ask [question]",
	Short: "Ask a one-shot question and get an answer",
	Long:  "Ask a question, get an answer, and exit. No interactive session.",
	Args:  cobra.MinimumNArgs(1),
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

		question := strings.Join(args, " ")

		client := api.NewClient(cfg.APIKey)
		registry := tools.NewRegistry()
		conv := ctxpkg.NewConversation(cfg.CustomSystemPrompt)
		checker := safety.NewChecker(yoloMode)

		conv.AddUserMessage(question)

		var output strings.Builder

		err = agent.Run(context.Background(), agent.Options{
			Client:        client,
			Registry:      registry,
			Conversation:  conv,
			Safety:        checker,
			Model:         model,
			MaxIterations: cfg.MaxIterations,
			OnEvent: func(evt agent.Event) {
				switch evt.Type {
				case agent.EventText:
					output.WriteString(evt.Content)
					fmt.Print(evt.Content) // Stream to stdout
				case agent.EventToolCall:
					fmt.Fprintf(os.Stderr, "%s\n",
						ui.ToolCallStyle.Render(fmt.Sprintf("🔧 %s", evt.ToolName)))
				case agent.EventError:
					fmt.Fprintf(os.Stderr, "%s\n",
						ui.ErrorStyle.Render("Error: "+evt.Content))
				}
			},
		})

		fmt.Println() // Final newline
		return err
	},
}

func init() {
	rootCmd.AddCommand(askCmd)
}
