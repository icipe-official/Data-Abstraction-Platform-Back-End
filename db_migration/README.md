## Pre-requisites
1. Postgres Database [(docker version here)](https://hub.docker.com/_/postgres).
2. [golang-migrate cli](https://github.com/golang-migrate/migrate/releases).
## Environment variables required
1. DATABSE_URL: eg. postgres://postgres:password@localhost:5432/example?sslmode=disable
## Run the migrations
1. Execute the migrate script.
```
bash migrate.sh
```
## Migrate CLI Commands
Generate migration file:
```
migrate create -ext sql -dir <MIGRATIONS_DIRECTORY> <MIGRATIONS_NAME>
```

Run migration:
```
migrate -path <MIGRATIONS_DIRECTORY> -database $DATABASE_URL up
```

Revert migration:
```
migrate -path <MIGRATIONS_DIRECTORY> -database $DATABASE_URL down <NO_OF_MIGRATIONS_TO_REVERT>
```

Fix dirty migrations:
```
migrate -path <MIGRATIONS_DIRECTORY> -database $DATABASE_URL force <VERSION_OF_MIGRATION>
```

## Linux cli install (replace <CLI_VERSION> with the desired cli to download e.g. v4.14.1)
1. curl -L https://github.com/golang-migrate/migrate/releases/download/<CLI_VERSION>/migrate.linux-amd64.tar.gz | tar xvz
2. move cli to desired location like the go bin path.