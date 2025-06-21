-- abstractions_directory_groups table
CREATE TABLE public.abstractions_directory_groups
(
    directory_groups_id uuid NOT NULL,
    metadata_models_id uuid NOT NULL,
    description text,
    abstraction_review_quorum integer DEFAULT 0,
    view_authorized boolean NOT NULL DEFAULT TRUE,
    view_unauthorized boolean NOT NULL DEFAULT FALSE,
    created_on timestamp without time zone NOT NULL DEFAULT NOW(),
    last_updated_on timestamp without time zone NOT NULL DEFAULT NOW(),
    deactivated_on timestamp without time zone,
    PRIMARY KEY (directory_groups_id),
    CONSTRAINT directory_groups_id FOREIGN KEY (directory_groups_id)
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

ALTER TABLE IF EXISTS public.abstractions_directory_groups
    OWNER to pg_database_owner;

-- abstractions_directory_groups trigger to update last_updated_on column
CREATE TRIGGER abstractions_directory_groups_update_last_updated_on
    BEFORE UPDATE OF description, abstraction_review_quorum
    ON public.abstractions_directory_groups
    FOR EACH ROW
    EXECUTE FUNCTION public.update_last_updated_on();




-- abstractions_directory_groups_authorization_ids table
CREATE TABLE public.abstractions_directory_groups_authorization_ids
(
    directory_groups_id uuid NOT NULL,
    creation_iam_group_authorizations_id uuid,
    deactivation_iam_group_authorizations_id uuid,
    PRIMARY KEY (directory_groups_id),
    CONSTRAINT directory_groups_id FOREIGN KEY (directory_groups_id)
        REFERENCES public.abstractions_directory_groups (directory_groups_id) MATCH SIMPLE
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

ALTER TABLE IF EXISTS public.abstractions_directory_groups_authorization_ids
    OWNER to pg_database_owner;
