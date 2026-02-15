package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
	"github.com/kartoza/kartoza-cloudbench/internal/api"
	"github.com/kartoza/kartoza-cloudbench/internal/config"
	"github.com/kartoza/kartoza-cloudbench/internal/models"
	"github.com/kartoza/kartoza-cloudbench/internal/tui/icons"
	"github.com/kartoza/kartoza-cloudbench/internal/tui/styles"
)

// SearchResult represents a single search result
type SearchResult struct {
	Type         string
	Name         string
	Workspace    string
	StoreName    string
	StoreType    string
	ConnectionID string
	ServerName   string
	Tags         []string
	Icon         string
}

// SearchSelectMsg is sent when a search result is selected
type SearchSelectMsg struct {
	Result SearchResult
}

// SearchAnimationMsg is sent to update animation state
type SearchAnimationMsg struct{}

// SearchResultsMsg is sent when search results are received
type SearchResultsMsg struct {
	Results []SearchResult
	Query   string
	Err     error
}

// SearchModal is a modal dialog for universal search
type SearchModal struct {
	config       *config.Config
	clients      map[string]*api.Client
	searchInput  textinput.Model
	results      []SearchResult
	selectedIdx  int
	width        int
	height       int
	visible      bool
	loading      bool
	lastQuery    string
	onSelect     func(SearchResult)
	onCancel     func()

	// Animation
	spring        harmonica.Spring
	animScale     float64
	animVelocity  float64
	animOpacity   float64
	targetScale   float64
	targetOpacity float64
	animating     bool
	closing       bool
}

// NewSearchModal creates a new search modal
func NewSearchModal(cfg *config.Config, clients map[string]*api.Client) *SearchModal {
	ti := textinput.New()
	ti.Placeholder = "Search workspaces, layers, styles..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	return &SearchModal{
		config:        cfg,
		clients:       clients,
		searchInput:   ti,
		results:       []SearchResult{},
		visible:       true,
		spring:        harmonica.NewSpring(harmonica.FPS(60), 6.0, 0.5),
		animScale:     0.0,
		animOpacity:   0.0,
		targetScale:   1.0,
		targetOpacity: 1.0,
		animating:     true,
	}
}

// SetSize sets the modal size
func (m *SearchModal) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.searchInput.Width = width/2 - 10
	if m.searchInput.Width < 30 {
		m.searchInput.Width = 30
	}
}

// SetCallbacks sets the select and cancel callbacks
func (m *SearchModal) SetCallbacks(onSelect func(SearchResult), onCancel func()) {
	m.onSelect = onSelect
	m.onCancel = onCancel
}

// IsVisible returns whether the modal is visible
func (m *SearchModal) IsVisible() bool {
	return m.visible
}

// animateCmd returns a command to continue the animation
func (m *SearchModal) animateCmd() tea.Cmd {
	return tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
		return SearchAnimationMsg{}
	})
}

// Init initializes the modal and starts the opening animation
func (m *SearchModal) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.animateCmd(),
	)
}

// Update handles messages
func (m *SearchModal) Update(msg tea.Msg) (*SearchModal, tea.Cmd) {
	switch msg := msg.(type) {
	case SearchAnimationMsg:
		return m.updateAnimation()

	case SearchResultsMsg:
		m.loading = false
		if msg.Err == nil && msg.Query == strings.ToLower(m.searchInput.Value()) {
			m.results = msg.Results
			m.lastQuery = msg.Query
			m.selectedIdx = 0
		}
		return m, nil

	case tea.KeyMsg:
		if !m.visible || m.animating {
			return m, nil
		}

		switch msg.String() {
		case "esc":
			if m.onCancel != nil {
				m.onCancel()
			}
			return m, m.startCloseAnimation()

		case "enter":
			if len(m.results) > 0 && m.selectedIdx < len(m.results) {
				result := m.results[m.selectedIdx]
				if m.onSelect != nil {
					m.onSelect(result)
				}
				return m, m.startCloseAnimation()
			}

		case "up", "ctrl+p":
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
			return m, nil

		case "down", "ctrl+n":
			if m.selectedIdx < len(m.results)-1 {
				m.selectedIdx++
			}
			return m, nil

		case "pgup":
			m.selectedIdx -= 5
			if m.selectedIdx < 0 {
				m.selectedIdx = 0
			}
			return m, nil

		case "pgdown":
			m.selectedIdx += 5
			if m.selectedIdx >= len(m.results) {
				m.selectedIdx = len(m.results) - 1
			}
			if m.selectedIdx < 0 {
				m.selectedIdx = 0
			}
			return m, nil

		default:
			// Forward to text input
			var cmd tea.Cmd
			oldValue := m.searchInput.Value()
			m.searchInput, cmd = m.searchInput.Update(msg)

			// If the query changed, trigger a new search
			if m.searchInput.Value() != oldValue {
				cmds := []tea.Cmd{cmd}
				if len(m.searchInput.Value()) >= 2 {
					m.loading = true
					cmds = append(cmds, m.doSearch())
				} else {
					m.results = []SearchResult{}
					m.selectedIdx = 0
				}
				return m, tea.Batch(cmds...)
			}
			return m, cmd
		}
	}

	return m, nil
}

