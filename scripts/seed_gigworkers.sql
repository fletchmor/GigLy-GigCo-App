-- Additional seed data for GigWorkers table to support Postman testing
-- This adds more test data specifically for the /api/v1/gigworkers endpoints

-- Insert additional sample gig workers
INSERT INTO gigworkers (
    name, email, phone, address, latitude, longitude, 
    role, is_active, email_verified, phone_verified, bio, 
    hourly_rate, experience_years, verification_status, 
    service_radius_miles, availability_notes,
    emergency_contact_name, emergency_contact_phone, emergency_contact_relationship
) VALUES 
    ('Alice Thompson', 'alice.thompson@example.com', '+1-555-1001', '123 Maple Street, Springfield, IL 62701', 
     39.7817, -89.6501, 'gig_worker', true, true, true, 
     'Professional house cleaner with 7 years experience. Specializes in eco-friendly cleaning solutions.', 
     24.00, 7, 'verified', 20.0, 'Available weekdays 9 AM - 5 PM',
     'John Thompson', '+1-555-1002', 'Husband'),
     
    ('Michael Rodriguez', 'mike.rodriguez@example.com', '+1-555-1003', '456 Oak Avenue, Chicago, IL 60601', 
     41.8781, -87.6298, 'gig_worker', true, true, false, 
     'Experienced handyman and maintenance worker. Certified electrician with plumbing skills.', 
     32.00, 12, 'verified', 30.0, 'Available evenings and weekends',
     'Maria Rodriguez', '+1-555-1004', 'Wife',
     'Maria Rodriguez', '+1-555-1004', 'Wife'),
     
    ('Sarah Chen', 'sarah.chen@example.com', '+1-555-1005', '789 Pine Road, Austin, TX 78701', 
     30.2672, -97.7431, 'gig_worker', true, false, true, 
     'Pet care specialist and dog trainer. Certified in pet first aid and behavior modification.', 
     22.00, 4, 'pending', 25.0, 'Flexible schedule, specializes in large breeds',
     'David Chen', '+1-555-1006', 'Brother',
     'David Chen', '+1-555-1006', 'Brother'),
     
    ('James Wilson', 'james.wilson@example.com', '+1-555-1007', '321 Cedar Lane, Denver, CO 80201', 
     39.7392, -104.9903, 'gig_worker', false, true, true, 
     'Freelance tutor and academic support specialist. Masters in Education.', 
     28.00, 8, 'verified', 15.0, 'Currently unavailable due to full-time commitment',
     'Linda Wilson', '+1-555-1008', 'Mother',
     'Linda Wilson', '+1-555-1008', 'Mother'),
     
    ('Emma Johnson', 'emma.johnson@example.com', '+1-555-1009', '654 Birch Boulevard, Seattle, WA 98101', 
     47.6062, -122.3321, 'gig_worker', true, true, true, 
     'Tech support specialist and computer repair technician. CompTIA A+ certified.', 
     35.00, 6, 'verified', 40.0, 'Available for remote and on-site support',
     'Robert Johnson', '+1-555-1010', 'Father',
     'Robert Johnson', '+1-555-1010', 'Father'),
     
    ('Carlos Martinez', 'carlos.martinez@example.com', '+1-555-1011', '987 Elm Street, Phoenix, AZ 85001', 
     33.4484, -112.0740, 'gig_worker', true, true, true, 
     'Delivery driver and courier service. Clean driving record with commercial license.', 
     18.00, 3, 'verified', 50.0, 'Available 7 days a week, own vehicle',
     'Ana Martinez', '+1-555-1012', 'Sister',
     'Ana Martinez', '+1-555-1012', 'Sister'),
     
    ('Jessica Brown', 'jessica.brown@example.com', '+1-555-1013', '147 Willow Way, Portland, OR 97201', 
     45.5152, -122.6784, 'gig_worker', true, false, false, 
     'Personal assistant and errand runner. Excellent organizational skills and time management.', 
     20.00, 2, 'pending', 20.0, 'New to the platform, eager to build reputation',
     'Mark Brown', '+1-555-1014', 'Spouse',
     'Mark Brown', '+1-555-1014', 'Spouse'),
     
    ('Anthony Davis', 'anthony.davis@example.com', '+1-555-1015', '258 Spruce Street, Miami, FL 33101', 
     25.7617, -80.1918, 'gig_worker', true, true, true, 
     'Landscaping and yard maintenance professional. Licensed pesticide applicator.', 
     26.00, 10, 'verified', 35.0, 'Weather dependent availability',
     'Patricia Davis', '+1-555-1016', 'Wife',
     'Patricia Davis', '+1-555-1016', 'Wife'),
     
    ('Rachel Kim', 'rachel.kim@example.com', '+1-555-1017', '369 Chestnut Circle, Boston, MA 02101', 
     42.3601, -71.0589, 'gig_worker', true, true, true, 
     'Childcare provider and babysitter. CPR certified with early childhood education background.', 
     25.00, 5, 'verified', 15.0, 'Available for evening and weekend childcare',
     'Kevin Kim', '+1-555-1018', 'Husband',
     'Kevin Kim', '+1-555-1018', 'Husband'),
     
    ('Tyler Green', 'tyler.green@example.com', '+1-555-1019', '741 Poplar Place, Nashville, TN 37201', 
     36.1627, -86.7816, 'gig_worker', true, true, false, 
     'Event setup and breakdown specialist. Experience with weddings and corporate events.', 
     21.00, 4, 'pending', 25.0, 'Weekend availability preferred',
     'Ashley Green', '+1-555-1020', 'Partner',
     'Ashley Green', '+1-555-1020', 'Partner')

