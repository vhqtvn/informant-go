package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed assets/informant.hook
var hookContent []byte

var (
	installForce bool
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install pacman hook for system integration",
	Long: `Install the pacman hook to automatically check for Arch Linux news
during package installations and upgrades.

The hook will be installed to /usr/share/libalpm/hooks/00-informant.hook
and will interrupt pacman transactions when there are unread news items.

This command requires root privileges to install the system-wide hook.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if running with appropriate privileges
		if os.Geteuid() != 0 {
			return fmt.Errorf("this command requires root privileges. Please run with sudo")
		}

		// Get the current binary path
		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get executable path: %w", err)
		}

		// Resolve any symlinks to get the actual binary path
		actualPath, err := filepath.EvalSymlinks(execPath)
		if err != nil {
			return fmt.Errorf("failed to resolve executable path: %w", err)
		}

		hookDir := "/usr/share/libalpm/hooks"
		hookPath := filepath.Join(hookDir, "00-informant.hook")

		// Create hooks directory if it doesn't exist
		if err := os.MkdirAll(hookDir, 0755); err != nil {
			return fmt.Errorf("failed to create hooks directory: %w", err)
		}

		// Check if hook already exists
		if _, err := os.Stat(hookPath); err == nil && !installForce {
			return fmt.Errorf("hook already exists at %s. Use --force to overwrite", hookPath)
		}

		// Replace the hardcoded path with the actual binary path
		hookContentStr := string(hookContent)
		hookContentStr = strings.Replace(hookContentStr, "/usr/bin/informant check", actualPath+" check", 1)

		// Write the hook file
		if err := os.WriteFile(hookPath, []byte(hookContentStr), 0644); err != nil {
			return fmt.Errorf("failed to write hook file: %w", err)
		}

		fmt.Printf("Successfully installed pacman hook to %s\n", hookPath)
		fmt.Printf("Hook configured to use binary at: %s\n", actualPath)
		fmt.Println("\nThe hook will now:")
		fmt.Println("• Check for unread Arch Linux news before package installations/upgrades")
		fmt.Println("• Interrupt pacman transactions if unread news items are found")
		fmt.Println("• Ensure you stay informed about important system updates")
		fmt.Println("\nTo read news items, use: informant read")
		fmt.Println("To list news items, use: informant list")
		fmt.Println("To use the interactive TUI, use: informant tui")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	installCmd.Flags().BoolVar(&installForce, "force", false, "overwrite existing hook file")
}
