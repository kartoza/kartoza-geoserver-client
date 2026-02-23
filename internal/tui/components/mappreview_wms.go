package components

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// fetchMap fetches the WMS GetMap image
func (m *MapPreview) fetchMap() tea.Cmd {
	return func() tea.Msg {
		m.loading = true

		// Build WMS GetMap URL
		style := ""
		if len(m.styles) > 0 && m.currentStyle < len(m.styles) {
			style = m.styles[m.currentStyle]
		}

		// Build layer list and styles
		var layers string
		var layerStyles string
		if m.isLayerGroup && m.canToggleLayers() && len(m.groupLayers) > 0 {
			// For layer groups with togglable layers, request only enabled layers
			enabledLayers := []string{}
			enabledStyles := []string{}
			for _, layer := range m.groupLayers {
				if layer.Enabled {
					enabledLayers = append(enabledLayers, layer.Name)
					// Add the style for this layer (empty string uses default)
					enabledStyles = append(enabledStyles, layer.CurrentStyle)
				}
			}
			if len(enabledLayers) == 0 {
				// No layers enabled, show an empty/placeholder
				return MapPreviewMsg{Error: fmt.Errorf("no layers enabled")}
			}
			layers = strings.Join(enabledLayers, ",")
			layerStyles = strings.Join(enabledStyles, ",")
		} else {
			// Single layer or layer group as whole
			layers = fmt.Sprintf("%s:%s", m.workspace, m.layerName)
			layerStyles = style
		}

		wmsURL := fmt.Sprintf("%s/wms?SERVICE=WMS&VERSION=1.1.1&REQUEST=GetMap&LAYERS=%s&STYLES=%s&FORMAT=%s&TRANSPARENT=true&SRS=EPSG:4326&WIDTH=%d&HEIGHT=%d&BBOX=%f,%f,%f,%f",
			m.geoserverURL, url.QueryEscape(layers), url.QueryEscape(layerStyles), url.QueryEscape("image/png"), m.imgWidth, m.imgHeight,
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

// fetchFeatureInfo performs a WMS GetFeatureInfo request at the crosshair location
func (m *MapPreview) fetchFeatureInfo() tea.Cmd {
	return func() tea.Msg {
		// Get pixel position
		pixelX, pixelY := m.getCrosshairPixelPosition()

		// Build layer name
		var layers string
		if m.isLayerGroup && m.canToggleLayers() && len(m.groupLayers) > 0 {
			enabledLayers := []string{}
			for _, layer := range m.groupLayers {
				if layer.Enabled {
					enabledLayers = append(enabledLayers, layer.Name)
				}
			}
			if len(enabledLayers) == 0 {
				return FeatureInfoMsg{Error: fmt.Errorf("no layers enabled")}
			}
			layers = strings.Join(enabledLayers, ",")
		} else {
			layers = fmt.Sprintf("%s:%s", m.workspace, m.layerName)
		}

		// Build WMS GetFeatureInfo URL
		wfsURL := fmt.Sprintf("%s/wms?SERVICE=WMS&VERSION=1.1.1&REQUEST=GetFeatureInfo&LAYERS=%s&QUERY_LAYERS=%s&INFO_FORMAT=%s&SRS=EPSG:4326&WIDTH=%d&HEIGHT=%d&BBOX=%f,%f,%f,%f&X=%d&Y=%d",
			m.geoserverURL,
			url.QueryEscape(layers),
			url.QueryEscape(layers),
			url.QueryEscape("text/plain"),
			m.imgWidth, m.imgHeight,
			m.bbox[0], m.bbox[1], m.bbox[2], m.bbox[3],
			pixelX, pixelY)

		// Create HTTP request with auth
		req, err := http.NewRequest("GET", wfsURL, nil)
		if err != nil {
			return FeatureInfoMsg{Error: err}
		}
		req.SetBasicAuth(m.username, m.password)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return FeatureInfoMsg{Error: err}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return FeatureInfoMsg{Error: fmt.Errorf("WMS error (%d): %s", resp.StatusCode, string(body))}
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return FeatureInfoMsg{Error: err}
		}

		info := strings.TrimSpace(string(data))
		if info == "" {
			info = "No features found at this location"
		}

		return FeatureInfoMsg{Info: info}
	}
}

// fetchLegend performs a WMS GetLegendGraphic request
func (m *MapPreview) fetchLegend() tea.Cmd {
	return func() tea.Msg {
		// Build layer name
		var layer string
		if m.workspace != "" {
			layer = fmt.Sprintf("%s:%s", m.workspace, m.layerName)
		} else {
			layer = m.layerName
		}

		// Get current style if any
		style := ""
		if len(m.styles) > 0 && m.currentStyle < len(m.styles) {
			style = m.styles[m.currentStyle]
		}

		// Build WMS GetLegendGraphic URL
		legendURL := fmt.Sprintf("%s/wms?SERVICE=WMS&VERSION=1.1.1&REQUEST=GetLegendGraphic&LAYER=%s&FORMAT=image/png&WIDTH=20&HEIGHT=20",
			m.geoserverURL,
			url.QueryEscape(layer))

		if style != "" {
			legendURL += "&STYLE=" + url.QueryEscape(style)
		}

		// Create HTTP request with auth
		req, err := http.NewRequest("GET", legendURL, nil)
		if err != nil {
			return LegendMsg{Error: err}
		}
		req.SetBasicAuth(m.username, m.password)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return LegendMsg{Error: err}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return LegendMsg{Error: fmt.Errorf("legend request failed: %d", resp.StatusCode)}
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return LegendMsg{Error: err}
		}

		return LegendMsg{ImageData: data}
	}
}
