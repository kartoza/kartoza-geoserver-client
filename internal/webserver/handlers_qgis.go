package webserver

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kartoza/kartoza-cloudbench/internal/config"
)

// handleQGISProjects handles QGIS project API endpoints
// GET /api/qgis/projects - List all projects
// POST /api/qgis/projects - Add a new project
func (s *Server) handleQGISProjects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listQGISProjects(w, r)
	case http.MethodPost:
		s.createQGISProject(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleQGISProjectByID handles single QGIS project operations
// GET /api/qgis/projects/{id} - Get project details
// PUT /api/qgis/projects/{id} - Update project
// DELETE /api/qgis/projects/{id} - Remove project from list
// GET /api/qgis/projects/{id}/file - Get project file content
// GET /api/qgis/projects/{id}/metadata - Get parsed project metadata with layers
func (s *Server) handleQGISProjectByID(w http.ResponseWriter, r *http.Request) {
	// Parse project ID from path: /api/qgis/projects/{id} or /api/qgis/projects/{id}/file
	path := strings.TrimPrefix(r.URL.Path, "/api/qgis/projects/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Project ID required", http.StatusBadRequest)
		return
	}

	projectID := parts[0]

	// Check if this is a file request
	if len(parts) > 1 && parts[1] == "file" {
		s.getQGISProjectFile(w, r, projectID)
		return
	}

	// Check if this is a metadata request
	if len(parts) > 1 && parts[1] == "metadata" {
		s.getQGISProjectMetadata(w, r, projectID)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getQGISProject(w, r, projectID)
	case http.MethodPut:
		s.updateQGISProject(w, r, projectID)
	case http.MethodDelete:
		s.deleteQGISProject(w, r, projectID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listQGISProjects returns all registered QGIS projects
func (s *Server) listQGISProjects(w http.ResponseWriter, r *http.Request) {
	projects := s.config.QGISProjects
	if projects == nil {
		projects = []config.QGISProject{}
	}

	// Update file info for each project
	for i := range projects {
		if info, err := os.Stat(projects[i].Path); err == nil {
			projects[i].Size = info.Size()
			projects[i].LastModified = info.ModTime().Format(time.RFC3339)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

// createQGISProject adds a new QGIS project via file upload
func (s *Server) createQGISProject(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 100MB)
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, "Failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get project name from form field
	name := r.FormValue("name")

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".qgs" && ext != ".qgz" {
		http.Error(w, "File must be a QGIS project file (.qgs or .qgz)", http.StatusBadRequest)
		return
	}

	// Get the QGIS projects directory
	projectsDir, err := config.QGISProjectsDir()
	if err != nil {
		http.Error(w, "Failed to get projects directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate unique ID for the project
	projectID := uuid.New().String()

	// Use filename as name if not provided
	if name == "" {
		name = strings.TrimSuffix(header.Filename, ext)
	}

	// Create unique filename to avoid collisions
	destFilename := projectID + ext
	destPath := filepath.Join(projectsDir, destFilename)

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "Failed to create destination file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	// Copy uploaded file to destination
	written, err := io.Copy(destFile, file)
	if err != nil {
		os.Remove(destPath) // Clean up on error
		http.Error(w, "Failed to save uploaded file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	project := config.QGISProject{
		ID:           projectID,
		Name:         name,
		Path:         destPath,
		LastModified: time.Now().Format(time.RFC3339),
		Size:         written,
	}

	// Add to config and save
	s.config.QGISProjects = append(s.config.QGISProjects, project)
	if err := s.config.Save(); err != nil {
		os.Remove(destPath) // Clean up on error
		http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

// getQGISProject returns details for a single project
func (s *Server) getQGISProject(w http.ResponseWriter, r *http.Request, projectID string) {
	for _, project := range s.config.QGISProjects {
		if project.ID == projectID {
			// Update file info
			if info, err := os.Stat(project.Path); err == nil {
				project.Size = info.Size()
				project.LastModified = info.ModTime().Format(time.RFC3339)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(project)
			return
		}
	}

	http.Error(w, "Project not found", http.StatusNotFound)
}

// updateQGISProject updates a project's metadata
func (s *Server) updateQGISProject(w http.ResponseWriter, r *http.Request, projectID string) {
	var input struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	for i := range s.config.QGISProjects {
		if s.config.QGISProjects[i].ID == projectID {
			if input.Name != "" {
				s.config.QGISProjects[i].Name = input.Name
			}
			if input.Path != "" {
				// Validate new path
				ext := strings.ToLower(filepath.Ext(input.Path))
				if ext != ".qgs" && ext != ".qgz" {
					http.Error(w, "Path must be a QGIS project file (.qgs or .qgz)", http.StatusBadRequest)
					return
				}

				info, err := os.Stat(input.Path)
				if err != nil {
					http.Error(w, "Cannot access file: "+err.Error(), http.StatusBadRequest)
					return
				}

				s.config.QGISProjects[i].Path = input.Path
				s.config.QGISProjects[i].LastModified = info.ModTime().Format(time.RFC3339)
				s.config.QGISProjects[i].Size = info.Size()
			}

			if err := s.config.Save(); err != nil {
				http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(s.config.QGISProjects[i])
			return
		}
	}

	http.Error(w, "Project not found", http.StatusNotFound)
}

// deleteQGISProject removes a project from the list and deletes the uploaded file
func (s *Server) deleteQGISProject(w http.ResponseWriter, r *http.Request, projectID string) {
	for i, project := range s.config.QGISProjects {
		if project.ID == projectID {
			// Check if the file is in our managed projects directory and delete it
			projectsDir, err := config.QGISProjectsDir()
			if err == nil && strings.HasPrefix(project.Path, projectsDir) {
				// This is an uploaded file, safe to delete
				if err := os.Remove(project.Path); err != nil && !os.IsNotExist(err) {
					// Log but don't fail - the file might already be gone
					// TODO: Add proper logging
				}
			}

			// Remove from slice
			s.config.QGISProjects = append(s.config.QGISProjects[:i], s.config.QGISProjects[i+1:]...)

			if err := s.config.Save(); err != nil {
				http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	http.Error(w, "Project not found", http.StatusNotFound)
}

// QGISProjectMetadata contains parsed information from a QGIS project file
type QGISProjectMetadata struct {
	Title       string       `json:"title"`
	CRS         string       `json:"crs"`
	Extent      *QGISExtent  `json:"extent,omitempty"`
	Layers      []QGISLayer  `json:"layers"`
	Version     string       `json:"version"`
	SaveUser    string       `json:"saveUser,omitempty"`
	SaveDate    string       `json:"saveDate,omitempty"`
}

// QGISExtent represents the map extent
type QGISExtent struct {
	XMin float64 `json:"xMin"`
	YMin float64 `json:"yMin"`
	XMax float64 `json:"xMax"`
	YMax float64 `json:"yMax"`
}

// QGISLayer represents a layer in the project
type QGISLayer struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`       // "raster", "vector", "xyz", "wms", etc.
	Provider   string `json:"provider"`   // "gdal", "ogr", "wms", etc.
	Source     string `json:"source"`     // Layer source/path
	Visible    bool   `json:"visible"`
	TileURL    string `json:"tileUrl,omitempty"`    // For XYZ/TMS layers
	WMSURL     string `json:"wmsUrl,omitempty"`     // For WMS layers
	WMSLayers  string `json:"wmsLayers,omitempty"`  // WMS layer names
}

// getQGISProjectMetadata parses a QGIS project file and returns metadata
func (s *Server) getQGISProjectMetadata(w http.ResponseWriter, r *http.Request, projectID string) {
	for _, project := range s.config.QGISProjects {
		if project.ID == projectID {
			metadata, err := parseQGISProject(project.Path)
			if err != nil {
				http.Error(w, "Failed to parse project: "+err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(metadata)
			return
		}
	}

	http.Error(w, "Project not found", http.StatusNotFound)
}

// parseQGISProject reads and parses a QGIS project file (.qgs or .qgz)
func parseQGISProject(path string) (*QGISProjectMetadata, error) {
	ext := strings.ToLower(filepath.Ext(path))

	var xmlData []byte
	var err error

	if ext == ".qgz" {
		// QGZ is a ZIP file containing the .qgs XML
		xmlData, err = readQGZProject(path)
	} else {
		xmlData, err = os.ReadFile(path)
	}

	if err != nil {
		return nil, err
	}

	return parseQGISXML(xmlData)
}

// readQGZProject extracts the .qgs file from a .qgz archive
func readQGZProject(path string) ([]byte, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.HasSuffix(strings.ToLower(f.Name), ".qgs") {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}

	return nil, io.EOF // No .qgs file found
}

// parseQGISXML parses the QGIS project XML
func parseQGISXML(data []byte) (*QGISProjectMetadata, error) {
	// QGIS XML structure (simplified)
	type qgisProject struct {
		XMLName     xml.Name `xml:"qgis"`
		Version     string   `xml:"version,attr"`
		ProjectName string   `xml:"projectname,attr"`
		SaveUser    string   `xml:"saveUser,attr"`
		SaveDate    string   `xml:"saveDateTime,attr"`
		Title       string   `xml:"title"`
		ProjectCRS  struct {
			AuthID string `xml:"spatialrefsys>authid"`
		} `xml:"projectCrs"`
		MapCanvas struct {
			Extent struct {
				XMin string `xml:"xmin"`
				YMin string `xml:"ymin"`
				XMax string `xml:"xmax"`
				YMax string `xml:"ymax"`
			} `xml:"extent"`
		} `xml:"mapcanvas"`
		LayerTreeGroup struct {
			Layers []struct {
				ID       string `xml:"id,attr"`
				Name     string `xml:"name,attr"`
				Source   string `xml:"source,attr"`
				Provider string `xml:"providerKey,attr"`
				Checked  string `xml:"checked,attr"`
			} `xml:"layer-tree-layer"`
		} `xml:"layer-tree-group"`
		ProjectLayers struct {
			MapLayers []struct {
				ID         string `xml:"id"`
				LayerName  string `xml:"layername"`
				DataSource string `xml:"datasource"`
				Provider   string `xml:"provider"`
				Type       string `xml:"type,attr"`
			} `xml:"maplayer"`
		} `xml:"projectlayers"`
	}

	var proj qgisProject
	if err := xml.Unmarshal(data, &proj); err != nil {
		return nil, err
	}

	metadata := &QGISProjectMetadata{
		Title:    proj.Title,
		CRS:      proj.ProjectCRS.AuthID,
		Version:  proj.Version,
		SaveUser: proj.SaveUser,
		SaveDate: proj.SaveDate,
		Layers:   []QGISLayer{},
	}

	if metadata.Title == "" {
		metadata.Title = proj.ProjectName
	}

	// Parse extent
	if proj.MapCanvas.Extent.XMin != "" {
		metadata.Extent = &QGISExtent{}
		parseFloat := func(s string) float64 {
			var f float64
			// Use json to parse float - it handles scientific notation etc.
			_ = json.Unmarshal([]byte(s), &f)
			return f
		}
		metadata.Extent.XMin = parseFloat(proj.MapCanvas.Extent.XMin)
		metadata.Extent.YMin = parseFloat(proj.MapCanvas.Extent.YMin)
		metadata.Extent.XMax = parseFloat(proj.MapCanvas.Extent.XMax)
		metadata.Extent.YMax = parseFloat(proj.MapCanvas.Extent.YMax)
	}

	// Build layer visibility map from layer tree
	visibilityMap := make(map[string]bool)
	for _, l := range proj.LayerTreeGroup.Layers {
		visibilityMap[l.ID] = l.Checked == "Qt::Checked"
	}

	// Parse layers from projectlayers
	for _, ml := range proj.ProjectLayers.MapLayers {
		layer := QGISLayer{
			ID:       ml.ID,
			Name:     ml.LayerName,
			Type:     ml.Type,
			Provider: ml.Provider,
			Source:   ml.DataSource,
			Visible:  visibilityMap[ml.ID],
		}

		// Parse XYZ tile URLs
		if ml.Provider == "wms" && strings.Contains(ml.DataSource, "type=xyz") {
			layer.Type = "xyz"
			layer.TileURL = extractXYZURL(ml.DataSource)
		} else if ml.Provider == "wms" {
			layer.Type = "wms"
			layer.WMSURL, layer.WMSLayers = extractWMSInfo(ml.DataSource)
		}

		metadata.Layers = append(metadata.Layers, layer)
	}

	return metadata, nil
}

// extractXYZURL extracts the tile URL from a QGIS XYZ layer source string
func extractXYZURL(source string) string {
	// Source format: "type=xyz&url=https://example.com/{z}/{x}/{y}.png&zmax=19&zmin=0"
	params, err := url.ParseQuery(source)
	if err != nil {
		// Try regex fallback
		re := regexp.MustCompile(`url=([^&]+)`)
		if match := re.FindStringSubmatch(source); len(match) > 1 {
			decoded, _ := url.QueryUnescape(match[1])
			return decoded
		}
		return ""
	}

	tileURL := params.Get("url")
	// Convert QGIS placeholders to standard format
	tileURL = strings.ReplaceAll(tileURL, "%7Bz%7D", "{z}")
	tileURL = strings.ReplaceAll(tileURL, "%7Bx%7D", "{x}")
	tileURL = strings.ReplaceAll(tileURL, "%7By%7D", "{y}")
	tileURL = strings.ReplaceAll(tileURL, "{z}", "{z}")
	tileURL = strings.ReplaceAll(tileURL, "{x}", "{x}")
	tileURL = strings.ReplaceAll(tileURL, "{y}", "{y}")

	return tileURL
}

// extractWMSInfo extracts WMS URL and layer names from a QGIS WMS source string
func extractWMSInfo(source string) (string, string) {
	params, err := url.ParseQuery(source)
	if err != nil {
		return "", ""
	}
	return params.Get("url"), params.Get("layers")
}

// getQGISProjectFile returns the project file content
func (s *Server) getQGISProjectFile(w http.ResponseWriter, r *http.Request, projectID string) {
	for _, project := range s.config.QGISProjects {
		if project.ID == projectID {
			file, err := os.Open(project.Path)
			if err != nil {
				http.Error(w, "Cannot open file: "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer file.Close()

			// Set content type based on extension
			ext := strings.ToLower(filepath.Ext(project.Path))
			if ext == ".qgz" {
				w.Header().Set("Content-Type", "application/zip")
			} else {
				w.Header().Set("Content-Type", "application/xml")
			}

			w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(project.Path)+"\"")

			io.Copy(w, file)
			return
		}
	}

	http.Error(w, "Project not found", http.StatusNotFound)
}
