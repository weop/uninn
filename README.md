# Uninn - Universal Package Uninstaller

A TUI app for managing and uninstalling applications installed through various package managers on Linux.

## Features

- Support for multiple package managers:
  - APT
  - Pacman
  - Flatpak
  - Snap
  - AppImages
- Interactive TUI with search functionality
- Automatic detection of available package managers
- Real-time package list updates after uninstallation

## Installation

```bash
# Clone the repository
git clone https://github.com/weop/uninn.git
cd uninn

# Build the application
go build -o uninn

# Make it executable
chmod +x uninn

# Optionally, move to PATH
sudo mv uninn /usr/local/bin/
```

## Usage

Simply run the application:

```bash
./uninn
```

Or if installed to PATH:

```bash
uninn
```

### Controls

- `↑/↓` or `j/k`: Navigate through the package list
- `Enter`: Select package for uninstallation
- `/`: Search packages
- `y`: Confirm uninstallation
- `n` or `Esc`: Cancel uninstallation
- `q` or `Ctrl+C`: Quit the application

## Requirements

- Go 1.21 or higher
- Linux operating system
- One or more of the supported package managers installed
- `pkexec` for privilege escalation (usually comes with PolicyKit)

## Security

The application uses `pkexec` to request administrative privileges only when needed for uninstalling packages. It never runs entirely as root.

## Building from Source

```bash
go mod download
go build -o uninn
```
