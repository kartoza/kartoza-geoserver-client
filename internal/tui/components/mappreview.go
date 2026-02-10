package components

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kartoza/kartoza-geoserver-client/internal/tui/styles"
)

// ImageProtocol represents the terminal image rendering protocol
type ImageProtocol int

const (
	ProtocolNone ImageProtocol = iota
	ProtocolKitty
	ProtocolSixel
	ProtocolChafa
	ProtocolASCII
)

// MapPreviewKeyMap defines key bindings for the map preview
type MapPreviewKeyMap struct {
	Close     key.Binding
	ZoomIn    key.Binding
	ZoomOut   key.Binding
	PanUp     key.Binding
	PanDown   key.Binding
	PanLeft   key.Binding
	PanRight  key.Binding
	Refresh   key.Binding
	NextStyle key.Binding
	PrevStyle key.Binding
}

// DefaultMapPreviewKeyMap returns the default key bindings
func DefaultMapPreviewKeyMap() MapPreviewKeyMap {
	return MapPreviewKeyMap{
		Close: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc/q", "close"),
		),
		ZoomIn: key.NewBinding(
			key.WithKeys("+", "="),
			key.WithHelp("+", "zoom in"),
		),
		ZoomOut: key.NewBinding(
			key.WithKeys("-", "_"),
			key.WithHelp("-", "zoom out"),
		),
		PanUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("up/k", "pan up"),
		),
		PanDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("down/j", "pan down"),
		),
		PanLeft: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("left/h", "pan left"),
		),
		PanRight: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("right/l", "pan right"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		NextStyle: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "next style"),
		),
		PrevStyle: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "prev style"),
		),
	}
}

// MapPreview is a TUI component for previewing map layers
type MapPreview struct {
	// Configuration
	geoserverURL string
	username     string
	password     string
	workspace    string
	layerName    string
	styles       []string
	currentStyle int

	// View state
	width      int
	height     int
	visible    bool
	loading    bool
	errorMsg   string
	statusMsg  string
	imageData  []byte
	renderedImage string

	// Map state
	centerX    float64 // Center longitude
	centerY    float64 // Center latitude
	zoomLevel  int     // Zoom level (higher = more zoomed in)
	bbox       [4]float64 // Current bounding box [minX, minY, maxX, maxY]

	// Image rendering
	protocol   ImageProtocol
	imgWidth   int
	imgHeight  int

	// Components
	keyMap  MapPreviewKeyMap
	spinner spinner.Model

	// Callbacks
	onClose func()
}

// MapPreviewMsg is sent when the map image is loaded
type MapPreviewMsg struct {
	ImageData []byte
	Error     error
}

// MapPreviewCloseMsg is sent when the preview should close
type MapPreviewCloseMsg struct{}

// MapPreviewMetadataMsg is sent when layer metadata (bounds/styles) is ready
type MapPreviewMetadataMsg struct {
	Bounds *[4]float64 // [minX, minY, maxX, maxY] or nil if not available
	Styles []string
}

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
		zoomLevel:    2,
		centerX:      0,
		centerY:      0,
		bbox:         [4]float64{-180, -90, 180, 90}, // World extent
		imgWidth:     800,
		imgHeight:    600,
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
	m.imgWidth = (width - 20) * 8  // Approximate pixel width
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
		switch {
		case key.Matches(msg, m.keyMap.Close):
			m.visible = false
			if m.onClose != nil {
				m.onClose()
			}
			return m, nil

		case key.Matches(msg, m.keyMap.ZoomIn):
			m.zoomIn()
			return m, m.fetchMap()

		case key.Matches(msg, m.keyMap.ZoomOut):
			m.zoomOut()
			return m, m.fetchMap()

		case key.Matches(msg, m.keyMap.PanUp):
			m.panUp()
			return m, m.fetchMap()

		case key.Matches(msg, m.keyMap.PanDown):
			m.panDown()
			return m, m.fetchMap()

		case key.Matches(msg, m.keyMap.PanLeft):
			m.panLeft()
			return m, m.fetchMap()

		case key.Matches(msg, m.keyMap.PanRight):
			m.panRight()
			return m, m.fetchMap()

		case key.Matches(msg, m.keyMap.Refresh):
			m.loading = true
			return m, m.fetchMap()

		case key.Matches(msg, m.keyMap.NextStyle):
			m.nextStyle()
			return m, m.fetchMap()

		case key.Matches(msg, m.keyMap.PrevStyle):
			m.prevStyle()
			return m, m.fetchMap()
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case MapPreviewMsg:
		m.loading = false
		if msg.Error != nil {
			m.errorMsg = msg.Error.Error()
			m.imageData = nil
			m.renderedImage = ""
		} else {
			m.errorMsg = ""
			m.imageData = msg.ImageData
			m.renderedImage = m.renderImage()
		}

	case MapPreviewMetadataMsg:
		// Apply metadata and fetch map
		if msg.Bounds != nil {
			m.SetBounds(msg.Bounds[0], msg.Bounds[1], msg.Bounds[2], msg.Bounds[3])
		}
		if len(msg.Styles) > 0 {
			m.SetStyles(msg.Styles)
		}
		// Now fetch the map with proper bounds
		m.loading = true
		return m, m.fetchMap()
	}

	return m, tea.Batch(cmds...)
}

