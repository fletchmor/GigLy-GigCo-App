-- Migration: Add dual completion confirmation for jobs
-- This allows both worker and consumer to confirm job completion

-- Add completion confirmation columns
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'jobs' AND column_name = 'worker_completed_at') THEN
        ALTER TABLE jobs ADD COLUMN worker_completed_at TIMESTAMP WITH TIME ZONE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'jobs' AND column_name = 'consumer_completed_at') THEN
        ALTER TABLE jobs ADD COLUMN consumer_completed_at TIMESTAMP WITH TIME ZONE;
    END IF;
END $$;

-- Add indexes for completion tracking
CREATE INDEX IF NOT EXISTS idx_jobs_worker_completed
ON jobs(worker_completed_at)
WHERE worker_completed_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_jobs_consumer_completed
ON jobs(consumer_completed_at)
WHERE consumer_completed_at IS NOT NULL;

-- Add comment explaining the dual completion process
COMMENT ON COLUMN jobs.worker_completed_at IS 'Timestamp when the gig worker marked the job as completed';
COMMENT ON COLUMN jobs.consumer_completed_at IS 'Timestamp when the consumer confirmed the job completion';

DO $$
BEGIN
    RAISE NOTICE 'Dual completion columns added successfully!';
    RAISE NOTICE 'Jobs can now be marked complete by both worker and consumer';
    RAISE NOTICE 'Job status changes to "completed" when both parties confirm';
END $$;
