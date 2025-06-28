# InformantGo

A modern Go rewrite of [informant](https://github.com/bradford-smith94/informant) - An Arch Linux News reader and pacman hook.

## Quick Start

```bash
# Install latest version with one command
curl -fsSL https://raw.githubusercontent.com/vhqtvn/informant-go/main/install.sh | sudo bash
```

Then use:
- `informant check` - Check for unread news (used by pacman hook)
- `informant tui` - Launch interactive terminal UI
- `informant list` - List news items
- `informant read` - Read news interactively

## Overview

InformantGo is a fast, feature-rich rewrite of the original Python informant tool. It maintains full compatibility with the original while adding new features like an interactive Terminal User Interface (TUI) and improved performance.

Like the original, InformantGo is designed to interrupt pacman transactions to ensure you have read important Arch Linux news before proceeding with system updates.

## Key Features

- **Full CLI Compatibility** - Drop-in replacement for the original informant commands
- **Interactive TUI Mode** - Browse and read news with a modern terminal interface  
- **Multiple Feed Support** - Configure multiple RSS/Atom feeds beyond just Arch Linux News
- **Fast Performance** - Written in Go for speed and efficiency
- **Pacman Hook Integration** - Seamlessly integrates with pacman to interrupt transactions
- **Cross-platform** - Works on any system with Go support (though designed for Arch Linux)

## Installation

### Prerequisites

- Go 1.19 or later
- git

### Quick Install (Recommended)

Install with a single command using our install script:

```bash
# Install latest version with automatic hook setup
curl -fsSL https://raw.githubusercontent.com/vhqtvn/informant-go/main/install.sh | sudo bash

# Or install without pacman hook
curl -fsSL https://raw.githubusercontent.com/vhqtvn/informant-go/main/install.sh | bash -s -- --no-hook

# Or install to user directory (non-root)
curl -fsSL https://raw.githubusercontent.com/vhqtvn/informant-go/main/install.sh | bash
```

The install script will:
- Auto-detect your architecture (AMD64/ARM64)
- Download the latest release binary
- Verify SHA256 checksums for security
- Install to `/usr/local/bin` (with sudo) or `~/.local/bin` (without)
- Automatically setup pacman hook on Arch Linux (if running as root)

**For detailed installation options and troubleshooting, see [INSTALL.md](INSTALL.md)**

### Manual Installation from GitHub Releases

Download pre-built binaries from the [GitHub Releases](../../releases) page:

```bash
# Download the latest release for your architecture
wget https://github.com/vhqtvn/informant-go/releases/latest/download/informant-linux-amd64
chmod +x informant-linux-amd64
sudo mv informant-linux-amd64 /usr/local/bin/informant

# Install the pacman hook
sudo informant install
```

### Build from Source

```bash
git clone https://github.com/vhqtvn/informant-go.git
cd informant-go
go build -o informant .
sudo cp informant /usr/local/bin/
```

### Development Setup

```bash
git clone https://github.com/vhqtvn/informant-go.git
cd informant-go
go mod download
go run . --help
```

### Install Script Testing

You can test the install script locally:

```bash
# Test the script (dry run)
./install.sh --help

# Test local installation
sudo ./install.sh

# Test without hook installation
./install.sh --no-hook
```

## Usage

InformantGo provides several commands for managing Arch Linux news:

### Commands

#### `informant check`
Check for unread news items. If there's exactly one unread item, it will be displayed and marked as read. The command exits with a return code equal to the number of unread items.

```bash
informant check
```

This is the command used by the pacman hook to interrupt transactions.

#### `informant list`
List news item titles with their read status and indices.

```bash
informant list                    # Show all items
informant list --unread          # Show only unread items  
informant list --reverse         # Show oldest to newest
```

#### `informant read`
Read specific news items or interactively read all unread items.

```bash
informant read                    # Interactive mode for unread items
informant read 3                  # Read item #3 (from list output)
informant read "kernel"           # Read item matching "kernel" in title
informant read --all              # Mark all items as read without displaying
```

#### `informant tui`
Launch the interactive Terminal User Interface for browsing news.

```bash
informant tui
```

**TUI Key Bindings:**
- `j/↓` - Move down
- `k/↑` - Move up  
- `Enter` - Read selected item
- `r` - Toggle read/unread status
- `q` - Quit
- `?` - Show help

#### `informant install`
Install the pacman hook for automatic news checking during package operations.

```bash
sudo informant install              # Install the pacman hook
sudo informant install --force     # Overwrite existing hook
```

This command embeds the hook file within the binary and installs it to `/usr/share/libalpm/hooks/00-informant.hook`.

#### `informant uninstall`
Remove the pacman hook from the system.

```bash
sudo informant uninstall           # Remove the pacman hook
```

### Global Options

```bash
informant --config /path/to/config.json    # Use custom config file
informant --verbose                         # Enable verbose output
informant --help                           # Show help
informant --version                        # Show version
```

## Configuration

InformantGo looks for configuration files in the following order:

1. Path specified with `--config` flag
2. `$HOME/.informantrc.json`
3. `$XDG_CONFIG_HOME/informantrc.json`
4. `/etc/informantrc.json`
5. Each directory in `$XDG_CONFIG_DIRS`

### Configuration Format

```json
{
  "feeds": [
    {
      "name": "Arch Linux News",
      "url": "https://archlinux.org/feeds/news/",
      "title-key": "title",
      "body-key": "summary", 
      "timestamp-key": "published"
    },
    {
      "name": "Arch Linux 32 News",
      "url": "https://archlinux32.org/news/feed/",
      "title-key": "title",
      "body-key": "summary",
      "timestamp-key": "published"
    }
  ]
}
```

### Configuration Fields

- `name` (optional) - Display name for the feed
- `url` (required) - RSS/Atom feed URL
- `title-key` (optional) - Key for item title in feed (default: "title")
- `body-key` (optional) - Key for item content in feed (default: "summary") 
- `timestamp-key` (optional) - Key for item date in feed (default: "published")

**Note:** For pacman hook integration, place your config in `/etc/informantrc.json` so it's accessible when running as root.

## Pacman Hook Integration

InformantGo is designed to work as a pacman PreTransaction hook. When you install packages or perform updates, it will:

1. Check for unread Arch Linux news
2. If unread items exist, interrupt the transaction 
3. Display the number of unread items
4. Exit with a non-zero code to halt pacman

### Easy Installation

The simplest way to set up the pacman hook is using the built-in install command:

```bash
sudo informant install
```

This will automatically install the hook to `/usr/share/libalpm/hooks/00-informant.hook`.

### Hook Configuration

The installed hook contains the following configuration:

```ini
[Trigger]
Operation = Install
Operation = Upgrade
Type = Package
Target = *
Target = !informant

[Action]
Description = Checking Arch News with Informant...
When = PreTransaction
Exec = <binary-path> check
AbortOnFail
```

**Note:** `<binary-path>` is automatically replaced with the actual path of the informant binary during installation.

**Key features of the hook:**
- Triggers on package installs and upgrades
- Excludes itself (`!informant`) to prevent recursion during updates
- Uses `AbortOnFail` to halt pacman if unread news exists
- Automatically uses the current binary path (no hardcoded paths)

### Managing the Hook

```bash
sudo informant install           # Install the hook
sudo informant install --force  # Overwrite existing hook
sudo informant uninstall        # Remove the hook
```

### Manual Hook Management

To temporarily disable the hook without uninstalling:
```bash
sudo ln -s /dev/null /etc/pacman.d/hooks/00-informant.hook
```

To re-enable after disabling:
```bash
sudo rm /etc/pacman.d/hooks/00-informant.hook
```

## Architecture

The project is organized as follows:

```
cmd/           # CLI commands (cobra)
├── assets/    # Embedded assets
│   └── informant.hook  # Pacman hook configuration
├── root.go    # Root command and config initialization
├── check.go   # Check command for pacman hook
├── list.go    # List command for displaying items
├── read.go    # Read command for reading items
├── tui.go     # TUI command for interactive mode
├── install.go # Install command for pacman hook
└── uninstall.go # Uninstall command for pacman hook

internal/      # Internal packages
├── config/    # Configuration management
├── feed/      # RSS/Atom feed parsing
├── storage/   # Read status tracking
└── tui/       # Terminal UI components

.github/workflows/  # CI/CD automation
└── release.yml     # Automated build and release

install.sh     # One-line installation script
INSTALL.md     # Comprehensive installation guide
Makefile       # Build automation
main.go        # Application entry point
informantrc.json.example  # Configuration example
```

## Dependencies

- **[Cobra](https://github.com/spf13/cobra)** - CLI framework
- **[Viper](https://github.com/spf13/viper)** - Configuration management
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - TUI framework
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - Terminal styling

## Development

### Building

```bash
make build           # Build the binary
make build-upx       # Build and compress with UPX
go build -o informant .  # Alternative direct build
```

**UPX Compression**: To use UPX compression locally, install it first:
```bash
sudo apt-get install upx-ucl  # Ubuntu/Debian
```

### Testing

```bash
make test           # Run tests
go test ./...       # Alternative direct test
```

### Local Installation

```bash
make install        # Install to /usr/local/bin (requires sudo)
make uninstall      # Remove from /usr/local/bin (requires sudo)
```

### Creating Releases

To create a new release:

1. **Tag the release:**
   ```bash
   git tag v1.4.2
   git push origin v1.4.2
   ```

2. **GitHub Actions automatically:**
   - Builds optimized binaries for Linux AMD64 and ARM64
   - Compresses binaries with UPX for smaller download sizes
   - Creates SHA256 checksums
   - Generates comprehensive release notes
   - Publishes the release with all artifacts

3. **Manual local release build:**
   ```bash
   make release        # Build release binaries locally
   make release-upx    # Build and compress with UPX (requires: sudo apt-get install upx-ucl)
   ```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Differences from Original

### New Features
- **TUI Mode** - Interactive terminal interface for browsing news
- **Easy Installation** - Built-in `install`/`uninstall` commands for pacman hook management
- **Embedded Hook** - Hook configuration is embedded in the binary for easier deployment
- **Improved Performance** - Go's compiled nature provides faster execution
- **Better Error Handling** - More robust error messages and recovery
- **Modern Dependencies** - Uses current, well-maintained Go libraries

### Compatibility
- All original command-line options and behavior are preserved
- Configuration file format is identical
- Pacman hook integration works the same way
- Exit codes and output formats match the original

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Original [informant](https://github.com/bradford-smith94/informant) project by [@bradford-smith94](https://github.com/bradford-smith94)
- The Go community for excellent libraries and tools
- Arch Linux community for the news feed infrastructure

## Repository

- **GitHub**: https://github.com/vhqtvn/informant-go
- **Issues**: https://github.com/vhqtvn/informant-go/issues  
- **Releases**: https://github.com/vhqtvn/informant-go/releases
- **Install Script**: `curl -fsSL https://raw.githubusercontent.com/vhqtvn/informant-go/main/install.sh | sudo bash`

## CI/CD

This project uses GitHub Actions for automated building and releasing:

- **Continuous Integration**: On every push, the code is built and tested
- **Automated Releases**: When you push a tag starting with `v` (e.g., `v1.4.2`), GitHub Actions automatically:
  - Builds binaries for Linux AMD64 and ARM64
  - Creates SHA256 checksums
  - Generates comprehensive release notes
  - Publishes a GitHub release with all artifacts

### Supported Architectures

- **Linux AMD64** (`informant-linux-amd64`) - Standard 64-bit Linux
- **Linux ARM64** (`informant-linux-arm64`) - ARM 64-bit Linux (e.g., Raspberry Pi 4+)

All release binaries are compressed with UPX for significantly smaller download sizes (60-70% reduction) while maintaining full functionality.
