package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kartoza/kartoza-geoserver-client/internal/models"
	"github.com/kartoza/kartoza-geoserver-client/internal/tui/styles"
)

// CacheWizardOperation represents the cache operation type
type CacheWizardOperation int

const (
	CacheOperationSeed CacheWizardOperation = iota
	CacheOperationReseed
	CacheOperationTruncate
)

// CacheWizardResult contains the result of the cache wizard
type CacheWizardResult struct {
	Confirmed    bool
	Operation    CacheWizardOperation
	GridSet      string
	Format       string
	ZoomStart    int
	ZoomStop     int
	ThreadCount  int
	LayerName    string
	ConnectionID string
	Workspace    string
}

// CacheWizardKeyMap defines the key bindings
type CacheWizardKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Left    key.Binding
	Right   key.Binding
	Confirm key.Binding
	Cancel  key.Binding
	Tab     key.Binding
}

// DefaultCacheWizardKeyMap returns the default key bindings
func DefaultCacheWizardKeyMap() CacheWizardKeyMap {
	return CacheWizardKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←", "decrease"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→", "increase"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc", "cancel"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next"),
		),
	}
}

// CacheWizard is a dialog for managing tile cache
type CacheWizard struct {
	keyMap       CacheWizardKeyMap
	visible      bool
	animating    bool
	animProgress float64

	// Layer info
	layerName    string
	connectionID string
	workspace    string

	// Options
	gridSets []string
	formats  []string

	// Current selections
	selectedGridSet int
	selectedFormat  int
	operation       CacheWizardOperation
	zoomStart       int
	zoomStop        int
	threadCount     int

	// Current field (for navigation)
	currentField int

	// Dimensions
	width  int
	height int

	// Loading state
	loading bool
	spinner spinner.Model

	// Callbacks
	onConfirm func(CacheWizardResult)
	onCancel  func()
}

// NewCacheWizard creates a new cache wizard
func NewCacheWizard(node *models.TreeNode, gridSets, formats []string) *CacheWizard {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.LoadingStyle

	// Default formats if none provided
	if len(formats) == 0 {
		formats = []string{"image/png", "image/jpeg", "image/png8"}
	}
	// Default grid sets if none provided
	if len(gridSets) == 0 {
		gridSets = []string{"EPSG:4326", "EPSG:900913", "EPSG:3857"}
	}

	layerName := node.Name
	if node.Workspace != "" {
		layerName = node.Workspace + ":" + node.Name
	}

	return &CacheWizard{
		keyMap:          DefaultCacheWizardKeyMap(),
		visible:         true,
		animating:       true,
		animProgress:    0,
		layerName:       layerName,
		connectionID:    node.ConnectionID,
		workspace:       node.Workspace,
		gridSets:        gridSets,
		formats:         formats,
		selectedGridSet: 0,
		selectedFormat:  0,
		operation:       CacheOperationSeed,
		zoomStart:       0,
		zoomStop:        10,
		threadCount:     2,
		currentField:    0,
		spinner:         s,
	}
}

// CacheWizardAnimationMsg is sent to animate the wizard
type CacheWizardAnimationMsg struct{}

// SetCallbacks sets the confirmation and cancellation callbacks
func (w *CacheWizard) SetCallbacks(onConfirm func(CacheWizardResult), onCancel func()) {
	w.onConfirm = onConfirm
	w.onCancel = onCancel
}

// SetSize sets the wizard dimensions
func (w *CacheWizard) SetSize(width, height int) {
	w.width = width
	w.height = height
}

// IsVisible returns whether the wizard is visible
func (w *CacheWizard) IsVisible() bool {
	return w.visible
}

// Init initializes the wizard
func (w *CacheWizard) Init() tea.Cmd {
	return tea.Batch(
		w.spinner.Tick,
		func() tea.Msg {
			return CacheWizardAnimationMsg{}
		},
	)
}

// Update handles messages
func (w *CacheWizard) Update(msg tea.Msg) (*CacheWizard, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case CacheWizardAnimationMsg:
		if w.animating {
			w.animProgress += 0.15
			if w.animProgress >= 1.0 {
				w.animProgress = 1.0
				w.animating = false
			} else {
				cmds = append(cmds, func() tea.Msg {
					return CacheWizardAnimationMsg{}
				})
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		w.spinner, cmd = w.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, w.keyMap.Cancel):
			w.close()
			if w.onCancel != nil {
				w.onCancel()
			}

		case key.Matches(msg, w.keyMap.Confirm):
			result := CacheWizardResult{
				Confirmed:    true,
				Operation:    w.operation,
				GridSet:      w.gridSets[w.selectedGridSet],
				Format:       w.formats[w.selectedFormat],
				ZoomStart:    w.zoomStart,
				ZoomStop:     w.zoomStop,
				ThreadCount:  w.threadCount,
				LayerName:    w.layerName,
				ConnectionID: w.connectionID,
				Workspace:    w.workspace,
			}
			w.close()
			if w.onConfirm != nil {
				w.onConfirm(result)
			}

		case key.Matches(msg, w.keyMap.Tab), key.Matches(msg, w.keyMap.Down):
			w.currentField = (w.currentField + 1) % 6

		case key.Matches(msg, w.keyMap.Up):
			w.currentField = (w.currentField + 5) % 6

		case key.Matches(msg, w.keyMap.Left):
			w.decreaseCurrentField()

		case key.Matches(msg, w.keyMap.Right):
			w.increaseCurrentField()
		}
	}

	return w, tea.Batch(cmds...)
}

