package models

import (
	"encoding/json"
	"time"
)

// Product represents a hosted product (GeoServer, GeoNode, PostGIS).
type Product struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	Slug                string    `json:"slug"`
	Description         string    `json:"description,omitempty"`
	ShortDescription    string    `json:"short_description,omitempty"`
	ImageURL            string    `json:"image_url,omitempty"`
	IconName            string    `json:"icon_name,omitempty"`
	DocumentationURL    string    `json:"documentation_url,omitempty"`
	IsAvailable         bool      `json:"is_available"`
	VaultCredentialPath string    `json:"vault_credential_path,omitempty"`
	SortOrder           int       `json:"sort_order"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`

	// Related data (populated by queries)
	Packages []Package `json:"packages,omitempty"`
}

// Package represents a pricing tier for a product.
type Package struct {
	ID                     string          `json:"id"`
	ProductID              string          `json:"product_id"`
	Name                   string          `json:"name"`
	Slug                   string          `json:"slug"`
	Description            string          `json:"description,omitempty"`
	PriceMonthly           float64         `json:"price_monthly"`
	PriceYearly            float64         `json:"price_yearly"`
	Features               json.RawMessage `json:"features"`
	CPULimit               string          `json:"cpu_limit,omitempty"`
	MemoryLimit            string          `json:"memory_limit,omitempty"`
	StorageLimit           string          `json:"storage_limit,omitempty"`
	ConcurrentUsers        *int            `json:"concurrent_users,omitempty"`
	IsPopular              bool            `json:"is_popular"`
	IsAvailable            bool            `json:"is_available"`
	StripePriceMonthlyID   string          `json:"stripe_price_monthly_id,omitempty"`
	StripePriceYearlyID    string          `json:"stripe_price_yearly_id,omitempty"`
	PaystackPlanMonthlyID  string          `json:"paystack_plan_monthly_id,omitempty"`
	PaystackPlanYearlyID   string          `json:"paystack_plan_yearly_id,omitempty"`
	SortOrder              int             `json:"sort_order"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`

	// Related data (populated by queries)
	Product *Product `json:"product,omitempty"`
}

// GetFeatures returns the features as a string slice.
func (p *Package) GetFeatures() []string {
	var features []string
	if err := json.Unmarshal(p.Features, &features); err != nil {
		return nil
	}
	return features
}

// SetFeatures sets the features from a string slice.
func (p *Package) SetFeatures(features []string) error {
	data, err := json.Marshal(features)
	if err != nil {
		return err
	}
	p.Features = data
	return nil
}

// YearlySavings returns the percentage saved when paying yearly.
func (p *Package) YearlySavings() float64 {
	monthlyTotal := p.PriceMonthly * 12
	if monthlyTotal == 0 {
		return 0
	}
	return ((monthlyTotal - p.PriceYearly) / monthlyTotal) * 100
}

// YearlySavingsAmount returns the absolute amount saved when paying yearly.
func (p *Package) YearlySavingsAmount() float64 {
	return (p.PriceMonthly * 12) - p.PriceYearly
}

// CreateProductRequest represents a request to create a product.
type CreateProductRequest struct {
	Name             string `json:"name"`
	Slug             string `json:"slug"`
	Description      string `json:"description,omitempty"`
	ShortDescription string `json:"short_description,omitempty"`
	ImageURL         string `json:"image_url,omitempty"`
	IconName         string `json:"icon_name,omitempty"`
	IsAvailable      bool   `json:"is_available"`
}

// CreatePackageRequest represents a request to create a package.
type CreatePackageRequest struct {
	ProductID              string   `json:"product_id"`
	Name                   string   `json:"name"`
	Slug                   string   `json:"slug"`
	Description            string   `json:"description,omitempty"`
	PriceMonthly           float64  `json:"price_monthly"`
	PriceYearly            float64  `json:"price_yearly"`
	Features               []string `json:"features"`
	CPULimit               string   `json:"cpu_limit,omitempty"`
	MemoryLimit            string   `json:"memory_limit,omitempty"`
	StorageLimit           string   `json:"storage_limit,omitempty"`
	ConcurrentUsers        *int     `json:"concurrent_users,omitempty"`
	IsPopular              bool     `json:"is_popular"`
	StripePriceMonthlyID   string   `json:"stripe_price_monthly_id,omitempty"`
	StripePriceYearlyID    string   `json:"stripe_price_yearly_id,omitempty"`
	PaystackPlanMonthlyID  string   `json:"paystack_plan_monthly_id,omitempty"`
	PaystackPlanYearlyID   string   `json:"paystack_plan_yearly_id,omitempty"`
}

// Cluster represents a deployment region/cluster.
type Cluster struct {
	ID              string    `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Region          string    `json:"region,omitempty"`
	Country         string    `json:"country,omitempty"`
	Domain          string    `json:"domain"`
	VaultURL        string    `json:"vault_url,omitempty"`
	VaultTokenPath  string    `json:"vault_token_path,omitempty"`
	JenkinsURL      string    `json:"jenkins_url,omitempty"`
	JenkinsJobName  string    `json:"jenkins_job_name,omitempty"`
	ArgoCDURL       string    `json:"argocd_url,omitempty"`
	IsActive        bool      `json:"is_active"`
	CapacityUsed    int       `json:"capacity_used"`
	CapacityTotal   int       `json:"capacity_total"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CapacityPercent returns the percentage of capacity used.
func (c *Cluster) CapacityPercent() float64 {
	if c.CapacityTotal == 0 {
		return 0
	}
	return float64(c.CapacityUsed) / float64(c.CapacityTotal) * 100
}

// HasCapacity returns true if the cluster has available capacity.
func (c *Cluster) HasCapacity() bool {
	return c.CapacityUsed < c.CapacityTotal
}
