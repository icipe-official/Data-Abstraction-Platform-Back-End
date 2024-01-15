-- directory table
CREATE TABLE public.directory
(
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    name text NOT NULL,
    contacts text[],
    created_on timestamp with time zone NOT NULL DEFAULT NOW(),
    last_updated_on timestamp with time zone NOT NULL DEFAULT NOW(),
    directory_vector tsvector NOT NULL,
    PRIMARY KEY (id)
);

ALTER TABLE IF EXISTS public.directory
    OWNER to pg_database_owner;

COMMENT ON TABLE public.directory
    IS 'Organisation users of the system';

CREATE INDEX directory_vector
    ON public.directory USING gin
    (directory_vector);

-- trigger function to update directory_vector
CREATE FUNCTION public.directory_update_vector()
    RETURNS trigger
    LANGUAGE 'plpgsql'
    NOT LEAKPROOF
AS $BODY$
DECLARE name text;
DECLARE contacts text[];

BEGIN
	IF LENGTH(NEW.name) > 0 THEN
		name = NEW.name;
	ELSE
		name = OLD.name;
	END IF;
    IF array_length(NEW.contacts, 1) > 0 THEN
        contacts = NEW.contacts;
    ELSE
        contacts = OLD.contacts;
    END IF;
	NEW.directory_vector = 
        setweight(to_tsvector(coalesce(name,'')),'A') ||
        setweight(to_tsvector(coalesce(array_to_string(contacts,' ','*'),'')),'B');
	RETURN NEW;
END
$BODY$;

ALTER FUNCTION public.directory_update_vector()
    OWNER TO pg_database_owner;

COMMENT ON FUNCTION public.directory_update_vector()
    IS 'Update directory_vector column when name and contacts change';

CREATE TRIGGER directory_update_vector
    BEFORE INSERT OR UPDATE OF name, contacts
    ON public.directory
    FOR EACH ROW
    EXECUTE FUNCTION public.directory_update_vector();

COMMENT ON TRIGGER directory_update_vector ON public.directory
    IS 'trigger to update directory_vector';

-- trigger to update last_updated_on column
CREATE TRIGGER directory_update_last_updated_on
    BEFORE UPDATE OF name, contacts
    ON public.directory
    FOR EACH ROW
    EXECUTE FUNCTION public.update_last_updated_on();

COMMENT ON TRIGGER directory_update_last_updated_on ON public.directory
    IS 'update timestamp upon update on relevant columns';

-- directory_iam_ticket_types table
CREATE TABLE public.directory_iam_ticket_types
(
    id text NOT NULL,
    description text NOT NULL,
    PRIMARY KEY (id)
);

ALTER TABLE IF EXISTS public.directory_iam_ticket_types
    OWNER to pg_database_owner;

COMMENT ON TABLE public.directory_iam_ticket_types
    IS 'Ticket types for iam requests';

-- directory_iam table
CREATE TABLE public.directory_iam
(
    directory_id uuid NOT NULL,
    email text,
    password text,
    is_email_verified boolean NOT NULL DEFAULT false,
    directory_iam_ticket_id text,
    ticket_number text,
    pin text,
    created_on timestamp with time zone NOT NULL DEFAULT NOW(),
    last_updated_on timestamp with time zone NOT NULL DEFAULT NOW(),
    PRIMARY KEY (directory_id),
    CONSTRAINT email_unique UNIQUE (email),
    CONSTRAINT directory_id FOREIGN KEY (directory_id)
        REFERENCES public.directory (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT directory_iam_ticket_id FOREIGN KEY (directory_iam_ticket_id)
        REFERENCES public.directory_iam_ticket_types (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS public.directory_iam
    OWNER to pg_database_owner;

COMMENT ON TABLE public.directory_iam
    IS 'Credentials for directory users to access the platform using email and password';

-- trigger to update last_updated_on column
CREATE TRIGGER directory_iam_update_last_updated_on
    BEFORE UPDATE OF email, password, is_email_verified, ticket_number, pin
    ON public.directory_iam
    FOR EACH ROW
    EXECUTE FUNCTION public.update_last_updated_on();

COMMENT ON TRIGGER directory_iam_update_last_updated_on ON public.directory_iam
    IS 'update timestamp upon update on relevant columns';

-- function to generate password salt
CREATE FUNCTION public.directory_gen_password_salt()
    RETURNS trigger
    LANGUAGE 'plpgsql'
    NOT LEAKPROOF
AS $BODY$
BEGIN
	IF LENGTH(NEW.password) > 0 THEN
		NEW.password = crypt(NEW.password, gen_salt('bf'));
    END IF;	
	RETURN NEW;
END
$BODY$;

ALTER FUNCTION public.directory_gen_password_salt()
    OWNER TO pg_database_owner;

COMMENT ON FUNCTION public.directory_gen_password_salt()
    IS 'Generate password hash when a user is created or when the password is updated';

CREATE TRIGGER directory_gen_password_salt
    BEFORE INSERT OR UPDATE OF password
    ON public.directory_iam
    FOR EACH ROW
    EXECUTE FUNCTION public.directory_gen_password_salt();

COMMENT ON TRIGGER directory_gen_password_salt ON public.directory_iam
    IS 'trigger to compute password salt';

-- function to generate pin salt
CREATE FUNCTION public.directory_gen_pin_salt()
    RETURNS trigger
    LANGUAGE 'plpgsql'
    VOLATILE NOT LEAKPROOF
AS $BODY$
BEGIN
	IF LENGTH(NEW.pin) > 0 THEN
		NEW.pin = crypt(NEW.pin, gen_salt('bf'));
    END IF;
	RETURN NEW;
END
$BODY$;

ALTER FUNCTION public.directory_gen_pin_salt()
    OWNER TO pg_database_owner;

COMMENT ON FUNCTION public.directory_gen_pin_salt()
    IS 'Generate pin hash when a pin and ticket is generated';

CREATE OR REPLACE TRIGGER directory_gen_pin_salt
    BEFORE UPDATE OF pin
    ON public.directory_iam
    FOR EACH ROW
    EXECUTE FUNCTION public.directory_gen_pin_salt();

COMMENT ON TRIGGER directory_gen_pin_salt ON public.directory_iam
    IS 'trigger to compute pin salt';

-- directory system users table
CREATE TABLE public.directory_system_users
(
    directory_id uuid NOT NULL,
    created_on timestamp with time zone NOT NULL DEFAULT NOW(),
    PRIMARY KEY (directory_id),
    CONSTRAINT directory_id FOREIGN KEY (directory_id)
        REFERENCES public.directory (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS public.directory_system_users
    OWNER to pg_database_owner;

COMMENT ON TABLE public.directory_system_users
    IS 'System users';