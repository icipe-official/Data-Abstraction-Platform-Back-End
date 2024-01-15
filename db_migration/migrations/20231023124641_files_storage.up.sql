-- storage_types table
CREATE TABLE public.storage_types
(
    id text NOT NULL,
    properties json NOT NULL,
    PRIMARY KEY (id)
);

ALTER TABLE IF EXISTS public.storage_types
    OWNER to pg_database_owner;

COMMENT ON TABLE public.storage_types
    IS 'Template for different file storage modes supported by the system';

-- storage table
CREATE TABLE public.storage
(
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    storage_type_id text NOT NULL,
    name text NOT NULL,
    storage json NOT NULL,
    is_active boolean NOT NULL DEFAULT false,
    created_on timestamp with time zone NOT NULL DEFAULT NOW(),
    last_updated_on timestamp with time zone NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id),
    CONSTRAINT storage_type_id FOREIGN KEY (storage_type_id)
        REFERENCES public.storage_types (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS public.storage
    OWNER to pg_database_owner;

COMMENT ON TABLE public.storage
    IS 'File storage options available for project users';

CREATE OR REPLACE TRIGGER storage_update_last_updated_on
    BEFORE UPDATE OF storage_type_id, name, storage, is_active
    ON public.storage
    FOR EACH ROW
    EXECUTE FUNCTION public.update_last_updated_on();

COMMENT ON TRIGGER storage_update_last_updated_on ON public.storage
    IS 'update timestamp upon update on relevant columns';

-- files table
CREATE TABLE public.files
(
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    storage_id uuid NOT NULL,
    project_id uuid NOT NULL,
    directory_id uuid NOT NULL,
    content_type text NOT NULL,
    tags text NOT NULL,
    created_on timestamp with time zone NOT NULL DEFAULT NOW(),
    file_vector tsvector NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT storage_id FOREIGN KEY (storage_id)
        REFERENCES public.storage (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT project_id FOREIGN KEY (project_id)
        REFERENCES public.projects (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT directory_id FOREIGN KEY (directory_id)
        REFERENCES public.directory (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS public.files
    OWNER to pg_database_owner;

COMMENT ON TABLE public.files
    IS 'Files stored in the system';

CREATE INDEX file_vector
    ON public.files USING gin
    (file_vector);

-- trigger function to update catalogue_vector
CREATE FUNCTION public.file_update_vector()
    RETURNS trigger
    LANGUAGE 'plpgsql'
    NOT LEAKPROOF
AS $BODY$
DECLARE content_type text;
DECLARE tags text;

BEGIN
	IF LENGTH(NEW.content_type) > 0 THEN
		content_type = NEW.content_type;
	ELSE
		content_type = OLD.content_type;
	END IF;
    IF LENGTH(NEW.tags) > 0 THEN
        tags = NEW.tags;
    ELSE
        tags = OLD.tags;
    END IF;
	NEW.file_vector = 
        setweight(to_tsvector(coalesce(content_type,'')),'A') ||
        setweight(to_tsvector(coalesce(tags,'')),'B');
	RETURN NEW;
END
$BODY$;

ALTER FUNCTION public.file_update_vector()
    OWNER TO pg_database_owner;

COMMENT ON FUNCTION public.file_update_vector()
    IS 'Update file_vector column when content_type and tags change';

CREATE TRIGGER file_update_vector
    BEFORE INSERT OR UPDATE OF content_type, tags
    ON public.files
    FOR EACH ROW
    EXECUTE FUNCTION public.file_update_vector();

COMMENT ON TRIGGER file_update_vector ON public.files
    IS 'trigger to update file_vector';

-- storage_projects table
CREATE TABLE public.storage_projects
(
    storage_id uuid NOT NULL,
    project_id uuid NOT NULL,
    created_on timestamp with time zone NOT NULL DEFAULT NOW(),
    CONSTRAINT storage_project PRIMARY KEY (project_id, storage_id),
    CONSTRAINT project_id FOREIGN KEY (project_id)
        REFERENCES public.projects (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT storage_id FOREIGN KEY (storage_id)
        REFERENCES public.storage (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS public.storage_projects
    OWNER to pg_database_owner;

COMMENT ON TABLE public.storage_projects
    IS 'Storage options available to projects';