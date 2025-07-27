-- Flyway Migration: V3__add_indexes_and_constraints.sql
-- Description: Add comprehensive indexes and performance optimizations
-- Author: GigCo Development Team
-- Created: 2025-07-27

-- =====================================================
-- INDEXES FOR PEOPLE TABLE (Standard indexes - transactional)
-- =====================================================
CREATE INDEX IF NOT EXISTS idx_people_email ON people(email);
CREATE INDEX IF NOT EXISTS idx_people_uuid ON people(uuid);
CREATE INDEX IF NOT EXISTS idx_people_role ON people(role);
CREATE INDEX IF NOT EXISTS idx_people_active ON people(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_people_location ON people(latitude, longitude) WHERE latitude IS NOT NULL AND longitude IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_people_created_at ON people(created_at);
CREATE INDEX IF NOT EXISTS idx_people_phone ON people(phone) WHERE phone IS NOT NULL;

-- =====================================================
-- INDEXES FOR JOBS TABLE (Standard indexes - transactional)
-- =====================================================
CREATE INDEX IF NOT EXISTS idx_jobs_uuid ON jobs(uuid);
CREATE INDEX IF NOT EXISTS idx_jobs_consumer_id ON jobs(consumer_id);
CREATE INDEX IF NOT EXISTS idx_jobs_gig_worker_id ON jobs(gig_worker_id) WHERE gig_worker_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_category ON jobs(category) WHERE category IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_jobs_location ON jobs(location_latitude, location_longitude) WHERE location_latitude IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_jobs_scheduled_start ON jobs(scheduled_start) WHERE scheduled_start IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at);
CREATE INDEX IF NOT EXISTS idx_jobs_pay_rate ON jobs(pay_rate_per_hour) WHERE pay_rate_per_hour IS NOT NULL;

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_jobs_status_created ON jobs(status, created_at);
CREATE INDEX IF NOT EXISTS idx_jobs_consumer_status ON jobs(consumer_id, status);
CREATE INDEX IF NOT EXISTS idx_jobs_worker_status ON jobs(gig_worker_id, status) WHERE gig_worker_id IS NOT NULL;

-- =====================================================
-- INDEXES FOR TRANSACTIONS TABLE (Standard indexes - transactional)
-- =====================================================
CREATE INDEX IF NOT EXISTS idx_transactions_uuid ON transactions(uuid);
CREATE INDEX IF NOT EXISTS idx_transactions_job_id ON transactions(job_id);
CREATE INDEX IF NOT EXISTS idx_transactions_consumer_id ON transactions(consumer_id);
CREATE INDEX IF NOT EXISTS idx_transactions_gig_worker_id ON transactions(gig_worker_id);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_payment_intent ON transactions(payment_intent_id) WHERE payment_intent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_payment_method ON transactions(payment_method) WHERE payment_method IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_transactions_settlement_batch ON transactions(settlement_batch_id) WHERE settlement_batch_id IS NOT NULL;

-- Composite indexes for financial queries
CREATE INDEX IF NOT EXISTS idx_transactions_status_created ON transactions(status, created_at);
CREATE INDEX IF NOT EXISTS idx_transactions_consumer_status ON transactions(consumer_id, status);
CREATE INDEX IF NOT EXISTS idx_transactions_worker_status ON transactions(gig_worker_id, status);

