package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove pacman hook from system",
	Long: `Remove the pacman hook that was installed for system integration.

This will remove /usr/share/libalpm/hooks/00-informant.hook and disable
automatic news checking during pacman transactions.

This command requires root privileges to remove the system-wide hook.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if running with appropriate privileges
		if os.Geteuid() != 0 {
			return fmt.Errorf("this command requires root privileges. Please run with sudo")
		}

		hookPath := "/usr/share/libalpm/hooks/00-informant.hook"

		// Check if hook exists
		if _, err := os.Stat(hookPath); os.IsNotExist(err) {
			fmt.Println("Pacman hook is not installed.")
			return nil
		}

		// Remove the hook file
		if err := os.Remove(hookPath); err != nil {
			return fmt.Errorf("failed to remove hook file: %w", err)
		}

		fmt.Printf("Successfully removed pacman hook from %s\n", hookPath)
		fmt.Println("\nPacman transactions will no longer check for Arch Linux news automatically.")
		fmt.Println("You can still manually check for news using:")
		fmt.Println("• informant check")
		fmt.Println("• informant list")
		fmt.Println("• informant tui")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
