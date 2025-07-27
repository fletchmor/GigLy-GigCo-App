-- Create additional enum types
CREATE TYPE notification_type AS ENUM ('job_posted', 'job_accepted', 'job_completed', 'payment_received', 'system_message');
CREATE TYPE notification_status AS ENUM ('unread', 'read', 'archived');

-- Payment providers table
CREATE TABLE payment_providers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL, -- 'clover', 'stripe', 'square'
    display_name VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    api_endpoint VARCHAR(255),
    webhook_endpoint VARCHAR(255),
    config_json JSONB, -- Store provider-specific configuration
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- User payment methods
CREATE TABLE user_payment_methods (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    user_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    provider_id INTEGER NOT NULL REFERENCES payment_providers(id),
    external_id VARCHAR(255) NOT NULL, -- Provider's payment method ID
    type VARCHAR(50) NOT NULL, -- 'card', 'bank_account', 'digital_wallet'
    last_four VARCHAR(4),
    brand VARCHAR(50), -- 'visa', 'mastercard', etc.
    is_default BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    expires_at DATE,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Settlement batches for payment reconciliation
CREATE TABLE settlement_batches (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    provider_id INTEGER NOT NULL REFERENCES payment_providers(id),
    batch_date DATE NOT NULL,
    batch_reference VARCHAR(255), -- Provider's batch ID
    total_amount DECIMAL(12, 2) NOT NULL,
    transaction_count INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'processed', 'failed'
    processed_at TIMESTAMP WITH TIME ZONE,
    reconciled_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Link transactions to settlement batches
ALTER TABLE transactions 
ADD COLUMN settlement_batch_id INTEGER REFERENCES settlement_batches(id);

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
    action_url VARCHAR(255), -- URL for user to take action
    scheduled_for TIMESTAMP WITH TIME ZONE DEFAULT NOW(), -- For delayed notifications
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
    UNIQUE(job_id, reviewer_id) -- One review per person per job
);

-- Apply updated_at triggers to new tables
CREATE TRIGGER update_payment_providers_updated_at BEFORE UPDATE ON payment_providers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_payment_methods_updated_at BEFORE UPDATE ON user_payment_methods FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_settlement_batches_updated_at BEFORE UPDATE ON settlement_batches FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_notifications_updated_at BEFORE UPDATE ON notifications FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_notification_preferences_updated_at BEFORE UPDATE ON notification_preferences FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_job_reviews_updated_at BEFORE UPDATE ON job_reviews FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default payment providers
INSERT INTO payment_providers (name, display_name, is_active) VALUES 
('clover', 'Clover', true),
('stripe', 'Stripe', false),
('square', 'Square', false);

-- Create default notification preferences for existing users
INSERT INTO notification_preferences (user_id, type, email_enabled, push_enabled, sms_enabled)
SELECT 
    p.id,
    nt.type,
    true, -- email enabled by default
    true, -- push enabled by default  
    false -- SMS disabled by default
FROM people p
CROSS JOIN (
    SELECT unnest(enum_range(NULL::notification_type)) as type
) nt;