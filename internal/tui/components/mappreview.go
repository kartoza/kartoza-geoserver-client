package components

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"  // Register GIF decoder
	_ "image/jpeg" // Register JPEG decoder
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kartoza/kartoza-cloudbench/internal/tui/styles"
)

// NewMapPreview creates a new map preview component
func NewMapPreview(geoserverURL, username, password, workspace, layerName string) *MapPreview {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.LoadingStyle

	return &MapPreview{
		geoserverURL: strings.TrimSuffix(geoserverURL, "/"),
		username:     username,
		password:     password,
		workspace:    workspace,
		layerName:    layerName,
		styles:       []string{}, // Will be populated when layer info is loaded
		currentStyle: 0,
		visible:      true,
		loading:      true,
		zoomLevel:    2.0,
		centerX:      0,
		centerY:      0,
		bbox:         [4]float64{-180, -90, 180, 90}, // World extent
		imgWidth:     800,
		imgHeight:    600,
		crosshairX:   0.5, // Start in center
		crosshairY:   0.5, // Start in center
		showOverlay:  true, // Show overlay by default
		keyMap:       DefaultMapPreviewKeyMap(),
		spinner:      s,
		protocol:     detectImageProtocol(),
	}
}

// SetStyles sets the available styles for the layer
func (m *MapPreview) SetStyles(styleNames []string) {
	m.styles = styleNames
	if len(m.styles) > 0 {
		m.currentStyle = 0
	}
}

// SetBounds sets the bounding box for the map
func (m *MapPreview) SetBounds(minX, minY, maxX, maxY float64) {
	m.bbox = [4]float64{minX, minY, maxX, maxY}
	m.centerX = (minX + maxX) / 2
	m.centerY = (minY + maxY) / 2
}

// SetOnClose sets the callback for when the preview closes
func (m *MapPreview) SetOnClose(fn func()) {
	m.onClose = fn
}

// SetSize sets the component size
func (m *MapPreview) SetSize(width, height int) {
	m.width = width
	m.height = height
	// Calculate image size based on terminal size
	// Leave room for controls and padding
	m.imgWidth = (width - 20) * 8   // Approximate pixel width
	m.imgHeight = (height - 10) * 16 // Approximate pixel height
	if m.imgWidth > 1024 {
		m.imgWidth = 1024
	}
	if m.imgHeight > 768 {
		m.imgHeight = 768
	}
	if m.imgWidth < 256 {
		m.imgWidth = 256
	}
	if m.imgHeight < 192 {
		m.imgHeight = 192
	}
}

// IsVisible returns whether the preview is visible
func (m *MapPreview) IsVisible() bool {
	return m.visible
}

// Init initializes the component - does NOT fetch map yet, waits for metadata
func (m *MapPreview) Init() tea.Cmd {
	// Only start spinner - map fetch is triggered by MapPreviewMetadataMsg
	return m.spinner.Tick
}

