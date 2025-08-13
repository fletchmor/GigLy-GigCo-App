-- Simple additional seed data for testing

-- Add some sample jobs with proper enum casting
INSERT INTO jobs (
    consumer_id, title, description, category, location_address,
    pay_rate_per_hour, estimated_duration_hours, total_pay, status, notes
) VALUES 
    (1, 'House Deep Cleaning', 'Complete deep cleaning of 3-bedroom house including bathrooms, kitchen, and living areas', 'cleaning', '123 Test Street, Test City, TC 12345', 22.00, 4.0, 88.00, 'posted'::job_status, 'Please bring your own supplies'),
    (2, 'Lawn Mowing Service', 'Weekly lawn mowing and edge trimming for medium-sized yard', 'maintenance', '456 Sample Avenue, Sample Town, ST 67890', 25.00, 2.0, 50.00, 'posted'::job_status, 'Equipment will be provided'),
    (1, 'Pet Sitting Weekend', 'Weekend pet sitting for two friendly dogs', 'pet_care', '789 Demo Drive, Demo City, DC 11111', 20.00, 8.0, 160.00, 'accepted'::job_status, 'Dogs are very friendly, instructions provided'),
    (2, 'Computer Setup Help', 'Help setting up new laptop and transferring files', 'tech_support', '321 Mock Road, Mock County, MC 22222', 30.00, 2.0, 60.00, 'posted'::job_status, 'Remote assistance available if needed'),
    (1, 'Grocery Shopping', 'Weekly grocery shopping and delivery', 'delivery', '654 Example Lane, Example City, EC 33333', 18.00, 1.5, 27.00, 'completed'::job_status, 'Shopping list will be provided')
ON CONFLICT DO NOTHING;

-- Add schedules for existing gig workers (IDs 3, 4, 5 from people table)
INSERT INTO schedules (
    gig_worker_id, title, start_time, end_time, is_available, notes
) VALUES 
    (3, 'Morning Cleaning Availability', NOW() + interval '1 day 9 hours', NOW() + interval '1 day 13 hours', true, 'Available for house cleaning'),
    (4, 'Evening Tech Support', NOW() + interval '2 days 17 hours', NOW() + interval '2 days 21 hours', true, 'Remote and on-site support'),
    (5, 'Weekend Pet Care', NOW() + interval '3 days 8 hours', NOW() + interval '3 days 18 hours', true, 'Pet sitting and walking services'),
    (3, 'Afternoon Availability', NOW() + interval '7 days 13 hours', NOW() + interval '7 days 17 hours', false, 'Not available - personal appointment')
ON CONFLICT DO NOTHING;

-- Add some transactions with proper enum casting
INSERT INTO transactions (
    job_id, consumer_id, gig_worker_id, amount, currency, status,
    payment_method, notes
) 
SELECT 
    j.id,
    j.consumer_id,
    COALESCE(j.gig_worker_id, 3), -- Default to gig worker ID 3 if not assigned
    j.total_pay,
    'USD',
    CASE 
        WHEN j.status = 'completed'::job_status THEN 'completed'::transaction_status
        WHEN j.status = 'accepted'::job_status THEN 'pending'::transaction_status
        ELSE 'pending'::transaction_status
    END,
    'credit_card',
    'Test transaction for job: ' || j.title
FROM jobs j 
WHERE j.total_pay IS NOT NULL
LIMIT 5
ON CONFLICT DO NOTHING;

-- Show results
SELECT 'Jobs created' as type, count(*) as count FROM jobs
UNION ALL
SELECT 'Schedules created', count(*) FROM schedules
UNION ALL
SELECT 'Transactions created', count(*) FROM transactions
UNION ALL
SELECT 'GigWorkers available', count(*) FROM gigworkers WHERE is_active = true
UNION ALL
SELECT 'Customers available', count(*) FROM customers;