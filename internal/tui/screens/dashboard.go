package screens

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kartoza/kartoza-geoserver-client/internal/api"
	"github.com/kartoza/kartoza-geoserver-client/internal/config"
)

// ServerStatus holds the status of a single server
type ServerStatus struct {
	ConnectionID   string
	ConnectionName string
	URL            string
	Online         bool
	ResponseTimeMs int64
	MemoryUsed     int64
	MemoryTotal    int64
	MemoryUsedPct  float64
	WorkspaceCount int
	LayerCount     int
	DataStoreCount int
	StyleCount     int
	Version        string
	Error          string
}

// DashboardKeyMap defines the key bindings
type DashboardKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Refresh key.Binding
	Enter   key.Binding
	Escape  key.Binding
}

// DefaultDashboardKeyMap returns the default key bindings
func DefaultDashboardKeyMap() DashboardKeyMap {
	return DashboardKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
	}
}

// DashboardRefreshMsg triggers a status refresh
type DashboardRefreshMsg struct{}

// DashboardStatusMsg contains updated server statuses
type DashboardStatusMsg struct {
	Statuses []ServerStatus
}

// DashboardScreen shows server status overview
type DashboardScreen struct {
	config         *config.Config
	keys           DashboardKeyMap
	width          int
	height         int
	selectedIdx    int
	statuses       []ServerStatus
	loading        bool
	lastRefresh    time.Time
	spinner        spinner.Model
	mu             sync.RWMutex
}

// NewDashboardScreen creates a new dashboard screen
func NewDashboardScreen(cfg *config.Config) *DashboardScreen {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#38B2AC"))

	return &DashboardScreen{
		config:   cfg,
		keys:     DefaultDashboardKeyMap(),
		statuses: make([]ServerStatus, 0),
		loading:  true,
		spinner:  s,
	}
}

// Init initializes the dashboard
func (d *DashboardScreen) Init() tea.Cmd {
	return tea.Batch(
		d.spinner.Tick,
		d.refreshStatus(),
	)
}

// Update handles messages
func (d *DashboardScreen) Update(msg tea.Msg) (*DashboardScreen, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		d.spinner, cmd = d.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case DashboardStatusMsg:
		d.mu.Lock()
		d.statuses = msg.Statuses
		d.loading = false
		d.lastRefresh = time.Now()
		d.mu.Unlock()

	case DashboardRefreshMsg:
		d.loading = true
		cmds = append(cmds, d.refreshStatus())

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, d.keys.Up):
			if d.selectedIdx > 0 {
				d.selectedIdx--
			}
		case key.Matches(msg, d.keys.Down):
			if d.selectedIdx < len(d.statuses)-1 {
				d.selectedIdx++
			}
		case key.Matches(msg, d.keys.Refresh):
			d.loading = true
			cmds = append(cmds, d.refreshStatus())
		}
	}

	return d, tea.Batch(cmds...)
}

// refreshStatus fetches status for all connections
func (d *DashboardScreen) refreshStatus() tea.Cmd {
	return func() tea.Msg {
		var wg sync.WaitGroup
		statusChan := make(chan ServerStatus, len(d.config.Connections))

		for _, conn := range d.config.Connections {
			wg.Add(1)
			go func(conn config.Connection) {
				defer wg.Done()
				status := d.fetchServerStatus(&conn)
				statusChan <- status
			}(conn)
		}

		go func() {
			wg.Wait()
			close(statusChan)
		}()

		var statuses []ServerStatus
		for status := range statusChan {
			statuses = append(statuses, status)
		}

		return DashboardStatusMsg{Statuses: statuses}
	}
}

