package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/api"
	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/config"
	"github.com/saudalghamdi/groq-vibe-coding-cli/internal/ui"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available Groq models",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if cfg.APIKey == "" {
			fmt.Fprintln(os.Stderr, "No API key found. Set GROQ_API_KEY or run 'groqcode init'.")
			os.Exit(1)
		}

		client := api.NewClient(cfg.APIKey)
		models, err := client.ListModels(context.Background())
		if err != nil {
			return fmt.Errorf("fetching models: %w", err)
		}

		sort.Slice(models, func(i, j int) bool {
			return models[i].ID < models[j].ID
		})

		fmt.Println(ui.HeaderStyle.Render("Available Groq Models"))
		fmt.Println()

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "  MODEL ID\tOWNER\n")
		fmt.Fprintf(w, "  --------\t-----\n")
		for _, m := range models {
			marker := "  "
			if m.ID == cfg.Model {
				marker = ui.SuccessStyle.Render("→ ")
			}
			fmt.Fprintf(w, "%s%s\t%s\n", marker, m.ID, m.OwnedBy)
		}
		w.Flush()

		fmt.Printf("\nCurrent model: %s\n", ui.SuccessStyle.Render(cfg.Model))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(modelsCmd)
}
