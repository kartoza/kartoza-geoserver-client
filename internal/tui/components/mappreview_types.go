package components

import (
	"image"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
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
	Close           key.Binding
	ZoomIn          key.Binding
	ZoomOut         key.Binding
	PanUp           key.Binding
	PanDown         key.Binding
	PanLeft         key.Binding
	PanRight        key.Binding
	Refresh         key.Binding
	NextStyle       key.Binding
	PrevStyle       key.Binding
	ToggleLayers    key.Binding
	CrosshairUp     key.Binding
	CrosshairDown   key.Binding
	CrosshairLeft   key.Binding
	CrosshairRight  key.Binding
	GetFeatureInfo  key.Binding
	ToggleOverlay   key.Binding
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
			key.WithKeys("right"),
			key.WithHelp("right", "pan right"),
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
		ToggleLayers: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "toggle layers"),
		),
		CrosshairUp: key.NewBinding(
			key.WithKeys("shift+up", "K"),
			key.WithHelp("Shift+↑/K", "crosshair up"),
		),
		CrosshairDown: key.NewBinding(
			key.WithKeys("shift+down", "J"),
			key.WithHelp("Shift+↓/J", "crosshair down"),
		),
		CrosshairLeft: key.NewBinding(
			key.WithKeys("shift+left", "H"),
			key.WithHelp("Shift+←/H", "crosshair left"),
		),
		CrosshairRight: key.NewBinding(
			key.WithKeys("shift+right", "L"),
			key.WithHelp("Shift+→/L", "crosshair right"),
		),
		GetFeatureInfo: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "feature info"),
		),
		ToggleOverlay: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "toggle overlay"),
		),
	}
}

// LayerToggleItem represents a layer that can be toggled on/off in layer group preview
type LayerToggleItem struct {
	Name            string
	Enabled         bool
	CurrentStyle    string   // Currently selected style
	AvailableStyles []string // Available styles for this layer
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

	// Layer group support
	isLayerGroup     bool              // Whether we're viewing a layer group
	layerGroupMode   string            // SINGLE, NAMED, CONTAINER, EO
	groupLayers      []LayerToggleItem // Layers in the group with enabled state
	showLayerPanel   bool              // Whether layer toggle panel is visible
	layerPanelCursor int               // Cursor position in layer panel

	// View state
	width         int
	height        int
	visible       bool
	loading       bool
	errorMsg      string
	statusMsg     string
	imageData     []byte
	renderedImage string

	// Double buffering - keep previous image while loading new one
	previousImage string

	// Map state
	centerX   float64    // Center longitude
	centerY   float64    // Center latitude
	zoomLevel float64    // Zoom level (higher = more zoomed in)
	bbox      [4]float64 // Current bounding box [minX, minY, maxX, maxY]

	// Image rendering
	protocol  ImageProtocol
	imgWidth  int
	imgHeight int

	// Crosshair state - position in terminal character coordinates (0.0-1.0 normalized)
	crosshairX      float64 // Normalized X position (0.0 = left, 1.0 = right)
	crosshairY      float64 // Normalized Y position (0.0 = top, 1.0 = bottom)
	showOverlay     bool    // Whether to show overlay controls on map
	featureInfo     string  // Last GetFeatureInfo result
	showFeatureInfo bool    // Whether to show feature info popup

	// Cached base image for overlay rendering (without crosshair)
	baseImage      image.Image // The original decoded PNG from WMS (without crosshair)
	compositeImage []byte      // PNG data with crosshair drawn in, ready for rendering

	// Legend image cached from GetLegendGraphic
	legendImage   image.Image // Cached legend image
	legendFetched bool        // Whether legend has been fetched

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

// FeatureInfoMsg is sent when GetFeatureInfo results are ready
type FeatureInfoMsg struct {
	Info  string
	Error error
}

// LegendMsg is sent when GetLegendGraphic results are ready
type LegendMsg struct {
	ImageData []byte
	Error     error
}

// LayerGroupLayerInfo contains info about a layer in a layer group
type LayerGroupLayerInfo struct {
	Name            string
	DefaultStyle    string   // Style assigned in the layer group
	AvailableStyles []string // All available styles for this layer
}

// MapPreviewMetadataMsg is sent when layer metadata (bounds/styles) is ready
type MapPreviewMetadataMsg struct {
	Bounds *[4]float64 // [minX, minY, maxX, maxY] or nil if not available
	Styles []string
	// Layer group specific
	IsLayerGroup   bool
	LayerGroupMode string                // SINGLE, NAMED, CONTAINER, EO
	GroupLayers    []LayerGroupLayerInfo // Layers in the group with style info
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
