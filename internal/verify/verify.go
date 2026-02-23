// Copyright 2026 Kartoza
// SPDX-License-Identifier: MIT

package verify

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
)

// LayerInfo contains metadata about a layer
type LayerInfo struct {
	FeatureCount int64
	BBox         BoundingBox
	GeometryType string
	Attributes   []Attribute
	CRS          string
}

// BoundingBox represents a geographic bounding box
type BoundingBox struct {
	MinX float64
	MinY float64
	MaxX float64
	MaxY float64
}

// Attribute represents a layer attribute/field
type Attribute struct {
	Name string
	Type string
}

// VerificationResult contains the comparison results
type VerificationResult struct {
	Success        bool
	FeatureCountOK bool
	BBoxOK         bool
	GeometryTypeOK bool
	AttributesOK   bool
	LocalInfo      *LayerInfo
	RemoteInfo     *LayerInfo
	Errors         []string
	Warnings       []string
}

// GetLocalLayerInfo reads metadata from a local geospatial file using ogrinfo
func GetLocalLayerInfo(filePath string) (*LayerInfo, error) {
	// Use ogrinfo to get layer information
	// -al: all layers, -so: summary only, -json: JSON output
	cmd := exec.Command("ogrinfo", "-al", "-so", "-json", filePath)
	output, err := cmd.Output()
	if err != nil {
		// Try without -json flag for older GDAL versions
		return getLocalLayerInfoLegacy(filePath)
	}

	var result struct {
		Layers []struct {
			Name         string `json:"name"`
			FeatureCount int64  `json:"featureCount"`
			GeometryType string `json:"geometryType"`
			Fields       []struct {
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"fields"`
			GeometryFields []struct {
				CoordinateSystem struct {
					Wkt string `json:"wkt"`
				} `json:"coordinateSystem"`
			} `json:"geometryFields"`
		} `json:"layers"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return getLocalLayerInfoLegacy(filePath)
	}

	if len(result.Layers) == 0 {
		return nil, fmt.Errorf("no layers found in file")
	}

	layer := result.Layers[0]
	info := &LayerInfo{
		FeatureCount: layer.FeatureCount,
		GeometryType: normalizeGeometryType(layer.GeometryType),
		Attributes:   make([]Attribute, len(layer.Fields)),
	}

	for i, f := range layer.Fields {
		info.Attributes[i] = Attribute{
			Name: f.Name,
			Type: normalizeFieldType(f.Type),
		}
	}

	// Get bounding box using ogrinfo extent
	bbox, err := getLocalBBox(filePath)
	if err == nil {
		info.BBox = bbox
	}

	return info, nil
}

// getLocalLayerInfoLegacy uses text parsing for older GDAL versions
func getLocalLayerInfoLegacy(filePath string) (*LayerInfo, error) {
	cmd := exec.Command("ogrinfo", "-al", "-so", filePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ogrinfo failed: %w", err)
	}

	info := &LayerInfo{
		Attributes: []Attribute{},
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Feature Count:") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				count, _ := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
				info.FeatureCount = count
			}
		}

		if strings.HasPrefix(line, "Geometry:") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				info.GeometryType = normalizeGeometryType(strings.TrimSpace(parts[1]))
			}
		}

		if strings.HasPrefix(line, "Extent:") {
			// Parse: Extent: (minx, miny) - (maxx, maxy)
			line = strings.TrimPrefix(line, "Extent:")
			line = strings.ReplaceAll(line, "(", "")
			line = strings.ReplaceAll(line, ")", "")
			line = strings.ReplaceAll(line, " ", "")
			parts := strings.Split(line, "-")
			if len(parts) == 2 {
				minParts := strings.Split(parts[0], ",")
				maxParts := strings.Split(parts[1], ",")
				if len(minParts) == 2 && len(maxParts) == 2 {
					info.BBox.MinX, _ = strconv.ParseFloat(minParts[0], 64)
					info.BBox.MinY, _ = strconv.ParseFloat(minParts[1], 64)
					info.BBox.MaxX, _ = strconv.ParseFloat(maxParts[0], 64)
					info.BBox.MaxY, _ = strconv.ParseFloat(maxParts[1], 64)
				}
			}
		}

		// Parse field definitions (e.g., "name: String (80.0)")
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "Layer") &&
			!strings.HasPrefix(line, "Feature") && !strings.HasPrefix(line, "Geometry") &&
			!strings.HasPrefix(line, "Extent") && !strings.HasPrefix(line, "INFO") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				fieldName := strings.TrimSpace(parts[0])
				fieldType := strings.TrimSpace(parts[1])
				// Extract just the type name
				if idx := strings.Index(fieldType, " "); idx > 0 {
					fieldType = fieldType[:idx]
				}
				if fieldName != "" && fieldType != "" && !strings.HasPrefix(fieldName, "INFO") {
					info.Attributes = append(info.Attributes, Attribute{
						Name: fieldName,
						Type: normalizeFieldType(fieldType),
					})
				}
			}
		}
	}

	return info, nil
}

