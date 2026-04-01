package webserver

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Mock hosting data for demo purposes
// In production, this would connect to the hosting database

type hostingProduct struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Description string   `json:"description"`
	ImageURL    string   `json:"image_url,omitempty"`
	Icon        string   `json:"icon,omitempty"`
	IsAvailable bool     `json:"is_available"`
	Features    []string `json:"features,omitempty"`
	CreatedAt   string   `json:"created_at"`
}

type hostingPackage struct {
	ID           string   `json:"id"`
	ProductID    string   `json:"product_id"`
	Name         string   `json:"name"`
	Slug         string   `json:"slug"`
	PriceMonthly float64  `json:"price_monthly"`
	PriceYearly  float64  `json:"price_yearly"`
	Features     []string `json:"features"`
	CPULimit     string   `json:"cpu_limit,omitempty"`
	MemoryLimit  string   `json:"memory_limit,omitempty"`
	StorageLimit string   `json:"storage_limit,omitempty"`
	IsPopular    bool     `json:"is_popular"`
	SortOrder    int      `json:"sort_order"`
}

type hostingCluster struct {
	ID            string `json:"id"`
	Code          string `json:"code"`
	Name          string `json:"name"`
	Region        string `json:"region"`
	Domain        string `json:"domain"`
	IsActive      bool   `json:"is_active"`
	InstanceCount int    `json:"instance_count,omitempty"`
	Capacity      int    `json:"capacity,omitempty"`
}

var mockProducts = []hostingProduct{
	{
		ID:          "prod-geoserver",
		Name:        "GeoServer",
		Slug:        "geoserver",
		Description: "Enterprise-grade web mapping server for publishing and sharing geospatial data. Supports WMS, WFS, WCS, and more.",
		Icon:        "server",
		IsAvailable: true,
		Features:    []string{"OGC Compliant", "REST API", "SLD Styling", "Tile Caching"},
		CreatedAt:   "2024-01-01T00:00:00Z",
	},
	{
		ID:          "prod-geonode",
		Name:        "GeoNode",
		Slug:        "geonode",
		Description: "Open source geospatial content management system. Create, share, and manage geospatial data and maps collaboratively.",
		Icon:        "globe",
		IsAvailable: true,
		Features:    []string{"Data Management", "Map Composer", "User Management", "Metadata Catalog"},
		CreatedAt:   "2024-01-01T00:00:00Z",
	},
	{
		ID:          "prod-postgis",
		Name:        "PostGIS",
		Slug:        "postgis",
		Description: "Powerful spatial database extension for PostgreSQL. Store, query, and analyze your geospatial data.",
		Icon:        "database",
		IsAvailable: true,
		Features:    []string{"Spatial Indexing", "Vector/Raster", "Topology", "3D Support"},
		CreatedAt:   "2024-01-01T00:00:00Z",
	},
}

