-- 20231023084410

DROP TABLE IF EXISTS public.directory_iam;
DROP TABLE IF EXISTS public.directory_iam_ticket_types;
DROP TABLE IF EXISTS public.directory_system_users;
DROP TABLE IF EXISTS public.directory;
DROP FUNCTION IF EXISTS public.directory_gen_pin_salt();
DROP FUNCTION IF EXISTS public.directory_gen_password_salt();
DROP FUNCTION IF EXISTS public.directory_update_vector();