// getLocalBBox gets the bounding box from a local file
func getLocalBBox(filePath string) (BoundingBox, error) {
	cmd := exec.Command("ogrinfo", "-al", "-so", filePath)
	output, err := cmd.Output()
	if err != nil {
		return BoundingBox{}, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "Extent:") {
			// Parse: Extent: (minx, miny) - (maxx, maxy)
			line = strings.TrimPrefix(strings.TrimSpace(line), "Extent:")
			line = strings.ReplaceAll(line, "(", "")
			line = strings.ReplaceAll(line, ")", "")
			line = strings.ReplaceAll(line, " ", "")
			parts := strings.Split(line, "-")
			if len(parts) == 2 {
				minParts := strings.Split(parts[0], ",")
				maxParts := strings.Split(parts[1], ",")
				if len(minParts) == 2 && len(maxParts) == 2 {
					bbox := BoundingBox{}
					bbox.MinX, _ = strconv.ParseFloat(minParts[0], 64)
					bbox.MinY, _ = strconv.ParseFloat(minParts[1], 64)
					bbox.MaxX, _ = strconv.ParseFloat(maxParts[0], 64)
					bbox.MaxY, _ = strconv.ParseFloat(maxParts[1], 64)
					return bbox, nil
				}
			}
		}
	}

	return BoundingBox{}, fmt.Errorf("extent not found")
}

