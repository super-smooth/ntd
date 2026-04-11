package flake

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
)

// Flake represents a Nix flake
type Flake struct {
	Path    string
	Outputs []string
}

// Load parses a flake and extracts nixosConfigurations
func Load(path string) (*Flake, error) {
	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve flake path: %w", err)
	}

	// Run nix flake show
	cmd := exec.Command("nix", "flake", "show", "--json", absPath)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("nix flake show failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run nix flake show: %w", err)
	}

	// Parse JSON output
	var flakeData map[string]interface{}
	if err := json.Unmarshal(output, &flakeData); err != nil {
		return nil, fmt.Errorf("failed to parse flake output: %w", err)
	}

	// Extract nixosConfigurations
	nixosConfs, ok := flakeData["nixosConfigurations"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no nixosConfigurations found in flake")
	}

	outputs := make([]string, 0, len(nixosConfs))
	for name := range nixosConfs {
		outputs = append(outputs, name)
	}

	if len(outputs) == 0 {
		return nil, fmt.Errorf("no nixosConfigurations found in flake")
	}

	return &Flake{
		Path:    absPath,
		Outputs: outputs,
	}, nil
}
