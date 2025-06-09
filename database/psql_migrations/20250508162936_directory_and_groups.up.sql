-- directory_groups table
CREATE TABLE directory_groups
(
    id uuid NOT NULL DEFAULT uuidv7_sub_ms(),
    display_name text NOT NULL UNIQUE,
    description text,
    data jsonb,
    created_on timestamp without time zone NOT NULL DEFAULT NOW(),
    last_updated_on timestamp without time zone NOT NULL DEFAULT NOW(),
    deactivated_on timestamp without time zone,
    full_text_search tsvector,
    PRIMARY KEY (id)
);

ALTER TABLE IF EXISTS directory_groups
    OWNER to pg_database_owner;

CREATE INDEX directory_groups_full_text_search_index
    ON directory_groups USING gin
    (full_text_search);

CREATE INDEX directory_groups_data_jsonb_index
    ON directory_groups USING gin
    (data);

-- directory_groups trigger to update last_updated_on column
CREATE TRIGGER directory_groups_update_last_updated_on
    BEFORE UPDATE OF display_name, description, data
    ON directory_groups
    FOR EACH ROW
    EXECUTE FUNCTION update_last_updated_on();

COMMENT ON TRIGGER directory_groups_update_last_updated_on ON directory_groups
    IS 'update timestamp upon update on relevant columns';

-- directory_groups_sub_groups table
CREATE TABLE directory_groups_sub_groups
(
    parent_group_id uuid NOT NULL,
    sub_group_id uuid NOT NULL,
    PRIMARY KEY (parent_group_id, sub_group_id),
    CONSTRAINT group_id FOREIGN KEY (parent_group_id)
        REFERENCES directory_groups (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT sub_group_id FOREIGN KEY (sub_group_id)
        REFERENCES directory_groups (id) MATCH SIMPLE
        ON UPDATE RESTRICT
        ON DELETE RESTRICT
        NOT VALID
);

-- directory table
CREATE TABLE directory
(
    id uuid NOT NULL DEFAULT uuidv7_sub_ms(),
    directory_groups_id uuid NOT NULL,
    display_name text NOT NULL UNIQUE,
    data jsonb,
    created_on timestamp without time zone NOT NULL DEFAULT NOW(),
    last_updated_on timestamp without time zone NOT NULL DEFAULT NOW(),
    deactivated_on timestamp without time zone,
    full_text_search tsvector,
    PRIMARY KEY (id),
    CONSTRAINT directory_groups_id FOREIGN KEY (directory_groups_id)
        REFERENCES directory_groups (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS directory
    OWNER to pg_database_owner;

CREATE INDEX directory_full_text_search_index
    ON directory USING gin
    (full_text_search);

CREATE INDEX directory_data_jsonb_index
    ON directory USING gin
    (data);

-- directory trigger to update last_updated_on column
CREATE TRIGGER directory_update_last_updated_on
    BEFORE UPDATE OF display_name, data
    ON directory
    FOR EACH ROW
    EXECUTE FUNCTION update_last_updated_on();

COMMENT ON TRIGGER directory_update_last_updated_on ON directory
    IS 'update timestamp upon update on relevant columns';