// Update handles messages
func (m *MapPreview) Update(msg tea.Msg) (*MapPreview, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If layer panel is open, handle its keys first
		if m.showLayerPanel {
			switch {
			case key.Matches(msg, m.keyMap.Close), key.Matches(msg, m.keyMap.ToggleLayers):
				m.showLayerPanel = false
				return m, nil
			case msg.String() == "up" || msg.String() == "k":
				if m.layerPanelCursor > 0 {
					m.layerPanelCursor--
				}
				return m, nil
			case msg.String() == "down" || msg.String() == "j":
				if m.layerPanelCursor < len(m.groupLayers)-1 {
					m.layerPanelCursor++
				}
				return m, nil
			case msg.String() == " ", msg.String() == "enter":
				// Toggle layer enabled state
				if m.layerPanelCursor < len(m.groupLayers) {
					m.groupLayers[m.layerPanelCursor].Enabled = !m.groupLayers[m.layerPanelCursor].Enabled
					// Refresh map with new layer selection
					m.showLayerPanel = false
					m.loading = true
					// Double buffer: save current image before fetching new one
					if m.renderedImage != "" {
						m.previousImage = m.renderedImage
					}
					return m, m.fetchMap()
				}
				return m, nil
			case msg.String() == "left" || msg.String() == "h":
				// Previous style for current layer
				if m.layerPanelCursor < len(m.groupLayers) {
					layer := &m.groupLayers[m.layerPanelCursor]
					if len(layer.AvailableStyles) > 1 {
						currentIdx := m.findStyleIndex(layer.CurrentStyle, layer.AvailableStyles)
						currentIdx--
						if currentIdx < 0 {
							currentIdx = len(layer.AvailableStyles) - 1
						}
						layer.CurrentStyle = layer.AvailableStyles[currentIdx]
					}
				}
				return m, nil
			case msg.String() == "right" || msg.String() == "l":
				// Next style for current layer (note: 'l' is overloaded here but panel is open)
				if m.layerPanelCursor < len(m.groupLayers) {
					layer := &m.groupLayers[m.layerPanelCursor]
					if len(layer.AvailableStyles) > 1 {
						currentIdx := m.findStyleIndex(layer.CurrentStyle, layer.AvailableStyles)
						currentIdx++
						if currentIdx >= len(layer.AvailableStyles) {
							currentIdx = 0
						}
						layer.CurrentStyle = layer.AvailableStyles[currentIdx]
					}
				}
				return m, nil
			case msg.String() == "a":
				// Apply style changes and refresh map
				m.showLayerPanel = false
				m.loading = true
				if m.renderedImage != "" {
					m.previousImage = m.renderedImage
				}
				return m, m.fetchMap()
			}
			return m, nil
		}

		switch {
		case key.Matches(msg, m.keyMap.Close):
			m.visible = false
			if m.onClose != nil {
				m.onClose()
			}
			return m, tea.ClearScreen

		case key.Matches(msg, m.keyMap.ToggleLayers):
			// Only show layer panel for layer groups that support layer toggling
			if m.isLayerGroup && m.canToggleLayers() && len(m.groupLayers) > 0 {
				m.showLayerPanel = true
				m.layerPanelCursor = 0
			}
			return m, nil

		case key.Matches(msg, m.keyMap.ZoomIn):
			m.zoomIn()
			m.loading = true
			// Double buffer: save current image before fetching new one
			if m.renderedImage != "" {
				m.previousImage = m.renderedImage
			}
			return m, m.fetchMap()

		case key.Matches(msg, m.keyMap.ZoomOut):
			m.zoomOut()
			m.loading = true
			if m.renderedImage != "" {
				m.previousImage = m.renderedImage
			}
			return m, m.fetchMap()

		case key.Matches(msg, m.keyMap.PanUp):
			m.panUp()
			m.loading = true
			if m.renderedImage != "" {
				m.previousImage = m.renderedImage
			}
			return m, m.fetchMap()

		case key.Matches(msg, m.keyMap.PanDown):
			m.panDown()
			m.loading = true
			if m.renderedImage != "" {
				m.previousImage = m.renderedImage
			}
			return m, m.fetchMap()

		case key.Matches(msg, m.keyMap.PanLeft):
			m.panLeft()
			m.loading = true
			if m.renderedImage != "" {
				m.previousImage = m.renderedImage
			}
			return m, m.fetchMap()

		case key.Matches(msg, m.keyMap.PanRight):
			m.panRight()
			m.loading = true
			if m.renderedImage != "" {
				m.previousImage = m.renderedImage
			}
			return m, m.fetchMap()

		case key.Matches(msg, m.keyMap.Refresh):
			m.loading = true
			if m.renderedImage != "" {
				m.previousImage = m.renderedImage
			}
			return m, m.fetchMap()

		case key.Matches(msg, m.keyMap.NextStyle):
			m.nextStyle()
			m.loading = true
			if m.renderedImage != "" {
				m.previousImage = m.renderedImage
			}
			return m, m.fetchMap()

		case key.Matches(msg, m.keyMap.PrevStyle):
			m.prevStyle()
			m.loading = true
			if m.renderedImage != "" {
				m.previousImage = m.renderedImage
			}
			return m, m.fetchMap()

		case key.Matches(msg, m.keyMap.CrosshairUp):
			m.moveCrosshair(0, -1) // Move 1 pixel up
			m.updateCompositeImage()
			m.renderedImage = m.renderCompositeImage()
			return m, nil

		case key.Matches(msg, m.keyMap.CrosshairDown):
			m.moveCrosshair(0, 1) // Move 1 pixel down
			m.updateCompositeImage()
			m.renderedImage = m.renderCompositeImage()
			return m, nil

		case key.Matches(msg, m.keyMap.CrosshairLeft):
			m.moveCrosshair(-1, 0) // Move 1 pixel left
			m.updateCompositeImage()
			m.renderedImage = m.renderCompositeImage()
			return m, nil

		case key.Matches(msg, m.keyMap.CrosshairRight):
			m.moveCrosshair(1, 0) // Move 1 pixel right
			m.updateCompositeImage()
			m.renderedImage = m.renderCompositeImage()
			return m, nil

		case key.Matches(msg, m.keyMap.GetFeatureInfo):
			// Perform GetFeatureInfo at crosshair location
			m.showFeatureInfo = false // Hide any previous info
			m.statusMsg = "Fetching feature info..."
			return m, m.fetchFeatureInfo()

		case key.Matches(msg, m.keyMap.ToggleOverlay):
			m.showOverlay = !m.showOverlay
			m.updateCompositeImage()
			m.renderedImage = m.renderCompositeImage()
			return m, nil
		}

	case FeatureInfoMsg:
		if msg.Error != nil {
			m.featureInfo = "Error: " + msg.Error.Error()
			m.statusMsg = "FeatureInfo error: " + msg.Error.Error()
		} else {
			m.featureInfo = msg.Info
			m.statusMsg = fmt.Sprintf("FeatureInfo: %d chars", len(msg.Info))
		}
		m.showFeatureInfo = true
		// Re-render composite with feature info overlay
		m.updateCompositeImage()
		m.renderedImage = m.renderCompositeImage()
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case MapPreviewMsg:
		m.loading = false
		if msg.Error != nil {
			m.errorMsg = msg.Error.Error()
			m.imageData = nil
			m.baseImage = nil
			m.compositeImage = nil
			m.renderedImage = ""
			m.previousImage = ""
		} else {
			m.errorMsg = ""
			m.imageData = msg.ImageData
			// Decode and cache the base image (supports PNG, GIF, JPEG)
			if img, _, err := image.Decode(bytes.NewReader(msg.ImageData)); err == nil {
				m.baseImage = img
			}
			// Render composite with crosshair and legend
			m.updateCompositeImage()
			m.renderedImage = m.renderCompositeImage()
			m.previousImage = ""
			// Fetch legend on first successful map load
			if !m.legendFetched {
				m.legendFetched = true
				return m, m.fetchLegend()
			}
		}

	case LegendMsg:
		if msg.Error != nil {
			m.statusMsg = "Legend error: " + msg.Error.Error()
		} else if len(msg.ImageData) > 0 {
			// Decode and cache the legend image (supports PNG, GIF, JPEG)
			if img, _, err := image.Decode(bytes.NewReader(msg.ImageData)); err == nil {
				m.legendImage = img
				m.statusMsg = fmt.Sprintf("Legend loaded: %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
				// Re-render composite with legend
				m.updateCompositeImage()
				m.renderedImage = m.renderCompositeImage()
			} else {
				m.statusMsg = "Legend decode error: " + err.Error()
			}
		}
		return m, nil

	case MapPreviewMetadataMsg:
		// Apply metadata and fetch map
		if msg.Bounds != nil {
			m.SetBounds(msg.Bounds[0], msg.Bounds[1], msg.Bounds[2], msg.Bounds[3])
		}
		if len(msg.Styles) > 0 {
			m.SetStyles(msg.Styles)
		}
		// Apply layer group info
		if msg.IsLayerGroup {
			m.isLayerGroup = true
			m.layerGroupMode = msg.LayerGroupMode
			m.groupLayers = make([]LayerToggleItem, len(msg.GroupLayers))
			for i, layerInfo := range msg.GroupLayers {
				m.groupLayers[i] = LayerToggleItem{
					Name:            layerInfo.Name,
					Enabled:         true, // All layers enabled by default
					CurrentStyle:    layerInfo.DefaultStyle,
					AvailableStyles: layerInfo.AvailableStyles,
				}
			}
		}
		// Now fetch the map with proper bounds
		m.loading = true
		return m, m.fetchMap()
	}

	return m, tea.Batch(cmds...)
}

