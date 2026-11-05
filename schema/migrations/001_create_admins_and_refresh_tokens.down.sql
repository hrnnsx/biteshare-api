-- Drop the child table first because it references admins.
DROP TABLE IF EXISTS admin_refresh_tokens;

-- Drop the admins table after its dependent table is removed.
DROP TABLE IF EXISTS admins;
