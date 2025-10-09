-- Migration script to update job_status enum and add temporal columns
-- Run this on existing databases to bring them up to date with the latest schema

-- Step 1: Add new status values to the job_status enum
-- PostgreSQL doesn't allow dropping/recreating enums with existing data,
-- so we add new values one at a time

DO $$
BEGIN
    -- Add new status values if they don't exist
    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'offer_sent' AND enumtypid = 'job_status'::regtype) THEN
        ALTER TYPE job_status ADD VALUE 'offer_sent' BEFORE 'accepted';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'rejected' AND enumtypid = 'job_status'::regtype) THEN
        ALTER TYPE job_status ADD VALUE 'rejected' AFTER 'accepted';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'worker_assigned' AND enumtypid = 'job_status'::regtype) THEN
        ALTER TYPE job_status ADD VALUE 'worker_assigned' AFTER 'rejected';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'scheduled' AND enumtypid = 'job_status'::regtype) THEN
        ALTER TYPE job_status ADD VALUE 'scheduled' AFTER 'worker_assigned';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'paid' AND enumtypid = 'job_status'::regtype) THEN
        ALTER TYPE job_status ADD VALUE 'paid' AFTER 'completed';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'review_pending' AND enumtypid = 'job_status'::regtype) THEN
        ALTER TYPE job_status ADD VALUE 'review_pending' AFTER 'paid';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'closed' AND enumtypid = 'job_status'::regtype) THEN
        ALTER TYPE job_status ADD VALUE 'closed' AFTER 'review_pending';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'no_worker_available' AND enumtypid = 'job_status'::regtype) THEN
        ALTER TYPE job_status ADD VALUE 'no_worker_available' AFTER 'cancelled';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'payment_failed' AND enumtypid = 'job_status'::regtype) THEN
        ALTER TYPE job_status ADD VALUE 'payment_failed' AFTER 'no_worker_available';
    END IF;
END $$;

-- Step 2: Add temporal workflow columns to jobs table if they don't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'jobs' AND column_name = 'temporal_workflow_id') THEN
        ALTER TABLE jobs ADD COLUMN temporal_workflow_id VARCHAR(255);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'jobs' AND column_name = 'temporal_run_id') THEN
        ALTER TABLE jobs ADD COLUMN temporal_run_id VARCHAR(255);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'jobs' AND column_name = 'workflow_started_at') THEN
        ALTER TABLE jobs ADD COLUMN workflow_started_at TIMESTAMP WITH TIME ZONE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'jobs' AND column_name = 'workflow_completed_at') THEN
        ALTER TABLE jobs ADD COLUMN workflow_completed_at TIMESTAMP WITH TIME ZONE;
    END IF;
END $$;

-- Step 3: Create index for temporal workflow lookups if it doesn't exist
CREATE INDEX IF NOT EXISTS idx_jobs_temporal_workflow
ON jobs(temporal_workflow_id)
WHERE temporal_workflow_id IS NOT NULL;

-- Step 4: Add password_hash column to people table if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'people' AND column_name = 'password_hash') THEN
        ALTER TABLE people ADD COLUMN password_hash VARCHAR(255);
    END IF;
END $$;

-- Display completion message
DO $$
BEGIN
    RAISE NOTICE 'Migration completed successfully!';
    RAISE NOTICE 'Job status enum now includes: posted, offer_sent, accepted, rejected, worker_assigned, scheduled, in_progress, completed, paid, review_pending, closed, cancelled, no_worker_available, payment_failed';
    RAISE NOTICE 'Jobs table now has temporal workflow tracking columns';
    RAISE NOTICE 'People table now has password_hash column';
END $$;