// updateAnimation updates the harmonica physics animation
func (m *SearchModal) updateAnimation() (*SearchModal, tea.Cmd) {
	if !m.animating {
		return m, nil
	}

	// Update scale using spring physics
	m.animScale, m.animVelocity = m.spring.Update(m.animScale, m.animVelocity, m.targetScale)

	// Update opacity
	opacityStep := 0.1
	if m.closing {
		m.animOpacity -= opacityStep
		if m.animOpacity < 0 {
			m.animOpacity = 0
		}
	} else {
		m.animOpacity += opacityStep
		if m.animOpacity > 1 {
			m.animOpacity = 1
		}
	}

	// Check if animation is complete
	scaleClose := abs(m.animScale-m.targetScale) < 0.01 && abs(m.animVelocity) < 0.01
	opacityClose := m.closing && m.animOpacity <= 0.01 || !m.closing && m.animOpacity >= 0.99

	if scaleClose && opacityClose {
		m.animating = false
		m.animScale = m.targetScale
		m.animOpacity = m.targetOpacity

		if m.closing {
			m.visible = false
			return m, nil
		}
	}

	return m, m.animateCmd()
}

// startCloseAnimation starts the closing animation
func (m *SearchModal) startCloseAnimation() tea.Cmd {
	m.closing = true
	m.targetScale = 0.0
	m.targetOpacity = 0.0
	m.animating = true
	return m.animateCmd()
}

// doSearch performs the search across all connections
func (m *SearchModal) doSearch() tea.Cmd {
	query := strings.ToLower(m.searchInput.Value())
	return func() tea.Msg {
		var results []SearchResult

		for _, conn := range m.config.Connections {
			client := m.clients[conn.ID]
			if client == nil {
				continue
			}

			// Search workspaces
			workspaces, err := client.GetWorkspaces()
			if err == nil {
				for _, ws := range workspaces {
					if matchesQuery(ws.Name, query) {
						results = append(results, SearchResult{
							Type:         "workspace",
							Name:         ws.Name,
							ConnectionID: conn.ID,
							ServerName:   conn.Name,
							Tags:         []string{"Workspace"},
							Icon:         icons.Folder,
						})
					}

					// Search data stores
					dataStores, _ := client.GetDataStores(ws.Name)
					for _, ds := range dataStores {
						if matchesQuery(ds.Name, query) {
							tags := []string{"Data Store"}
							if ds.Type != "" {
								tags = append(tags, ds.Type)
							}
							results = append(results, SearchResult{
								Type:         "datastore",
								Name:         ds.Name,
								Workspace:    ws.Name,
								ConnectionID: conn.ID,
								ServerName:   conn.Name,
								Tags:         tags,
								Icon:         icons.Database,
							})
						}
					}

					// Search coverage stores
					coverageStores, _ := client.GetCoverageStores(ws.Name)
					for _, cs := range coverageStores {
						if matchesQuery(cs.Name, query) {
							tags := []string{"Coverage Store"}
							if cs.Type != "" {
								tags = append(tags, cs.Type)
							}
							results = append(results, SearchResult{
								Type:         "coveragestore",
								Name:         cs.Name,
								Workspace:    ws.Name,
								ConnectionID: conn.ID,
								ServerName:   conn.Name,
								Tags:         tags,
								Icon:         icons.Image,
							})
						}
					}

					// Search layers
					layers, _ := client.GetLayers(ws.Name)
					for _, layer := range layers {
						if matchesQuery(layer.Name, query) {
							tags := []string{"Layer"}
							if layer.Type != "" {
								tags = append(tags, layer.Type)
							}
							results = append(results, SearchResult{
								Type:         "layer",
								Name:         layer.Name,
								Workspace:    ws.Name,
								ConnectionID: conn.ID,
								ServerName:   conn.Name,
								Tags:         tags,
								Icon:         icons.Layers,
							})
						}
					}

					// Search styles
					wsStyles, _ := client.GetStyles(ws.Name)
					for _, style := range wsStyles {
						if matchesQuery(style.Name, query) {
							tags := []string{"Style"}
							if style.Format != "" {
								tags = append(tags, strings.ToUpper(style.Format))
							}
							results = append(results, SearchResult{
								Type:         "style",
								Name:         style.Name,
								Workspace:    ws.Name,
								ConnectionID: conn.ID,
								ServerName:   conn.Name,
								Tags:         tags,
								Icon:         icons.Palette,
							})
						}
					}

					// Search layer groups
					layerGroups, _ := client.GetLayerGroups(ws.Name)
					for _, lg := range layerGroups {
						if matchesQuery(lg.Name, query) {
							results = append(results, SearchResult{
								Type:         "layergroup",
								Name:         lg.Name,
								Workspace:    ws.Name,
								ConnectionID: conn.ID,
								ServerName:   conn.Name,
								Tags:         []string{"Layer Group"},
								Icon:         icons.Book,
							})
						}
					}
				}
			}

			// Search global styles
			globalStyles, _ := client.GetStyles("")
			for _, style := range globalStyles {
				if matchesQuery(style.Name, query) {
					tags := []string{"Style", "Global"}
					if style.Format != "" {
						tags = append(tags, strings.ToUpper(style.Format))
					}
					results = append(results, SearchResult{
						Type:         "style",
						Name:         style.Name,
						ConnectionID: conn.ID,
						ServerName:   conn.Name,
						Tags:         tags,
						Icon:         icons.Palette,
					})
				}
			}

			// Search global layer groups
			globalLayerGroups, _ := client.GetLayerGroups("")
			for _, lg := range globalLayerGroups {
				if matchesQuery(lg.Name, query) {
					results = append(results, SearchResult{
						Type:         "layergroup",
						Name:         lg.Name,
						ConnectionID: conn.ID,
						ServerName:   conn.Name,
						Tags:         []string{"Layer Group", "Global"},
						Icon:         icons.Book,
					})
				}
			}
		}

		// Limit results
		if len(results) > 50 {
			results = results[:50]
		}

		return SearchResultsMsg{
			Results: results,
			Query:   query,
		}
	}
}

