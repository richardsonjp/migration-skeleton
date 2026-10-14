# Migration Skeleton

Template for a Go CLI that runs PostgreSQL migrations per schema with
[golang-migrate](https://github.com/golang-migrate/migrate). Each schema has its own
migration directory and its own `schema_migrations` table.

```
cmd/apiserver/
├── main.go                  # cobra entry point (`migrate` command)
└── app/
    ├── migrator.go          # parses up/down/version/force + --schema
    ├── db/db_service.go     # golang-migrate wrapper
    └── migrations/
        └── <schema>/        # one directory per schema
            ├── 000001_name.up.sql
            └── 000001_name.down.sql
config/config.go             # env / .env loading
```

## Using this template

```bash
cp -r migration-skeleton acme-migration && cd acme-migration
rm -rf .git && git init

# rename the Go module (optional)
go mod edit -module acme-migration
grep -rl 'migration-skeleton' --include='*.go' --include=Makefile . | xargs sed -i '' 's/migration-skeleton/acme-migration/g'   # drop the '' on Linux

cp .env.example .env         # fill in DB credentials and DB_SCHEMAS
rm -rf cmd/apiserver/app/migrations/app   # remove the example schema
make migration-create NAME=base SCHEMA=acme
```

Requires Go 1.21+ and PostgreSQL.

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_USERNAME` | Database user | - |
| `DB_PASSWORD` | Database password (special characters are fine) | - |
| `DB_NAME` | Database name | - |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_SSL_MODE` | `sslmode` for the connection | `disable` |
| `DB_SCHEMAS` | Comma-separated schemas to manage | `public` |

Variables are read from the environment, or from `.env` if present.

## Commands

Without `SCHEMA`, every command runs for all schemas in `DB_SCHEMAS`, in order.

| Make | Go |
|------|----|
| `make migration-up [SCHEMA=x] [N=1]` | `go run ./cmd/apiserver migrate up [--schema=x] [N]` |
| `make migration-down [SCHEMA=x] [N=1]` | `go run ./cmd/apiserver migrate down [--schema=x] [N]` |
| `make migration-version [SCHEMA=x]` | `go run ./cmd/apiserver migrate version [--schema=x]` |
| `make migration-force V=3 [SCHEMA=x]` | `go run ./cmd/apiserver migrate force [--schema=x] 3` |
| `make migration-create NAME=add_users SCHEMA=x` | creates the next numbered `.up.sql` / `.down.sql` pair |

`--schema` must come after the command and before positional arguments.
`down` without `N` rolls back **everything** in the schema.

## How schemas work

- `cmd/apiserver/app/migrations/<schema>/` holds the migrations for `<schema>`.
- The schema is created (`CREATE SCHEMA IF NOT EXISTS`) before migrating.
- The connection uses `search_path=<schema>,public`, so unqualified names land in
  `<schema>` and `<schema>.schema_migrations` tracks its version.

## Writing migrations

- Prefer idempotent DDL (`CREATE TABLE IF NOT EXISTS`, `ADD COLUMN IF NOT EXISTS`).
- Every `.up.sql` needs a `.down.sql` that reverses it.
- Never edit a migration that has already run somewhere; add a new one.
- A failed migration leaves the schema dirty: fix the SQL, then
  `make migration-force V=<last good version> SCHEMA=x` and run up again.
