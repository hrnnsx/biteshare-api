ALTER TABLE payments
    ADD COLUMN IF NOT EXISTS proof_image_url TEXT,
    ADD COLUMN IF NOT EXISTS proof_file_id VARCHAR(255);
