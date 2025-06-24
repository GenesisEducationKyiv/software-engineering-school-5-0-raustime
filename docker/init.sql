-- docker/init.sql
SELECT 'CREATE DATABASE weatherdb_test'
WHERE NOT EXISTS (
  SELECT FROM pg_database WHERE datname = 'weatherdb_test'
)\gexec