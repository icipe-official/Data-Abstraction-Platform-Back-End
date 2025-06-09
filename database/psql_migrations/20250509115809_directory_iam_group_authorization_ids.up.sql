-- directory_groups_authorization_ids table
CREATE TABLE directory_groups_authorization_ids
(
    id uuid NOT NULL,
    creation_iam_group_authorizations_id uuid NOT NULL,
    deactivation_iam_group_authorizations_id uuid,
    PRIMARY KEY (id),
    CONSTRAINT source_id FOREIGN KEY (id)
        REFERENCES directory_groups (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
        NOT VALID,
    CONSTRAINT creation_iam_group_authorizations_id FOREIGN KEY (creation_iam_group_authorizations_id)
        REFERENCES iam_group_authorizations (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT deactivation_iam_group_authorizations_id FOREIGN KEY (deactivation_iam_group_authorizations_id)
        REFERENCES iam_group_authorizations (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS directory_groups_authorization_ids
    OWNER to pg_database_owner;




-- directory_authorization_ids table
CREATE TABLE directory_authorization_ids
(
    id uuid NOT NULL,
    creation_iam_group_authorizations_id uuid NOT NULL,
    deactivation_iam_group_authorizations_id uuid,
    PRIMARY KEY (id),
    CONSTRAINT source_id FOREIGN KEY (id)
        REFERENCES directory (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
        NOT VALID,
    CONSTRAINT creation_iam_group_authorizations_id FOREIGN KEY (creation_iam_group_authorizations_id)
        REFERENCES iam_group_authorizations (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT deactivation_iam_group_authorizations_id FOREIGN KEY (deactivation_iam_group_authorizations_id)
        REFERENCES iam_group_authorizations (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS directory_authorization_ids
    OWNER to pg_database_owner;




-- group_rule_authorizations_ids table
CREATE TABLE group_rule_authorizations_ids
(
    id uuid NOT NULL,
    creation_iam_group_authorizations_id uuid NOT NULL,
    deactivation_iam_group_authorizations_id uuid,
    PRIMARY KEY (id),
    CONSTRAINT source_id FOREIGN KEY (id)
        REFERENCES group_rule_authorizations (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
        NOT VALID,
    CONSTRAINT creation_iam_group_authorizations_id FOREIGN KEY (creation_iam_group_authorizations_id)
        REFERENCES iam_group_authorizations (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT deactivation_iam_group_authorizations_id FOREIGN KEY (deactivation_iam_group_authorizations_id)
        REFERENCES iam_group_authorizations (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS group_rule_authorizations_ids
    OWNER to pg_database_owner;




-- iam_group_authorizations_ids table
CREATE TABLE iam_group_authorizations_ids
(
    id uuid NOT NULL,
    creation_iam_group_authorizations_id uuid NOT NULL,
    deactivation_iam_group_authorizations_id uuid,
    PRIMARY KEY (id),
    CONSTRAINT source_id FOREIGN KEY (id)
        REFERENCES iam_group_authorizations (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
        NOT VALID,
    CONSTRAINT creation_iam_group_authorizations_id FOREIGN KEY (creation_iam_group_authorizations_id)
        REFERENCES iam_group_authorizations (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT deactivation_iam_group_authorizations_id FOREIGN KEY (deactivation_iam_group_authorizations_id)
        REFERENCES iam_group_authorizations (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS iam_group_authorizations_ids
    OWNER to pg_database_owner;