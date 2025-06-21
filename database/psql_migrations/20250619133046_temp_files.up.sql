-- storage_files_temporary table
CREATE TABLE public.storage_files_temporary
(
    id uuid NOT NULL DEFAULT uuidv7_sub_ms(),
    original_name text,
    storage_file_mime_type text,
    tags text[],
    size_in_bytes bigint NOT NULL,
    created_on timestamp without time zone NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id)
);

ALTER TABLE IF EXISTS public.storage_files_temporary
    OWNER to pg_database_owner;