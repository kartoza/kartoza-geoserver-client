-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users and Authentication
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    avatar_url TEXT,
    is_active BOOLEAN DEFAULT true,
    is_admin BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_is_active ON users(is_active);

CREATE TABLE user_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    token_type VARCHAR(50) NOT NULL DEFAULT 'auth', -- 'auth', 'refresh', 'reset'
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_user_tokens_token ON user_tokens(token);
CREATE INDEX idx_user_tokens_user_id ON user_tokens(user_id);
CREATE INDEX idx_user_tokens_expires_at ON user_tokens(expires_at);

CREATE TABLE user_billing_info (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    company_name VARCHAR(255),
    tax_number VARCHAR(50),
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100),
    postal_code VARCHAR(20),
    phone VARCHAR(50),
    stripe_customer_id VARCHAR(255),
    paystack_customer_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_user_billing_stripe ON user_billing_info(stripe_customer_id);
CREATE INDEX idx_user_billing_paystack ON user_billing_info(paystack_customer_id);

-- Product Catalog
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    short_description VARCHAR(255),
    image_url TEXT,
    icon_name VARCHAR(50),
    documentation_url TEXT,
    is_available BOOLEAN DEFAULT true,
    vault_credential_path VARCHAR(255),
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_products_slug ON products(slug);
CREATE INDEX idx_products_is_available ON products(is_available);

CREATE TABLE packages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    price_monthly DECIMAL(10,2) NOT NULL,
    price_yearly DECIMAL(10,2) NOT NULL,
    features JSONB DEFAULT '[]'::jsonb,
    cpu_limit VARCHAR(20),
    memory_limit VARCHAR(20),
    storage_limit VARCHAR(20),
    concurrent_users INT,
    is_popular BOOLEAN DEFAULT false,
    is_available BOOLEAN DEFAULT true,
    stripe_price_monthly_id VARCHAR(255),
    stripe_price_yearly_id VARCHAR(255),
    paystack_plan_monthly_id VARCHAR(255),
    paystack_plan_yearly_id VARCHAR(255),
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(product_id, slug)
);

CREATE INDEX idx_packages_product_id ON packages(product_id);
CREATE INDEX idx_packages_is_available ON packages(is_available);

-- Clusters/Regions
CREATE TABLE clusters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    region VARCHAR(100),
    country VARCHAR(100),
    domain VARCHAR(255) NOT NULL,
    vault_url VARCHAR(255),
    vault_token_path VARCHAR(255),
    jenkins_url VARCHAR(255),
    jenkins_job_name VARCHAR(255),
    argocd_url VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    capacity_used INT DEFAULT 0,
    capacity_total INT DEFAULT 100,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_clusters_code ON clusters(code);
CREATE INDEX idx_clusters_is_active ON clusters(is_active);

-- Orders and Subscriptions
CREATE TABLE sales_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    package_id UUID REFERENCES packages(id),
    cluster_id UUID REFERENCES clusters(id),
    app_name VARCHAR(100),
    billing_cycle VARCHAR(20) NOT NULL CHECK (billing_cycle IN ('monthly', 'yearly')),
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'payment_pending', 'paid', 'deploying', 'deployed', 'cancelled', 'refunded', 'failed')),
    subtotal_amount DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    total_amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    coupon_id UUID,
    payment_method VARCHAR(50),
    payment_id VARCHAR(255),
    stripe_session_id VARCHAR(255),
    stripe_payment_intent_id VARCHAR(255),
    paystack_reference VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_sales_orders_user_id ON sales_orders(user_id);
CREATE INDEX idx_sales_orders_status ON sales_orders(status);
CREATE INDEX idx_sales_orders_created_at ON sales_orders(created_at);
CREATE INDEX idx_sales_orders_stripe_session ON sales_orders(stripe_session_id);

-- Instances
CREATE TABLE instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    sales_order_id UUID REFERENCES sales_orders(id),
    product_id UUID REFERENCES products(id),
    package_id UUID REFERENCES packages(id),
    cluster_id UUID REFERENCES clusters(id),
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'deploying', 'starting_up', 'online', 'offline', 'error', 'maintenance', 'deleting', 'deleted')),
    url VARCHAR(255),
    internal_url VARCHAR(255),
    vault_path VARCHAR(255),
    admin_username VARCHAR(100),
    health_status VARCHAR(50) DEFAULT 'unknown' CHECK (health_status IN ('unknown', 'healthy', 'degraded', 'unhealthy')),
    health_message TEXT,
    last_health_check TIMESTAMP,
    cpu_usage DECIMAL(5,2),
    memory_usage DECIMAL(5,2),
    storage_usage DECIMAL(5,2),
    expires_at TIMESTAMP,
    deleted_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(cluster_id, name)
);

