# ntd - NixOS Tailscale Deployer

A CLI/TUI tool that generates `nixos-rebuild` commands for deploying NixOS configurations to Tailscale-connected hosts.

## Features

- **Interactive TUI**: Select flake outputs and Tailscale hosts through a beautiful Bubble Tea interface
- **Smart Filtering**: Only shows Linux hosts from your Tailscale network (potential NixOS targets)
- **Current Host Detection**: Automatically detects and marks your current machine
- **Command Generation**: Generates proper `nixos-rebuild` commands with `--target-host` for remote hosts
- **Local Deployments**: Generates `sudo nixos-rebuild` commands for the current host
- **Recent Selections**: Remembers your last (output, host) combinations
- **Shell Integration**: Optional shell wrapper for automatic command history insertion

## Installation

### Homebrew

```bash
brew install super-smooth/tap/ntd
```

### From Source

```bash
git clone https://github.com/super-smooth/ntd
cd ntd
nix develop  # Enters dev shell and auto-generates gomod2nix.toml
nix build .#default
# Binary will be at result/bin/ntd
```

## Setup

### Shell Integration (Optional but Recommended)

To have commands automatically added to your shell history, add this to your shell config:

**For zsh (~/.zshrc):**
```bash
eval "$(ntd init)"
```

**For bash (~/.bashrc):**
```bash
eval "$(ntd init)"
```

Then reload your shell:
```bash
source ~/.zshrc  # or ~/.bashrc
```

Without the shell integration, `ntd` will simply output the command to stdout for you to copy/paste.

## Usage

### Interactive Mode (TUI)

Simply run `ntd` to start the interactive interface:

```bash
ntd
```

This will:
1. Show available NixOS configurations from your flake
2. Show Linux hosts from your Tailscale network (with current host marked)
3. Generate and output the appropriate `nixos-rebuild` command

**Navigation:**
- `↑/↓` or `j/k` - Navigate items
- `Enter` - Select item
- `/` - Filter/search
- `Esc` - Go back
- `Ctrl+C` - Quit

### CLI Mode

```bash
# Generate command for specific output and host
ntd --output desktop --host laptop

# Use a different flake path
ntd --flake ~/nixos-config --output server --host vps

# Generate command without sudo (for local deployments)
ntd --output desktop --host mymachine --no-sudo
```

### Environment Variables

- `NTD_FLAKE`: Default path to your NixOS flake (default: current directory)

## Generated Commands

**Local deployment** (when you select the current host):
```bash
# With sudo (default)
sudo nixos-rebuild switch --flake .#desktop

# Without sudo (--no-sudo flag)
nixos-rebuild switch --flake .#desktop
```

**Remote deployment** (when you select another host):
```bash
nixos-rebuild switch --flake .#desktop --target-host laptop --use-remote-sudo
```

## Requirements

- Nix with flakes enabled
- Tailscale CLI (`tailscale`) installed and configured
- SSH access to target hosts (Tailscale SSH or key-based) for remote deployments
- Target hosts must be running NixOS

## How It Works

1. Parse your flake to find `nixosConfigurations`
2. Query Tailscale for available Linux hosts
3. Let you select flake output and target host
4. Generate the appropriate `nixos-rebuild` command
5. Output the command (with shell integration: add to history)

## Development

### Initial Setup

When setting up the project for the first time, you need to generate the `gomod2nix.toml` lock file:

```bash
# Enter dev shell (automatically installs git hooks)
nix develop

# Generate the dependency lock file (only needed once)
gomod2nix generate

# Now you can build with Nix
nix build .#default
```

### Subsequent Development

```bash
# Enter dev shell
nix develop

# Run tests
go test ./...

# Build locally
go build -o ntd ./cmd/ntd

# Build with Nix (requires gomod2nix.toml to be up to date)
nix build .#default
```

### Handling Dependency Updates

This project uses [gomod2nix](https://github.com/nix-community/gomod2nix) to manage Go dependencies in a Nix-compatible way. When dependencies change:

1. **Automatic (via lefthook)**: When you commit changes to `go.mod` or `go.sum`, the pre-commit hook automatically runs `gomod2nix generate` and stages the updated `gomod2nix.toml` file.

2. **Manual**: If you need to update manually:
   ```bash
   gomod2nix generate
   git add gomod2nix.toml
   ```

The CI will verify that `gomod2nix.toml` is always in sync with `go.mod`.

## License

MIT

## Author

Made with ❤️ by the super-smooth team.
