// Copyright 2026 Kartoza
// SPDX-License-Identifier: MIT

package screens

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kartoza/kartoza-cloudbench/internal/config"
	"github.com/kartoza/kartoza-cloudbench/internal/sync"
	"github.com/kartoza/kartoza-cloudbench/internal/tui/styles"
)

// SyncKeyMap defines the key bindings for the sync screen
type SyncKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Enter  key.Binding
	Space  key.Binding
	Start  key.Binding
	Stop   key.Binding
	Escape key.Binding
	Tab    key.Binding
}

// DefaultSyncKeyMap returns the default key bindings
func DefaultSyncKeyMap() SyncKeyMap {
	return SyncKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "source panel"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "dest panel"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Space: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle"),
		),
		Start: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "start sync"),
		),
		Stop: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "stop"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch panel"),
		),
	}
}

// SyncPanel represents which panel is active
type SyncPanel int

const (
	PanelSource SyncPanel = iota
	PanelDestinations
	PanelOptions
)

// SyncProgressMsg is sent when sync progress updates
type SyncProgressMsg struct{}

// SyncScreen manages server synchronization
type SyncScreen struct {
	config         *config.Config
	keys           SyncKeyMap
	width          int
	height         int
	activePanel    SyncPanel
	sourceIdx      int
	destIdx        int
	selectedDests  map[string]bool
	syncOptions    config.SyncOptions
	optionIdx      int
	spinner        spinner.Model
	isRunning      bool
	statusMessage  string
	lastUpdateTime time.Time
}

// NewSyncScreen creates a new sync screen
func NewSyncScreen(cfg *config.Config) *SyncScreen {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.KartozaBlue)

	return &SyncScreen{
		config:        cfg,
		keys:          DefaultSyncKeyMap(),
		activePanel:   PanelSource,
		selectedDests: make(map[string]bool),
		syncOptions:   config.DefaultSyncOptions(),
		spinner:       s,
	}
}

// Init initializes the sync screen
func (s *SyncScreen) Init() tea.Cmd {
	return s.spinner.Tick
}

// Update handles messages
func (s *SyncScreen) Update(msg tea.Msg) (*SyncScreen, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		s.spinner, cmd = s.spinner.Update(msg)
		cmds = append(cmds, cmd)
		if s.isRunning {
			cmds = append(cmds, s.tickProgress())
		}

	case SyncProgressMsg:
		s.updateProgress()
		if s.isRunning {
			cmds = append(cmds, s.tickProgress())
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, s.keys.Tab), key.Matches(msg, s.keys.Right):
			s.activePanel = (s.activePanel + 1) % 3

		case key.Matches(msg, s.keys.Left):
			if s.activePanel > 0 {
				s.activePanel--
			} else {
				s.activePanel = PanelOptions
			}

		case key.Matches(msg, s.keys.Up):
			s.navigateUp()

		case key.Matches(msg, s.keys.Down):
			s.navigateDown()

		case key.Matches(msg, s.keys.Space), key.Matches(msg, s.keys.Enter):
			s.toggleSelection()

		case key.Matches(msg, s.keys.Start):
			if !s.isRunning {
				cmds = append(cmds, s.startSync())
			}

		case key.Matches(msg, s.keys.Stop):
			if s.isRunning {
				s.stopSync()
			}
		}
	}

	return s, tea.Batch(cmds...)
}

func (s *SyncScreen) navigateUp() {
	switch s.activePanel {
	case PanelSource:
		if s.sourceIdx > 0 {
			s.sourceIdx--
		}
	case PanelDestinations:
		if s.destIdx > 0 {
			s.destIdx--
		}
	case PanelOptions:
		if s.optionIdx > 0 {
			s.optionIdx--
		}
	}
}

func (s *SyncScreen) navigateDown() {
	switch s.activePanel {
	case PanelSource:
		if s.sourceIdx < len(s.config.Connections)-1 {
			s.sourceIdx++
		}
	case PanelDestinations:
		if s.destIdx < len(s.config.Connections)-1 {
			s.destIdx++
		}
	case PanelOptions:
		if s.optionIdx < 5 {
			s.optionIdx++
		}
	}
}

