-- Clover Payment Integration Schema
-- Run this after init.sql to add payment processing capabilities

-- ==============================================
-- PAYMENT ENUMS
-- ==============================================

-- Payment transaction type enum
CREATE TYPE IF NOT EXISTS payment_transaction_type AS ENUM (
    'authorization',     -- Initial pre-auth to hold funds
    'capture',          -- Capture previously authorized funds
    'charge',           -- Direct charge (no pre-auth)
    'refund',           -- Refund transaction
    'void',             -- Void transaction (within 25 min window)
    'adjustment'        -- Payment adjustment
);

-- Payment split type enum
CREATE TYPE IF NOT EXISTS payment_split_type AS ENUM (
    'platform_fee',     -- Platform/service fee
    'worker_payment',   -- Payment to worker
    'tax',             -- Tax amount
    'tip',             -- Tip/gratuity
    'other'            -- Other splits
);

-- ==============================================
-- ENHANCED PAYMENT TABLES
-- ==============================================

-- Drop and recreate transactions table with Clover support
-- IMPORTANT: Back up data before running this in production!
ALTER TABLE IF EXISTS transactions RENAME TO transactions_backup;

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    job_id INTEGER NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    consumer_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    gig_worker_id INTEGER REFERENCES people(id) ON DELETE SET NULL,

    -- Amount tracking
    amount DECIMAL(10, 2) NOT NULL,                    -- Total transaction amount
    currency VARCHAR(3) DEFAULT 'USD',

    -- Transaction status and type
    status transaction_status DEFAULT 'pending',
    transaction_type payment_transaction_type NOT NULL DEFAULT 'charge',

    -- Clover-specific fields
    clover_charge_id VARCHAR(255),                     -- Clover charge ID (clv_xxx)
    clover_payment_id VARCHAR(255),                    -- Clover payment ID for capture
    clover_source_token VARCHAR(255),                  -- Tokenized card (clv_xxx)
    clover_refund_id VARCHAR(255),                     -- Refund ID if refunded
    clover_order_id VARCHAR(255),                      -- Clover order ID if applicable

    -- Authorization and capture tracking
    authorized_at TIMESTAMP WITH TIME ZONE,            -- When funds were authorized
    authorization_expires_at TIMESTAMP WITH TIME ZONE, -- When auth expires (card-dependent)
    captured_at TIMESTAMP WITH TIME ZONE,              -- When funds were captured
    capture_amount DECIMAL(10, 2),                     -- Amount captured (may differ from auth)

    -- Payment metadata
    payment_method_id INTEGER REFERENCES user_payment_methods(id),
    payment_method VARCHAR(50),                        -- Card type: visa, mastercard, etc
    last_four VARCHAR(4),                              -- Last 4 digits of card

    -- Fees and splits
    processing_fee DECIMAL(10, 2) DEFAULT 0.00,        -- Payment processor fee
    platform_fee DECIMAL(10, 2) DEFAULT 0.00,          -- Platform service fee
    net_amount DECIMAL(10, 2),                         -- Amount after all fees

    -- Escrow tracking
    escrow_held_at TIMESTAMP WITH TIME ZONE,           -- When put in escrow
    escrow_released_at TIMESTAMP WITH TIME ZONE,       -- When released from escrow

    -- Refund tracking
    refunded_at TIMESTAMP WITH TIME ZONE,
    refund_amount DECIMAL(10, 2),
    refund_reason TEXT,

    -- Reconciliation
    settlement_batch_id INTEGER REFERENCES settlement_batches(id),
    reconciled_at TIMESTAMP WITH TIME ZONE,

    -- Parent transaction reference (for refunds, captures, etc)
    parent_transaction_id INTEGER REFERENCES transactions(id),

    -- Additional metadata
    metadata JSONB,                                    -- Store additional Clover response data
    notes TEXT,
    failure_reason TEXT,                               -- If transaction failed

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Payment splits table for tracking fee distribution
CREATE TABLE IF NOT EXISTS payment_splits (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    transaction_id INTEGER NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,

    split_type payment_split_type NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    percentage DECIMAL(5, 2),                          -- If calculated as percentage

    -- Recipient tracking (where applicable)
    recipient_id INTEGER REFERENCES people(id),        -- For worker payments

    description TEXT,
    metadata JSONB,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Payment events log for audit trail
CREATE TABLE IF NOT EXISTS payment_events (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4() UNIQUE NOT NULL,
    transaction_id INTEGER NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,

    event_type VARCHAR(50) NOT NULL,                   -- authorize, capture, refund, etc
    event_status VARCHAR(50) NOT NULL,                 -- success, failed, pending

    -- Clover API response data
    clover_response JSONB,
    error_message TEXT,
    error_code VARCHAR(50),

    -- Request tracking
    idempotency_key VARCHAR(255),                      -- For retry safety
    user_id INTEGER REFERENCES people(id),             -- Who initiated the event

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Update user_payment_methods table to add Clover token
ALTER TABLE user_payment_methods
ADD COLUMN IF NOT EXISTS clover_token VARCHAR(255),              -- Clover card token
ADD COLUMN IF NOT EXISTS clover_customer_id VARCHAR(255),        -- Clover customer ID
ADD COLUMN IF NOT EXISTS fingerprint VARCHAR(100);               -- Card fingerprint for duplicate detection

-- ==============================================
-- INDEXES
-- ==============================================

-- Transactions indexes
CREATE INDEX IF NOT EXISTS idx_transactions_job_id ON transactions(job_id);
CREATE INDEX IF NOT EXISTS idx_transactions_consumer_id ON transactions(consumer_id);
CREATE INDEX IF NOT EXISTS idx_transactions_gig_worker_id ON transactions(gig_worker_id) WHERE gig_worker_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions(transaction_type);
CREATE INDEX IF NOT EXISTS idx_transactions_clover_charge_id ON transactions(clover_charge_id) WHERE clover_charge_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_clover_payment_id ON transactions(clover_payment_id) WHERE clover_payment_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_parent ON transactions(parent_transaction_id) WHERE parent_transaction_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_authorized_at ON transactions(authorized_at) WHERE authorized_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_escrow ON transactions(escrow_held_at, escrow_released_at) WHERE escrow_held_at IS NOT NULL;

-- Payment splits indexes
CREATE INDEX IF NOT EXISTS idx_payment_splits_transaction_id ON payment_splits(transaction_id);
CREATE INDEX IF NOT EXISTS idx_payment_splits_type ON payment_splits(split_type);
CREATE INDEX IF NOT EXISTS idx_payment_splits_recipient_id ON payment_splits(recipient_id) WHERE recipient_id IS NOT NULL;

-- Payment events indexes
CREATE INDEX IF NOT EXISTS idx_payment_events_transaction_id ON payment_events(transaction_id);
CREATE INDEX IF NOT EXISTS idx_payment_events_type ON payment_events(event_type);
CREATE INDEX IF NOT EXISTS idx_payment_events_status ON payment_events(event_status);
CREATE INDEX IF NOT EXISTS idx_payment_events_created_at ON payment_events(created_at);
CREATE INDEX IF NOT EXISTS idx_payment_events_idempotency ON payment_events(idempotency_key) WHERE idempotency_key IS NOT NULL;

-- User payment methods indexes
CREATE INDEX IF NOT EXISTS idx_user_payment_methods_clover_token ON user_payment_methods(clover_token) WHERE clover_token IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_user_payment_methods_fingerprint ON user_payment_methods(fingerprint) WHERE fingerprint IS NOT NULL;

-- ==============================================
-- CONSTRAINTS
-- ==============================================

-- Ensure amounts are positive
ALTER TABLE transactions
ADD CONSTRAINT IF NOT EXISTS chk_transactions_positive_amount
CHECK (amount > 0);

ALTER TABLE transactions
ADD CONSTRAINT IF NOT EXISTS chk_transactions_positive_capture
CHECK (capture_amount IS NULL OR capture_amount > 0);

ALTER TABLE transactions
ADD CONSTRAINT IF NOT EXISTS chk_transactions_positive_refund
CHECK (refund_amount IS NULL OR refund_amount > 0);

ALTER TABLE payment_splits
ADD CONSTRAINT IF NOT EXISTS chk_payment_splits_positive_amount
CHECK (amount > 0);

-- Ensure capture amount doesn't exceed authorization
ALTER TABLE transactions
ADD CONSTRAINT IF NOT EXISTS chk_transactions_capture_within_auth
CHECK (capture_amount IS NULL OR authorization_expires_at IS NULL OR capture_amount <= amount);

-- Ensure percentages are valid
ALTER TABLE payment_splits
ADD CONSTRAINT IF NOT EXISTS chk_payment_splits_valid_percentage
CHECK (percentage IS NULL OR (percentage >= 0 AND percentage <= 100));

-- ==============================================
-- TRIGGERS
-- ==============================================

-- Create updated_at trigger for new tables
CREATE TRIGGER update_transactions_updated_at
BEFORE UPDATE ON transactions
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payment_splits_updated_at
BEFORE UPDATE ON payment_splits
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger to automatically create payment splits when transaction is created
CREATE OR REPLACE FUNCTION create_payment_splits()
RETURNS TRIGGER AS $$
BEGIN
    -- Only create splits for captures or direct charges
    IF NEW.transaction_type IN ('capture', 'charge') AND NEW.status = 'completed' THEN
        -- Platform fee split (example: 10% platform fee)
        INSERT INTO payment_splits (transaction_id, split_type, amount, percentage, description)
        VALUES (
            NEW.id,
            'platform_fee',
            NEW.platform_fee,
            (NEW.platform_fee / NULLIF(NEW.amount, 0)) * 100,
            'Platform service fee'
        );

        -- Worker payment split
        IF NEW.gig_worker_id IS NOT NULL THEN
            INSERT INTO payment_splits (transaction_id, split_type, amount, recipient_id, description)
            VALUES (
                NEW.id,
                'worker_payment',
                NEW.net_amount,
                NEW.gig_worker_id,
                'Worker payment'
            );
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_create_payment_splits
AFTER INSERT ON transactions
FOR EACH ROW EXECUTE FUNCTION create_payment_splits();

-- ==============================================
-- HELPER FUNCTIONS
-- ==============================================

-- Function to calculate platform fee (10% default)
CREATE OR REPLACE FUNCTION calculate_platform_fee(
    amount DECIMAL,
    fee_percentage DECIMAL DEFAULT 10.0
)
RETURNS DECIMAL AS $$
BEGIN
    RETURN ROUND((amount * fee_percentage / 100.0)::NUMERIC, 2);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function to calculate processing fee (Clover typical: 2.6% + $0.10)
CREATE OR REPLACE FUNCTION calculate_processing_fee(
    amount DECIMAL,
    percentage DECIMAL DEFAULT 2.6,
    fixed_fee DECIMAL DEFAULT 0.10
)
RETURNS DECIMAL AS $$
BEGIN
    RETURN ROUND(((amount * percentage / 100.0) + fixed_fee)::NUMERIC, 2);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function to get transaction summary for a job
CREATE OR REPLACE FUNCTION get_job_payment_summary(job_id_param INTEGER)
RETURNS TABLE(
    total_authorized DECIMAL,
    total_captured DECIMAL,
    total_refunded DECIMAL,
    platform_fees DECIMAL,
    worker_payment DECIMAL,
    escrow_status TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COALESCE(SUM(CASE WHEN t.transaction_type = 'authorization' THEN t.amount ELSE 0 END), 0) as total_authorized,
        COALESCE(SUM(CASE WHEN t.transaction_type IN ('capture', 'charge') THEN t.capture_amount ELSE 0 END), 0) as total_captured,
        COALESCE(SUM(CASE WHEN t.transaction_type = 'refund' THEN t.refund_amount ELSE 0 END), 0) as total_refunded,
        COALESCE(SUM(t.platform_fee), 0) as platform_fees,
        COALESCE(SUM(t.net_amount), 0) as worker_payment,
        CASE
            WHEN MAX(t.escrow_held_at) IS NOT NULL AND MAX(t.escrow_released_at) IS NULL THEN 'held'
            WHEN MAX(t.escrow_released_at) IS NOT NULL THEN 'released'
            ELSE 'none'
        END as escrow_status
    FROM transactions t
    WHERE t.job_id = job_id_param;
END;
$$ LANGUAGE plpgsql;

-- ==============================================
-- SEED DATA
-- ==============================================

-- Ensure Clover is configured in payment providers
UPDATE payment_providers
SET
    api_endpoint = 'https://scl-sandbox.dev.clover.com',
    webhook_endpoint = '/api/v1/webhooks/clover',
    config_json = jsonb_build_object(
        'tokenization_endpoint', 'https://token-sandbox.dev.clover.com/v1/tokens',
        'pakms_endpoint', 'https://scl-sandbox.dev.clover.com/pakms/apikey',
        'environment', 'sandbox'
    )
WHERE name = 'clover';

-- Show completion message
DO $$
BEGIN
    RAISE NOTICE 'Clover payment schema created successfully!';
    RAISE NOTICE 'New tables: payment_splits, payment_events';
    RAISE NOTICE 'Enhanced: transactions, user_payment_methods';
    RAISE NOTICE 'Helper functions created for fee calculations and payment summaries';
END $$;