CREATE INDEX idx_instances_user_id ON instances(user_id);
CREATE INDEX idx_instances_status ON instances(status);
CREATE INDEX idx_instances_cluster_id ON instances(cluster_id);
CREATE INDEX idx_instances_health_status ON instances(health_status);

-- Subscriptions
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    instance_id UUID REFERENCES instances(id),
    package_id UUID REFERENCES packages(id),
    stripe_subscription_id VARCHAR(255),
    stripe_customer_id VARCHAR(255),
    paystack_subscription_id VARCHAR(255),
    paystack_customer_id VARCHAR(255),
    billing_cycle VARCHAR(20) NOT NULL CHECK (billing_cycle IN ('monthly', 'yearly')),
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'past_due', 'unpaid', 'cancelled', 'expired', 'trialing')),
    current_period_start TIMESTAMP,
    current_period_end TIMESTAMP,
    cancel_at_period_end BOOLEAN DEFAULT false,
    cancelled_at TIMESTAMP,
    trial_start TIMESTAMP,
    trial_end TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_instance_id ON subscriptions(instance_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_stripe ON subscriptions(stripe_subscription_id);
CREATE INDEX idx_subscriptions_current_period_end ON subscriptions(current_period_end);

-- Deployment Activities
CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instance_id UUID REFERENCES instances(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    activity_type VARCHAR(50) NOT NULL CHECK (activity_type IN ('create', 'delete', 'restart', 'upgrade', 'downgrade', 'backup', 'restore', 'scale')),
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'success', 'error', 'cancelled')),
    jenkins_build_number INT,
    jenkins_build_url VARCHAR(255),
    argocd_app_name VARCHAR(255),
    error_message TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_activities_instance_id ON activities(instance_id);
CREATE INDEX idx_activities_status ON activities(status);
CREATE INDEX idx_activities_created_at ON activities(created_at);

-- Coupons
CREATE TABLE coupon_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    discount_type VARCHAR(20) NOT NULL CHECK (discount_type IN ('percent', 'fixed')),
    discount_percent DECIMAL(5,2),
    discount_amount DECIMAL(10,2),
    currency VARCHAR(3) DEFAULT 'USD',
    duration_type VARCHAR(20) NOT NULL DEFAULT 'once' CHECK (duration_type IN ('once', 'repeating', 'forever')),
    duration_months INT,
    applies_to_products UUID[],
    min_order_amount DECIMAL(10,2),
    valid_from TIMESTAMP,
    valid_until TIMESTAMP,
    max_uses INT,
    max_uses_per_user INT DEFAULT 1,
    current_uses INT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    stripe_coupon_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_coupon_groups_is_active ON coupon_groups(is_active);
CREATE INDEX idx_coupon_groups_valid_dates ON coupon_groups(valid_from, valid_until);

CREATE TABLE coupons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID REFERENCES coupon_groups(id) ON DELETE CASCADE,
    code VARCHAR(50) UNIQUE NOT NULL,
    used_by_user_id UUID REFERENCES users(id),
    used_for_order_id UUID REFERENCES sales_orders(id),
    used_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_coupons_code ON coupons(code);
CREATE INDEX idx_coupons_group_id ON coupons(group_id);
CREATE INDEX idx_coupons_is_active ON coupons(is_active);

-- Update sales_orders to reference coupons
ALTER TABLE sales_orders ADD CONSTRAINT fk_sales_orders_coupon
    FOREIGN KEY (coupon_id) REFERENCES coupons(id);

-- Support Tickets
CREATE TABLE tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    instance_id UUID REFERENCES instances(id),
    subject VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'open' CHECK (status IN ('open', 'in_progress', 'waiting_response', 'resolved', 'closed')),
    priority VARCHAR(20) DEFAULT 'normal' CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    category VARCHAR(50),
    assigned_to UUID REFERENCES users(id),
    resolved_at TIMESTAMP,
    closed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_tickets_user_id ON tickets(user_id);
