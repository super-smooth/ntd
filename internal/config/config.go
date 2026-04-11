package config

import (
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	FlakePath string
	Output    string
	Host      string
	NoSudo    bool
}

// DefaultFlakePath returns the default flake path
func DefaultFlakePath() string {
	if env := os.Getenv("NTD_FLAKE"); env != "" {
		return env
	}
	return "."
}

// ConfigDir returns the directory for storing config files
func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".config", "ntd")
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() error {
	dir := ConfigDir()
	return os.MkdirAll(dir, 0o755)
}
