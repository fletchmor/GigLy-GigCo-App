-- Fix critical seed data issues for Postman tests

-- Ensure we have customers with IDs 1 and 2 (tests expect these)
INSERT INTO customers (id, name, address) VALUES 
    (1, 'Test Customer One', '123 Test Street, Test City, TC 12345'),
    (2, 'Test Customer Two', '456 Sample Avenue, Sample Town, ST 67890')
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    address = EXCLUDED.address;

-- Ensure we have people records with IDs 1 and 2 (for consumer roles)
INSERT INTO people (id, name, email, address, role, is_active, email_verified) VALUES 
    (1, 'Test Consumer One', 'test.consumer1@gigco.dev', '123 Test Street, Test City, TC 12345', 'consumer', true, true),
    (2, 'Test Consumer Two', 'test.consumer2@gigco.dev', '456 Sample Avenue, Sample Town, ST 67890', 'consumer', true, true),
    (3, 'Test GigWorker One', 'test.worker1@gigco.dev', '789 Worker Lane, Worker City, WC 11111', 'gig_worker', true, true),
    (4, 'Test GigWorker Two', 'test.worker2@gigco.dev', '321 Helper Road, Helper Town, HT 22222', 'gig_worker', true, true),
    (5, 'Test GigWorker Three', 'test.worker3@gigco.dev', '654 Service Street, Service City, SC 33333', 'gig_worker', true, true)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    email = EXCLUDED.email,
    address = EXCLUDED.address,
    role = EXCLUDED.role,
    is_active = EXCLUDED.is_active,
    email_verified = EXCLUDED.email_verified;

-- Ensure gigworker ID 1 exists (tests expect this)
INSERT INTO gigworkers (id, name, email, address, role, is_active, email_verified, verification_status) VALUES 
    (1, 'Test GigWorker One', 'test.gigworker1@gigco.dev', '789 Worker Lane, Worker City, WC 11111', 'gig_worker', true, true, 'verified')
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    email = EXCLUDED.email,
    address = EXCLUDED.address,
    is_active = EXCLUDED.is_active,
    verification_status = EXCLUDED.verification_status;

-- Ensure we have jobs with IDs 1, 3, 4, 39 (tests expect these)
INSERT INTO jobs (id, consumer_id, title, description, category, status, pay_rate_per_hour, estimated_duration_hours, total_pay) VALUES 
    (1, 1, 'Test Job One', 'Basic test job for API testing', 'cleaning', 'posted', 20.00, 2.0, 40.00),
    (3, 1, 'Test Job Three', 'Third test job for acceptance testing', 'maintenance', 'posted', 25.00, 3.0, 75.00),
    (4, 2, 'Test Job Four', 'Fourth test job for workflow testing', 'delivery', 'posted', 15.00, 1.0, 15.00),
    (39, 1, 'Test Job Thirty-Nine', 'Job for workflow endpoint testing', 'tech_support', 'posted', 30.00, 2.0, 60.00)
ON CONFLICT (id) DO UPDATE SET
    consumer_id = EXCLUDED.consumer_id,
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    status = EXCLUDED.status,
    pay_rate_per_hour = EXCLUDED.pay_rate_per_hour,
    estimated_duration_hours = EXCLUDED.estimated_duration_hours,
    total_pay = EXCLUDED.total_pay;

-- Reset sequences to ensure proper auto-increment
SELECT setval('customers_id_seq', (SELECT MAX(id) FROM customers));
SELECT setval('people_id_seq', (SELECT MAX(id) FROM people));  
SELECT setval('gigworkers_id_seq', (SELECT MAX(id) FROM gigworkers));
SELECT setval('jobs_id_seq', (SELECT MAX(id) FROM jobs));

-- Show what we have
SELECT 'Data Summary' as section, 'customers' as table_name, count(*) as count FROM customers
UNION ALL
SELECT 'Data Summary', 'people', count(*) FROM people  
UNION ALL  
SELECT 'Data Summary', 'gigworkers', count(*) FROM gigworkers
UNION ALL
SELECT 'Data Summary', 'jobs', count(*) FROM jobs;