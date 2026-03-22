package screens

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kartoza/kartoza-cloudbench/internal/tui/styles"
)

// HostingInstance represents a hosted instance
type HostingInstance struct {
	ID          string
	Name        string
	ProductName string
	PackageName string
	Status      string
	URL         string
	CreatedAt   time.Time
}

// HostingKeyMap defines the key bindings
type HostingKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Refresh key.Binding
	Shop    key.Binding
	Escape  key.Binding
}

// DefaultHostingKeyMap returns the default key bindings
func DefaultHostingKeyMap() HostingKeyMap {
	return HostingKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "view details"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Shop: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "open shop"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
	}
}

// HostingScreen is the TUI screen for geospatial hosting
type HostingScreen struct {
	width     int
	height    int
	spinner   spinner.Model
	instances []HostingInstance
	cursor    int
	loading   bool
	error     string
	keys      HostingKeyMap
}

// NewHostingScreen creates a new hosting screen
func NewHostingScreen() HostingScreen {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return HostingScreen{
		spinner:   s,
		instances: []HostingInstance{},
		cursor:    0,
		loading:   true,
		keys:      DefaultHostingKeyMap(),
	}
}

// Init initializes the hosting screen
func (h HostingScreen) Init() tea.Cmd {
	return tea.Batch(
		h.spinner.Tick,
		h.fetchInstances,
	)
}

// fetchInstances simulates fetching instances
func (h HostingScreen) fetchInstances() tea.Msg {
	// TODO: Implement actual API call to hosting backend
	// For now, return demo data
	time.Sleep(500 * time.Millisecond)

	return hostingFetchedMsg{
		instances: []HostingInstance{
			// Demo instances - replace with actual API call
		},
	}
}

type hostingFetchedMsg struct {
	instances []HostingInstance
	err       error
}

// Update handles messages
func (h HostingScreen) Update(msg tea.Msg) (HostingScreen, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, h.keys.Up):
			if h.cursor > 0 {
				h.cursor--
			}
		case key.Matches(msg, h.keys.Down):
			if h.cursor < len(h.instances)-1 {
				h.cursor++
			}
		case key.Matches(msg, h.keys.Refresh):
			h.loading = true
			return h, h.fetchInstances
		case key.Matches(msg, h.keys.Shop):
			// TODO: Open shop/browser
			return h, nil
		}

	case hostingFetchedMsg:
		h.loading = false
		if msg.err != nil {
			h.error = msg.err.Error()
		} else {
			h.instances = msg.instances
		}

	case spinner.TickMsg:
		if h.loading {
			h.spinner, cmd = h.spinner.Update(msg)
			return h, cmd
		}
	}

	return h, cmd
}

func (h HostingScreen) statusBadge(status string) string {
	switch status {
	case "online":
		return styles.SuccessStyle.Render("● Online")
	case "starting_up", "deploying":
		return styles.AccentStyle.Render("● Starting")
	case "offline":
		return styles.MutedStyle.Render("○ Offline")
	case "error":
		return styles.ErrorStyle.Render("● Error")
	default:
		return styles.MutedStyle.Render("○ " + status)
	}
}

// View renders the hosting screen
func (h HostingScreen) View() string {
	var b strings.Builder

	// Title
	b.WriteString(styles.DialogTitleStyle.Render("☁️  Geospatial Hosting"))
	b.WriteString("\n\n")

	if h.loading {
		b.WriteString(h.spinner.View() + " Loading instances...")
		b.WriteString("\n")
	} else if h.error != "" {
		b.WriteString(styles.ErrorStyle.Render(fmt.Sprintf("Error: %s", h.error)))
		b.WriteString("\n")
	} else if len(h.instances) == 0 {
		b.WriteString(styles.MutedStyle.Render("No instances found. Press 's' to open the shop and deploy your first instance."))
		b.WriteString("\n\n")

		// Show shop info
		b.WriteString(styles.AccentStyle.Render("Available Products:"))
		b.WriteString("\n")
		b.WriteString(styles.ItemStyle.Render("  • GeoServer - Web mapping server"))
		b.WriteString("\n")
		b.WriteString(styles.ItemStyle.Render("  • GeoNode - Geospatial CMS"))
		b.WriteString("\n")
		b.WriteString(styles.ItemStyle.Render("  • PostGIS - Spatial database"))
		b.WriteString("\n")
	} else {
		// Render instance list
		b.WriteString(h.renderInstanceList())
	}

	// Help
	b.WriteString("\n")
	b.WriteString(styles.HelpTextStyle.Render("↑/↓: navigate • r: refresh • s: shop • esc: back"))

	dialog := styles.DialogStyle.Width(h.width - 10).Render(b.String())
	return styles.Center(h.width, h.height, dialog)
}

// renderInstanceList renders the list of instances
func (h HostingScreen) renderInstanceList() string {
	var b strings.Builder

	for i, inst := range h.instances {
		isSelected := i == h.cursor

		// Build instance info
		info := fmt.Sprintf("%s (%s/%s)\n  %s\n  %s",
			inst.Name,
			inst.ProductName,
			inst.PackageName,
			styles.MutedStyle.Render(inst.URL),
			h.statusBadge(inst.Status),
		)

		// Apply styling
		var style lipgloss.Style
		if isSelected {
			style = styles.ActiveItemStyle
		} else {
			style = styles.ItemStyle
		}

		b.WriteString(style.Render(info))
		b.WriteString("\n\n")
	}

	return b.String()
}

// SelectedInstance returns the currently selected instance
func (h HostingScreen) SelectedInstance() *HostingInstance {
	if len(h.instances) == 0 {
		return nil
	}
	if h.cursor < 0 || h.cursor >= len(h.instances) {
		return nil
	}
	return &h.instances[h.cursor]
}

// SetSize sets the screen dimensions
func (h *HostingScreen) SetSize(width, height int) {
	h.width = width
	h.height = height
}
