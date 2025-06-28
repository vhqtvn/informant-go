package cmd

import (
	"bufio"
	"fmt"
	"informant/internal/config"
	"informant/internal/feed"
	"informant/internal/storage"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	readAll bool
)

// readCmd represents the read command
var readCmd = &cobra.Command{
	Use:   "read [item]",
	Short: "Read news items",
	Long: `Read news items and mark them as read. You can specify an item by:
- Index number (as shown in 'informant list')
- String matching the title

If no item is specified, will loop through all unread items with prompts.
Use --all to mark all items as read without displaying them.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		store, err := storage.NewWithConfirmation(!viper.GetBool("no-confirm"))
		if err != nil {
			return fmt.Errorf("failed to initialize storage: %w", err)
		}

		// Collect all items
		var allItems []feed.Item
		for _, feedCfg := range cfg.Feeds {
			items, err := feed.ParseFeedWithStorage(feedCfg.URL, store)
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

		// Sort by published date (newest first)
		// This matches the order shown in 'list' command
		for i := 0; i < len(allItems)-1; i++ {
			for j := i + 1; j < len(allItems); j++ {
				if allItems[i].Published.Before(allItems[j].Published) {
					allItems[i], allItems[j] = allItems[j], allItems[i]
				}
			}
		}

		if readAll {
			// Mark all items as read without displaying
			count := 0
			for _, item := range allItems {
				if !store.IsRead(item.ID) {
					if err := store.MarkAsRead(item.ID); err != nil {
						return fmt.Errorf("failed to mark item as read: %w", err)
					}
					count++
				}
			}
			fmt.Printf("Marked %d items as read.\n", count)
			return nil
		}

		if len(args) == 0 {
			// Interactive mode - loop through unread items
			return readUnreadInteractive(allItems, store)
		}

		// Read specific item
		return readSpecificItem(args[0], allItems, store)
	},
}

func readUnreadInteractive(allItems []feed.Item, store *storage.Storage) error {
	reader := bufio.NewReader(os.Stdin)

	for _, item := range allItems {
		if store.IsRead(item.ID) {
			continue
		}

		displayItem(item)

		fmt.Print("\nMark as read and continue? [Y/n]: ")
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response == "" || response == "y" || response == "yes" {
			if err := store.MarkAsRead(item.ID); err != nil {
				return fmt.Errorf("failed to mark item as read: %w", err)
			}
			fmt.Println("Marked as read.")
		} else {
			fmt.Println("Skipped.")
		}
		fmt.Println()
	}

	return nil
}

func readSpecificItem(itemRef string, allItems []feed.Item, store *storage.Storage) error {
	var targetItem *feed.Item

	// Try to parse as index first
	if index, err := strconv.Atoi(itemRef); err == nil {
		if index >= 1 && index <= len(allItems) {
			targetItem = &allItems[index-1]
		}
	} else {
		// Search by title
		itemRef = strings.ToLower(itemRef)
		for i, item := range allItems {
			if strings.Contains(strings.ToLower(item.Title), itemRef) {
				targetItem = &allItems[i]
				break
			}
		}
	}

	if targetItem == nil {
		return fmt.Errorf("item not found: %s", itemRef)
	}

	displayItem(*targetItem)

	if err := store.MarkAsRead(targetItem.ID); err != nil {
		return fmt.Errorf("failed to mark item as read: %w", err)
	}

	return nil
}

func displayItem(item feed.Item) {
	fmt.Printf("Title: %s\n", item.Title)
	fmt.Printf("Date: %s\n", item.Published.Format("2006-01-02 15:04:05"))
	if item.FeedName != "" {
		fmt.Printf("Feed: %s\n", item.FeedName)
	}
	fmt.Printf("\n%s\n", item.Content)

	// Check if content is long and offer pager
	lines := strings.Count(item.Content, "\n")
	if lines > 20 {
		fmt.Print("\nPress Enter to continue or 'p' to view in pager: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "p" {
			showInPager(fmt.Sprintf("Title: %s\nDate: %s\nFeed: %s\n\n%s",
				item.Title, item.Published.Format("2006-01-02 15:04:05"), item.FeedName, item.Content))
		}
	}
}

func showInPager(content string) {
	// Try to use system pager
	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "less"
	}

	cmd := exec.Command(pager)
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Fallback to simple output if pager fails
		fmt.Print(content)
	}
}

func init() {
	rootCmd.AddCommand(readCmd)

	readCmd.Flags().BoolVar(&readAll, "all", false, "mark all items as read without displaying them")
}
