-- metadata_models table
CREATE TABLE public.metadata_models
(
    id uuid NOT NULL DEFAULT uuidv7_sub_ms(),
    directory_groups_id uuid NOT NULL,
    directory_id uuid,
    name text NOT NULL UNIQUE,
    description text NOT NULL,
    edit_authorized boolean NOT NULL DEFAULT TRUE,
    edit_unauthorized boolean NOT NULL DEFAULT FALSE,
    view_authorized boolean NOT NULL DEFAULT TRUE,
    view_unauthorized boolean NOT NULL DEFAULT FALSE,
    tags text[],
    data json NOT NULL,
    created_on timestamp without time zone NOT NULL DEFAULT NOW(),
    last_updated_on timestamp without time zone NOT NULL DEFAULT NOW(),
    deactivated_on timestamp without time zone,
    full_text_search tsvector,
    PRIMARY KEY (id),
    CONSTRAINT group_id FOREIGN KEY (directory_groups_id)
        REFERENCES public.directory_groups (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT directory_id FOREIGN KEY (directory_id)
        REFERENCES public.directory (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS public.metadata_models
    OWNER to pg_database_owner;

COMMENT ON TABLE public.metadata_models
    IS 'Metadata-Models for different purposes such as dynamic description of inventory items, directory user information etc.';

CREATE INDEX metadata_models_full_text_search_index
    ON public.metadata_models USING gin
    (full_text_search);

-- metadata_models trigger to update last_updated_on column
CREATE TRIGGER metadata_models_update_last_updated_on
    BEFORE UPDATE OF name, description, edit_authorized, edit_unauthorized, view_authorized, view_unauthorized, tags, data
    ON public.metadata_models
    FOR EACH ROW
    EXECUTE FUNCTION public.update_last_updated_on();

COMMENT ON TRIGGER metadata_models_update_last_updated_on ON public.metadata_models
    IS 'update timestamp upon update on relevant columns';

-- function and trigger to update metadata_models->full_text_search
CREATE FUNCTION public.metadata_models_update_full_text_search_index()
    RETURNS trigger
    LANGUAGE 'plpgsql'
    NOT LEAKPROOF
AS $BODY$
DECLARE name text;
DECLARE description text;
DECLARE tags text[];

BEGIN
	IF NEW.name IS DISTINCT FROM OLD.name THEN
		name = NEW.name;
	ELSE
		name = OLD.name;
	END IF;
	IF NEW.description IS DISTINCT FROM OLD.description THEN
		description = NEW.description;
	ELSE
		description = OLD.description;
	END IF;
	IF array_length(NEW.tags,1) > 0 THEN
		tags = NEW.tags;
	ELSE
        IF OLD.tags IS NOT NULL THEN
		    tags = OLD.tags;
        ELSE
            tags = '{}';
        END IF;
	END IF;

    NEW.full_text_search = 
        setweight(to_tsvector(coalesce(name,'')),'A') ||
        setweight(to_tsvector(coalesce(description,'')),'B') ||
        setweight(to_tsvector(coalesce(array_to_string(tags,' ','*'),'')),'C');
    	
	RETURN NEW;   
END
$BODY$;

ALTER FUNCTION public.metadata_models_update_full_text_search_index()
    OWNER TO pg_database_owner;

COMMENT ON FUNCTION public.metadata_models_update_full_text_search_index()
    IS 'Update full_text_search column in metadata_models when name, description, and tags change';

CREATE TRIGGER metadata_models_update_full_text_search_index
    BEFORE INSERT OR UPDATE OF name, description, tags
    ON public.metadata_models
    FOR EACH ROW
    EXECUTE FUNCTION public.metadata_models_update_full_text_search_index();

COMMENT ON TRIGGER metadata_models_update_full_text_search_index ON public.metadata_models
    IS 'trigger to update full_text_search column';

-- metadata_models_authorization_ids table
CREATE TABLE public.metadata_models_authorization_ids
(
    id uuid NOT NULL,
    creation_iam_group_authorizations_id uuid,
    deactivation_iam_group_authorizations_id uuid,
    PRIMARY KEY (id),
    CONSTRAINT group_rule_authorizations_id FOREIGN KEY (id)
        REFERENCES public.metadata_models (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID, 
    CONSTRAINT creation_iam_group_authorizations_id FOREIGN KEY (creation_iam_group_authorizations_id)
        REFERENCES public.iam_group_authorizations (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT deactivation_iam_group_authorizations_id FOREIGN KEY (deactivation_iam_group_authorizations_id)
        REFERENCES public.iam_group_authorizations (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS public.metadata_models_authorization_ids
    OWNER to pg_database_owner;

COMMENT ON TABLE public.metadata_models_authorization_ids
    IS 'authorization ids that were used to create and/or deactivate the resources';

-- metadata_models_directory table
CREATE TABLE public.metadata_models_directory
(
    directory_groups_id uuid NOT NULL,
    metadata_models_id uuid NOT NULL,
    created_on timestamp without time zone NOT NULL DEFAULT NOW(),
    last_updated_on timestamp without time zone NOT NULL DEFAULT NOW(),
    PRIMARY KEY (directory_groups_id),
    CONSTRAINT group_id FOREIGN KEY (directory_groups_id)
        REFERENCES public.directory_groups (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT metadata_models_id FOREIGN KEY (metadata_models_id)
        REFERENCES public.metadata_models (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS public.metadata_models_directory
    OWNER to pg_database_owner;

COMMENT ON TABLE public.metadata_models_directory
    IS 'Metadata model to use when working with directory data in a particular group';

-- metadata_models_directory trigger to update last_updated_on column
CREATE TRIGGER metadata_models_directory_update_last_updated_on
    BEFORE UPDATE OF metadata_models_id
    ON public.metadata_models_directory
    FOR EACH ROW
    EXECUTE FUNCTION public.update_last_updated_on();

COMMENT ON TRIGGER metadata_models_directory_update_last_updated_on ON public.metadata_models_directory
    IS 'update timestamp upon update on relevant columns';

-- metadata_models_directory_groups table
CREATE TABLE public.metadata_models_directory_groups
(
    directory_groups_id uuid NOT NULL,
    metadata_models_id uuid NOT NULL,
    created_on timestamp without time zone NOT NULL DEFAULT NOW(),
    last_updated_on timestamp without time zone NOT NULL DEFAULT NOW(),
    PRIMARY KEY (directory_groups_id),
    CONSTRAINT group_id FOREIGN KEY (directory_groups_id)
        REFERENCES public.directory_groups (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT metadata_models_id FOREIGN KEY (metadata_models_id)
        REFERENCES public.metadata_models (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS public.metadata_models_directory_groups
    OWNER to pg_database_owner;

COMMENT ON TABLE public.metadata_models_directory_groups
    IS 'Metadata model to use when working with directory groups data in a particular group';

-- metadata_models_directory_groups trigger to update last_updated_on column
CREATE TRIGGER metadata_models_directory_groups_update_last_updated_on
    BEFORE UPDATE OF metadata_models_id
    ON public.metadata_models_directory_groups
    FOR EACH ROW
    EXECUTE FUNCTION public.update_last_updated_on();

COMMENT ON TRIGGER metadata_models_directory_groups_update_last_updated_on ON public.metadata_models_directory_groups
    IS 'update timestamp upon update on relevant columns';