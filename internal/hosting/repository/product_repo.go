package repository

import (
	"context"
	"database/sql"

	"github.com/kartoza/kartoza-cloudbench/internal/hosting/db"
	"github.com/kartoza/kartoza-cloudbench/internal/hosting/models"
)

// ProductRepository provides data access for products and packages.
type ProductRepository struct {
	db *db.DB
}

// NewProductRepository creates a new product repository.
func NewProductRepository(database *db.DB) *ProductRepository {
	return &ProductRepository{db: database}
}

// ListProducts retrieves all available products.
func (r *ProductRepository) ListProducts(ctx context.Context, includeUnavailable bool) ([]*models.Product, error) {
	query := `
		SELECT id, name, slug, description, short_description, image_url, icon_name,
		       documentation_url, is_available, vault_credential_path, sort_order,
		       created_at, updated_at
		FROM products
	`
	if !includeUnavailable {
		query += ` WHERE is_available = true`
	}
	query += ` ORDER BY sort_order, name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		p := &models.Product{}
		var desc, shortDesc, imageURL, iconName, docURL, vaultPath sql.NullString
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Slug, &desc, &shortDesc, &imageURL, &iconName,
			&docURL, &p.IsAvailable, &vaultPath, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		p.Description = desc.String
		p.ShortDescription = shortDesc.String
		p.ImageURL = imageURL.String
		p.IconName = iconName.String
		p.DocumentationURL = docURL.String
		p.VaultCredentialPath = vaultPath.String
		products = append(products, p)
	}
	return products, rows.Err()
}

// GetProductByID retrieves a product by ID.
func (r *ProductRepository) GetProductByID(ctx context.Context, id string) (*models.Product, error) {
	query := `
		SELECT id, name, slug, description, short_description, image_url, icon_name,
		       documentation_url, is_available, vault_credential_path, sort_order,
		       created_at, updated_at
		FROM products WHERE id = $1
	`
	p := &models.Product{}
	var desc, shortDesc, imageURL, iconName, docURL, vaultPath sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Slug, &desc, &shortDesc, &imageURL, &iconName,
		&docURL, &p.IsAvailable, &vaultPath, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.Description = desc.String
	p.ShortDescription = shortDesc.String
	p.ImageURL = imageURL.String
	p.IconName = iconName.String
	p.DocumentationURL = docURL.String
	p.VaultCredentialPath = vaultPath.String
	return p, nil
}

// GetProductBySlug retrieves a product by slug.
func (r *ProductRepository) GetProductBySlug(ctx context.Context, slug string) (*models.Product, error) {
	query := `
		SELECT id, name, slug, description, short_description, image_url, icon_name,
		       documentation_url, is_available, vault_credential_path, sort_order,
		       created_at, updated_at
		FROM products WHERE slug = $1
	`
	p := &models.Product{}
	var desc, shortDesc, imageURL, iconName, docURL, vaultPath sql.NullString
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&p.ID, &p.Name, &p.Slug, &desc, &shortDesc, &imageURL, &iconName,
		&docURL, &p.IsAvailable, &vaultPath, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.Description = desc.String
	p.ShortDescription = shortDesc.String
	p.ImageURL = imageURL.String
	p.IconName = iconName.String
	p.DocumentationURL = docURL.String
	p.VaultCredentialPath = vaultPath.String
	return p, nil
}

// CreateProduct creates a new product.
func (r *ProductRepository) CreateProduct(ctx context.Context, p *models.Product) error {
	query := `
		INSERT INTO products (name, slug, description, short_description, image_url,
		                      icon_name, documentation_url, is_available, vault_credential_path, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		p.Name, p.Slug, nullString(p.Description), nullString(p.ShortDescription),
		nullString(p.ImageURL), nullString(p.IconName), nullString(p.DocumentationURL),
		p.IsAvailable, nullString(p.VaultCredentialPath), p.SortOrder,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

// UpdateProduct updates a product.
func (r *ProductRepository) UpdateProduct(ctx context.Context, p *models.Product) error {
	query := `
		UPDATE products SET
			name = $2, slug = $3, description = $4, short_description = $5,
			image_url = $6, icon_name = $7, documentation_url = $8,
			is_available = $9, vault_credential_path = $10, sort_order = $11
		WHERE id = $1
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		p.ID, p.Name, p.Slug, nullString(p.Description), nullString(p.ShortDescription),
		nullString(p.ImageURL), nullString(p.IconName), nullString(p.DocumentationURL),
		p.IsAvailable, nullString(p.VaultCredentialPath), p.SortOrder,
	).Scan(&p.UpdatedAt)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	return err
}

// DeleteProduct deletes a product.
func (r *ProductRepository) DeleteProduct(ctx context.Context, id string) error {
	query := `DELETE FROM products WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// ListPackages retrieves packages, optionally filtered by product.
func (r *ProductRepository) ListPackages(ctx context.Context, productID string, includeUnavailable bool) ([]*models.Package, error) {
	query := `
		SELECT id, product_id, name, slug, description, price_monthly, price_yearly,
		       features, cpu_limit, memory_limit, storage_limit, concurrent_users,
		       is_popular, is_available, stripe_price_monthly_id, stripe_price_yearly_id,
		       paystack_plan_monthly_id, paystack_plan_yearly_id, sort_order,
		       created_at, updated_at
		FROM packages WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	if productID != "" {
		query += ` AND product_id = $` + string(rune('0'+argNum))
		args = append(args, productID)
		argNum++
	}
	if !includeUnavailable {
		query += ` AND is_available = true`
	}
	query += ` ORDER BY sort_order, price_monthly`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packages []*models.Package
	for rows.Next() {
		pkg := &models.Package{}
		var desc, cpuLimit, memLimit, storageLimit sql.NullString
		var concurrentUsers sql.NullInt32
		var stripeMonthly, stripeYearly, paystackMonthly, paystackYearly sql.NullString

		if err := rows.Scan(
			&pkg.ID, &pkg.ProductID, &pkg.Name, &pkg.Slug, &desc,
			&pkg.PriceMonthly, &pkg.PriceYearly, &pkg.Features,
			&cpuLimit, &memLimit, &storageLimit, &concurrentUsers,
			&pkg.IsPopular, &pkg.IsAvailable, &stripeMonthly, &stripeYearly,
			&paystackMonthly, &paystackYearly, &pkg.SortOrder,
			&pkg.CreatedAt, &pkg.UpdatedAt,
		); err != nil {
			return nil, err
		}
		pkg.Description = desc.String
		pkg.CPULimit = cpuLimit.String
		pkg.MemoryLimit = memLimit.String
		pkg.StorageLimit = storageLimit.String
		if concurrentUsers.Valid {
			v := int(concurrentUsers.Int32)
			pkg.ConcurrentUsers = &v
		}
		pkg.StripePriceMonthlyID = stripeMonthly.String
		pkg.StripePriceYearlyID = stripeYearly.String
		pkg.PaystackPlanMonthlyID = paystackMonthly.String
		pkg.PaystackPlanYearlyID = paystackYearly.String
		packages = append(packages, pkg)
	}
	return packages, rows.Err()
}

// GetPackageByID retrieves a package by ID.
func (r *ProductRepository) GetPackageByID(ctx context.Context, id string) (*models.Package, error) {
	query := `
		SELECT id, product_id, name, slug, description, price_monthly, price_yearly,
		       features, cpu_limit, memory_limit, storage_limit, concurrent_users,
		       is_popular, is_available, stripe_price_monthly_id, stripe_price_yearly_id,
		       paystack_plan_monthly_id, paystack_plan_yearly_id, sort_order,
		       created_at, updated_at
		FROM packages WHERE id = $1
	`
	pkg := &models.Package{}
	var desc, cpuLimit, memLimit, storageLimit sql.NullString
	var concurrentUsers sql.NullInt32
	var stripeMonthly, stripeYearly, paystackMonthly, paystackYearly sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&pkg.ID, &pkg.ProductID, &pkg.Name, &pkg.Slug, &desc,
		&pkg.PriceMonthly, &pkg.PriceYearly, &pkg.Features,
		&cpuLimit, &memLimit, &storageLimit, &concurrentUsers,
		&pkg.IsPopular, &pkg.IsAvailable, &stripeMonthly, &stripeYearly,
		&paystackMonthly, &paystackYearly, &pkg.SortOrder,
		&pkg.CreatedAt, &pkg.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	pkg.Description = desc.String
	pkg.CPULimit = cpuLimit.String
	pkg.MemoryLimit = memLimit.String
	pkg.StorageLimit = storageLimit.String
	if concurrentUsers.Valid {
		v := int(concurrentUsers.Int32)
		pkg.ConcurrentUsers = &v
	}
	pkg.StripePriceMonthlyID = stripeMonthly.String
	pkg.StripePriceYearlyID = stripeYearly.String
	pkg.PaystackPlanMonthlyID = paystackMonthly.String
	pkg.PaystackPlanYearlyID = paystackYearly.String
	return pkg, nil
}

// CreatePackage creates a new package.
func (r *ProductRepository) CreatePackage(ctx context.Context, pkg *models.Package) error {
	query := `
		INSERT INTO packages (product_id, name, slug, description, price_monthly, price_yearly,
		                      features, cpu_limit, memory_limit, storage_limit, concurrent_users,
		                      is_popular, is_available, stripe_price_monthly_id, stripe_price_yearly_id,
		                      paystack_plan_monthly_id, paystack_plan_yearly_id, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		pkg.ProductID, pkg.Name, pkg.Slug, nullString(pkg.Description),
		pkg.PriceMonthly, pkg.PriceYearly, pkg.Features,
		nullString(pkg.CPULimit), nullString(pkg.MemoryLimit), nullString(pkg.StorageLimit),
		pkg.ConcurrentUsers, pkg.IsPopular, pkg.IsAvailable,
		nullString(pkg.StripePriceMonthlyID), nullString(pkg.StripePriceYearlyID),
		nullString(pkg.PaystackPlanMonthlyID), nullString(pkg.PaystackPlanYearlyID),
		pkg.SortOrder,
	).Scan(&pkg.ID, &pkg.CreatedAt, &pkg.UpdatedAt)
}

// UpdatePackage updates a package.
func (r *ProductRepository) UpdatePackage(ctx context.Context, pkg *models.Package) error {
	query := `
		UPDATE packages SET
			name = $2, slug = $3, description = $4, price_monthly = $5, price_yearly = $6,
			features = $7, cpu_limit = $8, memory_limit = $9, storage_limit = $10,
			concurrent_users = $11, is_popular = $12, is_available = $13,
			stripe_price_monthly_id = $14, stripe_price_yearly_id = $15,
			paystack_plan_monthly_id = $16, paystack_plan_yearly_id = $17, sort_order = $18
		WHERE id = $1
		RETURNING updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		pkg.ID, pkg.Name, pkg.Slug, nullString(pkg.Description),
		pkg.PriceMonthly, pkg.PriceYearly, pkg.Features,
		nullString(pkg.CPULimit), nullString(pkg.MemoryLimit), nullString(pkg.StorageLimit),
		pkg.ConcurrentUsers, pkg.IsPopular, pkg.IsAvailable,
		nullString(pkg.StripePriceMonthlyID), nullString(pkg.StripePriceYearlyID),
		nullString(pkg.PaystackPlanMonthlyID), nullString(pkg.PaystackPlanYearlyID),
		pkg.SortOrder,
	).Scan(&pkg.UpdatedAt)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	return err
}

// DeletePackage deletes a package.
func (r *ProductRepository) DeletePackage(ctx context.Context, id string) error {
	query := `DELETE FROM packages WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// ListClusters retrieves all clusters.
func (r *ProductRepository) ListClusters(ctx context.Context, activeOnly bool) ([]*models.Cluster, error) {
	query := `
		SELECT id, code, name, region, country, domain, vault_url, vault_token_path,
		       jenkins_url, jenkins_job_name, argocd_url, is_active,
		       capacity_used, capacity_total, created_at, updated_at
		FROM clusters
	`
	if activeOnly {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clusters []*models.Cluster
	for rows.Next() {
		c := &models.Cluster{}
		var region, country, vaultURL, vaultTokenPath, jenkinsURL, jenkinsJobName, argocdURL sql.NullString
		if err := rows.Scan(
			&c.ID, &c.Code, &c.Name, &region, &country, &c.Domain,
			&vaultURL, &vaultTokenPath, &jenkinsURL, &jenkinsJobName, &argocdURL,
			&c.IsActive, &c.CapacityUsed, &c.CapacityTotal, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		c.Region = region.String
		c.Country = country.String
		c.VaultURL = vaultURL.String
		c.VaultTokenPath = vaultTokenPath.String
		c.JenkinsURL = jenkinsURL.String
		c.JenkinsJobName = jenkinsJobName.String
		c.ArgoCDURL = argocdURL.String
		clusters = append(clusters, c)
	}
	return clusters, rows.Err()
}

// GetClusterByID retrieves a cluster by ID.
func (r *ProductRepository) GetClusterByID(ctx context.Context, id string) (*models.Cluster, error) {
	query := `
		SELECT id, code, name, region, country, domain, vault_url, vault_token_path,
		       jenkins_url, jenkins_job_name, argocd_url, is_active,
		       capacity_used, capacity_total, created_at, updated_at
		FROM clusters WHERE id = $1
	`
	c := &models.Cluster{}
	var region, country, vaultURL, vaultTokenPath, jenkinsURL, jenkinsJobName, argocdURL sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.Code, &c.Name, &region, &country, &c.Domain,
		&vaultURL, &vaultTokenPath, &jenkinsURL, &jenkinsJobName, &argocdURL,
		&c.IsActive, &c.CapacityUsed, &c.CapacityTotal, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	c.Region = region.String
	c.Country = country.String
	c.VaultURL = vaultURL.String
	c.VaultTokenPath = vaultTokenPath.String
	c.JenkinsURL = jenkinsURL.String
	c.JenkinsJobName = jenkinsJobName.String
	c.ArgoCDURL = argocdURL.String
	return c, nil
}

// GetClusterByCode retrieves a cluster by code.
func (r *ProductRepository) GetClusterByCode(ctx context.Context, code string) (*models.Cluster, error) {
	query := `
		SELECT id, code, name, region, country, domain, vault_url, vault_token_path,
		       jenkins_url, jenkins_job_name, argocd_url, is_active,
		       capacity_used, capacity_total, created_at, updated_at
		FROM clusters WHERE code = $1
	`
	c := &models.Cluster{}
	var region, country, vaultURL, vaultTokenPath, jenkinsURL, jenkinsJobName, argocdURL sql.NullString
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&c.ID, &c.Code, &c.Name, &region, &country, &c.Domain,
		&vaultURL, &vaultTokenPath, &jenkinsURL, &jenkinsJobName, &argocdURL,
		&c.IsActive, &c.CapacityUsed, &c.CapacityTotal, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	c.Region = region.String
	c.Country = country.String
	c.VaultURL = vaultURL.String
	c.VaultTokenPath = vaultTokenPath.String
	c.JenkinsURL = jenkinsURL.String
	c.JenkinsJobName = jenkinsJobName.String
	c.ArgoCDURL = argocdURL.String
	return c, nil
}

// IncrementClusterCapacity increments the capacity used for a cluster.
func (r *ProductRepository) IncrementClusterCapacity(ctx context.Context, id string) error {
	query := `UPDATE clusters SET capacity_used = capacity_used + 1 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// DecrementClusterCapacity decrements the capacity used for a cluster.
func (r *ProductRepository) DecrementClusterCapacity(ctx context.Context, id string) error {
	query := `UPDATE clusters SET capacity_used = GREATEST(0, capacity_used - 1) WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetProductWithPackages retrieves a product with its packages.
func (r *ProductRepository) GetProductWithPackages(ctx context.Context, slug string) (*models.Product, error) {
	product, err := r.GetProductBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	packages, err := r.ListPackages(ctx, product.ID, false)
	if err != nil {
		return nil, err
	}

	product.Packages = make([]models.Package, len(packages))
	for i, pkg := range packages {
		product.Packages[i] = *pkg
	}

	return product, nil
}