CREATE INDEX idx_tickets_status ON tickets(status);
CREATE INDEX idx_tickets_priority ON tickets(priority);
CREATE INDEX idx_tickets_instance_id ON tickets(instance_id);

CREATE TABLE ticket_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID REFERENCES tickets(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    is_staff BOOLEAN DEFAULT false,
    is_internal BOOLEAN DEFAULT false,
    message TEXT NOT NULL,
    attachments JSONB DEFAULT '[]'::jsonb,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_ticket_messages_ticket_id ON ticket_messages(ticket_id);
CREATE INDEX idx_ticket_messages_created_at ON ticket_messages(created_at);

-- Audit Log
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_audit_log_user_id ON audit_log(user_id);
CREATE INDEX idx_audit_log_entity ON audit_log(entity_type, entity_id);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);

-- Notifications
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    notification_type VARCHAR(50) NOT NULL,
    link_url TEXT,
    is_read BOOLEAN DEFAULT false,
    read_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);

-- Email Queue
CREATE TABLE email_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    to_email VARCHAR(255) NOT NULL,
    to_name VARCHAR(255),
    subject VARCHAR(255) NOT NULL,
    template_name VARCHAR(100) NOT NULL,
    template_data JSONB DEFAULT '{}'::jsonb,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'sending', 'sent', 'failed')),
    attempts INT DEFAULT 0,
    max_attempts INT DEFAULT 3,
    error_message TEXT,
    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_email_queue_status ON email_queue(status);
CREATE INDEX idx_email_queue_created_at ON email_queue(created_at);

-- Insert default products
INSERT INTO products (name, slug, description, short_description, icon_name, is_available, sort_order) VALUES
('GeoServer', 'geoserver', 'Open source server for sharing geospatial data. Publish maps and data from any major spatial data source using open standards.', 'OGC-compliant map server', 'server', true, 1),
('GeoNode', 'geonode', 'Geospatial content management system. A web-based platform for developing geospatial information systems and deploying spatial data infrastructures.', 'Geospatial CMS platform', 'globe', true, 2),
('PostGIS', 'postgis', 'Spatial database extender for PostgreSQL. Adds support for geographic objects allowing location queries to be run in SQL.', 'Spatial database', 'database', true, 3);

-- Insert default packages for each product
INSERT INTO packages (product_id, name, slug, description, price_monthly, price_yearly, features, cpu_limit, memory_limit, storage_limit, concurrent_users, is_popular, sort_order)
SELECT
    p.id,
    'Starter',
    'starter',
    'Perfect for small projects and development',
    19.99,
    199.99,
    '["5GB Storage", "2 Workspaces", "Basic Support", "1 User"]'::jsonb,
    '500m',
    '512Mi',
    '5Gi',
    1,
    false,
    1
FROM products p WHERE p.slug = 'geoserver';

INSERT INTO packages (product_id, name, slug, description, price_monthly, price_yearly, features, cpu_limit, memory_limit, storage_limit, concurrent_users, is_popular, sort_order)
SELECT
    p.id,
    'Professional',
    'professional',
    'For growing teams and production workloads',
    49.99,
    499.99,
    '["25GB Storage", "10 Workspaces", "Priority Support", "5 Users", "Custom Styles", "GeoWebCache"]'::jsonb,
    '1000m',
    '2Gi',
    '25Gi',
    5,
    true,
    2
FROM products p WHERE p.slug = 'geoserver';

INSERT INTO packages (product_id, name, slug, description, price_monthly, price_yearly, features, cpu_limit, memory_limit, storage_limit, concurrent_users, is_popular, sort_order)
SELECT
    p.id,
    'Enterprise',
    'enterprise',
    'For large organizations with high demands',
    149.99,
    1499.99,
    '["100GB Storage", "Unlimited Workspaces", "24/7 Support", "Unlimited Users", "Custom Domain", "SLA Guarantee", "Dedicated Resources"]'::jsonb,
    '4000m',
    '8Gi',
    '100Gi',
    NULL,
    false,
    3
FROM products p WHERE p.slug = 'geoserver';

