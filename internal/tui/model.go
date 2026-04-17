package tui

import (
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/super-smooth/ntd/internal/config"
	"github.com/super-smooth/ntd/internal/deploy"
	"github.com/super-smooth/ntd/internal/flake"
	"github.com/super-smooth/ntd/internal/history"
	"github.com/super-smooth/ntd/internal/tailscale"
)

// State represents the current TUI state
type State int

const (
	StateSelectingFlake State = iota
	StateSelectingHost
	StateCompleted
)

// Model represents the TUI model
type Model struct {
	State   State
	Config  *config.Config
	Flake   *flake.Flake
	Hosts   []tailscale.Host
	History *history.History

	// Selections
	SelectedOutput string
	SelectedHost   *tailscale.Host

	// Lists
	FlakeList list.Model
	HostList  list.Model

	// Result
	Deployer *deploy.Deployer

	// Styling
	AppStyle   lipgloss.Style
	TitleStyle lipgloss.Style
}

// NewModel creates a new TUI model
func NewModel(cfg *config.Config, flakeData *flake.Flake, hosts []tailscale.Host, hist *history.History) Model {
	m := Model{
		State:   StateSelectingFlake,
		Config:  cfg,
		Flake:   flakeData,
		Hosts:   hosts,
		History: hist,
	}

	// Initialize styles
	m.AppStyle = lipgloss.NewStyle().Padding(1, 2)
	m.TitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true).
		Padding(0, 1)

	// Initialize flake list
	m.FlakeList = newFlakeList(flakeData.Outputs, hist)

	// Initialize host list
	m.HostList = newHostList(hosts, hist)

	return m
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window size first
	if windowMsg, ok := msg.(tea.WindowSizeMsg); ok {
		h, v := m.AppStyle.GetFrameSize()
		m.FlakeList.SetSize(windowMsg.Width-h, windowMsg.Height-v)
		m.HostList.SetSize(windowMsg.Width-h, windowMsg.Height-v)
		return m, nil
	}

	// Handle global keys
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	// Handle state-specific updates
	switch m.State {
	case StateSelectingFlake:
		// Handle flake selection
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				if i, ok := m.FlakeList.SelectedItem().(flakeItem); ok {
					m.SelectedOutput = i.name
					m.State = StateSelectingHost
				}
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.FlakeList, cmd = m.FlakeList.Update(msg)
		return m, cmd

	case StateSelectingHost:
		// Handle host selection
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				if i, ok := m.HostList.SelectedItem().(hostItem); ok {
					m.SelectedHost = &i.host

					// Generate the deployer
					isLocal := m.SelectedHost.IsCurrent
					m.Deployer = deploy.NewDeployer(
						m.Flake.Path,
						m.SelectedOutput,
						m.SelectedHost.Hostname,
						isLocal,
						m.Config.NoSudo,
					)

					// Write command to file if NTD_OUTPUT_FILE is set (best effort)
					if outputFile := os.Getenv("NTD_OUTPUT_FILE"); outputFile != "" {
						_ = os.WriteFile(outputFile, []byte(m.Deployer.GenerateCommand()), 0o644)
					}

					// Output command and exit immediately
					return m, tea.Quit
				}
				return m, nil
			case "esc":
				m.State = StateSelectingFlake
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.HostList, cmd = m.HostList.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the current view
func (m Model) View() string {
	switch m.State {
	case StateSelectingFlake:
		return m.AppStyle.Render(m.FlakeList.View())
	case StateSelectingHost:
		return m.AppStyle.Render(m.HostList.View())
	case StateCompleted:
		if m.Deployer != nil {
			// Output only the command to stdout (for shell wrapper)
			return m.Deployer.GenerateCommand()
		}
		return ""
	default:
		return "Unknown state"
	}
}
