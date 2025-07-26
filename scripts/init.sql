-- Initialize GigCo Database
-- This script creates the basic tables needed for the current application

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create customers table (matches existing model)
CREATE TABLE IF NOT EXISTS customers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_customers_name ON customers(name);
CREATE INDEX IF NOT EXISTS idx_customers_created_at ON customers(created_at);

-- Insert sample data for testing
INSERT INTO customers (name, address) VALUES 
    ('John Doe', '123 Main St, Anytown, USA'),
    ('Jane Smith', '456 Oak Ave, Somewhere, USA'),
    ('Bob Johnson', '789 Pine Rd, Nowhere, USA')
ON CONFLICT DO NOTHING;

-- Create a health check function
CREATE OR REPLACE FUNCTION health_check() 
RETURNS TABLE(status TEXT, "timestamp" TIMESTAMP WITH TIME ZONE) AS $$
BEGIN
    RETURN QUERY SELECT 'healthy'::TEXT, NOW();
END;
$$ LANGUAGE plpgsql;