-- 1. Buyers Table
CREATE TABLE IF NOT EXISTS buyers (
    buyer_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    address TEXT NOT NULL
);

-- 2. Food Suppliers Table
CREATE TABLE IF NOT EXISTS food_suppliers (
    supplier_id SERIAL PRIMARY KEY,
    supplier_name VARCHAR(150) NOT NULL,
    email VARCHAR(100),
    phone VARCHAR(20),
    address TEXT,
    supplier_type VARCHAR(50),
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

-- 3. Food Products Table
CREATE TABLE IF NOT EXISTS food_products (
    product_id SERIAL PRIMARY KEY,
    admin_id INTEGER NOT NULL REFERENCES admins(admin_id) ON DELETE RESTRICT,
    supplier_id INTEGER REFERENCES food_suppliers(supplier_id) ON DELETE SET NULL,
    product_name VARCHAR(150) NOT NULL,
    description TEXT,
    price NUMERIC(12,2) NOT NULL CHECK (price >= 0),
    stock NUMERIC(10,2) NOT NULL DEFAULT 0 CHECK (stock >= 0),
    unit VARCHAR(30) NOT NULL,
    expiry_date DATE,
    status VARCHAR(30) NOT NULL DEFAULT 'available',
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

-- 4. Product Images Table
CREATE TABLE IF NOT EXISTS product_images (
    image_id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES food_products(product_id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    file_id VARCHAR(255),
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

-- 5. Orders Table
CREATE TABLE IF NOT EXISTS orders (
    order_id SERIAL PRIMARY KEY,
    buyer_id INTEGER NOT NULL REFERENCES buyers(buyer_id) ON DELETE RESTRICT,
    order_number VARCHAR(50) NOT NULL UNIQUE,
    delivery_address TEXT NOT NULL,
    total_amount NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (total_amount >= 0),
    delivery_fee NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (delivery_fee >= 0),
    order_status VARCHAR(30) NOT NULL DEFAULT 'pending',
    ordered_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 6. Order Items Table
CREATE TABLE IF NOT EXISTS order_items (
    order_item_id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(order_id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES food_products(product_id) ON DELETE RESTRICT,
    quantity NUMERIC(10,2) NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(12,2) NOT NULL CHECK (unit_price >= 0),
    subtotal NUMERIC(12,2) GENERATED ALWAYS AS (quantity * unit_price) STORED
);

-- 7. Payments Table
CREATE TABLE IF NOT EXISTS payments (
    payment_id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(order_id) ON DELETE CASCADE,
    payment_method VARCHAR(50) NOT NULL,
    amount NUMERIC(12,2) NOT NULL CHECK (amount >= 0),
    payment_status VARCHAR(30) NOT NULL DEFAULT 'pending',
    transaction_reference VARCHAR(100) UNIQUE,
    paid_at TIMESTAMP WITH TIME ZONE
);

-- 8. Deliveries Table
CREATE TABLE IF NOT EXISTS deliveries (
    delivery_id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL UNIQUE REFERENCES orders(order_id) ON DELETE CASCADE,
    courier_name VARCHAR(100),
    courier_phone VARCHAR(20),
    delivery_address TEXT NOT NULL,
    delivery_status VARCHAR(30) NOT NULL DEFAULT 'pending',
    shipped_at TIMESTAMP WITHOUT TIME ZONE,
    delivered_at TIMESTAMP WITHOUT TIME ZONE,
    notes TEXT
);

-- Indexes for ecommerce tables
CREATE INDEX IF NOT EXISTS idx_food_products_admin_id ON food_products(admin_id);
CREATE INDEX IF NOT EXISTS idx_food_products_supplier_id ON food_products(supplier_id);
CREATE INDEX IF NOT EXISTS idx_food_products_status ON food_products(status);
CREATE INDEX IF NOT EXISTS idx_product_images_product_id ON product_images(product_id);
CREATE INDEX IF NOT EXISTS idx_orders_buyer_id ON orders(buyer_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(order_status);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);
CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);

-- At most one primary photo per product
CREATE UNIQUE INDEX IF NOT EXISTS uq_product_images_one_primary ON product_images(product_id) WHERE is_primary = TRUE;