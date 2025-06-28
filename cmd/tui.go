package cmd

import (
	"fmt"
	"informant/internal/config"
	"informant/internal/feed"
	"informant/internal/storage"
	"informant/internal/tui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// tuiCmd represents the tui command
var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI mode",
	Long: `Launch the interactive Terminal User Interface for browsing and reading
news items. Use arrow keys or vim-style keys for navigation.

Key bindings:
- j/↓: Move down
- k/↑: Move up
- Enter: Read selected item
- r: Mark as read/unread
- q: Quit
- ?: Show help`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		store, err := storage.New()
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}

		// Collect all items
		var allItems []feed.Item
		for _, feedCfg := range cfg.Feeds {
			items, err := feed.ParseFeed(feedCfg.URL)
			if err != nil {
				if viper.GetBool("verbose") {
					fmt.Fprintf(os.Stderr, "Warning: Failed to parse feed %s: %v\n", feedCfg.Name, err)
				}
				continue
			}

			for i := range items {
				items[i].FeedName = feedCfg.Name
			}

			allItems = append(allItems, items...)
		}

		if len(allItems) == 0 {
			return fmt.Errorf("no news items found")
		}

		// Sort by published date (newest first)
		for i := 0; i < len(allItems)-1; i++ {
			for j := i + 1; j < len(allItems); j++ {
				if allItems[i].Published.Before(allItems[j].Published) {
					allItems[i], allItems[j] = allItems[j], allItems[i]
				}
			}
		}

		// Initialize and run TUI
		model := tui.NewModel(allItems, store)
		p := tea.NewProgram(model, tea.WithAltScreen())
		
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("TUI error: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