// View renders the component
func (m *MapPreview) View() string {
	if !m.visible {
		return ""
	}

	// Build the preview panel
	var content strings.Builder

	// Title
	styleName := "default"
	if len(m.styles) > 0 && m.currentStyle < len(m.styles) {
		styleName = m.styles[m.currentStyle]
	}
	title := styles.DialogTitleStyle.Render(fmt.Sprintf(" Layer Preview: %s:%s ", m.workspace, m.layerName))
	content.WriteString(title)
	content.WriteString("\n")

	// Style info bar
	styleInfo := styles.MutedStyle.Render(fmt.Sprintf("Style: %s  |  Zoom: %d  |  Protocol: %s",
		styleName, m.zoomLevel, m.protocolName()))
	content.WriteString(styleInfo)
	content.WriteString("\n\n")

	// Main content area with image and controls
	imagePanel := m.renderImagePanel()
	controlsPanel := m.renderControlsPanel()

	// Join image and controls horizontally
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, imagePanel, controlsPanel)
	content.WriteString(mainContent)
	content.WriteString("\n")

	// Status bar
	if m.errorMsg != "" {
		content.WriteString(styles.ErrorStyle.Render("Error: " + m.errorMsg))
	} else if m.loading {
		content.WriteString(m.spinner.View() + " Loading map...")
	} else {
		bbox := fmt.Sprintf("Bounds: [%.2f, %.2f, %.2f, %.2f]", m.bbox[0], m.bbox[1], m.bbox[2], m.bbox[3])
		content.WriteString(styles.MutedStyle.Render(bbox))
	}
	content.WriteString("\n")

	// Help bar
	helpItems := []string{
		styles.RenderHelpKey("arrows", "pan"),
		styles.RenderHelpKey("+/-", "zoom"),
		styles.RenderHelpKey("s/S", "style"),
		styles.RenderHelpKey("r", "refresh"),
		styles.RenderHelpKey("esc", "close"),
	}
	content.WriteString(strings.Join(helpItems, "  "))

	// Wrap in dialog box
	dialogWidth := m.width - 4
	if dialogWidth > 120 {
		dialogWidth = 120
	}

	dialog := styles.DialogBoxStyle.Width(dialogWidth).Render(content.String())

	return styles.Center(m.width, m.height, dialog)
}

// renderImagePanel renders the image area
func (m *MapPreview) renderImagePanel() string {
	// Calculate available image display size
	imgDisplayWidth := m.width - 30
	if imgDisplayWidth > 80 {
		imgDisplayWidth = 80
	}
	if imgDisplayWidth < 40 {
		imgDisplayWidth = 40
	}

	imgDisplayHeight := m.height - 15
	if imgDisplayHeight > 30 {
		imgDisplayHeight = 30
	}
	if imgDisplayHeight < 15 {
		imgDisplayHeight = 15
	}

	var imageContent string
	if m.loading {
		imageContent = styles.Center(imgDisplayWidth, imgDisplayHeight,
			m.spinner.View()+" Loading...")
	} else if m.errorMsg != "" {
		imageContent = styles.Center(imgDisplayWidth, imgDisplayHeight,
			styles.ErrorStyle.Render("Failed to load"))
	} else if m.renderedImage != "" {
		imageContent = m.renderedImage
	} else {
		imageContent = styles.Center(imgDisplayWidth, imgDisplayHeight,
			styles.MutedStyle.Render("[No image data]"))
	}

	// Wrap in a border
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Border).
		Width(imgDisplayWidth).
		Height(imgDisplayHeight)

	return panelStyle.Render(imageContent)
}