ON CONFLICT (email) DO UPDATE SET
    name = EXCLUDED.name,
    phone = EXCLUDED.phone,
    address = EXCLUDED.address,
    latitude = EXCLUDED.latitude,
    longitude = EXCLUDED.longitude,
    is_active = EXCLUDED.is_active,
    email_verified = EXCLUDED.email_verified,
    phone_verified = EXCLUDED.phone_verified,
    bio = EXCLUDED.bio,
    hourly_rate = EXCLUDED.hourly_rate,
    experience_years = EXCLUDED.experience_years,
    verification_status = EXCLUDED.verification_status,
    service_radius_miles = EXCLUDED.service_radius_miles,
    availability_notes = EXCLUDED.availability_notes,
    emergency_contact_name = EXCLUDED.emergency_contact_name,
    emergency_contact_phone = EXCLUDED.emergency_contact_phone,
    emergency_contact_relationship = EXCLUDED.emergency_contact_relationship,
    updated_at = NOW();

-- Insert additional customers for testing
INSERT INTO customers (name, address) VALUES 
    ('Test Customer Alpha', '100 Test Street, Test City, TS 12345'),
    ('Test Customer Beta', '200 Sample Avenue, Sample Town, ST 67890'),
    ('Test Customer Gamma', '300 Demo Drive, Demo Village, DV 11111'),
    ('Test Customer Delta', '400 Mock Road, Mock County, MC 22222'),
    ('Test Customer Echo', '500 Example Lane, Example City, EC 33333')
ON CONFLICT DO NOTHING;

-- Insert sample jobs for testing
INSERT INTO jobs (
    consumer_id, title, description, category, location_address,
    pay_rate_per_hour, estimated_duration_hours, total_pay, status,
    scheduled_start, scheduled_end, notes
)
SELECT 
    p.id,
    titles.title,
    descriptions.description,
    categories.category,
    addresses.address,
    rates.rate,
    durations.duration,
    rates.rate * durations.duration,
    statuses.status,
    starts.start_time,
    starts.start_time + (durations.duration || ' hours')::interval,
    notes.note
