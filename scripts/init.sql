-- GigCo Database Complete Setup Script (Fixed Version)
-- Run this in PgAdmin Query Tool to create the full schema

-- First, ensure we have the UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Drop existing tables if they exist (be careful in production!)
DROP TABLE IF EXISTS job_reviews CASCADE;
DROP TABLE IF EXISTS notifications CASCADE;
DROP TABLE IF EXISTS notification_preferences CASCADE;
DROP TABLE IF EXISTS user_payment_methods CASCADE;
DROP TABLE IF EXISTS settlement_batches CASCADE;
DROP TABLE IF EXISTS payment_providers CASCADE;
DROP TABLE IF EXISTS schedules CASCADE;
DROP TABLE IF EXISTS transactions CASCADE;
DROP TABLE IF EXISTS jobs CASCADE;
DROP TABLE IF EXISTS people CASCADE;
DROP TABLE IF EXISTS customers CASCADE;

-- Drop existing types if they exist
DROP TYPE IF EXISTS user_role CASCADE;
DROP TYPE IF EXISTS job_status CASCADE;
DROP TYPE IF EXISTS transaction_status CASCADE;
DROP TYPE IF EXISTS notification_type CASCADE;
DROP TYPE IF EXISTS notification_status CASCADE;

-- Create enum types for better data integrity
CREATE TYPE user_role AS ENUM ('consumer', 'gig_worker', 'admin');
CREATE TYPE job_status AS ENUM ('posted', 'accepted', 'in_progress', 'completed', 'cancelled');
CREATE TYPE transaction_status AS ENUM ('pending', 'completed', 'failed', 'refunded');
CREATE TYPE notification_type AS ENUM ('job_posted', 'job_accepted', 'job_completed', 'payment_received', 'system_message');
CREATE TYPE notification_status AS ENUM ('unread', 'read', 'archived');

-- ==============================================
-- CORE TABLES
-- ==============================================

-- People table (replaces and extends customers)
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

-- Payment providers table
CREATE TABLE payment_providers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    api_endpoint VARCHAR(255),
    webhook_endpoint VARCHAR(255),
    config_json JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Settlement batches for payment reconciliation
CREATE TABLE settlement_batches (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    provider_id INTEGER NOT NULL REFERENCES payment_providers(id),
    batch_date DATE NOT NULL,
    batch_reference VARCHAR(255),
    total_amount DECIMAL(12, 2) NOT NULL,
    transaction_count INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(50) DEFAULT 'pending',
    processed_at TIMESTAMP WITH TIME ZONE,
    reconciled_at TIMESTAMP WITH TIME ZONE,
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
    payment_intent_id VARCHAR(255),
    payment_method VARCHAR(50),
    escrow_released_at TIMESTAMP WITH TIME ZONE,
    processing_fee DECIMAL(10, 2) DEFAULT 0.00,
    net_amount DECIMAL(10, 2),
    settlement_batch_id INTEGER REFERENCES settlement_batches(id),
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
    is_available BOOLEAN DEFAULT true,
    job_id INTEGER REFERENCES jobs(id) ON DELETE SET NULL,
    recurring_pattern VARCHAR(50),
    recurring_until TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- User payment methods
CREATE TABLE user_payment_methods (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    user_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    provider_id INTEGER NOT NULL REFERENCES payment_providers(id),
    external_id VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    last_four VARCHAR(4),
    brand VARCHAR(50),
    is_default BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    expires_at DATE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Notifications table
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    user_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    type notification_type NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    status notification_status DEFAULT 'unread',
    related_job_id INTEGER REFERENCES jobs(id) ON DELETE SET NULL,
    related_transaction_id INTEGER REFERENCES transactions(id) ON DELETE SET NULL,
    action_url VARCHAR(255),
    scheduled_for TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    sent_at TIMESTAMP WITH TIME ZONE,
    read_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- User notification preferences
CREATE TABLE notification_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    type notification_type NOT NULL,
    email_enabled BOOLEAN DEFAULT true,
    push_enabled BOOLEAN DEFAULT true,
    sms_enabled BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, type)
);

-- Job reviews/ratings
CREATE TABLE job_reviews (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    job_id INTEGER NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    reviewer_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    reviewee_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    review_text TEXT,
    is_public BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(job_id, reviewer_id)
);

-- Create backward compatibility table for existing API
CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ==============================================
-- INDEXES FOR PERFORMANCE
-- ==============================================

-- People table indexes
CREATE INDEX idx_people_email ON people(email);
CREATE INDEX idx_people_uuid ON people(uuid);
CREATE INDEX idx_people_role ON people(role);
CREATE INDEX idx_people_active ON people(is_active) WHERE is_active = true;
CREATE INDEX idx_people_location ON people(latitude, longitude) WHERE latitude IS NOT NULL AND longitude IS NOT NULL;

-- Jobs table indexes
CREATE INDEX idx_jobs_consumer_id ON jobs(consumer_id);
CREATE INDEX idx_jobs_gig_worker_id ON jobs(gig_worker_id) WHERE gig_worker_id IS NOT NULL;
CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_category ON jobs(category) WHERE category IS NOT NULL;
CREATE INDEX idx_jobs_status_created ON jobs(status, created_at);

-- Transactions indexes
CREATE INDEX idx_transactions_job_id ON transactions(job_id);
CREATE INDEX idx_transactions_consumer_id ON transactions(consumer_id);
CREATE INDEX idx_transactions_gig_worker_id ON transactions(gig_worker_id);
CREATE INDEX idx_transactions_status ON transactions(status);

