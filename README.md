# Backend

This project contains the backend of the Data Abstraction Platform.

It is comprised of a collection of applications with the main goal is to run a http server which will offer API services. Contains a set of applications that serve different purposes as follows:

- [cmd_app_create_super_user](cmd/cmd_app_create_super_user/main.go) - cli app to create a system user with the necessary system administration roles in the system.
- [cmd_app_init_database](cmd/cmd_app_init_database/main.go) - cli app to initialize database with default values.
- [job_service](cmd/job_service/main.go) - service that combines all job services into one. 
- [web_service](cmd/web_service/main.go) - http server that combines all web api services into one.

This section contains information about setting up and running the various backend applications.

# Table of Contents

1. [Environment Variables](#environment-variables)
2. [Development](#development)
   1. [Setting Environment Variables](#set-environment-variables)
   2. [Database Migrations](#database-migrations)
      - [Create New Migrations](#create-new-migrations)
      - [Up migrations](#up-migrations)
      - [Down migrations](#down-migrations)
      - [Fix dirty migrations](#fix-dirty-migrations)
   3. [Applications](#applications)
      - [Web Service](#applications-web-service)
      - [Cmd App Create Super User](#applications-cmd-app-create-super-user)
      - [Cmd App Init Database](#applications-cmd-app-init-database)
3. [Containerization](#containerization)
   - [Web Service](#containerization-web-service)

# Containerization

### Containerization-Web Service

The [script](scripts/build/web_service_container_image.sh) builds a `data_abstraction_platform/web_service` container using this [`Dockerfile`](./build/Dockerfile.web_service).

The script accepts the following flags during execution.

<table>
    <thead>
        <th>Flag</th>
        <th>Example</th>
        <th>Default</th>
        <th>Purpose</th>
    </thead>
    <tbody>
        <tr>
            <td>-c</td>
            <td><code>podman</code></td>
            <td><code>docker</code></td>
            <td>
                <div>Container engine to use</div>
                <div>Optional</div>
            </td>
        </tr>
        <tr>
            <td>-t</td>
            <td><code>v1alpha3</code></td>
            <td><code>latest</code></td>
            <td>
                <div>Container Tag</div>
                <div>Optional</div>
            </td>
        </tr>
    </tbody>
</table>

Example:

```sh
#!/bin/bash

CONTAINER_TAG="v4.18.3"
CONTAINER_ENGINE="podman"

bash scripts/build/web_service_container_image.sh -t $CONTAINER_TAG -c $CONTAINER_ENGINE
```

# Development

The following pre-requisites must be installed/already setup:

- [Go](../docs/setup/go.md).
- [Postgres](../docs/setup/postgres.md).
- [OpenID Provider](../docs/setup/keycloak/README.md)
- Storage - [Local Folder](#local) or [S3](../docs/setup/s3.md).
- [Container Engine](../docs/setup/README.md#container-engines) - Optional. For building and runing container images.

**NB. For consistency, execute shell scripts at the root of the backend project.**

## Applications

Some [environment variables](#environment-variables) are required to be setup before each application runs otherwise it will crash.

Before the applications can be used, dependencies need to be downloaded.

```sh
#!/bin/bash

go mod tidy
```

#### Applications-Web Service

You can build and run the [`web_service`](cmd/web_service/main.go) application in the following ways:

1. Build+run. Useful for quick iterations

```sh
#!/bin/bash

go run cmd/web_service/main.go
```

2. Build alone using the [script](scripts/build/web_service.sh). Then run the application with the command: `bin/web_service`.

```sh
#!/bin/bash

bash scripts/build/web_service.sh
```

The application uses one port which defaults to `5174`.

#### Applications-Cmd App Create Super User

You can build and run the [`cmd_app_create_super_user`](cmd/cmd_app_create_super_user/main.go) application in the following ways:

1. Build+run. Useful for quick iterations

```sh
#!/bin/bash

go run cmd/cmd_app_create_super_user/main.go
```

2. Build alone using the [script](scripts/build/cmd_app_create_super_user.sh). Then run the application with the command: `bin/cmd_app_create_super_user`.

```sh
#!/bin/bash

bash scripts/build/cmd_app_create_super_user.sh
```

#### Applications-Cmd App Init Database

You can build and run the [`cmd_app_init_database`](cmd/cmd_app_init_database/main.go) application in the following ways:

1. Build+run. Useful for quick iterations

```sh
#!/bin/bash

go run cmd/cmd_app_init_database/main.go
```

2. Build alone using the [script](scripts/build/cmd_app_init_database.sh). Then run the application with the command: `bin/cmd_app_init_database`.

```sh
#!/bin/bash

bash scripts/build/cmd_app_init_database.sh
```

## Database Migrations

Before any of the services are ran, database migrations need to be executed against a postgres database.

Currently, the tool used to do so is called [golang migrate](https://github.com/golang-migrate/migrate).

This [script](scripts/golang_migrate/download.sh) can be used to download the cli tool into a `bin` folder (will automatically create if it does not exist) in the project.

Requires `tar` command to be available.

The script accepts the following flags during execution, refer to the [releases page](https://github.com/golang-migrate/migrate/releases) for more information.

<table>
    <thead>
        <th>Flag</th>
        <th>Example</th>
        <th>Accepted Values</th>
        <th>Purpose</th>
    </thead>
    <tbody>
        <tr>
            <td>-v</td>
            <td>v4.18.3</td>
            <td></td>
            <td>CLI Tool Version</td>
        </tr>
        <tr>
            <td>-o</td>
            <td>linux</td>
            <td>
                <div>darwin</div>
                <div>linux</div>
                <div>windows</div>
            </td>
            <td>Operating System</td>
        </tr>
        <tr>
            <td>-a</td>
            <td>amd64</td>
            <td>
                <div>amd64</div>
                <div>armv7</div>
                <div>386</div>
            </td>
            <td>Computer CPU Architecture</td>
        </tr>
    </tbody>
</table>

Below is a sample execution of the said script to be installed in a linux based machine that uses x86 architecture.

```sh
#!/bin/bash

VERSION="v4.18.3"
OS="linux"
ARCH="amd64"

bash scripts/download.sh -v $VERSION -o $OS -a $ARCH
```

All postgres database migrations are stored in [`database/psql_migrations/`](database/psql_migrations/) folder.

Once the cli tool has been installed, you can run it directly or use the helper scripts found [here](scripts/golang_migrate/).

#### Create new migrations

The [script](scripts/golang_migrate/create.sh) accepts the arguments and flags below:

<table>
    <thead>
        <th>Argument</th>
        <th>Example</th>
        <th>Purpose</th>
    </thead>
    <tbody>
        <tr>
            <td>$1</td>
            <td><code>table_name</code></td>
            <td>
                <div>Required</div>
                <div>name of migration</div>
                <div>no spaces</div>
                <div>Pass as first argument</div>
            </td>
        </tr>
    </tbody>
</table>

<table>
    <thead>
        <th>Flag</th>
        <th>Example</th>
        <th>Purpose</th>
    </thead>
    <tbody>
        <tr>
            <td>-m</td>
            <td><code>migrate</code></td>
            <td>
                <div>Optional. Will default to using the downloaded migrate tool in the <code>bin</code> folder.</div>
                <div>Path to the migrate cli tool.</div>
            </td>
        </tr>
    </tbody>
</table>

Example:

```sh
#!/bin/bash

MIGRATION_NAME="test_name"

bash scripts/golang_migrate/create.sh $MIGRATION_NAME

```

#### Up migrations

The [script](scripts/golang_migrate/up.sh) accepts the arguments and flags below:

<table>
    <thead>
        <th>Argument</th>
        <th>Example</th>
        <th>Purpose</th>
    </thead>
    <tbody>
        <tr>
            <td>$1</td>
            <td><code>2</code></td>
            <td>
                <div>Number of <code>up</code> migrations to run.</div>
                <div>Optional. Will run all remaining <code>up</code> migrations</div>
            </td>
        </tr>
    </tbody>
</table>

<table>
    <thead>
        <th>Flag</th>
        <th>Example</th>
        <th>Purpose</th>
    </thead>
    <tbody>
        <tr>
            <td>-m</td>
            <td><code>migrate</code></td>
            <td>
                <div>Optional. Will default to using the downloaded migrate tool in the <code>bin</code> folder.</div>
                <div>Path to the migrate cli tool.</div>
            </td>
        </tr>
    </tbody>
</table>

Example:

```sh
#!/bin/bash

NO_OF_MIGRATIONS="2"

bash scripts/golang_migrate/up.sh $NO_OF_MIGRATIONS

```

#### Down migrations

The [script](scripts/golang_migrate/down.sh) accepts the arguments and flags below:

<table>
    <thead>
        <th>Argument</th>
        <th>Example</th>
        <th>Purpose</th>
    </thead>
    <tbody>
        <tr>
            <td>$1</td>
            <td><code>2</code></td>
            <td>
                <div>Number of <code>down</code> migrations to run.</div>
                <div>Optional. Will run all remaining <code>down</code> migrations</div>
            </td>
        </tr>
    </tbody>
</table>

<table>
    <thead>
        <th>Flag</th>
        <th>Example</th>
        <th>Purpose</th>
    </thead>
    <tbody>
        <tr>
            <td>-m</td>
            <td><code>migrate</code></td>
            <td>
                <div>Optional. Will default to using the downloaded migrate tool in the <code>bin</code> folder.</div>
                <div>Path to the migrate cli tool.</div>
            </td>
        </tr>
    </tbody>
</table>

Example:

```sh
#!/bin/bash

MIGRATION_COUNT="3"

bash scripts/golang_migrate/down.sh $MIGRATION_COUNT

```

#### Fix dirty migrations

The [script](scripts/golang_migrate/force.sh) accepts the arguments and flags below:

<table>
    <thead>
        <th>Argument</th>
        <th>Example</th>
        <th>Purpose</th>
    </thead>
    <tbody>
        <tr>
            <td>$1</td>
            <td><code>20250402092853</code></td>
            <td>New migration version</td>
        </tr>
    </tbody>
</table>

<table>
    <thead>
        <th>Flag</th>
        <th>Example</th>
        <th>Purpose</th>
    </thead>
    <tbody>
        <tr>
            <td>-m</td>
            <td><code>migrate</code></td>
            <td>
                <div>Optional. Will default to using the downloaded migrate tool in the <code>bin</code> folder.</div>
                <div>Path to the migrate cli tool.</div>
            </td>
        </tr>
    </tbody>
</table>

Example:

```sh
#!/bin/bash

MIGRATION_VERSION="20250402092853"

bash scripts/golang_migrate/force.sh $MIGRATION_VERSION
```

## Set Environment Variables

Before any cli tool or service is ran, ensure the required environment variables are already set.

To quickly set the environment variables in the current terminal session, you can use the [template bash script](configs/env.sh.template). DO NOT edit the script. Instead copy it into a file such as `env.sh`, edit it appropriately the execute it in a shell/bash terminal e.g. `source env.sh`.

# Environment variables

## WEB_SERVICE

Used during startup of web services.

<table>
    <thead>
        <th>Name/Key</th>
        <th>Example</th>
        <th>Default</th>
        <th>Description</th>
        <th>Used In</th>
    </thead>
    <tbody>
        <tr>
            <td>WEB_SERVICE_CORS_URLS</td>
            <td>
                <div><code>http://0.0.0.0:5173 https://dap.icipe.org</code></div>
                <div><code>http://0.0.0.0:5173</code></div>
            </td>
            <td></td>
            <td>
                <div>List of server accepted origns.</div>
                <div>Separated by space.</div>
            </td>
            <td>
                <div>web_service</div>
            </td>
        </tr>
        <tr>
            <td>WEB_SERVICE_APP_PREFIX</td>
            <td><code>data_abstraction_platform</code></td>
            <td><code>data_abstraction_platform</code></td>
            <td>Used for purposes such as prefixing session id keys in the cache database.</td>
            <td>
                <div>web_service</div>
            </td>
        </tr>
        <tr>
            <td>WEB_SERVICE_PORT</td>
            <td><code>5174</code></td>
            <td><code>5174</code></td>
            <td>Port the web service should run on. </td>
            <td>
                <div>web_service</div>
            </td>
        </tr>
        <tr>
            <td>WEB_SERVICE_BASE_PATH</td>
            <td><code>/dap</code></td>
            <td><code>/</code></td>
            <td>Base path to listen for http requests.</td>
            <td>
                <div>web_service</div>
            </td>
        </tr>
    </tbody>
</table>

## PSQL

Used for connecting to a postgres database.

<table>
    <thead>
        <th>Name/Key</th>
        <th>Example</th>
        <th>Default</th>
        <th>Description</th>
        <th>Used In</th>
    </thead>
    <tbody>
        <tr>
            <td>PSQL_USER</td>
            <td><code>postgres</code></td>
            <td></td>
            <td>Postgres database user.</td>
            <td>
                <div>web_service</div>
                <div>cmd_app_create_super_user</div>
                <div>cmd_app_init_database</div>
                <div>golang migrate</div>
                <div>job_service</div>
            </td>
        </tr>
        <tr>
            <td>PSQL_PASSWORD</td>
            <td><code>postgres2025</code></td>
            <td></td>
            <td>Postgres database password</td>
            <td>
                <div>web_service</div>
                <div>cmd_app_create_super_user</div>
                <div>cmd_app_init_database</div>
                <div>golang migrate</div>
                <div>job_service</div>
            </td>
        </tr>
        <tr>
            <td>PSQL_HOST</td>
            <td><code>10.88.0.100</code></td>
            <td></td>
            <td>Postgres database host</td>
            <td>
                <div>web_service</div>
                <div>cmd_app_create_super_user</div>
                <div>cmd_app_init_database</div>
                <div>golang migrate</div>
                <div>job_service</div>
            </td>
        </tr>
        <tr>
            <td>PSQL_PORT</td>
            <td><code>5432</code></td>
            <td></td>
            <td>Postgres database port</td>
            <td>
                <div>web_service</div>
                <div>cmd_app_create_super_user</div>
                <div>cmd_app_init_database</div>
                <div>golang migrate</div>
                <div>job_service</div>
            </td>
        </tr>
        <tr>
            <td>PSQL_DATABASE</td>
            <td><code>data_abstraction_platform</code></td>
            <td></td>
            <td>Postgres database</td>
            <td>
                <div>web_service</div>
                <div>cmd_app_create_super_user</div>
                <div>cmd_app_init_database</div>
                <div>golang migrate</div>
            </td>
        </tr>
        <tr>
            <td>PSQL_SCHEMA</td>
            <td><code>public</code></td>
            <td><code>public</code></td>
            <td>Postgres database schema to use</td>
            <td>
                <div>web_service</div>
                <div>cmd_app_create_super_user</div>
                <div>cmd_app_init_database</div>
                <div>golang migrate</div>
                <div>job_service</div>
            </td>
        </tr>
        <tr>
            <td>PSQL_SEARCH_PARAMS</td>
            <td><code>sslmode=disable</code></td>
            <td></td>
            <td>
                <div>Optional</div>
                <div>List of postgres search params.</div>
                <div>Separated by space.</div>
                <div>golang migrate</div>
                <div>job_service</div>
            </td>
            <td>
                <div>web_service</div>
                <div>cmd_app_create_super_user</div>
                <div>cmd_app_init_database</div>
                <div>golang migrate</div>
                <div>job_service</div>
            </td>
        </tr>
    </tbody>
</table>

## Storage

Will look at S3 config first before local.

### Local

<table>
    <thead>
        <th>Name/Key</th>
        <th>Example</th>
        <th>Default</th>
        <th>Description</th>
        <th>Used In</th>
    </thead>
    <tbody>
        <tr>
            <td>STORAGE_LOCAL_FOLDER_PATH</td>
            <td>
                <div><code>/mnt/data</code></div>
            </td>
            <td></td>
            <td>Path to folder to store files</td>
            <td>
                <div>web_service</div>
                <div>job_service</div>
            </td>
        </tr>   
    </tbody>
</table>

### S3

<table>
    <thead>
        <th>Name/Key</th>
        <th>Example</th>
        <th>Default</th>
        <th>Description</th>
        <th>Used In</th>
    </thead>
    <tbody>
        <tr>
            <td>STORAGE_S3_ENDPOINT</td>
            <td>
                <div><code>play.min.io</code></div>
                <div><code>10.88.0.60:9000</code></div>
            </td>
            <td></td>
            <td></td>
            <td>
                <div>web_service</div>
                <div>job_service</div>
            </td>
        </tr>
        <tr>
            <td>STORAGE_S3_ACCESS_KEY</td>
            <td></td>
            <td></td>
            <td></td>
            <td>
                <div>web_service</div>
                <div>job_service</div>
            </td>
        </tr>
        <tr>
            <td>STORAGE_S3_SECRET_KEY</td>
            <td></td>
            <td></td>
            <td></td>
            <td>
                <div>web_service</div>
                <div>job_service</div>
            </td>
        </tr>
        <tr>
            <td>STORAGE_S3_USE_SSL</td>
            <td><code>true</code></td>
            <td><code>true</code></td>
            <td></td>
            <td>
                <div>web_service</div>
                <div>job_service</div>
            </td>
        </tr>
        <tr>
            <td>STORAGE_S3_BUCKET</td>
            <td><code>data-abstraction-platform</code></td>
            <td><code>data-abstraction-platform</code></td>
            <td>Bucket service will use to store its files.</td>
            <td>
                <div>web_service</div>
                <div>job_service</div>
            </td>
        </tr>        
    </tbody>
</table>

## IAM

Used for user authentication.

<table>
    <thead>
        <th>Name/Key</th>
        <th>Example</th>
        <th>Default</th>
        <th@>Description</th>
        <th>Used In</th>
    </thead>
    <tbody>
        <tr>
            <td>IAM_ENCRYPTION_KEY</td>
            <td><code>9lWLZlzJBCO4xuWe9hrLD97oI87EBdlL</code></td>
            <td></td>
            <td>
                <div>Random sequence of characters</div>
                <div>Used for encrypting data like JWT tokens.</div>
                <div>Key MUST be 16, 24, or 32 characters in length ONLY.</div>
            </td>
            <td>
                <div>web_service</div>
            </td>
        </tr>
         <tr>
            <td>IAM_SIGNING_KEY</td>
            <td><code>9lWLZlzJBCO4xuWe9hrLD97oI87EBdlL</code></td>
            <td></td>
            <td>
                <div>Random sequence of characters</div>
                <div>Used for signing JWT tokens.</div>
                <div>Key MUST be 16, 24, or 32 characters in length ONLY.</div>
            </td>
            <td>
                <div>web_service</div>
            </td>
        </tr>
        <tr>
            <td>IAM_COOKIE_HTTP_ONLY</td>
            <td><code>true</code></td>
            <td><code>true</code></td>
            <td></td>
            <td>
                <div>web_service</div>
            </td>
        </tr>
        <tr>
            <td>IAM_COOKIE_SAME_SITE</td>
            <td><code>1</code></td>
            <td><code>3</code></td>
            <td>
                <div>Accepted options:</div>
                <table>
                    <thead>
                        <th>Value</th>
                        <th>Represents</th>
                    </thead>
                    <tbody>
                        <tr>
                            <td>1</td>
                            <td>SameSiteDefaultMode</td>
                        </tr>
                        <tr>
                            <td>2</td>
                            <td>SameSiteLaxMode</td>
                        </tr>
                        <tr>
                            <td>3</td>
                            <td>SameSiteStrictMode</td>
                        </tr>
                        <tr>
                            <td>4</td>
                            <td>SameSiteNoneMode</td>
                        </tr>
                    </tbody>
                </table>
            </td>
            <td>
                <div>web_service</div>
            </td>
        </tr>
        <tr>
            <td>IAM_COOKIE_SECURE</td>
            <td><code>true</code></td>
            <td><code>true</code></td>
            <td>
                <div>Send cookie to server only if connection is secure</div>
            </td>
            <td>
                <div>web_service</div>
            </td>
        </tr>
        <tr>
            <td>IAM_COOKIE_DOMAIN</td>
            <td><code>localhost</code></td>
            <td></td>
            <td>
                <div>Match domain that hosts the website</div>
            </td>
            <td>
                <div>web_service</div>
            </td>
        </tr>
    </tbody>
</table>

## TELEMETRY

Setting up telemetry infrastructure.

<table>
    <thead>
        <th>Name/Key</th>
        <th>Example</th>
        <th>Default</th>
        <th>Description</th>
        <th>Used In</th>
    </thead>
    <tbody>
        <tr>
            <td>TELEMETRY_LOG_LEVEL</td>
            <td><code>0</code></td>
            <td><code>1</code></td>
            <td>
                <div>Level of detail of logs generated.</div>
                <table>
                    <thead>
                        <th>Range</th>
                        <th>Meaning</th>
                    </thead>
                    <tbody>
                        <tr>
                            <td>-4 to -1</td>
                            <td>debug</td>
                        </tr>
                        <tr>
                            <td>0 to 3</td>
                            <td>info</td>
                        </tr>
                        <tr>
                            <td>4 to 7</td>
                            <td>warning</td>
                        </tr>
                        <tr>
                            <td>8</td>
                            <td>error</td>
                        </tr>
                    </tbody>
                </table>
            </td>
            <td>
                <div>web_service</div>
                <div>cmd_app_create_super_user</div>
                <div>cmd_app_init_database</div>
                <div>job_service</div>
            </td>
        </tr>
        <tr>
            <td>TELEMETRY_LOG_USE_JSON</td>
            <td><code>false</code></td>
            <td><code>false</code></td>
            <td>Generate logs in JSON form</td>
            <td>
                <div>web_service</div>
                <div>cmd_app_create_super_user</div>
                <div>cmd_app_init_database</div>
                <div>job_service</div>
            </td>
        </tr>
        <tr>
            <td>TELEMETRY_LOG_COINCISE</td>
            <td><code>true</code></td>
            <td><code>true</code></td>
            <td>Emit non-detailed logs which excludes info like some http request details.</td>
            <td>
                <div>web_service</div>
                <div>cmd_app_create_super_user</div>
                <div>cmd_app_init_database</div>
                <div>job_service</div>
            </td>
        </tr>
        <tr>
            <td>TELEMETRY_LOG_REQUEST_HEADERS</td>
            <td><code>true</code></td>
            <td><code>true</code></td>
            <td>Logs should include http request details.</td>
            <td>
                <div>web_service</div>
                <div>cmd_app_create_super_user</div>
                <div>cmd_app_init_database</div>
                <div>job_service</div>
            </td>
        </tr>
        <tr>
            <td>TELEMETRY_LOG_APP_VERSION</td>
            <td><code>true</code></td>
            <td><code>true</code></td>
            <td>Version of the deployed applications.</td>
            <td>
                <div>web_service</div>
                <div>cmd_app_create_super_user</div>
                <div>cmd_app_init_database</div>
                <div>job_service</div>
            </td>
        </tr>
    </tbody>
</table>
