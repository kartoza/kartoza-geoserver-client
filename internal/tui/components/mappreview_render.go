package components

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kartoza/kartoza-cloudbench/internal/tui/styles"
)

// updateCompositeImage draws crosshair and overlays onto the base image
func (m *MapPreview) updateCompositeImage() {
	if m.baseImage == nil {
		m.compositeImage = m.imageData // Fallback to original
		return
	}

	bounds := m.baseImage.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create a new RGBA image to draw on
	composite := image.NewRGBA(bounds)

	// Copy base image to composite
	draw.Draw(composite, bounds, m.baseImage, bounds.Min, draw.Src)

	// Only draw crosshair if overlay is enabled
	if m.showOverlay {
		// Calculate crosshair position in pixels
		crossX := int(m.crosshairX * float64(width-1))
		crossY := int(m.crosshairY * float64(height-1))

		// Crosshair colors
		lineColor := color.RGBA{255, 68, 68, 255}     // Red
		centerColor := color.RGBA{255, 255, 255, 255} // White

		// Draw vertical line
		for y := 0; y < height; y++ {
			if y == crossY {
				continue // Skip center point
			}
			composite.Set(crossX, y, lineColor)
		}

		// Draw horizontal line
		for x := 0; x < width; x++ {
			if x == crossX {
				continue // Skip center point
			}
			composite.Set(x, crossY, lineColor)
		}

		// Draw center point (3x3 white square for visibility)
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				px := crossX + dx
				py := crossY + dy
				if px >= 0 && px < width && py >= 0 && py < height {
					composite.Set(px, py, centerColor)
				}
			}
		}

		// Draw coordinate text overlay at top-left
		m.drawTextOverlay(composite, width)
	}

	// Draw legend in bottom-left corner if available
	if m.legendImage != nil {
		m.drawLegendOverlay(composite, width, height)
	} else {
		// Draw a placeholder box to show where legend would go
		m.drawLegendPlaceholder(composite, width, height)
	}

	// Draw feature info overlay if showing
	if m.showFeatureInfo && m.featureInfo != "" {
		m.drawFeatureInfoOverlay(composite, width, height)
	}

	// Encode composite to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, composite); err == nil {
		m.compositeImage = buf.Bytes()
	} else {
		m.compositeImage = m.imageData // Fallback
	}
}

// drawTextOverlay draws coordinate and help text on the image
func (m *MapPreview) drawTextOverlay(img *image.RGBA, width int) {
	// Draw a semi-transparent black bar at the top for text
	barHeight := 20
	barColor := color.RGBA{0, 0, 0, 200} // Semi-transparent black

	bounds := img.Bounds()
	for y := 0; y < barHeight && y < bounds.Dy(); y++ {
		for x := 0; x < width; x++ {
			// Blend with existing pixel
			existing := img.RGBAAt(x, y)
			blended := blendColors(existing, barColor)
			img.Set(x, y, blended)
		}
	}

	// Note: Drawing actual text would require a font rendering library
	// For now, the bar provides visual separation and the coordinates
	// are shown in the terminal help bar rendered separately
}

// drawLegendOverlay draws the legend image in the bottom-left corner
func (m *MapPreview) drawLegendOverlay(composite *image.RGBA, width, height int) {
	if m.legendImage == nil {
		return
	}

	legendBounds := m.legendImage.Bounds()
	legendWidth := legendBounds.Dx()
	legendHeight := legendBounds.Dy()

	// Position legend in bottom-left with 10px padding
	padding := 10
	destX := padding
	destY := height - legendHeight - padding

	// Draw semi-transparent background behind legend
	bgColor := color.RGBA{0, 0, 0, 180}
	bgPadding := 5
	for y := destY - bgPadding; y < destY+legendHeight+bgPadding && y < height; y++ {
		if y < 0 {
			continue
		}
		for x := destX - bgPadding; x < destX+legendWidth+bgPadding && x < width; x++ {
			if x < 0 {
				continue
			}
			existing := composite.RGBAAt(x, y)
			blended := blendColors(existing, bgColor)
			composite.Set(x, y, blended)
		}
	}

	// Draw legend image
	for y := 0; y < legendHeight; y++ {
		for x := 0; x < legendWidth; x++ {
			destPx := destX + x
			destPy := destY + y
			if destPx >= 0 && destPx < width && destPy >= 0 && destPy < height {
				c := m.legendImage.At(legendBounds.Min.X+x, legendBounds.Min.Y+y)
				r, g, b, a := c.RGBA()
				if a > 0 {
					composite.Set(destPx, destPy, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)})
				}
			}
		}
	}
}

