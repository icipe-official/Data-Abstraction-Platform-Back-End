## Pre-requisites

1. [go](https://go.dev/dl/) version 1.21.1 or above.

## Environment variables

### Required

1. DOMAIN_URL: Required for cors.
2. PORT: Port that the server will run on.
3. PSQL_DBNAME: Postgres database name.
4. PSQL_HOST: Postgres database host.
5. PSQL_PORT: Postgres database port.
6. PSQL_USER: Postgres database user.
7. PSQL_PASS: Postgres database password.
8. PSQL_SSLMODE: Set to "disable".
9. PSQL_SCHEMA: Set to default "public" if schema being used is public. Used to generate models.
10. PSQL_DATABASE_DRIVE_NAME: Set to "postgres".
11. REDIS_HOST: Redis database host.
12. REDIS_PORT: Redis database port.
13. ACCESS_REFRESH_TOKEN: Secret key used to sign authentication jwt access-refresh token. Maximum length is 36 characters long.
14. ENCRYPTION_KEY: Secret key used to encrypt both the access token and refresh token. Key MUST be 16, 24, or 32 characters in length ONLY.
15. TMP_DIR: Directory to store temporary files.

### Optional

1. REDIS_SESSION_DB: Default is 15.
2. REDIS_USER: Redis database user (optional if database has no auth configured).
3. REDIS_PASSWORD: Redis database password (optional if database has no auth configured).
4. BASE_PATH: Subpath on which the API is hosted when using a shared domain.
5. TABLE_MODELS_DIRECTORY: Set to "$(pwd)/internal/pkg". Required if database schema has been modified.
6. GO_DEV: "true" or "false". Setting up authentication cookies in dev or production.
7. LOG_LEVEL: Default is 1 (debug, info, warning, and errors).
8. MAIL_HOST:
9. MAIL_PORT:
10. MAIL_USERNAME:
11. MAIL_PASSWORD:

## Instructions

### Populate database with necessary values

Necessary values include: Project Roles, Directory Iam Ticket Types, and Storage Types.

1. Set environment variables beginning with "PSQL\_".
2. Run the source file or executable to upload the defined values to the database.
   #### Option 1 (Recommended for production)
   a. Build the init_database cli tool and generate the executable.
   ```
   bash build_init_database.sh
   ```
   b. Run the executable and follow the command prompts.
   ```
   bin/init_database
   ```
   #### Option 2
   a. Run the source file directly and follow the command prompts (Requires a go environment already setup with the dependencies installed).
   ```
   go run cmd/init_database/main.go
   ```

### Create system user

1. Set environment variables beginning with "PSQL\_".
2. Run the source file or executable to create the supersuer.
   #### Option 1 (Recommended for production)
   a. Build the create_system_user cli tool and generate the executable.
   ```
   bash build_create_system_user.sh
   ```
   b. Run the executable and follow the command prompts.
   ```
   bin/create_system_user
   ```
   #### Option 2
   a. Run the source file directly and follow the command prompts (Requires a go environment already setup with the dependencies installed).
   ```
   go run cmd/create_system_user/main.go
   ```

### Build and run the api

1. Set required env variables as well as optional if necessary.

   Generate ACCESS_REFRESH_TOKEN (36) and ENCRYPTION_KEY (32) with openssl

   ```
   openssl rand -base64 <LENGTH_OF_KEY>
   ```

2. Download dependencies to cache.
   ```
   go mod download
   go mod tidy
   ```
3. Build api and generate executable.
   ```
   bash build_api.sh
   ```
4. Run the "api" generated executable in bin folder.
   ```
   bin/api
   ```

## Miscellaneous

### Generate/regenerate models and tables go

1. Set environment variables beginning with "PSQL" and the TABLES_MODELS_DIRECTORY env.
2. Execute command to generate/regenerate custom go code based on schema.
   ```
   go run cmd/gen_database_code/main.go
   ```
