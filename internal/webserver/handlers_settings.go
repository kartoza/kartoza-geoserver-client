package webserver

import (
	"encoding/json"
	"net/http"

	"github.com/kartoza/kartoza-geoserver-client/internal/api"
	"github.com/kartoza/kartoza-geoserver-client/internal/models"
)

// ContactResponse represents the contact information in API responses
type ContactResponse struct {
	ContactPerson       string `json:"contactPerson,omitempty"`
	ContactPosition     string `json:"contactPosition,omitempty"`
	ContactOrganization string `json:"contactOrganization,omitempty"`
	AddressType         string `json:"addressType,omitempty"`
	Address             string `json:"address,omitempty"`
	AddressCity         string `json:"addressCity,omitempty"`
	AddressState        string `json:"addressState,omitempty"`
	AddressPostCode     string `json:"addressPostalCode,omitempty"`
	AddressCountry      string `json:"addressCountry,omitempty"`
	ContactVoice        string `json:"contactVoice,omitempty"`
	ContactFax          string `json:"contactFacsimile,omitempty"`
	ContactEmail        string `json:"contactEmail,omitempty"`
	OnlineResource      string `json:"onlineResource,omitempty"`
	Welcome             string `json:"welcome,omitempty"`
}

// ContactUpdateRequest represents a contact update request
type ContactUpdateRequest struct {
	ContactPerson       string `json:"contactPerson,omitempty"`
	ContactPosition     string `json:"contactPosition,omitempty"`
	ContactOrganization string `json:"contactOrganization,omitempty"`
	AddressType         string `json:"addressType,omitempty"`
	Address             string `json:"address,omitempty"`
	AddressCity         string `json:"addressCity,omitempty"`
	AddressState        string `json:"addressState,omitempty"`
	AddressPostCode     string `json:"addressPostalCode,omitempty"`
	AddressCountry      string `json:"addressCountry,omitempty"`
	ContactVoice        string `json:"contactVoice,omitempty"`
	ContactFax          string `json:"contactFacsimile,omitempty"`
	ContactEmail        string `json:"contactEmail,omitempty"`
	OnlineResource      string `json:"onlineResource,omitempty"`
	Welcome             string `json:"welcome,omitempty"`
}

// handleSettings handles requests to /api/settings/{connId}
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	// Parse connection ID from path
	path := r.URL.Path
	connID := extractConnID(path, "/api/settings/")

	if connID == "" {
		s.jsonError(w, "Connection ID is required", http.StatusBadRequest)
		return
	}

	client := s.getClient(connID)
	if client == nil {
		s.jsonError(w, "Connection not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getContact(w, r, client)
	case http.MethodPut:
		s.updateContact(w, r, client)
	case http.MethodOptions:
		s.handleCORS(w)
	default:
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getContact fetches the GeoServer contact information
func (s *Server) getContact(w http.ResponseWriter, r *http.Request, client *api.Client) {
	contact, err := client.GetContact()
	if err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := ContactResponse{
		ContactPerson:       contact.ContactPerson,
		ContactPosition:     contact.ContactPosition,
		ContactOrganization: contact.ContactOrganization,
		AddressType:         contact.AddressType,
		Address:             contact.Address,
		AddressCity:         contact.AddressCity,
		AddressState:        contact.AddressState,
		AddressPostCode:     contact.AddressPostCode,
		AddressCountry:      contact.AddressCountry,
		ContactVoice:        contact.ContactVoice,
		ContactFax:          contact.ContactFax,
		ContactEmail:        contact.ContactEmail,
		OnlineResource:      contact.OnlineResource,
		Welcome:             contact.Welcome,
	}

	s.jsonResponse(w, response)
}

// updateContact updates the GeoServer contact information
func (s *Server) updateContact(w http.ResponseWriter, r *http.Request, client *api.Client) {
	var req ContactUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	contact := &models.GeoServerContact{
		ContactPerson:       req.ContactPerson,
		ContactPosition:     req.ContactPosition,
		ContactOrganization: req.ContactOrganization,
		AddressType:         req.AddressType,
		Address:             req.Address,
		AddressCity:         req.AddressCity,
		AddressState:        req.AddressState,
		AddressPostCode:     req.AddressPostCode,
		AddressCountry:      req.AddressCountry,
		ContactVoice:        req.ContactVoice,
		ContactFax:          req.ContactFax,
		ContactEmail:        req.ContactEmail,
		OnlineResource:      req.OnlineResource,
		Welcome:             req.Welcome,
	}

	if err := client.UpdateContact(contact); err != nil {
		s.jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the updated contact
	s.getContact(w, r, client)
}

// extractConnID extracts the connection ID from the path
func extractConnID(path, prefix string) string {
	if len(path) <= len(prefix) {
		return ""
	}
	return path[len(prefix):]
}
