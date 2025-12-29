# SeaPower-Calculator

A terminal-based calculator for Sea Power game weapon data analysis. This interactive TUI (Terminal User Interface) application allows you to browse weapons, calculate flight times and ranges so that all weapons reach their target at the same time.

## Features

- Browse and search through Sea Power game weapons
- Calculate weapon flight times based on range and velocity
- Delta time mode for comparing different ranges
- Interactive terminal interface with keyboard navigation
- Cross-platform support (Windows, macOS, Linux)

## Requirements

- **Sea Power game** installed on your system
- The game data files must be accessible in the Sea Power installation directory

## Download

### Option 1: Download Pre-built Executables (Recommended)

Download the latest release for your platform from the [Releases page](https://github.com/stevenzieske/SeaPower-Calculator/releases):

- **Windows**: `seapower-calculator-windows-amd64.exe`
- **macOS Intel**: `seapower-calculator-darwin-amd64`
- **macOS Apple Silicon**: `seapower-calculator-darwin-arm64`
- **Linux (64-bit)**: `seapower-calculator-linux-amd64`
- **Linux (ARM64)**: `seapower-calculator-linux-arm64`

### Option 2: Build from Source

If you have Go 1.25.4 or later installed:

```bash
git clone https://github.com/stevenzieske/SeaPower-Calculator.git
cd SeaPower-Calculator
go build -o seapower-calculator .
```

For cross-compilation to other platforms:

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o seapower-calculator.exe

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o seapower-calculator

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o seapower-calculator

# Linux
GOOS=linux GOARCH=amd64 go build -o seapower-calculator
```

## Setup

### 1. Locate Your Sea Power Installation

Find your Sea Power game installation directory. Common locations:

- **Windows**: `C:\Program Files (x86)\Steam\steamapps\common\`
- **macOS**: `~/Library/Application Support/Steam/steamapps/common/`
- **Linux**: `~/.steam/steam/steamapps/common/`

### 2. Configure the Game Directory

Create a `.env` file in the same directory as the `seapower-calculator` executable with the following content:

```env
GAME_DIR=/path/to/your/game/installation
```

**Example for Windows:**
```env
GAME_DIR=C:\Program Files (x86)\Steam\steamapps\common
```

**Example for macOS:**
```env
GAME_DIR=/Users/yourusername/Library/Application Support/Steam/steamapps/common
```

**Example for Linux:**
```env
GAME_DIR=/home/yourusername/.steam/steam/steamapps/common
```

The application will look for game data in: `GAME_DIR/Sea Power/Sea Power_Data/StreamingAssets/original/`

## Usage

### Running the Application

Simply execute the downloaded or compiled binary:

**Windows:**
```cmd
seapower-calculator-windows-amd64.exe
```

**macOS/Linux:**
```bash
./seapower-calculator-darwin-amd64
```

(Make sure to make it executable first on macOS/Linux: `chmod +x seapower-calculator-darwin-amd64`)

### Navigation

Once the application starts, you'll see an interactive table with weapon data:

#### Keyboard Controls

- **Arrow Keys (↑/↓) (j/k)**: Navigate through the weapon / selection list
- **Arrow Keys (↑/↓) (h/l)**: Navigate between left / right section
- **(g/G)**: Jump to the beginning / end of the table / selection
- **/**: In the right section: enter search input, in the left section: enter range input fields
- **Tab**: Switch between Flight Time and Delta Time calculation modes
- **Enter**: Select a weapon to view detailed calculations
- **Bacspace / Delete**: Remove Selection from list
- **Esc**: Exit text input
- **q**: Quit the application

#### Calculation Modes

**Flight Time Mode:**
- Enter a range value for the primary weapon to calculate the flight time. Based on the flight time, the range is calculated for all other weapons
- Shows: Range, Max Velocity, and calculated Flight Time

**Delta Time Mode:**
- Enter range values to compare flight times
- Shows: Ranges, velocities, flight times, and the delta (difference) between them
- Useful for understanding how flight time changes over different engagement ranges

### Understanding the Data

The application displays the following weapon properties:

- **Weapon Name**: Full designation of the weapon
- **Type**: Weapon classification
- **Target Type**: Intended target category
- **Max Velocity**: Maximum speed (knots)
- **Min/Max Launch Range**: Engagement envelope (nm)

## Troubleshooting

### "Empty table" or "No weapons found"

- Verify your `.env` file is in the same directory as the executable
- Check that `GAME_DIR` path is correct and points to your Steam common folder
- Ensure Sea Power game data files exist at: `GAME_DIR/Sea Power/Sea Power_Data/StreamingAssets/original/`

### Application doesn't start

- Ensure you have the correct executable for your platform and architecture
- On macOS/Linux, verify the file has execute permissions: `chmod +x seapower-calculator-*`

## Development

Built with:
- [Go](https://golang.org/) - Programming language
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions

## Disclosure

This project makes no claim to accuracy or future development.

This project is independent and not affiliated with the Sea Power game or its developers.

This project has been created with the use of Claude Sonnet 4.5