// matchesQuery checks if a name matches the search query
func matchesQuery(name, query string) bool {
	return strings.Contains(strings.ToLower(name), query)
}

// View renders the search modal
func (m *SearchModal) View() string {
	if !m.visible {
		return ""
	}

	modalWidth := m.width*2/3
	if modalWidth < 60 {
		modalWidth = 60
	}
	if modalWidth > 100 {
		modalWidth = 100
	}

	scaledWidth := int(float64(modalWidth) * m.animScale)
	if scaledWidth < 20 {
		scaledWidth = 20
	}

	// Modal styles
	modalStyle := lipgloss.NewStyle().
		Width(scaledWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Secondary).
		Padding(1, 2).
		Background(styles.Background)

	// When closing, render empty frame only
	if m.closing {
		modal := modalStyle.Render("")
		return styles.Center(m.width, m.height, modal)
	}

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.Secondary).
		Bold(true).
		MarginBottom(1)
	b.WriteString(titleStyle.Render(icons.Search + " Universal Search"))
	b.WriteString("\n\n")

	// Search input - Kartoza styled
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Border).
		Padding(0, 1).
		Width(scaledWidth - 6)
	b.WriteString(inputStyle.Render(m.searchInput.View()))
	b.WriteString("\n\n")

	// Results area
	maxResults := (m.height / 3) - 2
	if maxResults < 5 {
		maxResults = 5
	}
	if maxResults > 10 {
		maxResults = 10
	}

	currentQuery := strings.ToLower(m.searchInput.Value())
	if m.loading {
		loadingStyle := lipgloss.NewStyle().
			Foreground(styles.Muted).
			Italic(true)
		b.WriteString(loadingStyle.Render("Searching..."))
	} else if len(m.searchInput.Value()) < 2 {
		hintStyle := lipgloss.NewStyle().
			Foreground(styles.Muted).
			Align(lipgloss.Center).
			Width(scaledWidth - 6)
		b.WriteString(hintStyle.Render(icons.Search + " Type at least 2 characters to search"))
	} else if len(m.results) == 0 && m.lastQuery == currentQuery {
		// Only show "No results" if the results actually correspond to the current query
		emptyStyle := lipgloss.NewStyle().
			Foreground(styles.Muted).
			Italic(true)
		b.WriteString(emptyStyle.Render("No results found"))
	} else if len(m.results) == 0 {
		// Results don't match current query yet - show loading
		loadingStyle := lipgloss.NewStyle().
			Foreground(styles.Muted).
			Italic(true)
		b.WriteString(loadingStyle.Render("Searching..."))
	} else {
		// Calculate visible range
		startIdx := 0
		if m.selectedIdx >= maxResults {
			startIdx = m.selectedIdx - maxResults + 1
		}
		endIdx := startIdx + maxResults
		if endIdx > len(m.results) {
			endIdx = len(m.results)
		}

		for i := startIdx; i < endIdx; i++ {
			result := m.results[i]
			isSelected := i == m.selectedIdx

			b.WriteString(m.renderResultCard(result, isSelected, scaledWidth-6))
			b.WriteString("\n")
		}

		// Show total count
		if len(m.results) > maxResults {
			countStyle := lipgloss.NewStyle().
				Foreground(styles.Muted).
				Align(lipgloss.Right).
				Width(scaledWidth - 6)
			b.WriteString(countStyle.Render(fmt.Sprintf("Showing %d-%d of %d results", startIdx+1, endIdx, len(m.results))))
		}
	}

	// Help text
	b.WriteString("\n")
	helpStyle := lipgloss.NewStyle().
		Foreground(styles.Muted).
		Width(scaledWidth - 6)
	b.WriteString(helpStyle.Render("↑↓ Navigate  Enter Select  Esc Close"))

	modal := modalStyle.Render(b.String())

	// Apply opacity fade
	if m.animOpacity < 0.5 {
		// Dim the modal when fading
		modal = lipgloss.NewStyle().
			Foreground(styles.Border).
			Render(modal)
	}

	return styles.Center(m.width, m.height, modal)
}

