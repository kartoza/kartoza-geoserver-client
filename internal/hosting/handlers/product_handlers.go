package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/kartoza/kartoza-cloudbench/internal/hosting/auth"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/models"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/repository"
)

// ProductHandler handles product and package-related HTTP requests.
type ProductHandler struct {
	productRepo *repository.ProductRepository
}

// NewProductHandler creates a new product handler.
func NewProductHandler(productRepo *repository.ProductRepository) *ProductHandler {
	return &ProductHandler{productRepo: productRepo}
}

// HandleProducts handles GET /api/v1/products
func (h *ProductHandler) HandleProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if admin wants to see unavailable products
	includeUnavailable := false
	if auth.IsAdmin(r.Context()) && r.URL.Query().Get("include_unavailable") == "true" {
		includeUnavailable = true
	}

	products, err := h.productRepo.ListProducts(r.Context(), includeUnavailable)
	if err != nil {
		jsonError(w, "Failed to list products", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]interface{}{
		"products": products,
	}, http.StatusOK)
}

// HandleProductBySlug handles GET /api/v1/products/{slug}
func (h *ProductHandler) HandleProductBySlug(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract slug from path
	slug := strings.TrimPrefix(r.URL.Path, "/api/v1/products/")
	slug = strings.Split(slug, "/")[0]
	if slug == "" {
		jsonError(w, "Product slug required", http.StatusBadRequest)
		return
	}

	// Check if requesting packages
	if strings.Contains(r.URL.Path, "/packages") {
		h.handleProductPackages(w, r, slug)
		return
	}

	product, err := h.productRepo.GetProductWithPackages(r.Context(), slug)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			jsonError(w, "Product not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Failed to get product", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, product, http.StatusOK)
}

// handleProductPackages handles GET /api/v1/products/{slug}/packages
func (h *ProductHandler) handleProductPackages(w http.ResponseWriter, r *http.Request, slug string) {
	product, err := h.productRepo.GetProductBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			jsonError(w, "Product not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Failed to get product", http.StatusInternalServerError)
		return
	}

	includeUnavailable := false
	if auth.IsAdmin(r.Context()) && r.URL.Query().Get("include_unavailable") == "true" {
		includeUnavailable = true
	}

	packages, err := h.productRepo.ListPackages(r.Context(), product.ID, includeUnavailable)
	if err != nil {
		jsonError(w, "Failed to list packages", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]interface{}{
		"product":  product,
		"packages": packages,
	}, http.StatusOK)
}

// HandlePackageByID handles GET /api/v1/packages/{id}
func (h *ProductHandler) HandlePackageByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/packages/")
	id = strings.TrimSuffix(id, "/")
	if id == "" {
		jsonError(w, "Package ID required", http.StatusBadRequest)
		return
	}

	pkg, err := h.productRepo.GetPackageByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			jsonError(w, "Package not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Failed to get package", http.StatusInternalServerError)
		return
	}

	// Get the product for context
	product, _ := h.productRepo.GetProductByID(r.Context(), pkg.ProductID)
	pkg.Product = product

	jsonResponse(w, pkg, http.StatusOK)
}

// HandleClusters handles GET /api/v1/clusters
func (h *ProductHandler) HandleClusters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	activeOnly := true
	if auth.IsAdmin(r.Context()) && r.URL.Query().Get("include_inactive") == "true" {
		activeOnly = false
	}

	clusters, err := h.productRepo.ListClusters(r.Context(), activeOnly)
	if err != nil {
		jsonError(w, "Failed to list clusters", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]interface{}{
		"clusters": clusters,
	}, http.StatusOK)
}

// Admin handlers

// HandleAdminCreateProduct handles POST /api/v1/admin/products
func (h *ProductHandler) HandleAdminCreateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Slug == "" {
		jsonError(w, "Name and slug are required", http.StatusBadRequest)
		return
	}

	product := &models.Product{
		Name:             req.Name,
		Slug:             req.Slug,
		Description:      req.Description,
		ShortDescription: req.ShortDescription,
		ImageURL:         req.ImageURL,
		IconName:         req.IconName,
		IsAvailable:      req.IsAvailable,
	}

	if err := h.productRepo.CreateProduct(r.Context(), product); err != nil {
		jsonError(w, "Failed to create product", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, product, http.StatusCreated)
}

