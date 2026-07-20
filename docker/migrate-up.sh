#!/bin/sh
# Applies pending migrations using the discrete DB_* env vars the app
# itself already requires, so there is one source of truth for connection
# settings instead of a second DATABASE_URL to keep in sync.
set -eu

: "${DB_HOST:?DB_HOST is required}"
: "${DB_PORT:?DB_PORT is required}"
: "${DB_USER:?DB_USER is required}"
: "${DB_PASSWORD:?DB_PASSWORD is required}"
: "${DB_NAME:?DB_NAME is required}"

DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

if [ "$#" -eq 0 ]; then
  set -- up
fi

exec ./migrate -path ./migrations -database "$DATABASE_URL" "$@"