func (s *SyncScreen) toggleSelection() {
	switch s.activePanel {
	case PanelDestinations:
		if s.destIdx < len(s.config.Connections) {
			conn := s.config.Connections[s.destIdx]
			// Don't allow selecting source as destination
			if s.sourceIdx < len(s.config.Connections) && s.config.Connections[s.sourceIdx].ID == conn.ID {
				s.statusMessage = "Cannot sync to the same server"
				return
			}
			s.selectedDests[conn.ID] = !s.selectedDests[conn.ID]
		}
	case PanelOptions:
		switch s.optionIdx {
		case 0:
			s.syncOptions.Workspaces = !s.syncOptions.Workspaces
		case 1:
			s.syncOptions.DataStores = !s.syncOptions.DataStores
		case 2:
			s.syncOptions.CoverageStores = !s.syncOptions.CoverageStores
		case 3:
			s.syncOptions.Layers = !s.syncOptions.Layers
		case 4:
			s.syncOptions.Styles = !s.syncOptions.Styles
		case 5:
			s.syncOptions.LayerGroups = !s.syncOptions.LayerGroups
		}
	}
}

func (s *SyncScreen) tickProgress() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return SyncProgressMsg{}
	})
}

func (s *SyncScreen) updateProgress() {
	tasks := sync.DefaultManager.GetAllTasks()
	running := false
	for _, task := range tasks {
		if task.GetStatus() == "running" {
			running = true
			break
		}
	}
	s.isRunning = running
}

func (s *SyncScreen) startSync() tea.Cmd {
	if s.sourceIdx >= len(s.config.Connections) {
		s.statusMessage = "Please select a source server"
		return nil
	}

	sourceConn := &s.config.Connections[s.sourceIdx]

	destIDs := make([]string, 0)
	for id, selected := range s.selectedDests {
		if selected && id != sourceConn.ID {
			destIDs = append(destIDs, id)
		}
	}

	if len(destIDs) == 0 {
		s.statusMessage = "Please select at least one destination"
		return nil
	}

	s.isRunning = true
	s.statusMessage = "Starting sync..."

	// Start sync for each destination
	for _, destID := range destIDs {
		destConn := s.config.GetConnection(destID)
		if destConn != nil {
			sync.DefaultManager.StartSync(sourceConn, destConn, s.syncOptions, "")
		}
	}

	return s.tickProgress()
}

func (s *SyncScreen) stopSync() {
	sync.DefaultManager.StopAllTasks()
	s.isRunning = false
	s.statusMessage = "Sync stopped"
}

// View renders the sync screen
func (s *SyncScreen) View() string {
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.KartozaBlue).
		MarginBottom(1)

	header := headerStyle.Render("🔄 Server Synchronization")

	// Calculate panel widths
	panelWidth := (s.width - 6) / 3
	if panelWidth < 20 {
		panelWidth = 20
	}

	// Create panels
	sourcePanel := s.renderSourcePanel(panelWidth)
	destPanel := s.renderDestinationsPanel(panelWidth)
	optionsPanel := s.renderOptionsPanel(panelWidth)

	// Combine panels horizontally
	panels := lipgloss.JoinHorizontal(
		lipgloss.Top,
		sourcePanel,
		s.renderConnector(),
		destPanel,
		" ",
		optionsPanel,
	)

	// Progress section
	progress := s.renderProgress()

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(styles.Muted).
		MarginTop(1)

	var helpText string
	if s.isRunning {
		helpText = helpStyle.Render("x: stop • esc: back")
	} else {
		helpText = helpStyle.Render("tab: switch panel • space: toggle • s: start sync • esc: back")
	}

	// Status message
	statusStyle := lipgloss.NewStyle().
		Foreground(styles.KartozaOrange).
		MarginTop(1)

	status := ""
	if s.statusMessage != "" {
		status = statusStyle.Render(s.statusMessage)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		panels,
		progress,
		status,
		helpText,
	)
}

func (s *SyncScreen) renderSourcePanel(width int) string {
	borderColor := styles.Border
	if s.activePanel == PanelSource {
		borderColor = styles.KartozaBlue
	}

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width).
		Padding(0, 1)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Success).
		MarginBottom(1)

	title := titleStyle.Render("📤 Source (Read Only)")

	var items []string
	for i, conn := range s.config.Connections {
		marker := "  "
		if i == s.sourceIdx {
			marker = "→ "
		}
		style := lipgloss.NewStyle()
		if i == s.sourceIdx && s.activePanel == PanelSource {
			style = style.Background(styles.KartozaBlue).Foreground(styles.TextBright)
		}
		items = append(items, style.Render(marker+conn.Name))
	}

	content := title + "\n" + strings.Join(items, "\n")
	return panelStyle.Render(content)
}

func (s *SyncScreen) renderConnector() string {
	if s.isRunning {
		return s.spinner.View() + " → "
	}
	return "  →  "
}

