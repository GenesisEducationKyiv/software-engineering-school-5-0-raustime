#!/bin/sh

: "${PGHOST:=db}"
: "${PGPORT:=5432}"
: "${PGUSER:=postgres}"

echo "Waiting for postgres at $PGHOST:$PGPORT as user $PGUSER..."

until pg_isready -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" > /dev/null 2>&1; do
  echo "Waiting for postgres..."
  sleep 2
done

echo "Postgres is up - executing command"

# Перевірка чи існує база weatherdb_test
exists=$(psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -tAc "SELECT 1 FROM pg_database WHERE datname='weatherdb_test'")

if [ "$exists" != "1" ]; then
  echo "Creating database weatherdb_test..."
  createdb -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" weatherdb_test
else
  echo "Database weatherdb_test already exists."
fi

# Запускаємо команду, передану скрипту (наприклад go test)
exec "$@"