var mockPackages = map[string][]hostingPackage{
	"geoserver": {
		{
			ID:           "pkg-gs-starter",
			ProductID:    "prod-geoserver",
			Name:         "Starter",
			Slug:         "starter",
			PriceMonthly: 29,
			PriceYearly:  290,
			Features:     []string{"5 Workspaces", "10GB Storage", "100k Requests/month", "Email Support"},
			CPULimit:     "1 vCPU",
			MemoryLimit:  "2GB RAM",
			StorageLimit: "10GB",
			IsPopular:    false,
			SortOrder:    1,
		},
		{
			ID:           "pkg-gs-professional",
			ProductID:    "prod-geoserver",
			Name:         "Professional",
			Slug:         "professional",
			PriceMonthly: 79,
			PriceYearly:  790,
			Features:     []string{"Unlimited Workspaces", "50GB Storage", "500k Requests/month", "Priority Support", "Custom Domain"},
			CPULimit:     "2 vCPU",
			MemoryLimit:  "4GB RAM",
			StorageLimit: "50GB",
			IsPopular:    true,
			SortOrder:    2,
		},
		{
			ID:           "pkg-gs-enterprise",
			ProductID:    "prod-geoserver",
			Name:         "Enterprise",
			Slug:         "enterprise",
			PriceMonthly: 199,
			PriceYearly:  1990,
			Features:     []string{"Unlimited Everything", "200GB Storage", "Unlimited Requests", "24/7 Support", "SLA Guarantee", "Dedicated Resources"},
			CPULimit:     "4 vCPU",
			MemoryLimit:  "8GB RAM",
			StorageLimit: "200GB",
			IsPopular:    false,
			SortOrder:    3,
		},
	},
	"geonode": {
		{
			ID:           "pkg-gn-starter",
			ProductID:    "prod-geonode",
			Name:         "Starter",
			Slug:         "starter",
			PriceMonthly: 49,
			PriceYearly:  490,
			Features:     []string{"5 Users", "20GB Storage", "100 Layers", "Email Support"},
			CPULimit:     "2 vCPU",
			MemoryLimit:  "4GB RAM",
			StorageLimit: "20GB",
			IsPopular:    false,
			SortOrder:    1,
		},
		{
			ID:           "pkg-gn-professional",
			ProductID:    "prod-geonode",
			Name:         "Professional",
			Slug:         "professional",
			PriceMonthly: 129,
			PriceYearly:  1290,
			Features:     []string{"25 Users", "100GB Storage", "Unlimited Layers", "Priority Support", "Custom Branding"},
			CPULimit:     "4 vCPU",
			MemoryLimit:  "8GB RAM",
			StorageLimit: "100GB",
			IsPopular:    true,
			SortOrder:    2,
		},
		{
			ID:           "pkg-gn-enterprise",
			ProductID:    "prod-geonode",
			Name:         "Enterprise",
			Slug:         "enterprise",
			PriceMonthly: 299,
			PriceYearly:  2990,
			Features:     []string{"Unlimited Users", "500GB Storage", "Unlimited Layers", "24/7 Support", "SLA Guarantee", "SSO Integration"},
			CPULimit:     "8 vCPU",
			MemoryLimit:  "16GB RAM",
			StorageLimit: "500GB",
			IsPopular:    false,
			SortOrder:    3,
		},
	},
	"postgis": {
		{
			ID:           "pkg-pg-starter",
			ProductID:    "prod-postgis",
			Name:         "Starter",
			Slug:         "starter",
			PriceMonthly: 19,
			PriceYearly:  190,
			Features:     []string{"5 Databases", "10GB Storage", "10 Connections", "Daily Backups"},
			CPULimit:     "1 vCPU",
			MemoryLimit:  "2GB RAM",
			StorageLimit: "10GB",
			IsPopular:    false,
			SortOrder:    1,
		},
		{
			ID:           "pkg-pg-professional",
			ProductID:    "prod-postgis",
			Name:         "Professional",
			Slug:         "professional",
			PriceMonthly: 59,
			PriceYearly:  590,
			Features:     []string{"20 Databases", "100GB Storage", "50 Connections", "Hourly Backups", "Read Replicas"},
			CPULimit:     "2 vCPU",
			MemoryLimit:  "4GB RAM",
			StorageLimit: "100GB",
			IsPopular:    true,
			SortOrder:    2,
		},
		{
			ID:           "pkg-pg-enterprise",
			ProductID:    "prod-postgis",
			Name:         "Enterprise",
			Slug:         "enterprise",
			PriceMonthly: 149,
			PriceYearly:  1490,
			Features:     []string{"Unlimited Databases", "500GB Storage", "Unlimited Connections", "Point-in-time Recovery", "High Availability"},
			CPULimit:     "4 vCPU",
			MemoryLimit:  "8GB RAM",
			StorageLimit: "500GB",
			IsPopular:    false,
			SortOrder:    3,
		},
	},
}

var mockClusters = []hostingCluster{
	{
		ID:       "cluster-eu-west",
		Code:     "eu-west",
		Name:     "Europe (Frankfurt)",
		Region:   "eu-west-1",
		Domain:   "eu.kartoza.cloud",
		IsActive: true,
		Capacity: 100,
	},
	{
		ID:       "cluster-us-east",
		Code:     "us-east",
		Name:     "US East (Virginia)",
		Region:   "us-east-1",
		Domain:   "us.kartoza.cloud",
		IsActive: true,
		Capacity: 100,
	},
	{
		ID:       "cluster-af-south",
		Code:     "af-south",
		Name:     "Africa (Cape Town)",
		Region:   "af-south-1",
		Domain:   "za.kartoza.cloud",
		IsActive: true,
		Capacity: 50,
	},
}

// handleHostingProducts returns the list of available products
func (s *Server) handleHostingProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.jsonResponse(w, map[string]interface{}{
		"products": mockProducts,
	})
}

// handleHostingProductBySlug returns a product by slug, or packages for a product
func (s *Server) handleHostingProductBySlug(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/products/")
	parts := strings.Split(path, "/")
	slug := parts[0]

	// Check if requesting packages
	if len(parts) > 1 && parts[1] == "packages" {
		packages, ok := mockPackages[slug]
		if !ok {
			s.jsonError(w, "Product not found", http.StatusNotFound)
			return
		}
		s.jsonResponse(w, map[string]interface{}{
			"packages": packages,
		})
		return
	}

	// Find product by slug
	for _, p := range mockProducts {
		if p.Slug == slug {
			s.jsonResponse(w, p)
			return
		}
	}

	s.jsonError(w, "Product not found", http.StatusNotFound)
}

// handleHostingClusters returns the list of available clusters
func (s *Server) handleHostingClusters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Filter to only active clusters
	activeClusters := []hostingCluster{}
	for _, c := range mockClusters {
		if c.IsActive {
			activeClusters = append(activeClusters, c)
		}
	}

	s.jsonResponse(w, map[string]interface{}{
		"clusters": activeClusters,
	})
}

// hostingJSONResponse sends a JSON response (local helper to avoid import cycles)
func hostingJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
