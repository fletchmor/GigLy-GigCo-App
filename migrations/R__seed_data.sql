-- Clear existing test data (only for development)
-- DO NOT run this in production!
DO $$
BEGIN
    IF current_setting('app.environment', true) = 'development' THEN
        -- Clean up test data but preserve admin and migrated customers
        DELETE FROM job_reviews WHERE reviewer_id > 4;
        DELETE FROM notifications WHERE user_id > 4;
        DELETE FROM notification_preferences WHERE user_id > 4;
        DELETE FROM schedules WHERE gig_worker_id > 4;
        DELETE FROM transactions WHERE consumer_id > 4 OR gig_worker_id > 4;
        DELETE FROM jobs WHERE consumer_id > 4 OR gig_worker_id > 4;
        DELETE FROM user_payment_methods WHERE user_id > 4;
        DELETE FROM people WHERE id > 4; -- Keep admin (id=1) and migrated customers (ids 2-4)
    END IF;
END $$;

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