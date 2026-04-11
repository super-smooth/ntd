package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/super-smooth/ntd/internal/history"
)

// flakeItem represents a flake output item
type flakeItem struct {
	name   string
	recent bool
}

func (i flakeItem) Title() string {
	if i.recent {
		return "★ " + i.name + " (recent)"
	}
	return "  " + i.name
}

func (i flakeItem) Description() string { return "" }
func (i flakeItem) FilterValue() string { return i.name }

// newFlakeList creates a list of flake outputs
func newFlakeList(outputs []string, hist *history.History) list.Model {
	items := make([]list.Item, 0, len(outputs))

	// Add recent items first
	if hist.HasRecent() {
		recentMap := make(map[string]bool)
		for _, entry := range hist.GetRecent() {
			if !recentMap[entry.Output] {
				for _, output := range outputs {
					if output == entry.Output {
						items = append(items, flakeItem{name: output, recent: true})
						recentMap[output] = true
						break
					}
				}
			}
		}
	}

	// Add remaining items
	for _, output := range outputs {
		found := false
		for _, item := range items {
			if fi, ok := item.(flakeItem); ok && fi.name == output {
				found = true
				break
			}
		}
		if !found {
			items = append(items, flakeItem{name: output, recent: false})
		}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a NixOS configuration"
	l.Styles.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))

	return l
}
