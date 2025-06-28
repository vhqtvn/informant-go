package cmd

import (
	"fmt"
	"informant/internal/config"
	"informant/internal/feed"
	"informant/internal/storage"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for unread news items",
	Long: `Check for any unread news items. If there is only one unread item it will
print it and mark it as read. The command will exit with return code equal to
the number of unread news items.

This is the command used by the pacman hook to interrupt transactions when
there are unread news items.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		store, err := storage.New()
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}

		var unreadCount int
		var unreadItems []feed.Item

		for _, feedCfg := range cfg.Feeds {
			items, err := feed.ParseFeed(feedCfg.URL)
			if err != nil {
				if viper.GetBool("verbose") {
					fmt.Fprintf(os.Stderr, "Warning: Failed to parse feed %s: %v\n", feedCfg.Name, err)
				}
				continue
			}

			for _, item := range items {
				if !store.IsRead(item.ID) {
					unreadItems = append(unreadItems, item)
					unreadCount++
				}
			}
		}

		// If there's exactly one unread item, print it and mark as read
		if unreadCount == 1 {
			item := unreadItems[0]
			fmt.Printf("Title: %s\n", item.Title)
			fmt.Printf("Date: %s\n", item.Published.Format("2006-01-02 15:04:05"))
			if item.FeedName != "" {
				fmt.Printf("Feed: %s\n", item.FeedName)
			}
			fmt.Printf("\n%s\n", item.Content)

			if err := store.MarkAsRead(item.ID); err != nil {
				return fmt.Errorf("failed to mark item as read: %w", err)
			}
		} else if unreadCount > 1 {
			fmt.Printf("There are %d unread news items.\n", unreadCount)
			fmt.Println("Use 'informant list --unread' to see them or 'informant read' to read them.")
		}

		// Exit with the number of unread items for pacman hook integration
		os.Exit(unreadCount)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
