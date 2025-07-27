-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create enum types for better data integrity
CREATE TYPE user_role AS ENUM ('consumer', 'gig_worker', 'admin');
CREATE TYPE job_status AS ENUM ('posted', 'accepted', 'in_progress', 'completed', 'cancelled');
CREATE TYPE transaction_status AS ENUM ('pending', 'completed', 'failed', 'refunded');

-- People table (replaces and extends customers)
-- This is the base table for all users (consumers, gig workers, admins)
CREATE TABLE people (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    address TEXT,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    place_id VARCHAR(255),
    role user_role NOT NULL DEFAULT 'consumer',
    is_active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    phone_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Jobs table
CREATE TABLE jobs (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    consumer_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    gig_worker_id INTEGER REFERENCES people(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(100),
    location_address TEXT,
    location_latitude DECIMAL(10, 8),
    location_longitude DECIMAL(11, 8),
    estimated_duration_hours DECIMAL(4, 2),
    pay_rate_per_hour DECIMAL(10, 2),
    total_pay DECIMAL(10, 2),
    status job_status DEFAULT 'posted',
    scheduled_start TIMESTAMP WITH TIME ZONE,
    scheduled_end TIMESTAMP WITH TIME ZONE,
    actual_start TIMESTAMP WITH TIME ZONE,
    actual_end TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Transactions table for payment tracking
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    job_id INTEGER NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    consumer_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    gig_worker_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status transaction_status DEFAULT 'pending',
    payment_intent_id VARCHAR(255), -- External payment provider ID
    payment_method VARCHAR(50), -- 'clover', 'stripe', 'square', etc.
    escrow_released_at TIMESTAMP WITH TIME ZONE,
    processing_fee DECIMAL(10, 2) DEFAULT 0.00,
    net_amount DECIMAL(10, 2),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Schedules table for worker availability
CREATE TABLE schedules (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    gig_worker_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    title VARCHAR(255),
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    is_available BOOLEAN DEFAULT true, -- true = available, false = busy
    job_id INTEGER REFERENCES jobs(id) ON DELETE SET NULL, -- if linked to a job
    recurring_pattern VARCHAR(50), -- 'weekly', 'daily', 'monthly', null
    recurring_until TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at triggers to all tables
CREATE TRIGGER update_people_updated_at BEFORE UPDATE ON people FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_jobs_updated_at BEFORE UPDATE ON jobs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON transactions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_schedules_updated_at BEFORE UPDATE ON schedules FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert initial admin user
INSERT INTO people (email, name, role, is_active, email_verified) VALUES 
('admin@gigco.com', 'System Administrator', 'admin', true, true);

-- Migrate existing customers data if it exists
-- This handles the transition from your current customers table
INSERT INTO people (name, address, created_at, updated_at, email, role)
SELECT 
    name,
    address,
    created_at,
    COALESCE(updated_at, created_at),
    LOWER(REPLACE(REPLACE(name, ' ', '.'), '''', '')) || '@temp.gigco.com', -- Generate temporary email
    'consumer'
FROM customers
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'customers')
ON CONFLICT (email) DO NOTHING;