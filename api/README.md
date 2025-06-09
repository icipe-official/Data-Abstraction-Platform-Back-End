# API

This section contains information about the REST API for the `web_service`.

To run the requests, the following pre-requisites are needed:

1. bash/shell - Execute shell scripts.
2. curl - Execute http requests.
3. jq - Read json data.
4. Setup Enviornment variables.

The shell scripts created can be used to interact with the REST API of keycloak.

Online cURL converter (from cURL command to a language like JavaScript or JSON) can be found [here](https://curlconverter.com/).

# Table of Contents

- [Setup Environment Variables](#environment-variables)

- REST APIs.

  - [Iam](./iam/README.md) - Routes related to Identity and Access Management.

# Environment variables

Before the scripts are ran, a couple of environment variables need to be set in the following way.

1. Create a copy of the [env.sh](./env.sh.template) template script and name it `env.sh`.
2. Edit it appropriately.

This script is executed by [init_env.sh](./init_env.sh) which is executed before any of the API shell scripts are executed.
