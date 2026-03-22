-- Drop triggers first
DROP TRIGGER IF EXISTS update_tickets_updated_at ON tickets;
DROP TRIGGER IF EXISTS update_subscriptions_updated_at ON subscriptions;
DROP TRIGGER IF EXISTS update_instances_updated_at ON instances;
DROP TRIGGER IF EXISTS update_sales_orders_updated_at ON sales_orders;
DROP TRIGGER IF EXISTS update_clusters_updated_at ON clusters;
DROP TRIGGER IF EXISTS update_packages_updated_at ON packages;
DROP TRIGGER IF EXISTS update_products_updated_at ON products;
DROP TRIGGER IF EXISTS update_user_billing_info_updated_at ON user_billing_info;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS email_queue;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS ticket_messages;
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS coupons;
DROP TABLE IF EXISTS coupon_groups;
DROP TABLE IF EXISTS activities;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS instances;
DROP TABLE IF EXISTS sales_orders;
DROP TABLE IF EXISTS clusters;
DROP TABLE IF EXISTS packages;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS user_billing_info;
DROP TABLE IF EXISTS user_tokens;
DROP TABLE IF EXISTS users;
