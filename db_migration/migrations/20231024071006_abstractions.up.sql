-- abstractions table
CREATE TABLE public.abstractions
(
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    model_template_id uuid NOT NULL,
    file_id uuid NOT NULL,
    directory_id uuid NOT NULL,
    project_id uuid NOT NULL,
    tags text,
    abstraction jsonb NOT NULL,
    is_verified boolean NOT NULL DEFAULT FALSE,
    created_on timestamp with time zone NOT NULL DEFAULT NOW(),
    last_updated_on timestamp with time zone NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id),
    CONSTRAINT model_template_id FOREIGN KEY (model_template_id)
        REFERENCES public.model_templates (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT file_id FOREIGN KEY (file_id)
        REFERENCES public.files (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT directory_id FOREIGN KEY (directory_id)
        REFERENCES public.directory (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT project_id FOREIGN KEY (project_id)
        REFERENCES public.projects (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS public.abstractions
    OWNER to pg_database_owner;

COMMENT ON TABLE public.abstractions
    IS 'Abstractions created by abstractors';

CREATE TRIGGER abstractions_update_last_updated_on
    BEFORE UPDATE OF tags, abstraction, is_verified
    ON public.abstractions
    FOR EACH ROW
    EXECUTE FUNCTION public.update_last_updated_on();

COMMENT ON TRIGGER abstractions_update_last_updated_on ON public.abstractions
    IS 'update timestamp upon update on relevant columns';

-- abstraction_reviews table
CREATE TABLE public.abstraction_reviews
(
    abstraction_id uuid NOT NULL,
    directory_id uuid NOT NULL,
    review boolean NOT NULL,
    created_on timestamp with time zone NOT NULL DEFAULT NOW(),
    last_updated_on timestamp with time zone NOT NULL DEFAULT NOW(),
    CONSTRAINT abstraction_directory PRIMARY KEY (abstraction_id, directory_id),
    CONSTRAINT abstraction_id FOREIGN KEY (abstraction_id)
        REFERENCES public.abstractions (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT directory_id FOREIGN KEY (directory_id)
        REFERENCES public.directory (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS public.abstraction_reviews
    OWNER to pg_database_owner;

COMMENT ON TABLE public.abstraction_reviews
    IS 'Review made on abstractions by reviews';

-- abstraction_reviews_comments table
CREATE TABLE public.abstraction_reviews_comments
(
    id uuid NOT NULL DEFAULT uuid_generate_v4(),
    abstraction_id uuid NOT NULL,
    directory_id uuid NOT NULL,
    comment text NOT NULL,
    created_on timestamp with time zone NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id),
    CONSTRAINT abstraction_id FOREIGN KEY (abstraction_id)
        REFERENCES public.abstractions (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT directory_id FOREIGN KEY (directory_id)
        REFERENCES public.directory (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID    
);

ALTER TABLE IF EXISTS public.abstraction_reviews_comments
    OWNER to pg_database_owner;

COMMENT ON TABLE public.abstraction_reviews_comments
    IS 'Comments made by reviewers on abstractions';