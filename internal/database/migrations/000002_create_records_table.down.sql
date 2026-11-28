DROP TRIGGER IF EXISTS records_set_updated_at ON records;
DROP FUNCTION IF EXISTS set_updated_at();
DROP INDEX IF EXISTS idx_records_user_updated;
DROP TABLE IF EXISTS records;