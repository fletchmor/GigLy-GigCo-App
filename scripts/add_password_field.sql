-- Add password field to people table
-- This migration adds a password_hash field to store bcrypt hashed passwords

ALTER TABLE people
ADD COLUMN password_hash VARCHAR(255);

-- Update the table comment to reflect the password field
COMMENT ON COLUMN people.password_hash IS 'bcrypt hashed password for user authentication';