// renderControlsPanel renders the pan/zoom controls
func (m *MapPreview) renderControlsPanel() string {
	var controls strings.Builder

	// Zoom controls
	controls.WriteString(styles.DialogLabelStyle.Render(" Zoom"))
	controls.WriteString("\n")
	controls.WriteString(fmt.Sprintf("  %s\n", m.renderButton("+", "Zoom In", m.zoomLevel < 20)))
	controls.WriteString(fmt.Sprintf("  %s\n", m.renderButton("-", "Zoom Out", m.zoomLevel > 0)))
	controls.WriteString(fmt.Sprintf("  Level: %d\n", m.zoomLevel))
	controls.WriteString("\n")

	// Pan controls
	controls.WriteString(styles.DialogLabelStyle.Render(" Pan"))
	controls.WriteString("\n")
	controls.WriteString(fmt.Sprintf("     %s\n", m.renderButton("^", "Up", true)))
	controls.WriteString(fmt.Sprintf("  %s  %s  %s\n",
		m.renderButton("<", "Left", true),
		m.renderButton("o", "Center", true),
		m.renderButton(">", "Right", true)))
	controls.WriteString(fmt.Sprintf("     %s\n", m.renderButton("v", "Down", true)))
	controls.WriteString("\n")

	// Style selector
	controls.WriteString(styles.DialogLabelStyle.Render(" Style"))
	controls.WriteString("\n")
	if len(m.styles) > 0 {
		for i, style := range m.styles {
			prefix := "  "
			if i == m.currentStyle {
				prefix = "> "
				controls.WriteString(styles.AccentStyle.Render(prefix + style))
			} else {
				controls.WriteString(styles.MutedStyle.Render(prefix + style))
			}
			controls.WriteString("\n")
		}
	} else {
		controls.WriteString(styles.MutedStyle.Render("  (default)\n"))
	}
	controls.WriteString("\n")

	// Actions
	controls.WriteString(styles.DialogLabelStyle.Render(" Actions"))
	controls.WriteString("\n")
	controls.WriteString(fmt.Sprintf("  %s\n", m.renderButton("r", "Refresh", true)))
	controls.WriteString(fmt.Sprintf("  %s\n", m.renderButton("q", "Close", true)))

	// Wrap in panel style
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Border).
		Width(18).
		Padding(0, 1)

	return panelStyle.Render(controls.String())
}

// renderButton renders a control button
func (m *MapPreview) renderButton(key, _ string, enabled bool) string {
	if enabled {
		return lipgloss.NewStyle().
			Background(styles.SurfaceHigh).
			Foreground(styles.TextBright).
			Padding(0, 1).
			Render(key)
	}
	return lipgloss.NewStyle().
		Background(styles.Surface).
		Foreground(styles.Muted).
		Padding(0, 1).
		Render(key)
}

// fetchMap fetches the WMS GetMap image
func (m *MapPreview) fetchMap() tea.Cmd {
	return func() tea.Msg {
		m.loading = true

		// Build WMS GetMap URL
		style := ""
		if len(m.styles) > 0 && m.currentStyle < len(m.styles) {
			style = m.styles[m.currentStyle]
		}

		layers := fmt.Sprintf("%s:%s", m.workspace, m.layerName)
		wmsURL := fmt.Sprintf("%s/wms?SERVICE=WMS&VERSION=1.1.1&REQUEST=GetMap&LAYERS=%s&STYLES=%s&FORMAT=%s&TRANSPARENT=true&SRS=EPSG:4326&WIDTH=%d&HEIGHT=%d&BBOX=%f,%f,%f,%f",
			m.geoserverURL, url.QueryEscape(layers), url.QueryEscape(style), url.QueryEscape("image/png"), m.imgWidth, m.imgHeight,
			m.bbox[0], m.bbox[1], m.bbox[2], m.bbox[3])

		// Create HTTP request with auth
		req, err := http.NewRequest("GET", wmsURL, nil)
		if err != nil {
			return MapPreviewMsg{Error: err}
		}
		req.SetBasicAuth(m.username, m.password)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return MapPreviewMsg{Error: err}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return MapPreviewMsg{Error: fmt.Errorf("WMS error (%d): %s", resp.StatusCode, string(body))}
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return MapPreviewMsg{Error: err}
		}

		return MapPreviewMsg{ImageData: data}
	}
}

