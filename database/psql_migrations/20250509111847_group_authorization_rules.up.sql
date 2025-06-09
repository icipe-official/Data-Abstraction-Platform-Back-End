-- group_authorization_rules table
CREATE TABLE group_authorization_rules
(
    id text NOT NULL,
    rule_group text NOT NULL,
    description text NOT NULL,
    created_on timestamp without time zone NOT NULL DEFAULT NOW(),
    last_updated_on timestamp without time zone NOT NULL DEFAULT NOW(),
    full_text_search tsvector,
    PRIMARY KEY (id, rule_group)
);

ALTER TABLE IF EXISTS group_authorization_rules
    OWNER to pg_database_owner;

-- trigger to update last_updated_on column
CREATE TRIGGER group_authorization_rules_update_last_updated_on
    BEFORE UPDATE OF id, rule_group, description
    ON group_authorization_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_last_updated_on();

COMMENT ON TRIGGER group_authorization_rules_update_last_updated_on ON group_authorization_rules
    IS 'update timestamp upon update on relevant columns';

-- full_text_search
CREATE INDEX group_authorization_rules_full_text_search_index
    ON group_authorization_rules USING gin
    (full_text_search);

-- function and trigger to update group_authorization_rules->full_text_search
CREATE FUNCTION group_authorization_rules_update_full_text_search_index()
    RETURNS trigger
    LANGUAGE 'plpgsql'
    NOT LEAKPROOF
AS $BODY$

DECLARE id text;
DECLARE rule_group text;
DECLARE description text;

BEGIN
	IF NEW.id IS DISTINCT FROM OLD.id THEN
		id = NEW.id;
	ELSE
		id = OLD.id;
	END IF;

    IF NEW.rule_group IS DISTINCT FROM OLD.rule_group THEN
		rule_group = NEW.rule_group;
	ELSE
		rule_group = OLD.rule_group;
	END IF;

    IF NEW.description IS DISTINCT FROM OLD.description THEN
		description = NEW.description;
	ELSE
		description = OLD.description;
	END IF;

    NEW.full_text_search = 
        setweight(to_tsvector(coalesce(id,'')),'A') ||
        setweight(to_tsvector(coalesce(rule_group,'')),'B') ||
        setweight(to_tsvector(coalesce(description,'')),'C');
    	
	RETURN NEW;   
END
$BODY$;

ALTER FUNCTION group_authorization_rules_update_full_text_search_index()
    OWNER TO pg_database_owner;

COMMENT ON FUNCTION group_authorization_rules_update_full_text_search_index()
    IS 'Update full_text_search column in group_authorization_rules when id, rule_group, description change';

CREATE TRIGGER group_authorization_rules_update_full_text_search_index
    BEFORE INSERT OR UPDATE OF id, rule_group, description
    ON group_authorization_rules
    FOR EACH ROW
    EXECUTE FUNCTION group_authorization_rules_update_full_text_search_index();

COMMENT ON TRIGGER group_authorization_rules_update_full_text_search_index ON group_authorization_rules
    IS 'trigger to update full_text_search column';




-- group_authorization_rules_tags table
CREATE TABLE group_authorization_rules_tags
(
    id text NOT NULL,
    group_authorization_rules_id text NOT NULL,
    group_authorization_rules_group text NOT NULL,
    created_on timestamp without time zone NOT NULL DEFAULT NOW(),
    last_updated_on timestamp without time zone NOT NULL DEFAULT NOW(),
    full_text_search tsvector,
    PRIMARY KEY (id),
    CONSTRAINT group_authorization_rules_id FOREIGN KEY (group_authorization_rules_id, group_authorization_rules_group)
        REFERENCES group_authorization_rules (id, rule_group) MATCH SIMPLE
        ON UPDATE CASCADE
        ON DELETE CASCADE
        NOT VALID
);

ALTER TABLE IF EXISTS group_authorization_rules_tags
    OWNER to pg_database_owner;

-- trigger to update last_updated_on column
CREATE TRIGGER group_authorization_rules_tags_update_last_updated_on
    BEFORE UPDATE OF id, group_authorization_rules_id, group_authorization_rules_group
    ON group_authorization_rules_tags
    FOR EACH ROW
    EXECUTE FUNCTION update_last_updated_on();

COMMENT ON TRIGGER group_authorization_rules_tags_update_last_updated_on ON group_authorization_rules_tags
    IS 'update timestamp upon update on relevant columns';

-- full_text_search
CREATE INDEX group_authorization_rules_tags_full_text_search_index
    ON group_authorization_rules_tags USING gin
    (full_text_search);

-- function and trigger to update group_authorization_rules_tags->full_text_search
CREATE FUNCTION group_authorization_rules_tags_update_full_text_search_index()
    RETURNS trigger
    LANGUAGE 'plpgsql'
    NOT LEAKPROOF
AS $BODY$

DECLARE id text;
DECLARE group_authorization_rules_id text;
DECLARE group_authorization_rules_group text;

BEGIN
	IF NEW.id IS DISTINCT FROM OLD.id THEN
		id = NEW.id;
	ELSE
		id = OLD.id;
	END IF;

    IF NEW.group_authorization_rules_id IS DISTINCT FROM OLD.group_authorization_rules_id THEN
		group_authorization_rules_id = NEW.group_authorization_rules_id;
	ELSE
		group_authorization_rules_id = OLD.group_authorization_rules_id;
	END IF;

    IF NEW.group_authorization_rules_group IS DISTINCT FROM OLD.group_authorization_rules_group THEN
		group_authorization_rules_group = NEW.group_authorization_rules_group;
	ELSE
		group_authorization_rules_group = OLD.group_authorization_rules_group;
	END IF;

    NEW.full_text_search = 
        setweight(to_tsvector(coalesce(id,'')),'A') ||
        setweight(to_tsvector(coalesce(group_authorization_rules_id,'')),'B') ||
        setweight(to_tsvector(coalesce(group_authorization_rules_group,'')),'C');
    	
	RETURN NEW;   
END
$BODY$;

ALTER FUNCTION group_authorization_rules_tags_update_full_text_search_index()
    OWNER TO pg_database_owner;

COMMENT ON FUNCTION group_authorization_rules_tags_update_full_text_search_index()
    IS 'Update full_text_search column in group_authorization_rules_tags when id, group_authorization_rules_id, group_authorization_rules_group change';

CREATE TRIGGER group_authorization_rules_tags_update_full_text_search_index
    BEFORE INSERT OR UPDATE OF id, group_authorization_rules_id, group_authorization_rules_group
    ON group_authorization_rules_tags
    FOR EACH ROW
    EXECUTE FUNCTION group_authorization_rules_tags_update_full_text_search_index();

COMMENT ON TRIGGER group_authorization_rules_tags_update_full_text_search_index ON group_authorization_rules_tags
    IS 'trigger to update full_text_search column';