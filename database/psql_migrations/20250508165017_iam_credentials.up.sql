-- iam_credentials table
CREATE TABLE iam_credentials
(
    id uuid NOT NULL DEFAULT uuidv7_sub_ms(),
    directory_id uuid,
    openid_sub text NOT NULL UNIQUE,
    openid_preferred_username text NOT NULL,
    openid_email text NOT NULL,
    openid_email_verified bool,
    openid_given_name text,
    openid_family_name text,
    created_on timestamp without time zone NOT NULL DEFAULT NOW(),
    last_updated_on timestamp without time zone NOT NULL DEFAULT NOW(),
    deactivated_on timestamp without time zone,
    full_text_search tsvector,
    PRIMARY KEY (id),
    CONSTRAINT directory_id FOREIGN KEY (directory_id)
        REFERENCES directory (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS iam_credentials
    OWNER to pg_database_owner;




-- trigger to update last_updated_on column
CREATE TRIGGER iam_credentials_update_last_updated_on
    BEFORE UPDATE OF directory_id, openid_sub, openid_preferred_username, openid_email, openid_email_verified, openid_given_name, openid_family_name, deactivated_on
    ON iam_credentials
    FOR EACH ROW
    EXECUTE FUNCTION update_last_updated_on();

COMMENT ON TRIGGER iam_credentials_update_last_updated_on ON iam_credentials
    IS 'update timestamp upon update on relevant columns';




-- full_text_search
CREATE INDEX iam_credentials_full_text_search_index
    ON iam_credentials USING gin
    (full_text_search);

-- function and trigger to update iam_credentials->full_text_search
CREATE FUNCTION iam_credentials_update_full_text_search_index()
    RETURNS trigger
    LANGUAGE 'plpgsql'
    NOT LEAKPROOF
AS $BODY$

DECLARE openid_preferred_username text;
DECLARE openid_email text;
DECLARE openid_given_name text;
DECLARE openid_family_name text;

BEGIN
	IF NEW.openid_preferred_username IS DISTINCT FROM OLD.openid_preferred_username THEN
		openid_preferred_username = NEW.openid_preferred_username;
	ELSE
		openid_preferred_username = OLD.openid_preferred_username;
	END IF;
    IF NEW.openid_email IS DISTINCT FROM OLD.openid_email THEN
		openid_email = NEW.openid_email;
	ELSE
		openid_email = OLD.openid_email;
	END IF;
    IF NEW.openid_given_name IS DISTINCT FROM OLD.openid_given_name THEN
		openid_given_name = NEW.openid_given_name;
	ELSE
		openid_given_name = OLD.openid_given_name;
	END IF;
    IF NEW.openid_family_name IS DISTINCT FROM OLD.openid_family_name THEN
		openid_family_name = NEW.openid_family_name;
	ELSE
		openid_family_name = OLD.openid_family_name;
	END IF;

    NEW.full_text_search = 
        setweight(to_tsvector(coalesce(openid_preferred_username,'')),'A') ||
        setweight(to_tsvector(coalesce(openid_email,'')),'B') ||
        setweight(to_tsvector(coalesce(openid_given_name,'')),'C') ||
        setweight(to_tsvector(coalesce(openid_family_name,'')),'D');
    	
	RETURN NEW;   
END
$BODY$;

ALTER FUNCTION iam_credentials_update_full_text_search_index()
    OWNER TO pg_database_owner;

COMMENT ON FUNCTION iam_credentials_update_full_text_search_index()
    IS 'Update full_text_search column in iam_credentials when openid_preferred_username, openid_email, openid_given_name, openid_family_name change';

CREATE TRIGGER iam_credentials_update_full_text_search_index
    BEFORE INSERT OR UPDATE OF openid_preferred_username, openid_email, openid_given_name, openid_family_name
    ON iam_credentials
    FOR EACH ROW
    EXECUTE FUNCTION iam_credentials_update_full_text_search_index();

COMMENT ON TRIGGER iam_credentials_update_full_text_search_index ON iam_credentials
    IS 'trigger to update full_text_search column';