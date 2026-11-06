ALTER TABLE payments
    DROP COLUMN IF EXISTS proof_image_url,
    DROP COLUMN IF EXISTS proof_file_id;
