-- projects table
CREATE TABLE public.projects
(
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    directory_id uuid NOT NULL,
    name text NOT NULL,
    description text NOT NULL,
    created_on timestamp with time zone NOT NULL DEFAULT NOW(),
    last_updated_on timestamp with time zone NOT NULL DEFAULT NOW(),
    is_active boolean NOT NULL DEFAULT false,
    projects_vector tsvector NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT directory_id FOREIGN KEY (directory_id)
        REFERENCES public.directory (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS public.projects
    OWNER to pg_database_owner;

COMMENT ON TABLE public.projects
    IS 'Logical group of related catalogues, abstractions, models_templates, surveys etc.';

CREATE INDEX projects_vector
    ON public.projects USING gin
    (projects_vector);

CREATE TRIGGER projects_update_last_updated_on
    BEFORE UPDATE OF name, description
    ON public.projects
    FOR EACH ROW
    EXECUTE FUNCTION public.update_last_updated_on();

COMMENT ON TRIGGER projects_update_last_updated_on ON public.projects
    IS 'update timestamp upon update on relevant columns';

-- trigger function to update projects_vector
CREATE FUNCTION public.projects_update_vector()
    RETURNS trigger
    LANGUAGE 'plpgsql'
    NOT LEAKPROOF
AS $BODY$
DECLARE name text;
DECLARE description text;

BEGIN
	IF LENGTH(NEW.name) > 0 THEN
		name = NEW.name;
	ELSE
		name = OLD.name;
	END IF;
    IF LENGTH(NEW.description) > 0 THEN
        description = NEW.description;
    ELSE
        description = OLD.description;
    END IF;
	NEW.projects_vector = 
        setweight(to_tsvector(coalesce(name,'')),'A') ||
        setweight(to_tsvector(coalesce(description,'')),'B');
	RETURN NEW;
END
$BODY$;

ALTER FUNCTION public.projects_update_vector()
    OWNER TO pg_database_owner;

COMMENT ON FUNCTION public.projects_update_vector()
    IS 'Update projects_vector column when name and description change';

CREATE TRIGGER projects_update_vector
    BEFORE INSERT OR UPDATE OF name, description
    ON public.projects
    FOR EACH ROW
    EXECUTE FUNCTION public.projects_update_vector();

COMMENT ON TRIGGER projects_update_vector ON public.projects
    IS 'trigger to update projects_vector';

-- projects_roles table
CREATE TABLE public.project_roles
(
    id text NOT NULL,
    description text NOT NULL,
    PRIMARY KEY (id)
);

ALTER TABLE IF EXISTS public.project_roles
    OWNER to pg_database_owner;

COMMENT ON TABLE public.project_roles
    IS 'Different types of roles available in a project';

-- directory_projects_roles table
CREATE TABLE public.directory_projects_roles
(
    directory_id uuid NOT NULL,
    project_id uuid NOT NULL,
    project_role_id text NOT NULL,
    created_on timestamp with time zone NOT NULL DEFAULT NOW(),
    CONSTRAINT project_directory_role PRIMARY KEY (project_id, directory_id, project_role_id),
    CONSTRAINT project_id FOREIGN KEY (project_id)
        REFERENCES public.projects (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT directory_id FOREIGN KEY (directory_id)
        REFERENCES public.directory (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT project_role_id FOREIGN KEY (project_role_id)
        REFERENCES public.project_roles (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS public.directory_projects_roles
    OWNER to pg_database_owner;

COMMENT ON TABLE public.directory_projects_roles
    IS 'Different kinds of roles different users are playing in different projects';