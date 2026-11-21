CREATE OR REPLACE FUNCTION increment_record_version()
RETURNS TRIGGER AS $$
BEGIN
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER records_increment_version
    BEFORE UPDATE ON records
    FOR EACH ROW
    EXECUTE FUNCTION increment_record_version();