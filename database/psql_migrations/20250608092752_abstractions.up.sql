-- abstractions table
CREATE TABLE public.abstractions
(
    id uuid NOT NULL DEFAULT uuidv7_sub_ms(),
    abstractions_directory_groups_id uuid NOT NULL,
    directory_id uuid NOT NULL,
    storage_files_id uuid NOT NULL,
    data jsonb,
    completed boolean NOT NULL DEFAULT FALSE,
    review_pass boolean NOT NULL DEFAULT FALSE,
    tags text[],
    created_on timestamp without time zone NOT NULL DEFAULT NOW(),
    last_updated_on timestamp without time zone NOT NULL DEFAULT NOW(),
    deactivated_on timestamp without time zone,
    full_text_search tsvector,
    PRIMARY KEY (id),
    CONSTRAINT abstractions_directory_groups_id FOREIGN KEY (abstractions_directory_groups_id)
        REFERENCES public.abstractions_directory_groups (directory_groups_id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT directory_id FOREIGN KEY (directory_id)
        REFERENCES public.directory (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT storage_files_id FOREIGN KEY (storage_files_id)
        REFERENCES public.storage_files (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS public.abstractions
    OWNER to pg_database_owner;

-- abstractions trigger to update last_updated_on column
CREATE TRIGGER abstractions_update_last_updated_on
    BEFORE UPDATE OF directory_id, storage_files_id, tags, data, completed, review_pass
    ON public.abstractions
    FOR EACH ROW
    EXECUTE FUNCTION public.update_last_updated_on();




-- abstractions_authorization_ids table
CREATE TABLE public.abstractions_authorization_ids
(
    id uuid NOT NULL,
    creation_iam_group_authorizations_id uuid,
    deactivation_iam_group_authorizations_id uuid,
    PRIMARY KEY (id),
    CONSTRAINT abstractions_id FOREIGN KEY (id)
        REFERENCES public.abstractions (id) MATCH SIMPLE
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

ALTER TABLE IF EXISTS public.abstractions_authorization_ids
    OWNER to pg_database_owner;