// GetRemoteLayerInfo gets metadata from a GeoServer layer via WFS
func GetRemoteLayerInfo(geoserverURL, workspace, layerName, username, password string) (*LayerInfo, error) {
	info := &LayerInfo{
		Attributes: []Attribute{},
	}

	// Get feature count and attributes via WFS DescribeFeatureType and GetFeature
	wfsURL := fmt.Sprintf("%s/%s/wfs", geoserverURL, workspace)
	typeName := fmt.Sprintf("%s:%s", workspace, layerName)

	// Get feature count using resultType=hits
	countURL := fmt.Sprintf("%s?SERVICE=WFS&VERSION=2.0.0&REQUEST=GetFeature&TYPENAMES=%s&resultType=hits",
		wfsURL, url.QueryEscape(typeName))

	client := &http.Client{}
	req, err := http.NewRequest("GET", countURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if username != "" {
		req.SetBasicAuth(username, password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("WFS request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Parse numberMatched from response
	if strings.Contains(bodyStr, "numberMatched") {
		// Try to extract numberMatched attribute
		start := strings.Index(bodyStr, "numberMatched=\"")
		if start > 0 {
			start += len("numberMatched=\"")
			end := strings.Index(bodyStr[start:], "\"")
			if end > 0 {
				countStr := bodyStr[start : start+end]
				info.FeatureCount, _ = strconv.ParseInt(countStr, 10, 64)
			}
		}
	}

	// Get a sample feature to determine attributes and geometry type
	sampleURL := fmt.Sprintf("%s?SERVICE=WFS&VERSION=2.0.0&REQUEST=GetFeature&TYPENAMES=%s&COUNT=1&OUTPUTFORMAT=application/json",
		wfsURL, url.QueryEscape(typeName))

	req, err = http.NewRequest("GET", sampleURL, nil)
	if err != nil {
		return info, nil
	}
	if username != "" {
		req.SetBasicAuth(username, password)
	}

	resp, err = client.Do(req)
	if err != nil {
		return info, nil
	}
	defer resp.Body.Close()

	var geoJSON struct {
		Features []struct {
			Geometry struct {
				Type string `json:"type"`
			} `json:"geometry"`
			Properties map[string]interface{} `json:"properties"`
		} `json:"features"`
		BBox []float64 `json:"bbox"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geoJSON); err == nil {
		if len(geoJSON.Features) > 0 {
			info.GeometryType = normalizeGeometryType(geoJSON.Features[0].Geometry.Type)

			// Extract attribute names and types
			for name, value := range geoJSON.Features[0].Properties {
				attrType := "String"
				switch value.(type) {
				case float64:
					attrType = "Real"
				case int, int64:
					attrType = "Integer"
				case bool:
					attrType = "Integer"
				}
				info.Attributes = append(info.Attributes, Attribute{
					Name: name,
					Type: attrType,
				})
			}
		}

		if len(geoJSON.BBox) >= 4 {
			info.BBox = BoundingBox{
				MinX: geoJSON.BBox[0],
				MinY: geoJSON.BBox[1],
				MaxX: geoJSON.BBox[2],
				MaxY: geoJSON.BBox[3],
			}
		}
	}

	// If bbox not from GeoJSON, try WMS GetCapabilities
	if info.BBox.MinX == 0 && info.BBox.MaxX == 0 {
		bbox, err := getRemoteBBox(geoserverURL, workspace, layerName, username, password)
		if err == nil {
			info.BBox = bbox
		}
	}

	return info, nil
}

// getRemoteBBox gets bounding box from WMS GetCapabilities
func getRemoteBBox(geoserverURL, workspace, layerName, username, password string) (BoundingBox, error) {
	wmsURL := fmt.Sprintf("%s/%s/wms?SERVICE=WMS&VERSION=1.1.1&REQUEST=GetCapabilities", geoserverURL, workspace)

	client := &http.Client{}
	req, err := http.NewRequest("GET", wmsURL, nil)
	if err != nil {
		return BoundingBox{}, err
	}
	if username != "" {
		req.SetBasicAuth(username, password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return BoundingBox{}, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Find the layer section and extract LatLonBoundingBox
	layerStart := strings.Index(bodyStr, "<Name>"+layerName+"</Name>")
	if layerStart < 0 {
		layerStart = strings.Index(bodyStr, "<Name>"+workspace+":"+layerName+"</Name>")
	}

	if layerStart > 0 {
		searchArea := bodyStr[layerStart:]
		bboxStart := strings.Index(searchArea, "<LatLonBoundingBox")
		if bboxStart > 0 {
			bboxEnd := strings.Index(searchArea[bboxStart:], "/>")
			if bboxEnd > 0 {
				bboxStr := searchArea[bboxStart : bboxStart+bboxEnd]
				bbox := BoundingBox{}

				if minx := extractAttr(bboxStr, "minx"); minx != "" {
					bbox.MinX, _ = strconv.ParseFloat(minx, 64)
				}
				if miny := extractAttr(bboxStr, "miny"); miny != "" {
					bbox.MinY, _ = strconv.ParseFloat(miny, 64)
				}
				if maxx := extractAttr(bboxStr, "maxx"); maxx != "" {
					bbox.MaxX, _ = strconv.ParseFloat(maxx, 64)
				}
				if maxy := extractAttr(bboxStr, "maxy"); maxy != "" {
					bbox.MaxY, _ = strconv.ParseFloat(maxy, 64)
				}

				return bbox, nil
			}
		}
	}

	return BoundingBox{}, fmt.Errorf("bounding box not found")
}

func extractAttr(s, attr string) string {
	search := attr + "=\""
	start := strings.Index(s, search)
	if start < 0 {
		return ""
	}
	start += len(search)
	end := strings.Index(s[start:], "\"")
	if end < 0 {
		return ""
	}
	return s[start : start+end]
}

// VerifyUpload compares local and remote layer info
func VerifyUpload(local, remote *LayerInfo) *VerificationResult {
	result := &VerificationResult{
		Success:    true,
		LocalInfo:  local,
		RemoteInfo: remote,
		Errors:     []string{},
		Warnings:   []string{},
	}

	// Compare feature count
	if local.FeatureCount == remote.FeatureCount {
		result.FeatureCountOK = true
	} else {
		result.FeatureCountOK = false
		result.Success = false
		result.Errors = append(result.Errors,
			fmt.Sprintf("Feature count mismatch: local=%d, remote=%d",
				local.FeatureCount, remote.FeatureCount))
	}

	// Compare bounding box (with tolerance for floating point)
	tolerance := 0.0001
	bboxMatch := math.Abs(local.BBox.MinX-remote.BBox.MinX) < tolerance &&
		math.Abs(local.BBox.MinY-remote.BBox.MinY) < tolerance &&
		math.Abs(local.BBox.MaxX-remote.BBox.MaxX) < tolerance &&
		math.Abs(local.BBox.MaxY-remote.BBox.MaxY) < tolerance

	if bboxMatch {
		result.BBoxOK = true
	} else {
		// Check if it's significantly different or just floating point issues
		largeTolerance := 0.01
		roughMatch := math.Abs(local.BBox.MinX-remote.BBox.MinX) < largeTolerance &&
			math.Abs(local.BBox.MinY-remote.BBox.MinY) < largeTolerance &&
			math.Abs(local.BBox.MaxX-remote.BBox.MaxX) < largeTolerance &&
			math.Abs(local.BBox.MaxY-remote.BBox.MaxY) < largeTolerance

		if roughMatch {
			result.BBoxOK = true
			result.Warnings = append(result.Warnings,
				"Bounding box has minor differences (likely floating point precision)")
		} else {
			result.BBoxOK = false
			result.Success = false
			result.Errors = append(result.Errors,
				fmt.Sprintf("Bounding box mismatch: local=(%.4f,%.4f,%.4f,%.4f), remote=(%.4f,%.4f,%.4f,%.4f)",
					local.BBox.MinX, local.BBox.MinY, local.BBox.MaxX, local.BBox.MaxY,
					remote.BBox.MinX, remote.BBox.MinY, remote.BBox.MaxX, remote.BBox.MaxY))
		}
	}

	// Compare geometry type
	if local.GeometryType == remote.GeometryType ||
		strings.Contains(local.GeometryType, remote.GeometryType) ||
		strings.Contains(remote.GeometryType, local.GeometryType) {
		result.GeometryTypeOK = true
	} else {
		result.GeometryTypeOK = false
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Geometry type difference: local=%s, remote=%s",
				local.GeometryType, remote.GeometryType))
	}

	// Compare attributes (just check that remote has all local attributes)
	localAttrMap := make(map[string]string)
	for _, a := range local.Attributes {
		localAttrMap[strings.ToLower(a.Name)] = a.Type
	}

	remoteAttrMap := make(map[string]string)
	for _, a := range remote.Attributes {
		remoteAttrMap[strings.ToLower(a.Name)] = a.Type
	}

	missingAttrs := []string{}
	for name := range localAttrMap {
		if _, ok := remoteAttrMap[name]; !ok {
			missingAttrs = append(missingAttrs, name)
		}
	}

	if len(missingAttrs) == 0 {
		result.AttributesOK = true
	} else {
		result.AttributesOK = false
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Missing attributes in remote layer: %s", strings.Join(missingAttrs, ", ")))
	}

	return result
}

// normalizeGeometryType normalizes geometry type names
func normalizeGeometryType(geomType string) string {
	geomType = strings.ToLower(geomType)
	geomType = strings.ReplaceAll(geomType, "multi ", "multi")

	switch {
	case strings.Contains(geomType, "point"):
		if strings.Contains(geomType, "multi") {
			return "MultiPoint"
		}
		return "Point"
	case strings.Contains(geomType, "line"):
		if strings.Contains(geomType, "multi") {
			return "MultiLineString"
		}
		return "LineString"
	case strings.Contains(geomType, "polygon"):
		if strings.Contains(geomType, "multi") {
			return "MultiPolygon"
		}
		return "Polygon"
	default:
		return geomType
	}
}

// normalizeFieldType normalizes field type names
func normalizeFieldType(fieldType string) string {
	fieldType = strings.ToLower(fieldType)

	switch {
	case strings.Contains(fieldType, "int"):
		return "Integer"
	case strings.Contains(fieldType, "real"), strings.Contains(fieldType, "float"), strings.Contains(fieldType, "double"):
		return "Real"
	case strings.Contains(fieldType, "date"):
		return "Date"
	default:
		return "String"
	}
}

// FormatResult returns a human-readable summary of the verification result
func (r *VerificationResult) FormatResult() string {
	var sb strings.Builder

	if r.Success {
		sb.WriteString("✓ Upload verification PASSED\n\n")
	} else {
		sb.WriteString("✗ Upload verification FAILED\n\n")
	}

	sb.WriteString("Checks:\n")
	sb.WriteString(fmt.Sprintf("  Feature Count: %s (local: %d, remote: %d)\n",
		statusIcon(r.FeatureCountOK), r.LocalInfo.FeatureCount, r.RemoteInfo.FeatureCount))
	sb.WriteString(fmt.Sprintf("  Bounding Box:  %s\n", statusIcon(r.BBoxOK)))
	sb.WriteString(fmt.Sprintf("  Geometry Type: %s (local: %s, remote: %s)\n",
		statusIcon(r.GeometryTypeOK), r.LocalInfo.GeometryType, r.RemoteInfo.GeometryType))
	sb.WriteString(fmt.Sprintf("  Attributes:    %s (%d fields)\n",
		statusIcon(r.AttributesOK), len(r.LocalInfo.Attributes)))

	if len(r.Errors) > 0 {
		sb.WriteString("\nErrors:\n")
		for _, e := range r.Errors {
			sb.WriteString(fmt.Sprintf("  • %s\n", e))
		}
	}

	if len(r.Warnings) > 0 {
		sb.WriteString("\nWarnings:\n")
		for _, w := range r.Warnings {
			sb.WriteString(fmt.Sprintf("  • %s\n", w))
		}
	}

	return sb.String()
}

func statusIcon(ok bool) string {
	if ok {
		return "\uf00c" // fa-check
	}
	return "\uf00d" // fa-times
}
