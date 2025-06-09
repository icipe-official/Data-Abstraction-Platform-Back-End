#!/bin/bash

set -e # terminate script if commands fail

bash scripts/golang_migrate/up.sh
bash scripts/cmd_app_init_database.sh

bin/web_service