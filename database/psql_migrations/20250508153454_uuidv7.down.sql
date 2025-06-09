-- 20250508153454

DROP FUNCTION IF EXISTS uuidv7(timestamptz);
DROP FUNCTION IF EXISTS uuidv7_sub_ms(timestamptz);
DROP FUNCTION IF EXISTS uuidv7_extract_timestamp(uuid);
DROP FUNCTION IF EXISTS uuidv7_boundary(timestamptz);
DROP EXTENSION IF EXISTS "uuid-ossp";