// renderResultCard renders a single search result as a card - Kartoza styled
func (m *SearchModal) renderResultCard(result SearchResult, isSelected bool, width int) string {
	// Card styling
	var cardStyle lipgloss.Style
	if isSelected {
		cardStyle = lipgloss.NewStyle().
			Background(styles.SurfaceHigh).
			Foreground(styles.TextBright).
			Width(width).
			Padding(0, 1)
	} else {
		cardStyle = lipgloss.NewStyle().
			Width(width).
			Padding(0, 1)
	}

	// Icon
	iconStyle := lipgloss.NewStyle().
		Foreground(m.getTypeColor(result.Type)).
		MarginRight(1)

	// Name
	nameStyle := lipgloss.NewStyle().
		Bold(true)
	if isSelected {
		nameStyle = nameStyle.Foreground(styles.TextBright)
	} else {
		nameStyle = nameStyle.Foreground(styles.Text)
	}

	// Tags
	var tagStrs []string
	for _, tag := range result.Tags {
		tagStyle := lipgloss.NewStyle().
			Background(m.getTypeColor(result.Type)).
			Foreground(styles.TextBright).
			Padding(0, 1).
			MarginRight(1)
		tagStrs = append(tagStrs, tagStyle.Render(tag))
	}
	tags := strings.Join(tagStrs, "")

	// Location (server • workspace)
	locStyle := lipgloss.NewStyle().
		Foreground(styles.Muted).
		Italic(true)
	location := result.ServerName
	if result.Workspace != "" {
		location = location + " • " + result.Workspace
	}

	// Build the card content
	line1 := iconStyle.Render(result.Icon) + nameStyle.Render(result.Name) + "  " + tags
	line2 := locStyle.Render(location)

	// Selection indicator
	indicator := "  "
	if isSelected {
		indicator = iconStyle.Render(icons.ChevronRight) + " "
	}

	content := indicator + line1 + "\n   " + line2

	return cardStyle.Render(content)
}

// getTypeColor returns the color for a resource type - Kartoza branded
func (m *SearchModal) getTypeColor(resourceType string) lipgloss.Color {
	switch resourceType {
	case "workspace":
		return styles.KartozaBlue
	case "datastore":
		return styles.Success
	case "coveragestore":
		return styles.KartozaOrange
	case "layer":
		return styles.KartozaBlueLight
	case "style":
		return styles.StyleFile
	case "layergroup":
		return styles.KartozaBlue
	default:
		return styles.Muted
	}
}

// GetSelectedResult returns the currently selected result
func (m *SearchModal) GetSelectedResult() *SearchResult {
	if len(m.results) > 0 && m.selectedIdx < len(m.results) {
		return &m.results[m.selectedIdx]
	}
	return nil
}

// ToTreeNode converts a search result to a tree node for navigation
func (r *SearchResult) ToTreeNode() *models.TreeNode {
	var nodeType models.NodeType
	switch r.Type {
	case "workspace":
		nodeType = models.NodeTypeWorkspace
	case "datastore":
		nodeType = models.NodeTypeDataStore
	case "coveragestore":
		nodeType = models.NodeTypeCoverageStore
	case "layer":
		nodeType = models.NodeTypeLayer
	case "style":
		nodeType = models.NodeTypeStyle
	case "layergroup":
		nodeType = models.NodeTypeLayerGroup
	default:
		nodeType = models.NodeTypeRoot
	}

	node := models.NewTreeNode(r.Name, nodeType)
	node.ConnectionID = r.ConnectionID
	node.Workspace = r.Workspace
	node.StoreName = r.StoreName
	node.StoreType = r.StoreType
	return node
}
