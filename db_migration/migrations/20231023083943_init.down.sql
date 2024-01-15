-- 20231023083943
DROP EXTENSION IF EXISTS "uuid-ossp";
DROP EXTENSION IF EXISTS pgcrypto;
DROP FUNCTION IF EXISTS public.update_last_updated_on();
DROP FUNCTION IF EXISTS public.gen_random_string();