func (d *DashboardScreen) fetchServerStatus(conn *config.Connection) ServerStatus {
	client := api.NewClient(conn)
	status := ServerStatus{
		ConnectionID:   conn.ID,
		ConnectionName: conn.Name,
		URL:            conn.URL,
		Online:         false,
	}

	startTime := time.Now()

	// Get server status
	serverStatus, err := client.GetServerStatus()
	if err != nil {
		status.Error = err.Error()
		return status
	}

	status.ResponseTimeMs = time.Since(startTime).Milliseconds()
	status.Online = serverStatus.Online
	status.MemoryUsed = serverStatus.MemoryUsed
	status.MemoryTotal = serverStatus.MemoryTotal
	status.MemoryUsedPct = serverStatus.MemoryUsedPct
	status.Version = serverStatus.GeoServerVersion
	status.Error = serverStatus.Error

	if !status.Online {
		return status
	}

	// Get counts
	counts, _ := client.GetServerCounts()
	if counts != nil {
		status.WorkspaceCount = counts.WorkspaceCount
		status.LayerCount = counts.LayerCount
		status.DataStoreCount = counts.DataStoreCount
		status.StyleCount = counts.StyleCount
	}

	return status
}

// View renders the dashboard
func (d *DashboardScreen) View() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#38B2AC")).
		MarginBottom(1)

	header := headerStyle.Render("📊 Server Dashboard")

	// Loading indicator
	if d.loading && len(d.statuses) == 0 {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			d.spinner.View()+" Loading server status...",
		)
	}

	// Stats summary
	onlineCount := 0
	offlineCount := 0
	totalLayers := 0
	totalStores := 0

	for _, s := range d.statuses {
		if s.Online {
			onlineCount++
			totalLayers += s.LayerCount
			totalStores += s.DataStoreCount
		} else {
			offlineCount++
		}
	}

	summaryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		MarginBottom(1)

	summary := summaryStyle.Render(fmt.Sprintf(
		"✓ %d Online  ✗ %d Offline  📦 %d Layers  🗄 %d Stores",
		onlineCount, offlineCount, totalLayers, totalStores,
	))

	// Alert section for offline servers
	var alertSection string
	var alertServers []ServerStatus
	var healthyServers []ServerStatus

	for _, s := range d.statuses {
		if !s.Online {
			alertServers = append(alertServers, s)
		} else {
			healthyServers = append(healthyServers, s)
		}
	}

	if len(alertServers) > 0 {
		alertHeaderStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6B6B")).
			MarginTop(1).
			MarginBottom(1)

		alertSection = alertHeaderStyle.Render(fmt.Sprintf("⚠️  Servers Requiring Attention (%d)", len(alertServers)))

		for i, server := range alertServers {
			alertSection += "\n" + d.renderServerCard(server, d.selectedIdx == i, true)
		}
		alertSection += "\n"
	}

	// Healthy servers section
	var healthySection string
	if len(healthyServers) > 0 {
		healthyHeaderStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#48BB78")).
			MarginTop(1).
			MarginBottom(1)

		if len(alertServers) > 0 {
			healthySection = healthyHeaderStyle.Render(fmt.Sprintf("✓ Online Servers (%d)", len(healthyServers)))
		}

		startIdx := len(alertServers)
		for i, server := range healthyServers {
			healthySection += "\n" + d.renderServerCard(server, d.selectedIdx == startIdx+i, false)
		}
	}

	if len(d.statuses) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)
		return lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			emptyStyle.Render("No servers configured. Press 'c' to add a connection."),
		)
	}

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		MarginTop(2)

	var helpText string
	if d.loading {
		helpText = helpStyle.Render(d.spinner.View() + " Refreshing...")
	} else {
		helpText = helpStyle.Render(fmt.Sprintf("↑↓: navigate • r: refresh • Last update: %s", d.lastRefresh.Format("15:04:05")))
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		summary,
		alertSection,
		healthySection,
		helpText,
	)
}

