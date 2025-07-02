#!/bin/bash
set -e

host="${PGHOST:-db}"
port="${PGPORT:-5432}"
user="${PGUSER:-postgres}"

echo "Waiting for PostgreSQL at $host:$port as user $user..."
until pg_isready -h "$host" -p "$port" -U "$user"; do
  sleep 1
done

echo "PostgreSQL is up. Running command: $@"
exec "$@"