// drawLegendPlaceholder draws a placeholder box where the legend would go
func (m *MapPreview) drawLegendPlaceholder(composite *image.RGBA, width, height int) {
	// Draw a yellow box in bottom-left to show legend area (for debugging)
	boxWidth := 60
	boxHeight := 40
	padding := 10
	destX := padding
	destY := height - boxHeight - padding

	// Draw yellow background
	bgColor := color.RGBA{255, 200, 0, 200} // Yellow
	for y := destY; y < destY+boxHeight && y < height; y++ {
		if y < 0 {
			continue
		}
		for x := destX; x < destX+boxWidth && x < width; x++ {
			if x < 0 {
				continue
			}
			composite.Set(x, y, bgColor)
		}
	}

	// Draw border
	borderColor := color.RGBA{200, 150, 0, 255}
	for x := destX; x < destX+boxWidth && x < width; x++ {
		if destY >= 0 && destY < height {
			composite.Set(x, destY, borderColor)
		}
		endY := destY + boxHeight - 1
		if endY >= 0 && endY < height {
			composite.Set(x, endY, borderColor)
		}
	}
	for y := destY; y < destY+boxHeight && y < height; y++ {
		if y >= 0 {
			composite.Set(destX, y, borderColor)
			endX := destX + boxWidth - 1
			if endX < width {
				composite.Set(endX, y, borderColor)
			}
		}
	}
}

// drawFeatureInfoOverlay draws the feature info in the bottom-right corner
func (m *MapPreview) drawFeatureInfoOverlay(composite *image.RGBA, width, height int) {
	// Draw a semi-transparent box in bottom-right for feature info
	// Since we can't render text directly, we draw a colored box to indicate info is available
	boxWidth := 150
	boxHeight := 80
	padding := 10

	startX := width - boxWidth - padding
	startY := height - boxHeight - padding

	// Draw semi-transparent background
	bgColor := color.RGBA{0, 50, 100, 220} // Dark blue
	for y := startY; y < startY+boxHeight && y < height; y++ {
		if y < 0 {
			continue
		}
		for x := startX; x < startX+boxWidth && x < width; x++ {
			if x < 0 {
				continue
			}
			existing := composite.RGBAAt(x, y)
			blended := blendColors(existing, bgColor)
			composite.Set(x, y, blended)
		}
	}

	// Draw a border
	borderColor := color.RGBA{100, 150, 255, 255} // Light blue border
	// Top border
	for x := startX; x < startX+boxWidth && x < width; x++ {
		if x >= 0 && startY >= 0 && startY < height {
			composite.Set(x, startY, borderColor)
		}
	}
	// Bottom border
	endY := startY + boxHeight - 1
	for x := startX; x < startX+boxWidth && x < width; x++ {
		if x >= 0 && endY >= 0 && endY < height {
			composite.Set(x, endY, borderColor)
		}
	}
	// Left border
	for y := startY; y < startY+boxHeight && y < height; y++ {
		if startX >= 0 && y >= 0 {
			composite.Set(startX, y, borderColor)
		}
	}
	// Right border
	endX := startX + boxWidth - 1
	for y := startY; y < startY+boxHeight && y < height; y++ {
		if endX >= 0 && endX < width && y >= 0 {
			composite.Set(endX, y, borderColor)
		}
	}

	// Draw an "i" icon to indicate info (simple pixel art)
	iconX := startX + 10
	iconY := startY + 10
	iconColor := color.RGBA{255, 255, 255, 255}
	// Draw "i" - dot
	if iconX >= 0 && iconX < width && iconY >= 0 && iconY < height {
		composite.Set(iconX, iconY, iconColor)
		composite.Set(iconX+1, iconY, iconColor)
	}
	// Draw "i" - stem
	for dy := 3; dy < 12; dy++ {
		if iconX >= 0 && iconX < width && iconY+dy >= 0 && iconY+dy < height {
			composite.Set(iconX, iconY+dy, iconColor)
			composite.Set(iconX+1, iconY+dy, iconColor)
		}
	}
}

// blendColors blends two colors with alpha
func blendColors(base, overlay color.RGBA) color.RGBA {
	alpha := float64(overlay.A) / 255.0
	return color.RGBA{
		R: uint8(float64(base.R)*(1-alpha) + float64(overlay.R)*alpha),
		G: uint8(float64(base.G)*(1-alpha) + float64(overlay.G)*alpha),
		B: uint8(float64(base.B)*(1-alpha) + float64(overlay.B)*alpha),
		A: 255,
	}
}

// renderCompositeImage renders the composite image (with crosshair) to terminal
func (m *MapPreview) renderCompositeImage() string {
	if len(m.compositeImage) == 0 {
		return m.renderImage() // Fallback to original rendering
	}

	// Use the composite image data for rendering
	originalData := m.imageData
	m.imageData = m.compositeImage
	result := m.renderImage()
	m.imageData = originalData

	return result
}

// renderFeatureInfoPopup renders the GetFeatureInfo result as a popup
func (m *MapPreview) renderFeatureInfoPopup() string {
	// Truncate long info
	info := m.featureInfo
	lines := strings.Split(info, "\n")
	maxLines := 10
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, "... (truncated)")
	}
	info = strings.Join(lines, "\n")

	// Style the popup
	popupStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.KartozaOrange).
		Background(lipgloss.Color("#1a1a2e")).
		Foreground(styles.TextBright).
		Padding(0, 1).
		MaxWidth(m.width - 4)

	headerStyle := lipgloss.NewStyle().
		Foreground(styles.KartozaOrangeLight).
		Bold(true)

	content := headerStyle.Render("Feature Info") + "\n" + info

	return popupStyle.Render(content)
}

