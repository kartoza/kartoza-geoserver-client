package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/kartoza/kartoza-cloudbench/internal/models"
)

func (c *Client) GetStyles(workspace string) ([]models.Style, error) {
	var path string
	if workspace == "" {
		path = "/styles"
	} else {
		path = fmt.Sprintf("/workspaces/%s/styles", workspace)
	}

	resp, err := c.doRequest("GET", path, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get styles: %s", string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Fix GeoServer's empty string response quirk
	body = fixEmptyGeoServerResponse(body)

	var result struct {
		Styles struct {
			Style []models.Style `json:"style"`
		} `json:"styles"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode styles: %w", err)
	}

	for i := range result.Styles.Style {
		result.Styles.Style[i].Workspace = workspace
	}

	return result.Styles.Style, nil
}

func (c *Client) GetStyleSLD(workspace, styleName string) (string, error) {
	var path string
	if workspace == "" {
		path = fmt.Sprintf("/styles/%s.sld", styleName)
	} else {
		path = fmt.Sprintf("/workspaces/%s/styles/%s.sld", workspace, styleName)
	}

	url := c.baseURL + "/rest" + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Accept", "application/vnd.ogc.sld+xml")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get style SLD: %s", string(body))
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read style content: %w", err)
	}

	return string(content), nil
}

func (c *Client) CreateOrUpdateStyle(workspace, styleName, sldContent string) error {
	var basePath string
	if workspace == "" {
		basePath = "/styles"
	} else {
		basePath = fmt.Sprintf("/workspaces/%s/styles", workspace)
	}

	// First, try to check if the style exists
	checkPath := basePath + "/" + styleName
	checkResp, err := c.doRequest("GET", checkPath, nil, "")
	if err != nil {
		return err
	}
	checkResp.Body.Close()

	if checkResp.StatusCode == http.StatusOK {
		// Style exists, update it
		updatePath := basePath + "/" + styleName
		url := c.baseURL + "/rest" + updatePath
		req, err := http.NewRequest("PUT", url, strings.NewReader(sldContent))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.SetBasicAuth(c.username, c.password)
		req.Header.Set("Content-Type", "application/vnd.ogc.sld+xml")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to update style: %s", string(body))
		}
		return nil
	}

	// Style doesn't exist, create it
	// First create the style definition
	styleBody := map[string]interface{}{
		"style": map[string]interface{}{
			"name":     styleName,
			"filename": styleName + ".sld",
		},
	}
	resp, err := c.doJSONRequest("POST", basePath, styleBody)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create style definition: status %d", resp.StatusCode)
	}

	// Now upload the SLD content
	uploadPath := basePath + "/" + styleName
	url := c.baseURL + "/rest" + uploadPath
	req, err := http.NewRequest("PUT", url, strings.NewReader(sldContent))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/vnd.ogc.sld+xml")

	uploadResp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != http.StatusOK && uploadResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(uploadResp.Body)
		return fmt.Errorf("failed to upload style content: %s", string(body))
	}

	return nil
}

func (c *Client) GetStyleContent(workspace, styleName, format string) (string, error) {
	var path string
	var extension string
	var acceptHeader string

	switch format {
	case "css":
		extension = ".css"
		acceptHeader = "application/vnd.geoserver.geocss+css"
	case "mbstyle":
		extension = ".json"
		acceptHeader = "application/vnd.geoserver.mbstyle+json"
	default: // sld
		extension = ".sld"
		acceptHeader = "application/vnd.ogc.sld+xml"
	}

	if workspace == "" {
		path = fmt.Sprintf("/styles/%s%s", styleName, extension)
	} else {
		path = fmt.Sprintf("/workspaces/%s/styles/%s%s", workspace, styleName, extension)
	}

	url := c.baseURL + "/rest" + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Accept", acceptHeader)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get style content: %s", string(body))
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read style content: %w", err)
	}

	return string(content), nil
}

func (c *Client) UpdateStyleContent(workspace, styleName, content, format string) error {
	var path string
	var contentType string

	switch format {
	case "css":
		contentType = "application/vnd.geoserver.geocss+css"
	case "mbstyle":
		contentType = "application/vnd.geoserver.mbstyle+json"
	default: // sld
		contentType = "application/vnd.ogc.sld+xml"
	}

	if workspace == "" {
		path = fmt.Sprintf("/styles/%s", styleName)
	} else {
		path = fmt.Sprintf("/workspaces/%s/styles/%s", workspace, styleName)
	}

	url := c.baseURL + "/rest" + path
	req, err := http.NewRequest("PUT", url, strings.NewReader(content))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update style: %s", string(body))
	}

	return nil
}

func (c *Client) CreateStyle(workspace, styleName, content, format string) error {
	var basePath string
	var contentType string

	switch format {
	case "css":
		contentType = "application/vnd.geoserver.geocss+css"
	case "mbstyle":
		contentType = "application/vnd.geoserver.mbstyle+json"
	default: // sld
		contentType = "application/vnd.ogc.sld+xml"
	}

	if workspace == "" {
		basePath = "/styles"
	} else {
		basePath = fmt.Sprintf("/workspaces/%s/styles", workspace)
	}

	// Use the "raw" upload approach - POST the content directly with name parameter
	// This is simpler and more reliable than the two-step approach
	url := c.baseURL + "/rest" + basePath + "?name=" + neturl.QueryEscape(styleName)
	req, err := http.NewRequest("POST", url, strings.NewReader(content))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create style: status %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Client) UploadStyle(workspace, styleName, filePath string, format string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read style file: %w", err)
	}

	// First create the style
	var path string
	if workspace == "" {
		path = "/styles"
	} else {
		path = fmt.Sprintf("/workspaces/%s/styles", workspace)
	}

	// Determine content type based on format
	var contentType string
	switch strings.ToLower(format) {
	case "sld":
		contentType = "application/vnd.ogc.sld+xml"
	case "css":
		contentType = "application/vnd.geoserver.geocss+css"
	default:
		contentType = "application/vnd.ogc.sld+xml"
	}

	// Create style entry
	createBody := map[string]interface{}{
		"style": map[string]interface{}{
			"name":     styleName,
			"filename": filepath.Base(filePath),
		},
	}

	resp, err := c.doJSONRequest("POST", path, createBody)
	if err != nil {
		return err
	}
	resp.Body.Close()

	// Upload the actual style content
	uploadPath := path + "/" + styleName
	resp, err = c.doRequest("PUT", uploadPath, bytes.NewReader(content), contentType)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to upload style: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) DeleteStyle(workspace, name string, purge bool) error {
	var path string
	if workspace == "" {
		path = fmt.Sprintf("/styles/%s", name)
	} else {
		path = fmt.Sprintf("/workspaces/%s/styles/%s", workspace, name)
	}
	if purge {
		path += "?purge=true"
	}

	resp, err := c.doRequest("DELETE", path, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete style: %s", string(bodyBytes))
	}

	return nil
}

func (c *Client) DownloadStyle(workspace, name string) ([]byte, string, error) {
	// First get style info to determine format
	var infoPath string
	if workspace == "" {
		infoPath = fmt.Sprintf("/styles/%s", name)
	} else {
		infoPath = fmt.Sprintf("/workspaces/%s/styles/%s", workspace, name)
	}

	resp, err := c.doRequest("GET", infoPath, nil, "")
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("style not found: %s", name)
	}

	var styleInfo struct {
		Style struct {
			Format   string `json:"format"`
			Filename string `json:"filename"`
		} `json:"style"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&styleInfo); err != nil {
		return nil, "", err
	}

	// Now get the actual style content
	format := styleInfo.Style.Format
	if format == "" {
		format = "sld"
	}

	// Determine content type and extension
	var contentType, ext string
	switch format {
	case "css":
		contentType = "application/vnd.geoserver.geocss+css"
		ext = ".css"
	case "mbstyle":
		contentType = "application/vnd.geoserver.mbstyle+json"
		ext = ".json"
	default: // sld
		contentType = "application/vnd.ogc.sld+xml"
		ext = ".sld"
	}

	// Create request for style content
	url := c.baseURL + "/rest" + infoPath + ".sld"
	if format == "css" {
		url = c.baseURL + "/rest" + infoPath + ".css"
	} else if format == "mbstyle" {
		url = c.baseURL + "/rest" + infoPath + ".mbstyle"
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Accept", contentType)

	resp2, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to download style content")
	}

	data, err := io.ReadAll(resp2.Body)
	return data, ext, err
}

