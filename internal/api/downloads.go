package api

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (c *Client) DownloadLayerAsShapefile(workspace, layerName string) ([]byte, error) {
	// Build WFS GetFeature URL with SHAPE-ZIP output format
	wfsURL := fmt.Sprintf("%s/wfs?service=WFS&version=1.1.0&request=GetFeature&typeName=%s:%s&outputFormat=SHAPE-ZIP",
		c.baseURL, workspace, layerName)

	req, err := http.NewRequest("GET", wfsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("WFS request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, parseOWSException(bodyBytes, fmt.Sprintf("WFS request failed (%d)", resp.StatusCode))
	}

	// Check content type - GeoServer returns application/xml for errors even with 200 OK
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "xml") || strings.Contains(contentType, "text/") {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, parseOWSException(bodyBytes, "WFS download failed")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Verify we got a ZIP file (first 2 bytes should be PK)
	if len(data) < 4 || data[0] != 'P' || data[1] != 'K' {
		return nil, parseOWSException(data, "Invalid shapefile response")
	}

	return data, nil
}

func (c *Client) DownloadCoverageAsGeoTIFF(workspace, coverageName string) ([]byte, error) {
	// Build WCS GetCoverage URL - use workspace__coveragename format for CoverageId
	wcsURL := fmt.Sprintf("%s/wcs?service=WCS&version=2.0.1&request=GetCoverage&CoverageId=%s__%s&format=image/geotiff",
		c.baseURL, workspace, coverageName)

	req, err := http.NewRequest("GET", wcsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("WCS request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, parseOWSException(bodyBytes, fmt.Sprintf("WCS request failed (%d)", resp.StatusCode))
	}

	// Check content type - GeoServer returns application/xml for errors even with 200 OK
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "xml") || strings.Contains(contentType, "text/") {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, parseOWSException(bodyBytes, "WCS download failed")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Verify we got a TIFF file (first 2 bytes should be II or MM for little/big endian)
	if len(data) < 4 || !((data[0] == 'I' && data[1] == 'I') || (data[0] == 'M' && data[1] == 'M')) {
		return nil, parseOWSException(data, "Invalid GeoTIFF response")
	}

	return data, nil
}

