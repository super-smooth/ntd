package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/super-smooth/ntd/internal/history"
	"github.com/super-smooth/ntd/internal/tailscale"
)

// hostItem represents a Tailscale host item
type hostItem struct {
	host   tailscale.Host
	recent bool
}

func (i hostItem) Title() string {
	prefix := "  "
	if i.recent {
		prefix = "★ "
	}

	// Add indicator for current host
	hostLabel := i.host.Hostname
	if i.host.IsCurrent {
		hostLabel = hostLabel + " (current)"
	}

	return fmt.Sprintf("%s%s (%s)", prefix, hostLabel, i.host.IP)
}

func (i hostItem) Description() string { return "" }
func (i hostItem) FilterValue() string { return i.host.Hostname }

// newHostList creates a list of Tailscale hosts (filtered to Linux only)
func newHostList(hosts []tailscale.Host, hist *history.History) list.Model {
	// Filter to only Linux hosts
	linuxHosts := tailscale.FilterLinuxHosts(hosts)

	items := make([]list.Item, 0, len(linuxHosts))

	// Add recent items first (matching current flake output)
	if hist.HasRecent() {
		for _, entry := range hist.GetRecent() {
			for _, host := range linuxHosts {
				if host.Hostname == entry.Host {
					items = append(items, hostItem{host: host, recent: true})
					break
				}
			}
		}
	}

	// Add remaining hosts
	for _, host := range linuxHosts {
		found := false
		for _, item := range items {
			if hi, ok := item.(hostItem); ok && hi.host.Hostname == host.Hostname {
				found = true
				break
			}
		}
		if !found {
			items = append(items, hostItem{host: host, recent: false})
		}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a Tailscale host (Linux only)"
	l.Styles.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))

	return l
}
