# InformantGo Installation Guide

This guide covers all the different ways to install InformantGo on your system.

## Quick Install (Recommended)

The fastest way to get InformantGo installed is using our automated install script:

### With Root Privileges (System-wide Installation)

```bash
# Install with automatic pacman hook setup
curl -fsSL https://raw.githubusercontent.com/vhqtvn/informant-go/main/install.sh | sudo bash
```

This will:
- Auto-detect your architecture (AMD64/ARM64)
- Download the latest release binary
- Verify SHA256 checksums for security
- Install to `/usr/local/bin/informant`
- Automatically setup pacman hook on Arch Linux

### Without Root Privileges (User Installation)

```bash
# Install to ~/.local/bin (no pacman hook)
curl -fsSL https://raw.githubusercontent.com/vhqtvn/informant-go/main/install.sh | bash
```

This will:
- Install to `~/.local/bin/informant`
- Remind you to add `~/.local/bin` to your PATH if needed
- Show instructions for manually installing pacman hook later

### Install Options

```bash
# Skip pacman hook installation
curl -fsSL https://raw.githubusercontent.com/vhqtvn/informant-go/main/install.sh | bash -s -- --no-hook

# See all options
curl -fsSL https://raw.githubusercontent.com/vhqtvn/informant-go/main/install.sh | bash -s -- --help
```

## Manual Installation from GitHub Releases

### 1. Download Pre-built Binary

Visit the [GitHub Releases](https://github.com/vhqtvn/informant-go/releases) page or download directly:

```bash
# For AMD64 (most common)
wget https://github.com/vhqtvn/informant-go/releases/latest/download/informant-linux-amd64

# For ARM64 (Raspberry Pi, etc.)
wget https://github.com/vhqtvn/informant-go/releases/latest/download/informant-linux-arm64

# Download checksums for verification
wget https://github.com/vhqtvn/informant-go/releases/latest/download/checksums.txt
```

### 2. Verify Checksum (Recommended)

```bash
# Verify the binary
sha256sum -c checksums.txt --ignore-missing
```

### 3. Install Binary

```bash
# Make executable and install
chmod +x informant-linux-*
sudo mv informant-linux-* /usr/local/bin/informant

# Or install to user directory
mkdir -p ~/.local/bin
mv informant-linux-* ~/.local/bin/informant
```

### 4. Install Pacman Hook (Arch Linux only)

```bash
# Install the pacman hook
sudo informant install
```

## Build from Source

### Prerequisites

- Go 1.19 or later
- git
- (Optional) UPX for compression: `sudo apt-get install upx-ucl`

### Building

```bash
# Clone the repository
git clone https://github.com/vhqtvn/informant-go.git
cd informant-go

# Build with Make
make build              # Standard build
make build-upx          # Build with UPX compression
make release            # Build release binaries for multiple architectures

# Or build directly with Go
go build -o informant .
```

### Install Built Binary

```bash
# Install system-wide
sudo make install

# Or copy manually
sudo cp informant /usr/local/bin/

# Install pacman hook
sudo informant install
```

## Installation Verification

After installation, verify everything works:

```bash
# Check version
informant --version

# Test basic functionality
informant list

# Verify pacman hook (Arch Linux)
ls -la /usr/share/libalpm/hooks/00-informant.hook
```

## Troubleshooting

### Command Not Found

If you get "command not found" after installation:

**For system installation:**
```bash
# Check if /usr/local/bin is in PATH
echo $PATH | grep -o '/usr/local/bin'

# If not found, add to your shell profile
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

**For user installation:**
```bash
# Check if ~/.local/bin is in PATH
echo $PATH | grep -o "$HOME/.local/bin"

# If not found, add to your shell profile
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Permission Issues

```bash
# If you get permission errors during installation
sudo chown $(whoami):$(whoami) /usr/local/bin/informant
chmod +x /usr/local/bin/informant
```

### Pacman Hook Issues

```bash
# Check if hook is installed
ls -la /usr/share/libalpm/hooks/00-informant.hook

# Manually install hook
sudo informant install --force

# Remove hook if needed
sudo informant uninstall
```

## Updating

### Using Install Script

```bash
# The install script always downloads the latest version
curl -fsSL https://raw.githubusercontent.com/vhqtvn/informant-go/main/install.sh | sudo bash
```

### Manual Update

```bash
# Download new version and replace existing
wget https://github.com/vhqtvn/informant-go/releases/latest/download/informant-linux-amd64
chmod +x informant-linux-amd64
sudo mv informant-linux-amd64 /usr/local/bin/informant
```

## Uninstalling

```bash
# Remove pacman hook
sudo informant uninstall

# Remove binary
sudo rm /usr/local/bin/informant

# Or for user installation
rm ~/.local/bin/informant

# Remove config and read status (optional)
rm -f ~/.informantrc.json ~/.informant_read_status
```

## Supported Platforms

- **Linux AMD64** - Standard 64-bit Linux systems
- **Linux ARM64** - ARM 64-bit systems (Raspberry Pi 4+, etc.)

The install script automatically detects your architecture and downloads the appropriate binary.

All release binaries are:
- Compiled with Go's static linking
- Compressed with UPX for smaller downloads (60-70% size reduction)
- Verified with SHA256 checksums
- Signed releases with embedded hook configuration 