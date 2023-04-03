# pgmigrate

Wrapper for https://github.com/golang-migrate/migrate but only for PostgreSQL, and only as standalone CLI.

Added features:

- List pending migrations before actually applying them to the DB.
- Read connection string from [.pg_service.conf](https://www.postgresql.org/docs/current/libpq-pgservice.html).

Usage:

```sh
pgmigrate <pgservice_name> <migrations_dir>
```
