package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/super-smooth/ntd/internal/config"
)

// Entry represents a single history entry
type Entry struct {
	Output    string    `json:"output"`
	Host      string    `json:"host"`
	Timestamp time.Time `json:"timestamp"`
}

// History manages recent selections
type History struct {
	Recent []Entry `json:"recent"`
}

// Load reads the history from disk
func Load() (*History, error) {
	historyPath := getHistoryPath()

	data, err := os.ReadFile(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &History{Recent: []Entry{}}, nil
		}
		return nil, fmt.Errorf("failed to read history: %w", err)
	}

	var history History
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to parse history: %w", err)
	}

	return &history, nil
}

// Save writes the history to disk
func (h *History) Save() error {
	if err := config.EnsureConfigDir(); err != nil {
		return err
	}

	historyPath := getHistoryPath()
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	if err := os.WriteFile(historyPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write history: %w", err)
	}

	return nil
}

// Add adds a new entry to history, keeping only the 5 most recent unique entries
func (h *History) Add(output, host string) {
	// Remove existing entry if present
	for i, entry := range h.Recent {
		if entry.Output == output && entry.Host == host {
			h.Recent = append(h.Recent[:i], h.Recent[i+1:]...)
			break
		}
	}

	// Add new entry at the beginning
	newEntry := Entry{
		Output:    output,
		Host:      host,
		Timestamp: time.Now(),
	}
	h.Recent = append([]Entry{newEntry}, h.Recent...)

	// Keep only 5 entries
	if len(h.Recent) > 5 {
		h.Recent = h.Recent[:5]
	}
}

// GetRecent returns the recent entries
func (h *History) GetRecent() []Entry {
	return h.Recent
}

// HasRecent checks if there are any recent entries
func (h *History) HasRecent() bool {
	return len(h.Recent) > 0
}

func getHistoryPath() string {
	return filepath.Join(config.ConfigDir(), "history.json")
}