func (s *SyncScreen) renderDestinationsPanel(width int) string {
	borderColor := styles.Border
	if s.activePanel == PanelDestinations {
		borderColor = styles.KartozaBlue
	}

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width).
		Padding(0, 1)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.KartozaBlueLight).
		MarginBottom(1)

	title := titleStyle.Render("📥 Destinations")

	var sourceID string
	if s.sourceIdx < len(s.config.Connections) {
		sourceID = s.config.Connections[s.sourceIdx].ID
	}

	var items []string
	for i, conn := range s.config.Connections {
		if conn.ID == sourceID {
			continue // Don't show source as destination
		}

		marker := "[ ] "
		if s.selectedDests[conn.ID] {
			marker = "[\uf00c] " // fa-check
		}
		if i == s.destIdx && s.activePanel == PanelDestinations {
			marker = "\uf0da" + marker[1:] // fa-caret-right
		}

		style := lipgloss.NewStyle()
		if i == s.destIdx && s.activePanel == PanelDestinations {
			style = style.Background(styles.KartozaBlueLight).Foreground(styles.TextBright)
		}
		items = append(items, style.Render(marker+conn.Name))
	}

	if len(items) == 0 {
		items = append(items, lipgloss.NewStyle().Foreground(styles.Muted).Render("No other servers available"))
	}

	content := title + "\n" + strings.Join(items, "\n")
	return panelStyle.Render(content)
}

func (s *SyncScreen) renderOptionsPanel(width int) string {
	borderColor := styles.Border
	if s.activePanel == PanelOptions {
		borderColor = styles.KartozaBlue
	}

	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width).
		Padding(0, 1)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.KartozaOrange).
		MarginBottom(1)

	title := titleStyle.Render("\uf013  Sync Options") // fa-cog

	options := []struct {
		name    string
		enabled bool
	}{
		{"Workspaces", s.syncOptions.Workspaces},
		{"Data Stores", s.syncOptions.DataStores},
		{"Coverage Stores", s.syncOptions.CoverageStores},
		{"Layers", s.syncOptions.Layers},
		{"Styles", s.syncOptions.Styles},
		{"Layer Groups", s.syncOptions.LayerGroups},
	}

	var items []string
	for i, opt := range options {
		marker := "[ ] "
		if opt.enabled {
			marker = "[\uf00c] " // fa-check
		}
		if i == s.optionIdx && s.activePanel == PanelOptions {
			marker = "\uf0da" + marker[1:] // fa-caret-right
		}

		style := lipgloss.NewStyle()
		if i == s.optionIdx && s.activePanel == PanelOptions {
			style = style.Background(styles.KartozaOrange).Foreground(styles.TextBright)
		}
		items = append(items, style.Render(marker+opt.name))
	}

	content := title + "\n" + strings.Join(items, "\n")
	return panelStyle.Render(content)
}

func (s *SyncScreen) renderProgress() string {
	tasks := sync.DefaultManager.GetAllTasks()
	if len(tasks) == 0 {
		return ""
	}

	progressStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.KartozaBlue).
		Padding(0, 1).
		MarginTop(1)

	var lines []string
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render("Progress:"))

	for _, task := range tasks {
		status := task.GetStatus()
		progress := task.GetProgress()
		currentItem := task.GetCurrentItem()

		statusIcon := "\uf110" // fa-spinner
		switch status {
		case "completed":
			statusIcon = "\uf00c" // fa-check
		case "failed":
			statusIcon = "\uf00d" // fa-times
		case "stopped":
			statusIcon = "\uf04d" // fa-stop
		}

		progressBar := s.renderProgressBar(int(progress), 20)
		lines = append(lines, fmt.Sprintf("%s %s %.0f%% %s",
			statusIcon,
			progressBar,
			progress,
			currentItem,
		))
	}

	return progressStyle.Render(strings.Join(lines, "\n"))
}

func (s *SyncScreen) renderProgressBar(percent, width int) string {
	filled := (percent * width) / 100
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return "[" + bar + "]"
}

// GetSelectedSourceID returns the selected source connection ID
func (s *SyncScreen) GetSelectedSourceID() string {
	if s.sourceIdx < len(s.config.Connections) {
		return s.config.Connections[s.sourceIdx].ID
	}
	return ""
}

// GetSelectedDestIDs returns the selected destination connection IDs
func (s *SyncScreen) GetSelectedDestIDs() []string {
	var ids []string
	for id, selected := range s.selectedDests {
		if selected {
			ids = append(ids, id)
		}
	}
	return ids
}

// SetSize sets the screen size
func (s *SyncScreen) SetSize(width, height int) {
	s.width = width
	s.height = height
}
