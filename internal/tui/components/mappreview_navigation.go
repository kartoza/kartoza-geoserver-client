package components

import "math"

// Pan and zoom methods

func (m *MapPreview) zoomIn() {
	if m.zoomLevel < 20 {
		m.zoomLevel += 0.5 // Zoom in by half a level
		m.updateBBox()
	}
}

func (m *MapPreview) zoomOut() {
	if m.zoomLevel > 0 {
		m.zoomLevel -= 0.5 // Zoom out by half a level
		m.updateBBox()
	}
}

func (m *MapPreview) panUp() {
	height := m.bbox[3] - m.bbox[1]
	m.centerY += height * 0.125 // Move by 12.5% of view
	m.updateBBox()
}

func (m *MapPreview) panDown() {
	height := m.bbox[3] - m.bbox[1]
	m.centerY -= height * 0.125 // Move by 12.5% of view
	m.updateBBox()
}

func (m *MapPreview) panLeft() {
	width := m.bbox[2] - m.bbox[0]
	m.centerX -= width * 0.125 // Move by 12.5% of view
	m.updateBBox()
}

func (m *MapPreview) panRight() {
	width := m.bbox[2] - m.bbox[0]
	m.centerX += width * 0.125 // Move by 12.5% of view
	m.updateBBox()
}

func (m *MapPreview) updateBBox() {
	// Calculate bbox size based on zoom level
	// Zoom 0 = world extent, each level halves the extent
	worldWidth := 360.0
	worldHeight := 180.0

	// Use math.Pow for fractional zoom levels
	scale := 1.0 / math.Pow(2, m.zoomLevel)
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

// moveCrosshair moves the crosshair by the given delta (normalized 0-1)
func (m *MapPreview) moveCrosshair(dx, dy int) {
	// Move crosshair by pixel count - calculate normalized delta based on image dimensions
	if m.baseImage == nil {
		return
	}
	bounds := m.baseImage.Bounds()
	width := float64(bounds.Dx())
	height := float64(bounds.Dy())

	// Convert pixel delta to normalized coordinates
	m.crosshairX += float64(dx) / width
	m.crosshairY += float64(dy) / height

	// Clamp to valid range
	if m.crosshairX < 0 {
		m.crosshairX = 0
	}
	if m.crosshairX > 1 {
		m.crosshairX = 1
	}
	if m.crosshairY < 0 {
		m.crosshairY = 0
	}
	if m.crosshairY > 1 {
		m.crosshairY = 1
	}
	// Hide feature info when crosshair moves
	m.showFeatureInfo = false
}

// getCrosshairCoordinates returns the geographic coordinates at the crosshair position
func (m *MapPreview) getCrosshairCoordinates() (lon, lat float64) {
	// Convert normalized position to geographic coordinates
	lon = m.bbox[0] + (m.bbox[2]-m.bbox[0])*m.crosshairX
	lat = m.bbox[3] - (m.bbox[3]-m.bbox[1])*m.crosshairY // Y is inverted
	return lon, lat
}

// getCrosshairPixelPosition returns the pixel position for GetFeatureInfo request
func (m *MapPreview) getCrosshairPixelPosition() (x, y int) {
	x = int(m.crosshairX * float64(m.imgWidth))
	y = int(m.crosshairY * float64(m.imgHeight))
	return x, y
}
