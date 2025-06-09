-- postigs extension
-- CREATE EXTENSION postgis;

-- will be used to generate password hashes
-- CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- function to update last_updated_on column
CREATE FUNCTION update_last_updated_on()
    RETURNS trigger
    LANGUAGE 'plpgsql'
    NOT LEAKPROOF
AS $BODY$
BEGIN
    NEW.last_updated_on=NOW();
    RETURN NEW;
END
$BODY$;

ALTER FUNCTION update_last_updated_on()
    OWNER TO pg_database_owner;

COMMENT ON FUNCTION update_last_updated_on()
    IS 'update last_updated_on column when relevant columns in a particular row are updated';