// renderImage converts the PNG image data to terminal graphics
func (m *MapPreview) renderImage() string {
	if len(m.imageData) == 0 {
		return ""
	}

	switch m.protocol {
	case ProtocolKitty:
		return m.renderKitty()
	case ProtocolSixel:
		return m.renderSixel()
	case ProtocolChafa:
		return m.renderChafa()
	default:
		return m.renderASCII()
	}
}

// renderKitty renders using Kitty graphics protocol
func (m *MapPreview) renderKitty() string {
	// Kitty protocol: ESC_cs; payload ESC\
	// a=T for transmit, f=100 for PNG format
	encoded := base64.StdEncoding.EncodeToString(m.imageData)

	// Split into chunks of 4096 bytes for transmission
	var result strings.Builder
	chunkSize := 4096
	for i := 0; i < len(encoded); i += chunkSize {
		end := i + chunkSize
		if end > len(encoded) {
			end = len(encoded)
		}
		chunk := encoded[i:end]

		// m=1 for more chunks, m=0 for last chunk
		more := 1
		if end >= len(encoded) {
			more = 0
		}

		if i == 0 {
			// First chunk includes action and format
			result.WriteString(fmt.Sprintf("\x1b_Ga=T,f=100,m=%d;%s\x1b\\", more, chunk))
		} else {
			result.WriteString(fmt.Sprintf("\x1b_Gm=%d;%s\x1b\\", more, chunk))
		}
	}

	return result.String()
}

// renderSixel renders using Sixel graphics
func (m *MapPreview) renderSixel() string {
	// Try to use img2sixel if available
	cmd := exec.Command("img2sixel", "-")
	cmd.Stdin = bytes.NewReader(m.imageData)

	output, err := cmd.Output()
	if err != nil {
		return m.renderASCII()
	}

	return string(output)
}

