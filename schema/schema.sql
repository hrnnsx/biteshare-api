-- BiteShare schema for PostgreSQL 17.11
-- 16 tables, 15 foreign-key relationships. Execute on a new/empty database.
-- ERD2: donors, admins, buyers, pickup_schedules, food_products, pickups.
-- ERD1: orders, payments. Images: ERD2. JWT refresh tokens: ERD1.
-- PostgreSQL has no DATETIME type: TIMESTAMP WITHOUT TIME ZONE is used.
BEGIN;

CREATE TABLE donors (
 donor_id SERIAL PRIMARY KEY,
 name VARCHAR(100) NOT NULL,
 email VARCHAR(100) NOT NULL,
 phone VARCHAR(20) NOT NULL,
 address TEXT NOT NULL
);

CREATE TABLE admins (
 admin_id SERIAL PRIMARY KEY,
 name VARCHAR(100) NOT NULL,
 email VARCHAR(100) NOT NULL UNIQUE,
 password VARCHAR(255) NOT NULL
);

CREATE TABLE admin_refresh_tokens (
 token_id SERIAL PRIMARY KEY,
 admin_id INTEGER NOT NULL REFERENCES admins(admin_id) ON DELETE CASCADE,
 token_hash VARCHAR(255) NOT NULL UNIQUE,
 expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
 revoked_at TIMESTAMP WITH TIME ZONE,
 created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE buyers (
 buyer_id SERIAL PRIMARY KEY,
 name VARCHAR(100) NOT NULL,
 email VARCHAR(100) NOT NULL,
 phone VARCHAR(20) NOT NULL,
 address TEXT NOT NULL
);

CREATE TABLE donation_requests (
 request_id SERIAL PRIMARY KEY,
 donor_id INTEGER NOT NULL REFERENCES donors(donor_id),
 status VARCHAR(30) NOT NULL DEFAULT 'pending',
 notes TEXT,
 submitted_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE donated_foods (
 food_id SERIAL PRIMARY KEY,
 request_id INTEGER NOT NULL REFERENCES donation_requests(request_id),
 food_name VARCHAR(100) NOT NULL,
 quantity NUMERIC(10,2) NOT NULL CHECK (quantity > 0),
 unit VARCHAR(30) NOT NULL,
 expiry_date DATE,
 description TEXT
);

CREATE TABLE donation_images (
 image_id SERIAL PRIMARY KEY,
 food_id INTEGER NOT NULL REFERENCES donated_foods(food_id),
 image_url TEXT NOT NULL,
 file_id VARCHAR(255),
 is_primary BOOLEAN NOT NULL DEFAULT FALSE,
 created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE pickup_schedules (
 schedule_id SERIAL PRIMARY KEY,
 request_id INTEGER NOT NULL REFERENCES donation_requests(request_id),
 admin_id INTEGER NOT NULL REFERENCES admins(admin_id),
 pickup_date TIMESTAMP WITHOUT TIME ZONE NOT NULL,
 pickup_time VARCHAR(20) NOT NULL,
 status VARCHAR(30) NOT NULL DEFAULT 'scheduled',
 created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE pickups (
 pickup_id SERIAL PRIMARY KEY,
 schedule_id INTEGER NOT NULL UNIQUE REFERENCES pickup_schedules(schedule_id),
 picked_up_at TIMESTAMP WITHOUT TIME ZONE,
 status VARCHAR(30) NOT NULL DEFAULT 'pending',
 notes TEXT
);

CREATE TABLE food_suppliers (
 supplier_id SERIAL PRIMARY KEY,
 supplier_name VARCHAR(150) NOT NULL,
 email VARCHAR(100),
 phone VARCHAR(20),
 address TEXT,
 supplier_type VARCHAR(50),
 created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE food_products (
 product_id SERIAL PRIMARY KEY,
 admin_id INTEGER NOT NULL REFERENCES admins(admin_id),
 supplier_id INTEGER REFERENCES food_suppliers(supplier_id),
 product_name VARCHAR(150) NOT NULL,
 description TEXT,
 price NUMERIC(12,2) NOT NULL CHECK (price >= 0),
 stock NUMERIC(10,2) NOT NULL DEFAULT 0 CHECK (stock >= 0),
 unit VARCHAR(30) NOT NULL,
 expiry_date DATE,
 status VARCHAR(30) NOT NULL DEFAULT 'available',
 created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE product_images (
 image_id SERIAL PRIMARY KEY,
 product_id INTEGER NOT NULL REFERENCES food_products(product_id),
 image_url TEXT NOT NULL,
 file_id VARCHAR(255),
 is_primary BOOLEAN NOT NULL DEFAULT FALSE,
 created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE orders (
 order_id SERIAL PRIMARY KEY,
 buyer_id INTEGER NOT NULL REFERENCES buyers(buyer_id),
 order_number VARCHAR(50) NOT NULL UNIQUE,
 delivery_address TEXT NOT NULL,
 total_amount NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (total_amount >= 0),
 delivery_fee NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (delivery_fee >= 0),
 order_status VARCHAR(30) NOT NULL DEFAULT 'pending',
 ordered_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE order_items (
 order_item_id SERIAL PRIMARY KEY,
 order_id INTEGER NOT NULL REFERENCES orders(order_id),
 product_id INTEGER NOT NULL REFERENCES food_products(product_id),
 quantity NUMERIC(10,2) NOT NULL CHECK (quantity > 0),
 unit_price NUMERIC(12,2) NOT NULL CHECK (unit_price >= 0),
 subtotal NUMERIC(12,2) GENERATED ALWAYS AS (quantity * unit_price) STORED
);

CREATE TABLE payments (
 payment_id SERIAL PRIMARY KEY,
 order_id INTEGER NOT NULL REFERENCES orders(order_id),
 payment_method VARCHAR(50) NOT NULL,
 amount NUMERIC(12,2) NOT NULL CHECK (amount >= 0),
 payment_status VARCHAR(30) NOT NULL DEFAULT 'pending',
 transaction_reference VARCHAR(100) UNIQUE,
 paid_at TIMESTAMP WITH TIME ZONE,
 proof_image_url TEXT,
 proof_file_id VARCHAR(255)
);

CREATE TABLE deliveries (
 delivery_id SERIAL PRIMARY KEY,
 order_id INTEGER NOT NULL UNIQUE REFERENCES orders(order_id),
 courier_name VARCHAR(100),
 courier_phone VARCHAR(20),
 delivery_address TEXT NOT NULL,
 delivery_status VARCHAR(30) NOT NULL DEFAULT 'pending',
 shipped_at TIMESTAMP WITHOUT TIME ZONE,
 delivered_at TIMESTAMP WITHOUT TIME ZONE,
 notes TEXT
);

-- Index foreign keys and frequently queried filters.
CREATE INDEX idx_admin_refresh_tokens_admin_id ON admin_refresh_tokens(admin_id);
CREATE INDEX idx_admin_refresh_tokens_expires_at ON admin_refresh_tokens(expires_at);
CREATE INDEX idx_donation_requests_donor_id ON donation_requests(donor_id);
CREATE INDEX idx_donation_requests_status ON donation_requests(status);
CREATE INDEX idx_donated_foods_request_id ON donated_foods(request_id);
CREATE INDEX idx_donation_images_food_id ON donation_images(food_id);
CREATE INDEX idx_pickup_schedules_request_id ON pickup_schedules(request_id);
CREATE INDEX idx_pickup_schedules_admin_id ON pickup_schedules(admin_id);
CREATE INDEX idx_food_products_admin_id ON food_products(admin_id);
CREATE INDEX idx_food_products_supplier_id ON food_products(supplier_id);
CREATE INDEX idx_food_products_status ON food_products(status);
CREATE INDEX idx_product_images_product_id ON product_images(product_id);
CREATE INDEX idx_orders_buyer_id ON orders(buyer_id);
CREATE INDEX idx_orders_status ON orders(order_status);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);
CREATE INDEX idx_payments_order_id ON payments(order_id);
-- At most one primary photo per product or donated food.
CREATE UNIQUE INDEX uq_product_images_one_primary ON product_images(product_id) WHERE is_primary = TRUE;
CREATE UNIQUE INDEX uq_donation_images_one_primary ON donation_images(food_id) WHERE is_primary = TRUE;

COMMIT;

