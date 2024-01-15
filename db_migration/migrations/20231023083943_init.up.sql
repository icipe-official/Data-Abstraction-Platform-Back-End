-- will be used to generate uuid version 4
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- will be used to generate password hashes
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- function to update last_updated_on column
CREATE FUNCTION public.update_last_updated_on()
    RETURNS trigger
    LANGUAGE 'plpgsql'
    NOT LEAKPROOF
AS $BODY$
BEGIN
    NEW.last_updated_on=NOW();
    RETURN NEW;
END
$BODY$;

ALTER FUNCTION public.update_last_updated_on()
    OWNER TO pg_database_owner;

COMMENT ON FUNCTION public.update_last_updated_on()
    IS 'update last_updated_on column when relevant columns in a particular row are updated';

-- function to generate random alphanumeric string
CREATE FUNCTION public.gen_random_string()
    RETURNS character varying
    LANGUAGE 'sql'
    

RETURN encode(sha256(md5(random()::text)::bytea),'hex');

ALTER FUNCTION public.gen_random_string()
    OWNER TO pg_database_owner;

COMMENT ON FUNCTION public.gen_random_string()
    IS 'generate alphanumeric lowercase string that is 64 characters in length using sha256 and md5';