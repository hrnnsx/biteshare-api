DROP INDEX IF EXISTS uq_donation_images_one_primary;
DROP INDEX IF EXISTS idx_pickup_schedules_admin_id;
DROP INDEX IF EXISTS idx_pickup_schedules_request_id;
DROP INDEX IF EXISTS idx_donation_images_food_id;
DROP INDEX IF EXISTS idx_donated_foods_request_id;
DROP INDEX IF EXISTS idx_donation_requests_status;
DROP INDEX IF EXISTS idx_donation_requests_donor_id;

DROP TABLE IF EXISTS pickups;
DROP TABLE IF EXISTS pickup_schedules;
DROP TABLE IF EXISTS donation_images;
DROP TABLE IF EXISTS donated_foods;
DROP TABLE IF EXISTS donation_requests;
DROP TABLE IF EXISTS donors;