-- GeoNode packages
INSERT INTO packages (product_id, name, slug, description, price_monthly, price_yearly, features, cpu_limit, memory_limit, storage_limit, concurrent_users, is_popular, sort_order)
SELECT
    p.id,
    'Starter',
    'starter',
    'Perfect for small projects and development',
    29.99,
    299.99,
    '["10GB Storage", "5 Layers", "Basic Support", "3 Users"]'::jsonb,
    '1000m',
    '1Gi',
    '10Gi',
    3,
    false,
    1
FROM products p WHERE p.slug = 'geonode';

INSERT INTO packages (product_id, name, slug, description, price_monthly, price_yearly, features, cpu_limit, memory_limit, storage_limit, concurrent_users, is_popular, sort_order)
SELECT
    p.id,
    'Professional',
    'professional',
    'For growing teams and production workloads',
    79.99,
    799.99,
    '["50GB Storage", "Unlimited Layers", "Priority Support", "10 Users", "Custom Themes", "API Access"]'::jsonb,
    '2000m',
    '4Gi',
    '50Gi',
    10,
    true,
    2
FROM products p WHERE p.slug = 'geonode';

INSERT INTO packages (product_id, name, slug, description, price_monthly, price_yearly, features, cpu_limit, memory_limit, storage_limit, concurrent_users, is_popular, sort_order)
SELECT
    p.id,
    'Enterprise',
    'enterprise',
    'For large organizations with high demands',
    199.99,
    1999.99,
    '["500GB Storage", "Unlimited Layers", "24/7 Support", "Unlimited Users", "Custom Domain", "SSO Integration", "Dedicated Resources"]'::jsonb,
    '8000m',
    '16Gi',
    '500Gi',
    NULL,
    false,
    3
FROM products p WHERE p.slug = 'geonode';

-- PostGIS packages
INSERT INTO packages (product_id, name, slug, description, price_monthly, price_yearly, features, cpu_limit, memory_limit, storage_limit, concurrent_users, is_popular, sort_order)
SELECT
    p.id,
    'Starter',
    'starter',
    'Perfect for small projects and development',
    14.99,
    149.99,
    '["5GB Storage", "5 Connections", "Basic Support", "Daily Backups"]'::jsonb,
    '500m',
    '1Gi',
    '5Gi',
    5,
    false,
    1
FROM products p WHERE p.slug = 'postgis';

INSERT INTO packages (product_id, name, slug, description, price_monthly, price_yearly, features, cpu_limit, memory_limit, storage_limit, concurrent_users, is_popular, sort_order)
SELECT
    p.id,
    'Professional',
    'professional',
    'For growing teams and production workloads',
    39.99,
    399.99,
    '["25GB Storage", "25 Connections", "Priority Support", "Hourly Backups", "Point-in-time Recovery"]'::jsonb,
    '1000m',
    '4Gi',
    '25Gi',
    25,
    true,
    2
FROM products p WHERE p.slug = 'postgis';

INSERT INTO packages (product_id, name, slug, description, price_monthly, price_yearly, features, cpu_limit, memory_limit, storage_limit, concurrent_users, is_popular, sort_order)
SELECT
    p.id,
    'Enterprise',
    'enterprise',
    'For large organizations with high demands',
    99.99,
    999.99,
    '["100GB Storage", "Unlimited Connections", "24/7 Support", "Continuous Backups", "Read Replicas", "Dedicated Resources"]'::jsonb,
    '4000m',
    '16Gi',
    '100Gi',
    NULL,
    false,
    3
FROM products p WHERE p.slug = 'postgis';

-- Insert default cluster
INSERT INTO clusters (code, name, region, country, domain, is_active) VALUES
('eu-west-1', 'Europe (Ireland)', 'eu-west-1', 'Ireland', 'eu.geospatialhosting.io', true),
('us-east-1', 'US East (Virginia)', 'us-east-1', 'USA', 'us.geospatialhosting.io', true),
('ap-south-1', 'Asia Pacific (Mumbai)', 'ap-south-1', 'India', 'ap.geospatialhosting.io', true);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at trigger to all relevant tables
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_billing_info_updated_at BEFORE UPDATE ON user_billing_info
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_packages_updated_at BEFORE UPDATE ON packages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_clusters_updated_at BEFORE UPDATE ON clusters
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sales_orders_updated_at BEFORE UPDATE ON sales_orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_instances_updated_at BEFORE UPDATE ON instances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_subscriptions_updated_at BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tickets_updated_at BEFORE UPDATE ON tickets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
