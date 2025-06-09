-- group_rule_authorizations table
CREATE TABLE group_rule_authorizations
(
    id uuid NOT NULL DEFAULT uuidv7_sub_ms(),
    directory_groups_id uuid NOT NULL,
    group_authorization_rules_id text NOT NULL,
    group_authorization_rules_group text NOT NULL,
    created_on timestamp without time zone NOT NULL DEFAULT NOW(),
    deactivated_on timestamp without time zone,
    PRIMARY KEY (id),
    CONSTRAINT directory_groups_id FOREIGN KEY (directory_groups_id)
        REFERENCES directory_groups (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT group_authorization_rules_id FOREIGN KEY (group_authorization_rules_id, group_authorization_rules_group)
        REFERENCES group_authorization_rules (id, rule_group) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS group_rule_authorizations
    OWNER to pg_database_owner;

-- iam_group_authorizations table
CREATE TABLE iam_group_authorizations
(
    id uuid NOT NULL DEFAULT uuidv7_sub_ms(),
    iam_credentials_id uuid NOT NULL,
    group_rule_authorizations_id uuid NOT NULL,
    created_on timestamp without time zone NOT NULL DEFAULT NOW(),
    deactivated_on timestamp without time zone,
    PRIMARY KEY (id),
    CONSTRAINT iam_credentials_id FOREIGN KEY (iam_credentials_id)
        REFERENCES iam_credentials (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID,
    CONSTRAINT group_rule_authorizations_id FOREIGN KEY (group_rule_authorizations_id)
        REFERENCES group_rule_authorizations (id) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE RESTRICT
        NOT VALID
);

ALTER TABLE IF EXISTS iam_group_authorizations
    OWNER to pg_database_owner;