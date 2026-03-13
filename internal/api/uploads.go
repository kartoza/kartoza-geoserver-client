package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var uploadClient = &http.Client{Timeout: 30 * time.Minute}

func (c *Client) UploadShapefile(workspace, storeName, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	return c.UploadShapefileFrom(workspace, storeName, file, info.Size())
}

func (c *Client) UploadShapefileFrom(workspace, storeName string, r io.Reader, contentLength int64) error {
	path := fmt.Sprintf("/workspaces/%s/datastores/%s/file.shp", workspace, storeName)
	req, err := http.NewRequest("PUT", c.baseURL+"/rest"+path, r)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.ContentLength = contentLength
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/zip")
	resp, err := uploadClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s", string(bodyBytes))
	}
	return nil
}

func (c *Client) UploadGeoTIFF(workspace, storeName, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	return c.UploadGeoTIFFFrom(workspace, storeName, file, info.Size())
}

func (c *Client) UploadGeoTIFFFrom(workspace, storeName string, r io.Reader, contentLength int64) error {
	path := fmt.Sprintf("/workspaces/%s/coveragestores/%s/file.geotiff", workspace, storeName)
	req, err := http.NewRequest("PUT", c.baseURL+"/rest"+path, r)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.ContentLength = contentLength
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "image/tiff")
	resp, err := uploadClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s", string(bodyBytes))
	}
	return nil
}

func (c *Client) UploadGeoPackage(workspace, storeName, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	return c.UploadGeoPackageFrom(workspace, storeName, file, info.Size())
}

func (c *Client) UploadGeoPackageFrom(workspace, storeName string, r io.Reader, contentLength int64) error {
	path := fmt.Sprintf("/workspaces/%s/datastores/%s/file.gpkg", workspace, storeName)
	req, err := http.NewRequest("PUT", c.baseURL+"/rest"+path, r)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.ContentLength = contentLength
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/geopackage+sqlite3")
	resp, err := uploadClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s", string(bodyBytes))
	}
	return nil
}

func parseOWSException(data []byte, fallbackPrefix string) error {
	content := string(data)

	// Common error patterns and their user-friendly messages
	errorMappings := map[string]string{
		"idle-session timeout":       "The database connection timed out. Please try again.",
		"terminating connection":     "The database connection was lost. Please try again.",
		"Could not find layer":       "Layer not found. It may have been deleted or renamed.",
		"No such feature type":       "This layer type cannot be downloaded as a shapefile.",
		"Feature type not found":     "Layer not found on the server.",
		"Unknown coverage":           "Coverage not found on the server.",
		"InvalidParameterValue":      "Invalid request parameters.",
		"MissingParameterValue":      "Missing required parameters.",
		"OperationNotSupported":      "This operation is not supported for this layer.",
		"java.lang.OutOfMemoryError": "The server ran out of memory. The dataset may be too large to download.",
		"Connection refused":         "Cannot connect to the database server.",
		"authentication failed":      "Database authentication failed.",
		"does not exist":             "The requested resource does not exist.",
		"permission denied":          "Permission denied. Check your credentials.",
		"WFS is not enabled":         "WFS service is not enabled for this layer.",
		"WCS is not enabled":         "WCS service is not enabled for this coverage.",
	}

	// Check for known error patterns
	for pattern, message := range errorMappings {
		if strings.Contains(strings.ToLower(content), strings.ToLower(pattern)) {
			return fmt.Errorf("%s", message)
		}
	}

	// Try to extract ExceptionText from OWS XML
	if strings.Contains(content, "ExceptionText") {
		start := strings.Index(content, "<ows:ExceptionText>")
		if start == -1 {
			start = strings.Index(content, "<ExceptionText>")
		}
		if start != -1 {
			// Find the text content
			textStart := strings.Index(content[start:], ">") + start + 1
			end := strings.Index(content[textStart:], "</")
			if end != -1 {
				exceptionText := content[textStart : textStart+end]
				// Clean up the exception text
				exceptionText = strings.ReplaceAll(exceptionText, "\n", " ")
				exceptionText = strings.TrimSpace(exceptionText)
				// Truncate if too long
				if len(exceptionText) > 200 {
					exceptionText = exceptionText[:200] + "..."
				}
				return fmt.Errorf("%s: %s", fallbackPrefix, exceptionText)
			}
		}
	}

	// Fallback: just return a generic message
	return fmt.Errorf("%s: Server returned an error. Please check the layer configuration.", fallbackPrefix)
}

// DownloadLayerAsShapefile downloads a layer's data via WFS as a zipped shapefile
func (c *Client) UploadShapefileData(workspace, storeName string, data []byte) error {
	path := fmt.Sprintf("/workspaces/%s/datastores/%s/file.shp", workspace, storeName)

	req, err := http.NewRequest("PUT", c.baseURL+"/rest"+path, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/zip")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// DownloadCoverageAsGeoTIFF downloads a coverage via WCS as GeoTIFF
func (c *Client) UploadGeoTIFFData(workspace, storeName string, data []byte) error {
	path := fmt.Sprintf("/workspaces/%s/coveragestores/%s/file.geotiff", workspace, storeName)

	req, err := http.NewRequest("PUT", c.baseURL+"/rest"+path, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "image/tiff")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