func (w *CacheWizard) decreaseCurrentField() {
	switch w.currentField {
	case 0: // Operation
		if w.operation > 0 {
			w.operation--
		} else {
			w.operation = CacheOperationTruncate
		}
	case 1: // Grid Set
		if w.selectedGridSet > 0 {
			w.selectedGridSet--
		} else {
			w.selectedGridSet = len(w.gridSets) - 1
		}
	case 2: // Format
		if w.selectedFormat > 0 {
			w.selectedFormat--
		} else {
			w.selectedFormat = len(w.formats) - 1
		}
	case 3: // Zoom Start
		if w.zoomStart > 0 {
			w.zoomStart--
		}
	case 4: // Zoom Stop
		if w.zoomStop > w.zoomStart {
			w.zoomStop--
		}
	case 5: // Thread Count
		if w.threadCount > 1 {
			w.threadCount--
		}
	}
}

func (w *CacheWizard) increaseCurrentField() {
	switch w.currentField {
	case 0: // Operation
		if w.operation < CacheOperationTruncate {
			w.operation++
		} else {
			w.operation = CacheOperationSeed
		}
	case 1: // Grid Set
		w.selectedGridSet = (w.selectedGridSet + 1) % len(w.gridSets)
	case 2: // Format
		w.selectedFormat = (w.selectedFormat + 1) % len(w.formats)
	case 3: // Zoom Start
		if w.zoomStart < w.zoomStop {
			w.zoomStart++
		}
	case 4: // Zoom Stop
		if w.zoomStop < 25 {
			w.zoomStop++
		}
	case 5: // Thread Count
		if w.threadCount < 16 {
			w.threadCount++
		}
	}
}

func (w *CacheWizard) close() {
	w.visible = false
}

// View renders the wizard
func (w *CacheWizard) View() string {
	if !w.visible {
		return ""
	}

	var b strings.Builder

	// Title
	title := styles.DialogTitleStyle.Render("Tile Cache Management")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Layer name
	b.WriteString(styles.HelpTextStyle.Render("Layer: "))
	b.WriteString(styles.PanelHeaderStyle.Render(w.layerName))
	b.WriteString("\n\n")

	// Fields
	fields := []struct {
		label string
		value string
	}{
		{"Operation", w.getOperationName()},
		{"Grid Set", w.gridSets[w.selectedGridSet]},
		{"Format", w.formats[w.selectedFormat]},
		{"Zoom Start", fmt.Sprintf("%d", w.zoomStart)},
		{"Zoom Stop", fmt.Sprintf("%d", w.zoomStop)},
		{"Threads", fmt.Sprintf("%d", w.threadCount)},
	}

	for i, field := range fields {
		prefix := "  "
		if i == w.currentField {
			prefix = "▸ "
		}

		label := styles.HelpTextStyle.Width(12).Render(field.label + ":")
		value := field.value

		if i == w.currentField {
			value = styles.ActiveItemStyle.Render("◄ " + value + " ►")
		} else {
			value = styles.ItemStyle.Render("  " + value + "  ")
		}

		b.WriteString(prefix)
		b.WriteString(label)
		b.WriteString(value)
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Help
	if w.operation == CacheOperationTruncate {
		b.WriteString(styles.ErrorStyle.Render("⚠ Truncate will delete all cached tiles!"))
		b.WriteString("\n\n")
	}

	b.WriteString(styles.HelpTextStyle.Render("↑↓ navigate  ←→ change  Enter confirm  Esc cancel"))

	// Create dialog box
	dialogWidth := 50
	dialog := styles.DialogStyle.Width(dialogWidth).Render(b.String())

	// Apply animation
	if w.animating {
		scale := w.animProgress
		dialogHeight := lipgloss.Height(dialog)
		visibleLines := int(float64(dialogHeight) * scale)
		if visibleLines < 1 {
			visibleLines = 1
		}
		lines := strings.Split(dialog, "\n")
		if visibleLines < len(lines) {
			dialog = strings.Join(lines[:visibleLines], "\n")
		}
	}

	return styles.Center(w.width, w.height, dialog)
}

func (w *CacheWizard) getOperationName() string {
	switch w.operation {
	case CacheOperationSeed:
		return "Seed (generate new tiles)"
	case CacheOperationReseed:
		return "Reseed (regenerate all)"
	case CacheOperationTruncate:
		return "Truncate (clear cache)"
	default:
		return "Unknown"
	}
}
