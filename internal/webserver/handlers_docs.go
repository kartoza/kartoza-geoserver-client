package webserver

import (
	_ "embed"
	"encoding/json"
	"net/http"
)

//go:embed docs/SPECIFICATION.md
var specificationMD string

// DocumentationResponse is the API response for documentation
type DocumentationResponse struct {
	Content string `json:"content"`
	Title   string `json:"title"`
}

// handleDocumentation serves the embedded SPECIFICATION.md file
func (s *Server) handleDocumentation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := DocumentationResponse{
		Content: specificationMD,
		Title:   "Kartoza CloudBench Documentation",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
