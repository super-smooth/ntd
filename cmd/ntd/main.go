package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/super-smooth/ntd/internal/config"
	"github.com/super-smooth/ntd/internal/deploy"
	"github.com/super-smooth/ntd/internal/flake"
	"github.com/super-smooth/ntd/internal/history"
	"github.com/super-smooth/ntd/internal/tailscale"
	"github.com/super-smooth/ntd/internal/tui"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "ntd",
		Short: "NixOS Tailscale Deployer - Generate nixos-rebuild commands",
		Long: `ntd is a CLI/TUI tool that helps generate nixos-rebuild commands for Tailscale-connected hosts.

It provides an interactive interface to select flake outputs and Tailscale hosts,
then outputs the command to stdout for execution.

Usage:
  eval "$(ntd init)"  # Add to your shell config
  ntd                 # Select options interactively, command appears in shell buffer`,
		RunE: run,
	}

	// Define flags
	rootCmd.Flags().StringP("flake", "f", ".", "Path to NixOS flake (default: . or NTD_FLAKE env var)")
	rootCmd.Flags().StringP("output", "o", "", "NixOS configuration output name")
	rootCmd.Flags().String("host", "", "Target Tailscale hostname")
	rootCmd.Flags().Bool("no-sudo", false, "Omit sudo for local deployments")
	rootCmd.Flags().BoolP("version", "v", false, "Show version")

	// Add init command for shell integration
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Print shell function for integration",
		Long:  `Prints the shell function to be evaluated in your shell config (e.g., ~/.bashrc or ~/.zshrc)`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(shellFunc)
		},
	}
	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

const shellFunc = `
ntd() {
  # Create temp file for command output
  local tmpfile=$(mktemp)
  
  # Run ntd with temp file path, TUI displays normally
  NTD_OUTPUT_FILE="$tmpfile" command ntd "$@"
  
  # Read the command from temp file
  local cmd=""
  if [ -f "$tmpfile" ]; then
    cmd=$(cat "$tmpfile")
    rm -f "$tmpfile"
  fi
  
  # If we got a command back, push it to the input buffer
  if [ -n "$cmd" ]; then
    # zsh: push to input buffer using zle
    if [ -n "$ZSH_VERSION" ]; then
      print -z "$cmd"
    else
      # bash: print command for user to copy
      echo "$cmd"
    fi
  fi
}
`

func run(cmd *cobra.Command, args []string) error {
	// Handle version flag
	showVersion, _ := cmd.Flags().GetBool("version")
	if showVersion {
		fmt.Printf("ntd version %s\n", version)
		return nil
	}

	// Build config from flags
	cfg := &config.Config{
		FlakePath: config.DefaultFlakePath(),
	}

	if flakeFlag, _ := cmd.Flags().GetString("flake"); flakeFlag != "." {
		cfg.FlakePath = flakeFlag
	}
	cfg.Output, _ = cmd.Flags().GetString("output")
	cfg.Host, _ = cmd.Flags().GetString("host")
	cfg.NoSudo, _ = cmd.Flags().GetBool("no-sudo")

	// Load flake data
	flakeData, err := flake.Load(cfg.FlakePath)
	if err != nil {
		return fmt.Errorf("failed to load flake: %w", err)
	}

	// If output is specified, validate it
	if cfg.Output != "" {
		found := false
		for _, output := range flakeData.Outputs {
			if output == cfg.Output {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("output '%s' not found in flake", cfg.Output)
		}
	}

	// Get Tailscale hosts
	hosts, err := tailscale.GetHosts()
	if err != nil {
		return fmt.Errorf("failed to get Tailscale hosts: %w", err)
	}

	// Filter to Linux hosts only
	linuxHosts := tailscale.FilterLinuxHosts(hosts)
	if len(linuxHosts) == 0 {
		return fmt.Errorf("no Linux hosts found in Tailscale network")
	}

	// If host is specified, validate it
	if cfg.Host != "" {
		found := false
		for _, host := range linuxHosts {
			if host.Hostname == cfg.Host {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("host '%s' not found in Tailscale network (Linux hosts only)", cfg.Host)
		}
	}

	// Load history
	hist, err := history.Load()
	if err != nil {
		hist = &history.History{Recent: []history.Entry{}}
	}

	// Check if we can run in CLI-only mode
	if cfg.Output != "" && cfg.Host != "" {
		return runCLIMode(cfg, flakeData, linuxHosts, hist)
	}

	// Run TUI mode
	return runTUIMode(cfg, flakeData, linuxHosts, hist)
}

func runCLIMode(cfg *config.Config, flakeData *flake.Flake, hosts []tailscale.Host, hist *history.History) error {
	// Find the selected host
	var selectedHost *tailscale.Host
	for _, host := range hosts {
		if host.Hostname == cfg.Host {
			selectedHost = &host
			break
		}
	}

	if selectedHost == nil {
		return fmt.Errorf("host '%s' not found", cfg.Host)
	}

	// Generate command
	isLocal := selectedHost.IsCurrent
	deployer := deploy.NewDeployer(flakeData.Path, cfg.Output, selectedHost.Hostname, isLocal, cfg.NoSudo)
	cmd := deployer.GenerateCommand()

	// Write to file if NTD_OUTPUT_FILE is set, otherwise print to stdout
	if outputFile := os.Getenv("NTD_OUTPUT_FILE"); outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(cmd), 0o644); err != nil {
			return fmt.Errorf("failed to write command to file: %w", err)
		}
	} else {
		fmt.Println(cmd)
	}

	// Save to history
	hist.Add(cfg.Output, cfg.Host)
	_ = hist.Save()

	return nil
}

func runTUIMode(cfg *config.Config, flakeData *flake.Flake, hosts []tailscale.Host, hist *history.History) error {
	model := tui.NewModel(cfg, flakeData, hosts, hist)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Check final state and output command if successful
	if m, ok := finalModel.(tui.Model); ok {
		if m.State == tui.StateCompleted && m.Deployer != nil {
			// Output the generated command
			fmt.Println(m.Deployer.GenerateCommand())

			// Save to history
			if m.SelectedHost != nil {
				hist.Add(m.SelectedOutput, m.SelectedHost.Hostname)
				_ = hist.Save()
			}
		}
	}

	return nil
}
