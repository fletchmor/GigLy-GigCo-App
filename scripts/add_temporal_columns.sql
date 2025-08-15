-- Add Temporal workflow tracking columns to jobs table
ALTER TABLE jobs 
ADD COLUMN temporal_workflow_id VARCHAR(255),
ADD COLUMN temporal_run_id VARCHAR(255);

-- Add indexes for workflow lookups
CREATE INDEX idx_jobs_temporal_workflow 
ON jobs(temporal_workflow_id) 
WHERE temporal_workflow_id IS NOT NULL;

-- Add additional job status values for workflow states
-- Note: This assumes your jobs table already has a status column
-- You may need to modify this based on your current schema

-- Update existing 'posted' jobs to 'draft' status for workflow compatibility
-- (Only if you want to change the initial state)
-- UPDATE jobs SET status = 'draft' WHERE status = 'posted';

-- Add some additional useful columns for workflow tracking
ALTER TABLE jobs 
ADD COLUMN workflow_started_at TIMESTAMP,
ADD COLUMN workflow_completed_at TIMESTAMP;

-- Add skills and urgency columns if they don't exist
-- (These are used by the pricing activity)
DO $$ 
BEGIN
    -- Add skills column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'jobs' AND column_name = 'skills') THEN
        ALTER TABLE jobs ADD COLUMN skills TEXT;
    END IF;
    
    -- Add urgency column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'jobs' AND column_name = 'urgency') THEN
        ALTER TABLE jobs ADD COLUMN urgency VARCHAR(50) DEFAULT 'medium';
    END IF;
    
    -- Add duration column if it doesn't exist (in hours)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'jobs' AND column_name = 'duration') THEN
        ALTER TABLE jobs ADD COLUMN duration INTEGER DEFAULT 1;
    END IF;
END $$;

-- Ensure gigworkers table has the necessary columns for matching
DO $$ 
BEGIN
    -- Add rating column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'gigworkers' AND column_name = 'rating') THEN
        ALTER TABLE gigworkers ADD COLUMN rating DECIMAL(3,2) DEFAULT 5.0;
    END IF;
    
    -- Add skills column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'gigworkers' AND column_name = 'skills') THEN
        ALTER TABLE gigworkers ADD COLUMN skills TEXT;
    END IF;
    
    -- Add location column for matching if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'gigworkers' AND column_name = 'location') THEN
        ALTER TABLE gigworkers ADD COLUMN location VARCHAR(255);
    END IF;
END $$;

-- Create a simple reviews table for storing review data
-- (This is referenced in the SubmitReview endpoint)
CREATE TABLE IF NOT EXISTS reviews (
    id SERIAL PRIMARY KEY,
    job_id INTEGER NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    reviewer_id INTEGER NOT NULL,
    reviewee_id INTEGER NOT NULL,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index on reviews for faster job lookups
CREATE INDEX IF NOT EXISTS idx_reviews_job_id ON reviews(job_id);
CREATE INDEX IF NOT EXISTS idx_reviews_reviewer ON reviews(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_reviews_reviewee ON reviews(reviewee_id);

-- Update gigworkers table to have the address copied to location column
UPDATE gigworkers 
SET location = address 
WHERE location IS NULL AND address IS NOT NULL;

-- Set default values for existing jobs
UPDATE jobs 
SET urgency = 'medium' 
WHERE urgency IS NULL;

UPDATE jobs 
SET duration = 1 
WHERE duration IS NULL;

-- Sample data updates for testing (optional)
-- You can uncomment these if you want some test data

-- UPDATE jobs SET skills = 'cleaning,organizing' WHERE category = 'cleaning';
-- UPDATE jobs SET skills = 'gardening,landscaping' WHERE category = 'gardening';
-- UPDATE jobs SET urgency = 'high' WHERE total_pay > 100;

-- UPDATE gigworkers SET skills = 'cleaning,organizing,pet-care' WHERE id = 1;
-- UPDATE gigworkers SET skills = 'gardening,landscaping,handyman' WHERE id = 2;