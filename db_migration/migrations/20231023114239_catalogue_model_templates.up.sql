-- catalogue table
CREATE TABLE public.catalogue
(
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    project_id uuid NOT NULL,
    directory_id uuid NOT NULL,
    name text NOT NULL,
    description text NOT NULL,
    catalogue text[] NOT NULL,
    created_on timestamp with time zone NOT NULL DEFAULT NOW(),
    last_updated_on timestamp with time zone NOT NULL DEFAULT NOW(),
    can_public_view boolean NOT NULL DEFAULT false,
    catalogue_vector tsvector NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT project_id FOREIGN KEY (project_id)
        REFERENCES public.projects (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT directory_id FOREIGN KEY (directory_id)
        REFERENCES public.directory (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS public.catalogue
    OWNER to pg_database_owner;

COMMENT ON TABLE public.catalogue
    IS 'A collection of well known and established list of distinct world entities';

CREATE INDEX catalogue_vector
    ON public.catalogue USING gin
    (catalogue_vector);

CREATE TRIGGER catalogue_update_last_updated_on
    BEFORE UPDATE OF directory_id, name, description, catalogue, can_public_view
    ON public.catalogue
    FOR EACH ROW
    EXECUTE FUNCTION public.update_last_updated_on();

COMMENT ON TRIGGER catalogue_update_last_updated_on ON public.catalogue
    IS 'update timestamp upon update on relevant columns';

-- trigger function to update catalogue_vector
CREATE FUNCTION public.catalogue_update_vector()
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
	NEW.catalogue_vector = 
        setweight(to_tsvector(coalesce(name,'')),'A') ||
        setweight(to_tsvector(coalesce(description,'')),'B');
	RETURN NEW;
END
$BODY$;

ALTER FUNCTION public.catalogue_update_vector()
    OWNER TO pg_database_owner;

COMMENT ON FUNCTION public.catalogue_update_vector()
    IS 'Update catalogue_vector column when name and description change';

CREATE TRIGGER catalogue_update_vector
    BEFORE INSERT OR UPDATE OF name, description
    ON public.catalogue
    FOR EACH ROW
    EXECUTE FUNCTION public.catalogue_update_vector();

COMMENT ON TRIGGER catalogue_update_vector ON public.catalogue
    IS 'trigger to update catalogue_vector';

-- model_templates table
CREATE TABLE model_templates
(
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    project_id uuid NOT NULL,
    directory_id uuid NOT NULL,
    template_name text NOT NULL,
    data_name text NOT NULL,
    description text NOT NULL,
    model_template json NOT NULL,
    created_on timestamp with time zone NOT NULL DEFAULT NOW(),
    last_updated_on timestamp with time zone NOT NULL DEFAULT NOW(),
    verification_quorum integer NOT NULL DEFAULT 0,
    can_public_view boolean NOT NULL DEFAULT false,
    model_template_vector tsvector NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT project_id FOREIGN KEY (project_id)
        REFERENCES public.projects (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT directory_id FOREIGN KEY (directory_id)
        REFERENCES public.directory (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS public.model_templates
    OWNER to pg_database_owner;

COMMENT ON TABLE public.model_templates
    IS 'Template and model for the data to collect';

CREATE INDEX model_template_vector
    ON public.model_templates USING gin
    (model_template_vector);

CREATE TRIGGER model_templates_update_last_updated_on
    BEFORE UPDATE OF directory_id, template_name, data_name, description, model_template, verification_quorum, can_public_view
    ON public.model_templates
    FOR EACH ROW
    EXECUTE FUNCTION public.update_last_updated_on();

COMMENT ON TRIGGER model_templates_update_last_updated_on ON public.model_templates
    IS 'update timestamp upon update on relevant columns';

-- trigger function to update model_template_vector
CREATE FUNCTION public.model_templates_update_vector()
    RETURNS trigger
    LANGUAGE 'plpgsql'
    NOT LEAKPROOF
AS $BODY$
DECLARE template_name text;
DECLARE data_name text;
DECLARE description text;

BEGIN
	IF LENGTH(NEW.template_name) > 0 THEN
		template_name = NEW.template_name;
	ELSE
		template_name = OLD.template_name;
	END IF;
    IF LENGTH(NEW.data_name) > 0 THEN
        data_name = NEW.data_name;
    ELSE
        data_name = OLD.data_name;
    END IF;
    IF LENGTH(NEW.description) > 0 THEN
        description = NEW.description;
    ELSE
        description = OLD.description;
    END IF;
	NEW.model_template_vector = 
        setweight(to_tsvector(coalesce(template_name,'')),'A') ||
        setweight(to_tsvector(coalesce(data_name,'')),'B') ||
        setweight(to_tsvector(coalesce(description,'')),'C');
	RETURN NEW;
END
$BODY$;

ALTER FUNCTION public.model_templates_update_vector()
    OWNER TO pg_database_owner;

COMMENT ON FUNCTION public.model_templates_update_vector()
    IS 'Update model_template_vector column when name and description change';

CREATE TRIGGER model_templates_update_vector
    BEFORE INSERT OR UPDATE OF template_name, data_name, description
    ON public.model_templates
    FOR EACH ROW
    EXECUTE FUNCTION public.model_templates_update_vector();

COMMENT ON TRIGGER model_templates_update_vector ON public.model_templates
    IS 'trigger to update model_template_vector';