package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kartoza/kartoza-cloudbench/internal/models"
)

func (c *Client) GetGlobalSettings() (*models.GeoServerGlobalSettings, error) {
	resp, err := c.doRequest("GET", "/settings", nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get global settings: %s", string(body))
	}

	var result struct {
		Global struct {
			Settings struct {
				Charset           string `json:"charset"`
				NumDecimals       int    `json:"numDecimals"`
				OnlineResource    string `json:"onlineResource"`
				Verbose           bool   `json:"verbose"`
				VerboseExceptions bool   `json:"verboseExceptions"`
				ProxyBaseURL      string `json:"proxyBaseUrl"`
				Contact           struct {
					ContactPerson       string `json:"contactPerson"`
					ContactOrganization string `json:"contactOrganization"`
					ContactPosition     string `json:"contactPosition"`
					AddressType         string `json:"addressType"`
					Address             string `json:"address"`
					AddressCity         string `json:"addressCity"`
					AddressState        string `json:"addressState"`
					AddressPostCode     string `json:"addressPostalCode"`
					AddressCountry      string `json:"addressCountry"`
					ContactVoice        string `json:"contactVoice"`
					ContactFax          string `json:"contactFacsimile"`
					ContactEmail        string `json:"contactEmail"`
				} `json:"contact"`
			} `json:"settings"`
		} `json:"global"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode global settings: %w", err)
	}

	settings := &models.GeoServerGlobalSettings{
		Charset:           result.Global.Settings.Charset,
		NumDecimals:       result.Global.Settings.NumDecimals,
		OnlineResource:    result.Global.Settings.OnlineResource,
		Verbose:           result.Global.Settings.Verbose,
		VerboseExceptions: result.Global.Settings.VerboseExceptions,
		ProxyBaseURL:      result.Global.Settings.ProxyBaseURL,
		Contact: &models.GeoServerContact{
			ContactPerson:       result.Global.Settings.Contact.ContactPerson,
			ContactOrganization: result.Global.Settings.Contact.ContactOrganization,
			ContactPosition:     result.Global.Settings.Contact.ContactPosition,
			AddressType:         result.Global.Settings.Contact.AddressType,
			Address:             result.Global.Settings.Contact.Address,
			AddressCity:         result.Global.Settings.Contact.AddressCity,
			AddressState:        result.Global.Settings.Contact.AddressState,
			AddressPostCode:     result.Global.Settings.Contact.AddressPostCode,
			AddressCountry:      result.Global.Settings.Contact.AddressCountry,
			ContactVoice:        result.Global.Settings.Contact.ContactVoice,
			ContactFax:          result.Global.Settings.Contact.ContactFax,
			ContactEmail:        result.Global.Settings.Contact.ContactEmail,
		},
	}

	return settings, nil
}

func (c *Client) GetContact() (*models.GeoServerContact, error) {
	resp, err := c.doRequest("GET", "/settings/contact", nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get contact: %s", string(body))
	}

	var result struct {
		Contact struct {
			ContactPerson       string `json:"contactPerson"`
			ContactOrganization string `json:"contactOrganization"`
			ContactPosition     string `json:"contactPosition"`
			AddressType         string `json:"addressType"`
			Address             string `json:"address"`
			AddressCity         string `json:"addressCity"`
			AddressState        string `json:"addressState"`
			AddressPostCode     string `json:"addressPostalCode"`
			AddressCountry      string `json:"addressCountry"`
			ContactVoice        string `json:"contactVoice"`
			ContactFax          string `json:"contactFacsimile"`
			ContactEmail        string `json:"contactEmail"`
			OnlineResource      string `json:"onlineResource"`
			Welcome             string `json:"welcome"`
		} `json:"contact"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode contact: %w", err)
	}

	contact := &models.GeoServerContact{
		ContactPerson:       result.Contact.ContactPerson,
		ContactOrganization: result.Contact.ContactOrganization,
		ContactPosition:     result.Contact.ContactPosition,
		AddressType:         result.Contact.AddressType,
		Address:             result.Contact.Address,
		AddressCity:         result.Contact.AddressCity,
		AddressState:        result.Contact.AddressState,
		AddressPostCode:     result.Contact.AddressPostCode,
		AddressCountry:      result.Contact.AddressCountry,
		ContactVoice:        result.Contact.ContactVoice,
		ContactFax:          result.Contact.ContactFax,
		ContactEmail:        result.Contact.ContactEmail,
		OnlineResource:      result.Contact.OnlineResource,
		Welcome:             result.Contact.Welcome,
	}

	return contact, nil
}

func (c *Client) UpdateContact(contact *models.GeoServerContact) error {
	body := map[string]interface{}{
		"contact": map[string]interface{}{
			"contactPerson":       contact.ContactPerson,
			"contactOrganization": contact.ContactOrganization,
			"contactPosition":     contact.ContactPosition,
			"addressType":         contact.AddressType,
			"address":             contact.Address,
			"addressCity":         contact.AddressCity,
			"addressState":        contact.AddressState,
			"addressPostalCode":   contact.AddressPostCode,
			"addressCountry":      contact.AddressCountry,
			"contactVoice":        contact.ContactVoice,
			"contactFacsimile":    contact.ContactFax,
			"contactEmail":        contact.ContactEmail,
			"onlineResource":      contact.OnlineResource,
			"welcome":             contact.Welcome,
		},
	}

	resp, err := c.doJSONRequest("PUT", "/settings/contact", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update contact: %s", string(bodyBytes))
	}

	return nil
}

// ============================================================================
// Download Functions - Export resource configurations as JSON/SLD
// ============================================================================

