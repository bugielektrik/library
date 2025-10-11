-- Drop trigger first
DROP TRIGGER IF EXISTS trigger_callback_retries_updated_at ON callback_retries;

-- Drop function
DROP FUNCTION IF EXISTS update_callback_retries_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_callback_retries_created_at;
DROP INDEX IF EXISTS idx_callback_retries_status_next_retry;
DROP INDEX IF EXISTS idx_callback_retries_payment_id;

-- Drop table
DROP TABLE IF EXISTS callback_retries;