FROM people p
CROSS JOIN (VALUES 
    ('House Deep Cleaning', 'Complete deep cleaning of 3-bedroom house including bathrooms, kitchen, and living areas'),
    ('Lawn Mowing Service', 'Weekly lawn mowing and edge trimming for medium-sized yard'),
    ('Pet Sitting Overnight', 'Overnight pet sitting for two dogs while owners are traveling'),
    ('Computer Setup Help', 'Help setting up new laptop and transferring data from old computer'),
    ('Grocery Shopping', 'Weekly grocery shopping and delivery for elderly client')
) AS titles(title)
CROSS JOIN (VALUES 
    ('Thorough cleaning of all rooms, windows, and appliances. Bring own supplies.'),
    ('Mow grass, trim edges, and clean up clippings. Equipment provided.'),
    ('Feed pets, let out for walks, provide companionship. Instructions provided.'),
    ('Install software, transfer files, set up email and accounts. Remote support available.'),
    ('Shop from provided list, select quality items, deliver and put away groceries.')
) AS descriptions(description)
CROSS JOIN (VALUES ('cleaning'), ('maintenance'), ('pet_care'), ('tech_support'), ('delivery')) AS categories(category)
CROSS JOIN (VALUES 
    ('123 Residential Street, Suburb City, SC 12345'),
    ('456 Family Lane, Home Town, HT 67890'),
    ('789 Apartment Complex, Metro City, MC 11111')
) AS addresses(address)
CROSS JOIN (VALUES (20.00), (25.00), (18.00), (30.00), (15.00)) AS rates(rate)
CROSS JOIN (VALUES (3.0), (1.5), (8.0), (2.0), (1.0)) AS durations(duration)
CROSS JOIN (VALUES ('posted'::job_status), ('posted'::job_status), ('accepted'::job_status)) AS statuses(status)
CROSS JOIN (VALUES 
    (NOW() + interval '1 day'),
    (NOW() + interval '2 days'),
    (NOW() + interval '3 days')
) AS starts(start_time)
CROSS JOIN (VALUES 
    ('Please contact before arrival'),
    ('Key will be left under mat'),
    ('Ring doorbell, do not knock')
) AS notes(note)
WHERE p.role = 'consumer'
LIMIT 15
ON CONFLICT DO NOTHING;

-- Insert sample schedules for gig workers (using people table IDs)
INSERT INTO schedules (
    gig_worker_id, title, start_time, end_time, is_available, notes
)
SELECT 
    p.id,
    'Available for ' || categories.category,
    NOW() + (days.day || ' days')::interval + (hours.hour || ' hours')::interval,
    NOW() + (days.day || ' days')::interval + (hours.hour + 4 || ' hours')::interval,
    availability.available,
    notes.note
FROM people p
CROSS JOIN (VALUES ('cleaning'), ('maintenance'), ('delivery')) AS categories(category)
CROSS JOIN (VALUES (1), (2), (3), (7)) AS days(day)
CROSS JOIN (VALUES (9), (13), (17)) AS hours(hour)
CROSS JOIN (VALUES (true), (true), (false)) AS availability(available)
CROSS JOIN (VALUES 
    ('Flexible timing within window'),
    ('Prefer morning appointments'),
    ('Emergency availability only')
) AS notes(note)
WHERE p.role = 'gig_worker' AND p.is_active = true
LIMIT 5
ON CONFLICT DO NOTHING;

-- Create some sample transactions
INSERT INTO transactions (
    job_id, consumer_id, gig_worker_id, amount, currency, status,
    payment_method, notes
)
SELECT 
    j.id,
    j.consumer_id,
    COALESCE(j.gig_worker_id, (SELECT id FROM people WHERE role = 'gig_worker' AND is_active = true LIMIT 1)),
    j.total_pay,
    'USD',
    statuses.status,
    methods.method,
    notes.note
FROM jobs j
CROSS JOIN (VALUES ('completed'::transaction_status), ('pending'::transaction_status), ('completed'::transaction_status)) AS statuses(status)
CROSS JOIN (VALUES ('credit_card'), ('bank_transfer'), ('cash')) AS methods(method)
CROSS JOIN (VALUES 
    ('Payment for completed work'),
    ('Advance payment for service'),
    ('Final payment after completion')
) AS notes(note)
WHERE j.total_pay IS NOT NULL
LIMIT 10
ON CONFLICT DO NOTHING;

-- Display summary of inserted data
DO $$
BEGIN
    RAISE NOTICE 'Seed data insertion completed!';
    RAISE NOTICE 'GigWorkers in database: %', (SELECT count(*) FROM gigworkers);
    RAISE NOTICE 'Customers in database: %', (SELECT count(*) FROM customers);
    RAISE NOTICE 'Jobs in database: %', (SELECT count(*) FROM jobs);
    RAISE NOTICE 'Schedules in database: %', (SELECT count(*) FROM schedules);
    RAISE NOTICE 'Transactions in database: %', (SELECT count(*) FROM transactions);
END $$;