-- =====================================================
-- INDEXES FOR SCHEDULES TABLE (Standard indexes - transactional)
-- =====================================================
CREATE INDEX IF NOT EXISTS idx_schedules_uuid ON schedules(uuid);
CREATE INDEX IF NOT EXISTS idx_schedules_gig_worker_id ON schedules(gig_worker_id);
CREATE INDEX IF NOT EXISTS idx_schedules_time_range ON schedules(start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_schedules_availability ON schedules(is_available);
CREATE INDEX IF NOT EXISTS idx_schedules_job_id ON schedules(job_id) WHERE job_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_schedules_start_time ON schedules(start_time);
CREATE INDEX IF NOT EXISTS idx_schedules_recurring ON schedules(recurring_pattern) WHERE recurring_pattern IS NOT NULL;

-- Composite indexes for availability queries
CREATE INDEX IF NOT EXISTS idx_schedules_worker_availability ON schedules(gig_worker_id, is_available, start_time);
CREATE INDEX IF NOT EXISTS idx_schedules_worker_timerange ON schedules(gig_worker_id, start_time, end_time);

-- =====================================================
-- INDEXES FOR PAYMENT PROVIDERS TABLE (Standard indexes - transactional)
-- =====================================================
CREATE INDEX IF NOT EXISTS idx_payment_providers_name ON payment_providers(name);
CREATE INDEX IF NOT EXISTS idx_payment_providers_active ON payment_providers(is_active) WHERE is_active = true;

-- =====================================================
-- INDEXES FOR USER PAYMENT METHODS TABLE (Standard indexes - transactional)
-- =====================================================
CREATE INDEX IF NOT EXISTS idx_user_payment_methods_uuid ON user_payment_methods(uuid);
CREATE INDEX IF NOT EXISTS idx_user_payment_methods_user_id ON user_payment_methods(user_id);
CREATE INDEX IF NOT EXISTS idx_user_payment_methods_provider_id ON user_payment_methods(provider_id);
CREATE INDEX IF NOT EXISTS idx_user_payment_methods_external_id ON user_payment_methods(external_id);
CREATE INDEX IF NOT EXISTS idx_user_payment_methods_default ON user_payment_methods(user_id, is_default) WHERE is_default = true;
CREATE INDEX IF NOT EXISTS idx_user_payment_methods_active ON user_payment_methods(is_active) WHERE is_active = true;

-- =====================================================
-- INDEXES FOR SETTLEMENT BATCHES TABLE (Standard indexes - transactional)
-- =====================================================
CREATE INDEX IF NOT EXISTS idx_settlement_batches_uuid ON settlement_batches(uuid);
CREATE INDEX IF NOT EXISTS idx_settlement_batches_provider_id ON settlement_batches(provider_id);
CREATE INDEX IF NOT EXISTS idx_settlement_batches_date ON settlement_batches(batch_date);
CREATE INDEX IF NOT EXISTS idx_settlement_batches_status ON settlement_batches(status);
CREATE INDEX IF NOT EXISTS idx_settlement_batches_reference ON settlement_batches(batch_reference) WHERE batch_reference IS NOT NULL;

-- =====================================================
-- INDEXES FOR NOTIFICATIONS TABLE (Standard indexes - transactional)
-- =====================================================
CREATE INDEX IF NOT EXISTS idx_notifications_uuid ON notifications(uuid);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
CREATE INDEX IF NOT EXISTS idx_notifications_scheduled ON notifications(scheduled_for);
CREATE INDEX IF NOT EXISTS idx_notifications_job_id ON notifications(related_job_id) WHERE related_job_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_notifications_transaction_id ON notifications(related_transaction_id) WHERE related_transaction_id IS NOT NULL;

-- Composite indexes for notification queries
CREATE INDEX IF NOT EXISTS idx_notifications_user_status ON notifications(user_id, status);
CREATE INDEX IF NOT EXISTS idx_notifications_user_scheduled ON notifications(user_id, scheduled_for) WHERE status = 'unread';

-- =====================================================
-- INDEXES FOR NOTIFICATION PREFERENCES TABLE (Standard indexes - transactional)
-- =====================================================
CREATE INDEX IF NOT EXISTS idx_notification_preferences_user_id ON notification_preferences(user_id);
CREATE INDEX IF NOT EXISTS idx_notification_preferences_type ON notification_preferences(type);

-- =====================================================
-- INDEXES FOR JOB REVIEWS TABLE (Standard indexes - transactional)
-- =====================================================
CREATE INDEX IF NOT EXISTS idx_job_reviews_uuid ON job_reviews(uuid);
CREATE INDEX IF NOT EXISTS idx_job_reviews_job_id ON job_reviews(job_id);
CREATE INDEX IF NOT EXISTS idx_job_reviews_reviewer_id ON job_reviews(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_job_reviews_reviewee_id ON job_reviews(reviewee_id);
CREATE INDEX IF NOT EXISTS idx_job_reviews_rating ON job_reviews(rating);
CREATE INDEX IF NOT EXISTS idx_job_reviews_public ON job_reviews(is_public) WHERE is_public = true;

-- Composite indexes for review queries
CREATE INDEX IF NOT EXISTS idx_job_reviews_reviewee_rating ON job_reviews(reviewee_id, rating) WHERE is_public = true;

-- =====================================================
-- ADDITIONAL CONSTRAINTS AND VALIDATIONS
-- =====================================================

-- Ensure payment methods have valid expiration dates
ALTER TABLE user_payment_methods 
ADD CONSTRAINT chk_user_payment_methods_expiry 
CHECK (expires_at IS NULL OR expires_at > CURRENT_DATE);

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

ALTER TABLE settlement_batches
ADD CONSTRAINT chk_settlement_batches_positive_amount 
CHECK (total_amount >= 0);

-- Ensure only one default payment method per user
CREATE UNIQUE INDEX idx_user_payment_methods_one_default 
ON user_payment_methods(user_id) 
WHERE is_default = true;

-- =====================================================
-- PERFORMANCE ANALYSIS FUNCTIONS
-- =====================================================

-- Function to analyze query performance
CREATE OR REPLACE FUNCTION analyze_table_stats(table_name TEXT)
RETURNS TABLE(
    schemaname TEXT,
    tablename TEXT,
    n_tup_ins BIGINT,
    n_tup_upd BIGINT,
    n_tup_del BIGINT,
    n_live_tup BIGINT,
    n_dead_tup BIGINT,
    last_vacuum TIMESTAMP WITH TIME ZONE,
    last_autovacuum TIMESTAMP WITH TIME ZONE,
    last_analyze TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        s.schemaname::TEXT,
        s.tablename::TEXT,
        s.n_tup_ins,
        s.n_tup_upd,
        s.n_tup_del,
        s.n_live_tup,
        s.n_dead_tup,
        s.last_vacuum,
        s.last_autovacuum,
        s.last_analyze
    FROM pg_stat_user_tables s
    WHERE s.tablename = analyze_table_stats.table_name;
END;
$$ LANGUAGE plpgsql;