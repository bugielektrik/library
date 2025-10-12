-- Drop trigger first
DROP TRIGGER IF EXISTS trigger_receipts_updated_at ON receipts;

-- Drop function
DROP FUNCTION IF EXISTS update_receipts_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_receipts_created_at;
DROP INDEX IF EXISTS idx_receipts_payment_date;
DROP INDEX IF EXISTS idx_receipts_receipt_number;
DROP INDEX IF EXISTS idx_receipts_member_id;
DROP INDEX IF EXISTS idx_receipts_payment_id;

-- Drop table
DROP TABLE IF EXISTS receipts;