func (d *DashboardScreen) renderServerCard(server ServerStatus, selected bool, isAlert bool) string {
	cardWidth := d.width - 4
	if cardWidth < 40 {
		cardWidth = 40
	}
	if cardWidth > 80 {
		cardWidth = 80
	}

	// Border color based on status and selection
	var borderColor lipgloss.Color
	if selected {
		borderColor = lipgloss.Color("#38B2AC")
	} else if isAlert {
		borderColor = lipgloss.Color("#FF6B6B")
	} else {
		borderColor = lipgloss.Color("#48BB78")
	}

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(cardWidth).
		Padding(0, 1)

	if selected {
		cardStyle = cardStyle.Background(lipgloss.Color("#2D3748"))
	}

	// Server name and status
	nameStyle := lipgloss.NewStyle().Bold(true)
	statusIcon := "✓"
	statusColor := lipgloss.Color("#48BB78")
	if !server.Online {
		statusIcon = "✗"
		statusColor = lipgloss.Color("#FF6B6B")
	}

	statusStyle := lipgloss.NewStyle().Foreground(statusColor)

	header := nameStyle.Render(server.ConnectionName) + " " + statusStyle.Render(statusIcon)

	// URL
	urlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Italic(true)
	urlLine := urlStyle.Render(server.URL)

	if !server.Online {
		// Show error for offline servers
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
		errorMsg := server.Error
		if errorMsg == "" {
			errorMsg = "Server is offline"
		}
		if len(errorMsg) > cardWidth-4 {
			errorMsg = errorMsg[:cardWidth-7] + "..."
		}
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			urlLine,
			errorStyle.Render(errorMsg),
		)
		return cardStyle.Render(content)
	}

	// Version badge
	versionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4299E1")).
		Background(lipgloss.Color("#2B6CB0")).
		Padding(0, 1)

	version := ""
	if server.Version != "" {
		version = versionStyle.Render("v" + server.Version)
	}

	// Stats
	statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A0AEC0"))
	stats := statsStyle.Render(fmt.Sprintf(
		"📦 %d layers  🗄 %d stores  📁 %d workspaces",
		server.LayerCount, server.DataStoreCount, server.WorkspaceCount,
	))

	// Memory bar
	var memoryLine string
	if server.MemoryTotal > 0 {
		memoryLine = d.renderMemoryBar(server.MemoryUsedPct, cardWidth-4)
	}

	// Response time
	responseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#718096"))
	responseLine := responseStyle.Render(fmt.Sprintf("⏱ %dms", server.ResponseTimeMs))

	var content string
	if memoryLine != "" {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			header+" "+version,
			urlLine,
			stats,
			memoryLine,
			responseLine,
		)
	} else {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			header+" "+version,
			urlLine,
			stats,
			responseLine,
		)
	}

	return cardStyle.Render(content)
}

func (d *DashboardScreen) renderMemoryBar(percent float64, width int) string {
	barWidth := width - 20
	if barWidth < 10 {
		barWidth = 10
	}

	filled := int(percent * float64(barWidth) / 100)
	empty := barWidth - filled

	var barColor lipgloss.Color
	if percent > 80 {
		barColor = lipgloss.Color("#FF6B6B")
	} else if percent > 60 {
		barColor = lipgloss.Color("#FFD93D")
	} else {
		barColor = lipgloss.Color("#48BB78")
	}

	filledStyle := lipgloss.NewStyle().Foreground(barColor)
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4A5568"))

	bar := "[" + filledStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", empty)) + "]"

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A0AEC0"))
	return labelStyle.Render("Memory: ") + bar + labelStyle.Render(fmt.Sprintf(" %.0f%%", percent))
}

// SetSize sets the screen dimensions
func (d *DashboardScreen) SetSize(width, height int) {
	d.width = width
	d.height = height
}

// GetSelectedConnectionID returns the ID of the selected connection
func (d *DashboardScreen) GetSelectedConnectionID() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.selectedIdx >= 0 && d.selectedIdx < len(d.statuses) {
		return d.statuses[d.selectedIdx].ConnectionID
	}
	return ""
}

// IsLoading returns true if the dashboard is loading
func (d *DashboardScreen) IsLoading() bool {
	return d.loading
}

// TriggerRefresh triggers a status refresh
func (d *DashboardScreen) TriggerRefresh() tea.Cmd {
	return func() tea.Msg {
		return DashboardRefreshMsg{}
	}
}
