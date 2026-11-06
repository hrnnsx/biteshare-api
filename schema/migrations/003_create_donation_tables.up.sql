-- 1. Donors Table
CREATE TABLE IF NOT EXISTS donors (
    donor_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    address TEXT NOT NULL
);

-- 2. Donation Requests Table
CREATE TABLE IF NOT EXISTS donation_requests (
    request_id SERIAL PRIMARY KEY,
    donor_id INTEGER NOT NULL REFERENCES donors(donor_id) ON DELETE RESTRICT,
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    notes TEXT,
    submitted_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

-- 3. Donated Foods Table
CREATE TABLE IF NOT EXISTS donated_foods (
    food_id SERIAL PRIMARY KEY,
    request_id INTEGER NOT NULL REFERENCES donation_requests(request_id) ON DELETE CASCADE,
    food_name VARCHAR(100) NOT NULL,
    quantity NUMERIC(10,2) NOT NULL CHECK (quantity > 0),
    unit VARCHAR(30) NOT NULL,
    expiry_date DATE,
    description TEXT
);

-- 4. Donation Images Table
CREATE TABLE IF NOT EXISTS donation_images (
    image_id SERIAL PRIMARY KEY,
    food_id INTEGER NOT NULL REFERENCES donated_foods(food_id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    file_id VARCHAR(255),
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

-- 5. Pickup Schedules Table
CREATE TABLE IF NOT EXISTS pickup_schedules (
    schedule_id SERIAL PRIMARY KEY,
    request_id INTEGER NOT NULL REFERENCES donation_requests(request_id) ON DELETE CASCADE,
    admin_id INTEGER NOT NULL REFERENCES admins(admin_id) ON DELETE RESTRICT,
    pickup_date TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    pickup_time VARCHAR(20) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'scheduled',
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
);

-- 6. Pickups (Execution Result) Table
CREATE TABLE IF NOT EXISTS pickups (
    pickup_id SERIAL PRIMARY KEY,
    schedule_id INTEGER NOT NULL UNIQUE REFERENCES pickup_schedules(schedule_id) ON DELETE CASCADE,
    picked_up_at TIMESTAMP WITHOUT TIME ZONE,
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    notes TEXT
);

-- Indexes for donation tables
CREATE INDEX IF NOT EXISTS idx_donation_requests_donor_id ON donation_requests(donor_id);
CREATE INDEX IF NOT EXISTS idx_donation_requests_status ON donation_requests(status);
CREATE INDEX IF NOT EXISTS idx_donated_foods_request_id ON donated_foods(request_id);
CREATE INDEX IF NOT EXISTS idx_donation_images_food_id ON donation_images(food_id);
CREATE INDEX IF NOT EXISTS idx_pickup_schedules_request_id ON pickup_schedules(request_id);
CREATE INDEX IF NOT EXISTS idx_pickup_schedules_admin_id ON pickup_schedules(admin_id);

-- At most one primary photo per donated food
CREATE UNIQUE INDEX IF NOT EXISTS uq_donation_images_one_primary ON donation_images(food_id) WHERE is_primary = TRUE;