// HandleAdminUpdateProduct handles PUT /api/v1/admin/products/{id}
func (h *ProductHandler) HandleAdminUpdateProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/products/")
	id = strings.TrimSuffix(id, "/")
	if id == "" {
		jsonError(w, "Product ID required", http.StatusBadRequest)
		return
	}

	// Get existing product
	product, err := h.productRepo.GetProductByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			jsonError(w, "Product not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Failed to get product", http.StatusInternalServerError)
		return
	}

	// Decode update request
	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields
	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Slug != "" {
		product.Slug = req.Slug
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.ShortDescription != "" {
		product.ShortDescription = req.ShortDescription
	}
	if req.ImageURL != "" {
		product.ImageURL = req.ImageURL
	}
	if req.IconName != "" {
		product.IconName = req.IconName
	}
	product.IsAvailable = req.IsAvailable

	if err := h.productRepo.UpdateProduct(r.Context(), product); err != nil {
		jsonError(w, "Failed to update product", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, product, http.StatusOK)
}

// HandleAdminDeleteProduct handles DELETE /api/v1/admin/products/{id}
func (h *ProductHandler) HandleAdminDeleteProduct(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/products/")
	id = strings.TrimSuffix(id, "/")
	if id == "" {
		jsonError(w, "Product ID required", http.StatusBadRequest)
		return
	}

	if err := h.productRepo.DeleteProduct(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			jsonError(w, "Product not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Failed to delete product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleAdminCreatePackage handles POST /api/v1/admin/packages
func (h *ProductHandler) HandleAdminCreatePackage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreatePackageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ProductID == "" || req.Name == "" || req.Slug == "" {
		jsonError(w, "Product ID, name, and slug are required", http.StatusBadRequest)
		return
	}

	// Verify product exists
	_, err := h.productRepo.GetProductByID(r.Context(), req.ProductID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			jsonError(w, "Product not found", http.StatusBadRequest)
			return
		}
		jsonError(w, "Failed to verify product", http.StatusInternalServerError)
		return
	}

	pkg := &models.Package{
		ProductID:             req.ProductID,
		Name:                  req.Name,
		Slug:                  req.Slug,
		Description:           req.Description,
		PriceMonthly:          req.PriceMonthly,
		PriceYearly:           req.PriceYearly,
		CPULimit:              req.CPULimit,
		MemoryLimit:           req.MemoryLimit,
		StorageLimit:          req.StorageLimit,
		ConcurrentUsers:       req.ConcurrentUsers,
		IsPopular:             req.IsPopular,
		IsAvailable:           true,
		StripePriceMonthlyID:  req.StripePriceMonthlyID,
		StripePriceYearlyID:   req.StripePriceYearlyID,
		PaystackPlanMonthlyID: req.PaystackPlanMonthlyID,
		PaystackPlanYearlyID:  req.PaystackPlanYearlyID,
	}

	// Set features
	if len(req.Features) > 0 {
		if err := pkg.SetFeatures(req.Features); err != nil {
			jsonError(w, "Invalid features format", http.StatusBadRequest)
			return
		}
	}

	if err := h.productRepo.CreatePackage(r.Context(), pkg); err != nil {
		jsonError(w, "Failed to create package", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, pkg, http.StatusCreated)
}

// HandleAdminUpdatePackage handles PUT /api/v1/admin/packages/{id}
func (h *ProductHandler) HandleAdminUpdatePackage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/packages/")
	id = strings.TrimSuffix(id, "/")
	if id == "" {
		jsonError(w, "Package ID required", http.StatusBadRequest)
		return
	}

	// Get existing package
	pkg, err := h.productRepo.GetPackageByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			jsonError(w, "Package not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Failed to get package", http.StatusInternalServerError)
		return
	}

	// Decode update request
	var req models.CreatePackageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields
	if req.Name != "" {
		pkg.Name = req.Name
	}
	if req.Slug != "" {
		pkg.Slug = req.Slug
	}
	if req.Description != "" {
		pkg.Description = req.Description
	}
	if req.PriceMonthly > 0 {
		pkg.PriceMonthly = req.PriceMonthly
	}
	if req.PriceYearly > 0 {
		pkg.PriceYearly = req.PriceYearly
	}
	if len(req.Features) > 0 {
		if err := pkg.SetFeatures(req.Features); err != nil {
			jsonError(w, "Invalid features format", http.StatusBadRequest)
			return
		}
	}
	if req.CPULimit != "" {
		pkg.CPULimit = req.CPULimit
	}
	if req.MemoryLimit != "" {
		pkg.MemoryLimit = req.MemoryLimit
	}
	if req.StorageLimit != "" {
		pkg.StorageLimit = req.StorageLimit
	}
	pkg.ConcurrentUsers = req.ConcurrentUsers
	pkg.IsPopular = req.IsPopular

	if req.StripePriceMonthlyID != "" {
		pkg.StripePriceMonthlyID = req.StripePriceMonthlyID
	}
	if req.StripePriceYearlyID != "" {
		pkg.StripePriceYearlyID = req.StripePriceYearlyID
	}
	if req.PaystackPlanMonthlyID != "" {
		pkg.PaystackPlanMonthlyID = req.PaystackPlanMonthlyID
	}
	if req.PaystackPlanYearlyID != "" {
		pkg.PaystackPlanYearlyID = req.PaystackPlanYearlyID
	}

	if err := h.productRepo.UpdatePackage(r.Context(), pkg); err != nil {
		jsonError(w, "Failed to update package", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, pkg, http.StatusOK)
}

// HandleAdminDeletePackage handles DELETE /api/v1/admin/packages/{id}
func (h *ProductHandler) HandleAdminDeletePackage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/packages/")
	id = strings.TrimSuffix(id, "/")
	if id == "" {
		jsonError(w, "Package ID required", http.StatusBadRequest)
		return
	}

	if err := h.productRepo.DeletePackage(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			jsonError(w, "Package not found", http.StatusNotFound)
			return
		}
		jsonError(w, "Failed to delete package", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
