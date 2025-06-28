package cmd

import (
	"fmt"
	"informant/internal/config"
	"informant/internal/feed"
	"informant/internal/storage"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	listUnread  bool
	listReverse bool
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List news items",
	Long: `List the titles of the most recent news items. By default shows all items
regardless of read status, unless the --unread flag is used.

Items are shown with an index number that can be used with the 'read' command.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		store, err := storage.NewWithConfirmation(!viper.GetBool("no-confirm"))
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}

		var allItems []feed.Item

		for _, feedCfg := range cfg.Feeds {
			items, err := feed.ParseFeedWithStorage(feedCfg.URL, store)
			if err != nil {
				if viper.GetBool("verbose") {
					fmt.Fprintf(os.Stderr, "Warning: Failed to parse feed %s: %v\n", feedCfg.Name, err)
				}
				continue
			}

			// Add feed name to items
			for i := range items {
				items[i].FeedName = feedCfg.Name
			}

			allItems = append(allItems, items...)
		}

		// Sort by published date (newest first by default)
		sort.Slice(allItems, func(i, j int) bool {
			if listReverse {
				return allItems[i].Published.Before(allItems[j].Published)
			}
			return allItems[i].Published.After(allItems[j].Published)
		})

		// Filter by read status if requested
		var itemsToShow []feed.Item
		for _, item := range allItems {
			if listUnread && store.IsRead(item.ID) {
				continue
			}
			itemsToShow = append(itemsToShow, item)
		}

		if len(itemsToShow) == 0 {
			if listUnread {
				fmt.Println("No unread news items.")
			} else {
				fmt.Println("No news items found.")
			}
			return nil
		}

		// Display items with index
		for i, item := range itemsToShow {
			index := i + 1
			status := ""
			if store.IsRead(item.ID) {
				status = " [READ]"
			} else {
				status = " [UNREAD]"
			}

			dateStr := item.Published.Format("2006-01-02")
			feedInfo := ""
			if item.FeedName != "" {
				feedInfo = fmt.Sprintf(" (%s)", item.FeedName)
			}

			fmt.Printf("%d. %s %s%s%s\n", index, dateStr, item.Title, feedInfo, status)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVar(&listUnread, "unread", false, "only show unread items")
	listCmd.Flags().BoolVar(&listReverse, "reverse", false, "show items oldest to newest")
}
