package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/config"
	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/ui"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize groqcode configuration",
	Long:  "Create a configuration file with your Groq API key and preferences.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(ui.HeaderStyle.Render("groqcode setup"))
		fmt.Println()

		reader := bufio.NewReader(os.Stdin)

		// Check for existing config
		existing, _ := config.Load()

		// Get API key
		fmt.Print("Groq API key (gsk_...): ")
		apiKey, _ := reader.ReadString('\n')
		apiKey = strings.TrimSpace(apiKey)

		if apiKey == "" && existing.APIKey != "" {
			apiKey = existing.APIKey
			fmt.Println(ui.DimStyle.Render("  (keeping existing key)"))
		} else if apiKey == "" {
			fmt.Println(ui.ErrorStyle.Render("API key is required. Get one at https://console.groq.com"))
			return nil
		}

		// Get model preference
		fmt.Printf("Default model [%s]: ", config.DefaultModel)
		model, _ := reader.ReadString('\n')
		model = strings.TrimSpace(model)
		if model == "" {
			model = config.DefaultModel
		}

		cfg := &config.Config{
			APIKey:           apiKey,
			Model:            model,
			MaxIterations:    config.DefaultMaxIterations,
			AutoApproveReads: true,
			Theme:            "dark",
		}

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		fmt.Println()
		fmt.Println(ui.SuccessStyle.Render("✓ Config saved to " + config.ConfigDir() + "/config.yaml"))
		fmt.Println(ui.DimStyle.Render("  Run 'groqcode' to start chatting!"))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