// renderLayerPanel renders the layer toggle panel for layer groups
func (m *MapPreview) renderLayerPanel() string {
	var b strings.Builder

	// Panel header
	headerStyle := lipgloss.NewStyle().
		Background(styles.Accent).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1).
		Bold(true)
	b.WriteString(headerStyle.Render(" Layer Settings "))
	b.WriteString("\n")

	// Help text
	helpStyle := lipgloss.NewStyle().Foreground(styles.Muted).Italic(true)
	b.WriteString(helpStyle.Render("  ↑/↓ navigate  Space toggle  ←/→ change style  a apply  Esc close"))
	b.WriteString("\n\n")

	// Calculate max layer name width for alignment
	maxNameLen := 0
	for _, layer := range m.groupLayers {
		if len(layer.Name) > maxNameLen {
			maxNameLen = len(layer.Name)
		}
	}

	// Layer list
	for i, layer := range m.groupLayers {
		// Checkbox
		checkbox := "[ ]"
		if layer.Enabled {
			checkbox = "[✓]"
		}

		// Style indicator
		styleText := ""
		if len(layer.AvailableStyles) > 0 {
			styleName := layer.CurrentStyle
			if styleName == "" {
				styleName = "(default)"
			}
			// Show style with arrows if multiple styles available
			if len(layer.AvailableStyles) > 1 {
				styleText = fmt.Sprintf("  ◀ %s ▶", styleName)
			} else {
				styleText = fmt.Sprintf("  [%s]", styleName)
			}
		}

		// Style based on selection and enabled state
		var lineStyle, styleStyle lipgloss.Style
		if i == m.layerPanelCursor {
			lineStyle = lipgloss.NewStyle().
				Background(styles.KartozaBlueLight).
				Foreground(styles.TextBright).
				Bold(true)
			styleStyle = lipgloss.NewStyle().
				Background(styles.KartozaBlueLight).
				Foreground(styles.KartozaOrangeLight).
				Bold(true)
		} else if layer.Enabled {
			lineStyle = lipgloss.NewStyle().Foreground(styles.TextBright)
			styleStyle = lipgloss.NewStyle().Foreground(styles.KartozaOrange)
		} else {
			lineStyle = lipgloss.NewStyle().Foreground(styles.Muted)
			styleStyle = lipgloss.NewStyle().Foreground(styles.Muted)
		}

		// Pad layer name for alignment
		paddedName := layer.Name
		for len(paddedName) < maxNameLen {
			paddedName += " "
		}

		line := fmt.Sprintf("  %s %s", checkbox, paddedName)
		b.WriteString(lineStyle.Render(line))
		if styleText != "" {
			b.WriteString(styleStyle.Render(styleText))
		}
		b.WriteString("\n")
	}

	// Wrap in a border
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Accent).
		Padding(0, 1)

	return panelStyle.Render(b.String())
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

// renderKitty renders for Kitty terminal using chafa with kitty protocol
func (m *MapPreview) renderKitty() string {
	// Calculate display size - use most of screen since controls are at top
	displayWidth := m.width - 4
	if displayWidth > 120 {
		displayWidth = 120
	}
	if displayWidth < 40 {
		displayWidth = 40
	}
	displayHeight := m.height - 8 // Leave room for title and control bar
	if displayHeight > 50 {
		displayHeight = 50
	}
	if displayHeight < 15 {
		displayHeight = 15
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

	// Use chafa with kitty format for best quality in Kitty terminal
	// Use --clear flag to clear previous images and avoid ghosting
	cmd := exec.Command("chafa",
		"--format", "kitty",
		"--size", fmt.Sprintf("%dx%d", displayWidth, displayHeight),
		"--colors", "full",
		"--color-space", "rgb",
		"--clear",
		tmpFile.Name())

	output, err := cmd.Output()
	if err != nil {
		// Fallback to symbols format if kitty format fails
		cmd = exec.Command("chafa",
			"--format", "symbols",
			"--size", fmt.Sprintf("%dx%d", displayWidth, displayHeight),
			"--colors", "full",
			tmpFile.Name())
		output, err = cmd.Output()
		if err != nil {
			return m.renderASCII()
		}
	}

	return string(output)
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
	// Calculate display size - use most of screen since controls are at top
	displayWidth := m.width - 4
	if displayWidth > 120 {
		displayWidth = 120
	}
	if displayWidth < 40 {
		displayWidth = 40
	}
	displayHeight := m.height - 8
	if displayHeight > 50 {
		displayHeight = 50
	}
	if displayHeight < 15 {
		displayHeight = 15
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

	// Calculate ASCII size - use most of screen since controls are at top
	asciiWidth := m.width - 4
	if asciiWidth > 100 {
		asciiWidth = 100
	}
	if asciiWidth < 40 {
		asciiWidth = 40
	}
	asciiHeight := m.height - 8
	if asciiHeight > 40 {
		asciiHeight = 40
	}
	if asciiHeight < 15 {
		asciiHeight = 15
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

// decodeImage decodes image data to an image.Image
func decodeImage(data []byte) (image.Image, error) {
	return png.Decode(bytes.NewReader(data))
}