-- Schedules indexes
CREATE INDEX idx_schedules_gig_worker_id ON schedules(gig_worker_id);
CREATE INDEX idx_schedules_time_range ON schedules(start_time, end_time);
CREATE INDEX idx_schedules_worker_availability ON schedules(gig_worker_id, is_available, start_time);

-- Customers table indexes (for backward compatibility)
CREATE INDEX idx_customers_name ON customers(name);
CREATE INDEX idx_customers_created_at ON customers(created_at);

-- ==============================================
-- CONSTRAINTS
-- ==============================================

-- Ensure job schedules make sense
ALTER TABLE jobs
ADD CONSTRAINT chk_jobs_schedule_order 
CHECK (scheduled_end IS NULL OR scheduled_start IS NULL OR scheduled_end > scheduled_start);

ALTER TABLE jobs
ADD CONSTRAINT chk_jobs_actual_times 
CHECK (actual_end IS NULL OR actual_start IS NULL OR actual_end > actual_start);

-- Ensure schedule time ranges are valid
ALTER TABLE schedules
ADD CONSTRAINT chk_schedules_time_order 
CHECK (end_time > start_time);

-- Ensure transaction amounts are positive
ALTER TABLE transactions
ADD CONSTRAINT chk_transactions_positive_amount 
CHECK (amount > 0);

-- Ensure only one default payment method per user
CREATE UNIQUE INDEX idx_user_payment_methods_one_default 
ON user_payment_methods(user_id) 
WHERE is_default = true;

-- ==============================================
-- TRIGGERS FOR UPDATED_AT
-- ==============================================

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
CREATE TRIGGER update_payment_providers_updated_at BEFORE UPDATE ON payment_providers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_payment_methods_updated_at BEFORE UPDATE ON user_payment_methods FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_settlement_batches_updated_at BEFORE UPDATE ON settlement_batches FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_notifications_updated_at BEFORE UPDATE ON notifications FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_notification_preferences_updated_at BEFORE UPDATE ON notification_preferences FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_job_reviews_updated_at BEFORE UPDATE ON job_reviews FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_customers_updated_at BEFORE UPDATE ON customers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ==============================================
-- SEED DATA
-- ==============================================

-- Insert initial admin user
INSERT INTO people (email, name, role, is_active, email_verified) VALUES 
('admin@gigco.com', 'System Administrator', 'admin', true, true);

-- Insert default payment providers
INSERT INTO payment_providers (name, display_name, is_active) VALUES 
('clover', 'Clover', true),
('stripe', 'Stripe', false),
('square', 'Square', false);

-- Insert development test users
INSERT INTO people (email, name, phone, address, role, is_active, email_verified, phone_verified) VALUES 
    ('consumer1@gigco.dev', 'Alice Johnson', '+1-555-0101', '123 Oak Street, Downtown, City 12345', 'consumer', true, true, true),
    ('consumer2@gigco.dev', 'Bob Wilson', '+1-555-0102', '456 Pine Avenue, Uptown, City 12346', 'consumer', true, true, false),
    ('worker1@gigco.dev', 'Carol Davis', '+1-555-0201', '789 Elm Boulevard, Midtown, City 12347', 'gig_worker', true, true, true),
    ('worker2@gigco.dev', 'David Miller', '+1-555-0202', '321 Maple Drive, Riverside, City 12348', 'gig_worker', true, true, true),
    ('worker3@gigco.dev', 'Eva Garcia', '+1-555-0203', '654 Cedar Lane, Hillside, City 12349', 'gig_worker', true, true, false)
ON CONFLICT (email) DO UPDATE SET
    name = EXCLUDED.name,
    phone = EXCLUDED.phone,
    address = EXCLUDED.address,
    role = EXCLUDED.role,
    is_active = EXCLUDED.is_active,
    email_verified = EXCLUDED.email_verified,
    phone_verified = EXCLUDED.phone_verified;

-- Insert sample data into customers table for backward compatibility with existing API
INSERT INTO customers (name, address) VALUES 
    ('John Doe', '123 Main St, Anytown, USA'),
    ('Jane Smith', '456 Oak Ave, Somewhere, USA'),
    ('Bob Johnson', '789 Pine Rd, Nowhere, USA');

-- Create default notification preferences for existing users
INSERT INTO notification_preferences (user_id, type, email_enabled, push_enabled, sms_enabled)
SELECT 
    p.id,
    nt.type,
    true,
    true,
    false
FROM people p
CROSS JOIN (
    SELECT unnest(enum_range(NULL::notification_type)) as type
) nt
ON CONFLICT (user_id, type) DO NOTHING;

-- Create a health check function
CREATE OR REPLACE FUNCTION health_check() 
RETURNS TABLE(status TEXT, "timestamp" TIMESTAMP WITH TIME ZONE) AS $$
BEGIN
    RETURN QUERY SELECT 'healthy'::TEXT, NOW();
END;
$$ LANGUAGE plpgsql;

-- Show completion message
DO $$
BEGIN
    RAISE NOTICE 'GigCo database schema created successfully!';
    RAISE NOTICE 'Tables created: %', (SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public');
    RAISE NOTICE 'Users in people table: %', (SELECT count(*) FROM people);
    RAISE NOTICE 'Users in customers table: %', (SELECT count(*) FROM customers);
END $$;