// renderChafa renders using chafa
func (m *MapPreview) renderChafa() string {
	// Calculate display size in characters
	displayWidth := m.width - 30
	if displayWidth > 80 {
		displayWidth = 80
	}
	displayHeight := m.height - 15
	if displayHeight > 30 {
		displayHeight = 30
	}

	// Create temp file for image
	tmpFile, err := os.CreateTemp("", "geoserver-preview-*.png")
	if err != nil {
		return m.renderASCII()
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(m.imageData); err != nil {
		tmpFile.Close()
		return m.renderASCII()
	}
	tmpFile.Close()

	// Run chafa
	cmd := exec.Command("chafa",
		"--size", fmt.Sprintf("%dx%d", displayWidth, displayHeight),
		"--colors", "256",
		tmpFile.Name())

	output, err := cmd.Output()
	if err != nil {
		return m.renderASCII()
	}

	return string(output)
}

// renderASCII renders a simple ASCII representation
func (m *MapPreview) renderASCII() string {
	// Decode PNG to get dimensions
	img, err := png.Decode(bytes.NewReader(m.imageData))
	if err != nil {
		return "[Image decode error]"
	}

	bounds := img.Bounds()

	// Calculate ASCII size
	asciiWidth := m.width - 32
	if asciiWidth > 60 {
		asciiWidth = 60
	}
	if asciiWidth < 20 {
		asciiWidth = 20
	}
	asciiHeight := m.height - 16
	if asciiHeight > 25 {
		asciiHeight = 25
	}
	if asciiHeight < 10 {
		asciiHeight = 10
	}

	// Sample pixels and convert to ASCII art
	chars := " .:-=+*#%@"
	xScale := float64(bounds.Dx()) / float64(asciiWidth)
	yScale := float64(bounds.Dy()) / float64(asciiHeight)

	var result strings.Builder
	for y := 0; y < asciiHeight; y++ {
		for x := 0; x < asciiWidth; x++ {
			px := int(float64(x) * xScale)
			py := int(float64(y) * yScale)
			if px >= bounds.Max.X {
				px = bounds.Max.X - 1
			}
			if py >= bounds.Max.Y {
				py = bounds.Max.Y - 1
			}

			r, g, b, a := img.At(px+bounds.Min.X, py+bounds.Min.Y).RGBA()
			if a < 128 {
				result.WriteByte(' ')
				continue
			}

			// Convert to grayscale
			gray := (r + g + b) / 3
			// Map to character (0-65535 range from RGBA)
			idx := int(gray * uint32(len(chars)-1) / 65535)
			result.WriteByte(chars[idx])
		}
		result.WriteByte('\n')
	}

	return result.String()
}

// Pan and zoom methods
func (m *MapPreview) zoomIn() {
	if m.zoomLevel < 20 {
		m.zoomLevel++
		m.updateBBox()
	}
}

func (m *MapPreview) zoomOut() {
	if m.zoomLevel > 0 {
		m.zoomLevel--
		m.updateBBox()
	}
}

func (m *MapPreview) panUp() {
	height := m.bbox[3] - m.bbox[1]
	m.centerY += height * 0.25
	m.updateBBox()
}

func (m *MapPreview) panDown() {
	height := m.bbox[3] - m.bbox[1]
	m.centerY -= height * 0.25
	m.updateBBox()
}

func (m *MapPreview) panLeft() {
	width := m.bbox[2] - m.bbox[0]
	m.centerX -= width * 0.25
	m.updateBBox()
}

func (m *MapPreview) panRight() {
	width := m.bbox[2] - m.bbox[0]
	m.centerX += width * 0.25
	m.updateBBox()
}

func (m *MapPreview) updateBBox() {
	// Calculate bbox size based on zoom level
	// Zoom 0 = world extent, each level halves the extent
	worldWidth := 360.0
	worldHeight := 180.0

	scale := 1.0 / float64(int(1)<<m.zoomLevel)
	width := worldWidth * scale
	height := worldHeight * scale

	m.bbox[0] = m.centerX - width/2
	m.bbox[1] = m.centerY - height/2
	m.bbox[2] = m.centerX + width/2
	m.bbox[3] = m.centerY + height/2

	// Clamp to world bounds
	if m.bbox[0] < -180 {
		m.bbox[0] = -180
		m.bbox[2] = m.bbox[0] + width
	}
	if m.bbox[2] > 180 {
		m.bbox[2] = 180
		m.bbox[0] = m.bbox[2] - width
	}
	if m.bbox[1] < -90 {
		m.bbox[1] = -90
		m.bbox[3] = m.bbox[1] + height
	}
	if m.bbox[3] > 90 {
		m.bbox[3] = 90
		m.bbox[1] = m.bbox[3] - height
	}
}

func (m *MapPreview) nextStyle() {
	if len(m.styles) > 0 {
		m.currentStyle = (m.currentStyle + 1) % len(m.styles)
	}
}

func (m *MapPreview) prevStyle() {
	if len(m.styles) > 0 {
		m.currentStyle--
		if m.currentStyle < 0 {
			m.currentStyle = len(m.styles) - 1
		}
	}
}

func (m *MapPreview) protocolName() string {
	switch m.protocol {
	case ProtocolKitty:
		return "Kitty"
	case ProtocolSixel:
		return "Sixel"
	case ProtocolChafa:
		return "Chafa"
	default:
		return "ASCII"
	}
}

// detectImageProtocol detects the best available image rendering protocol
func detectImageProtocol() ImageProtocol {
	// Check for Kitty graphics support
	if isKittyTerminal() {
		return ProtocolKitty
	}

	// Check for Sixel support (img2sixel available)
	if _, err := exec.LookPath("img2sixel"); err == nil {
		return ProtocolSixel
	}

	// Check for chafa
	if _, err := exec.LookPath("chafa"); err == nil {
		return ProtocolChafa
	}

	// Fallback to ASCII art
	return ProtocolASCII
}

// isKittyTerminal checks if we're running in a Kitty terminal
func isKittyTerminal() bool {
	term := os.Getenv("TERM")
	kitty := os.Getenv("KITTY_WINDOW_ID")
	return strings.Contains(term, "kitty") || kitty != ""
}

// decodeImage decodes image data to an image.Image
func decodeImage(data []byte) (image.Image, error) {
	return png.Decode(bytes.NewReader(data))
}