// canToggleLayers returns true if the layer group mode supports toggling individual layers
func (m *MapPreview) canToggleLayers() bool {
	// NAMED mode allows individual layers to be visible separately
	// SINGLE mode merges all layers into one, so toggling doesn't make sense
	// CONTAINER is organizational only
	// EO is for Earth Observation
	return m.layerGroupMode == "NAMED" || m.layerGroupMode == "EO"
}

// findStyleIndex finds the index of a style in the available styles list
func (m *MapPreview) findStyleIndex(styleName string, styles []string) int {
	for i, s := range styles {
		if s == styleName {
			return i
		}
	}
	return 0
}

// View renders the component with full-screen map and crosshair overlay
func (m *MapPreview) View() string {
	if !m.visible {
		return ""
	}

	// If we have the layer panel open, show that instead of full-screen map
	if m.showLayerPanel {
		return m.renderLayerPanelView()
	}

	// Get the map image (or loading state)
	var mapImage string
	if m.errorMsg != "" {
		mapImage = styles.ErrorStyle.Render("Error: " + m.errorMsg)
	} else if m.loading {
		if m.previousImage != "" {
			mapImage = m.previousImage
		} else {
			mapImage = m.spinner.View() + " Loading map..."
		}
	} else if m.renderedImage != "" {
		mapImage = m.renderedImage
	}

	// Build final output
	var content strings.Builder

	// Add map image with overlay
	content.WriteString(mapImage)

	// Add status line below the map (always show legend status for debugging)
	legendStatus := m.statusMsg
	if legendStatus == "" {
		if m.legendImage != nil {
			legendStatus = fmt.Sprintf("Legend: loaded (%dx%d)", m.legendImage.Bounds().Dx(), m.legendImage.Bounds().Dy())
		} else if m.legendFetched {
			legendStatus = "Legend: fetch attempted but no image"
		} else {
			legendStatus = "Legend: not yet fetched"
		}
	}
	content.WriteString("\n")
	content.WriteString(styles.HelpBarStyle.Render(legendStatus))

	// Add coordinates and help
	lon, lat := m.getCrosshairCoordinates()
	coordStr := fmt.Sprintf("Crosshair: %.4f, %.4f | Press 'i' for info, 'o' toggle overlay, Shift+arrows move crosshair", lon, lat)
	content.WriteString("\n")
	content.WriteString(styles.HelpBarStyle.Render(coordStr))

	// If showing feature info, render it at the bottom
	if m.showFeatureInfo && m.featureInfo != "" {
		content.WriteString("\n")
		content.WriteString(m.renderFeatureInfoPopup())
	}

	return content.String()
}

// renderLayerPanelView renders the layer panel with the map visible behind
func (m *MapPreview) renderLayerPanelView() string {
	var content strings.Builder

	// Title bar
	titleText := fmt.Sprintf(" Layer Preview: %s:%s ", m.workspace, m.layerName)
	if m.isLayerGroup {
		titleText = fmt.Sprintf(" Layer Group Preview: %s:%s ", m.workspace, m.layerName)
	}
	title := styles.DialogTitleStyle.Render(titleText)
	content.WriteString(title)
	content.WriteString("\n\n")

	content.WriteString(m.renderLayerPanel())

	return content.String()
}

// renderKey renders a key hint in a compact style
func (m *MapPreview) renderKey(key string) string {
	return styles.AccentStyle.Copy().
		Bold(true).
		Render(key)
}
