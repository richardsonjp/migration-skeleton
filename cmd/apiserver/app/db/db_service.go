package db

import (
	"database/sql"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lib/pq"

	"migration-skeleton/config"
)

// VersionInfo is the migration state of a single schema.
type VersionInfo struct {
	Version uint
	Dirty   bool
	Error   error
}

// migrationSourceURL builds the file:// source URL for a schema's migration
// directory. filepath.Join emits OS-native separators, so on Windows the raw
// result would be "cmd\apiserver\..." and url.Parse rejects it with
// `invalid character "\\" in host name`. ToSlash normalises to forward slashes,
// which the golang-migrate file driver resolves back to a relative path (it
// concatenates host + path), so the URL stays valid on every platform.
func migrationSourceURL(schema string) string {
	return "file://" + filepath.ToSlash(filepath.Join("cmd/apiserver/app/migrations", schema))
}

// databaseURL builds the postgres URL. url.UserPassword escapes credentials so
// passwords containing '@', ':' or '/' don't break parsing.
func databaseURL(searchPath string) string {
	cfg := config.Config.DB
	q := url.Values{"sslmode": {cfg.SSLMode}}
	if searchPath != "" {
		q.Set("search_path", searchPath)
	}
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.Username, cfg.Password),
		Host:     cfg.Host + ":" + cfg.Port,
		Path:     "/" + cfg.Name,
		RawQuery: q.Encode(),
	}
	return u.String()
}

// ensureSchema creates the schema if missing. Without it, a missing schema makes
// search_path fall through to public, so schema_migrations is created in public
// and silently shared by every schema.
func ensureSchema(schema string) error {
	conn, err := sql.Open("postgres", databaseURL(""))
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Exec("CREATE SCHEMA IF NOT EXISTS " + pq.QuoteIdentifier(schema))
	return err
}

// newMigrate returns a migrator for the schema; the caller must Close it.
func newMigrate(schema string) (*migrate.Migrate, error) {
	if err := ensureSchema(schema); err != nil {
		return nil, fmt.Errorf("ensure schema %s failed: %w", schema, err)
	}

	m, err := migrate.New(migrationSourceURL(schema), databaseURL(schema+",public"))
	if err != nil {
		return nil, fmt.Errorf("migration init failed: %w", err)
	}
	return m, nil
}

// RunUp runs pending migrations in the specified schema.
// If steps > 0, it runs that many migrations. Otherwise it runs all.
func RunUp(schema string, steps int) error {
	m, err := newMigrate(schema)
	if err != nil {
		return err
	}
	defer m.Close()

	if steps > 0 {
		if err := m.Steps(steps); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("migrate up %d steps failed: %w", steps, err)
		}
	} else {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("migrate up failed: %w", err)
		}
	}

	version, dirty, _ := m.Version()
	if dirty {
		return fmt.Errorf("migration version %d is in a dirty state", version)
	}

	return nil
}

// RunUpAll runs migrations for all schemas defined in config
func RunUpAll(steps int) error {
	for _, schema := range config.Config.GetSchemas() {
		fmt.Printf("Running migrations for schema: %s\n", schema)
		if err := RunUp(schema, steps); err != nil {
			return fmt.Errorf("failed to run migrations for schema %s: %w", schema, err)
		}
		fmt.Printf("Successfully completed migrations for schema: %s\n", schema)
	}
	return nil
}

// RunDown rolls back migrations in the specified schema.
// If steps > 0, it rolls back that many migrations. Otherwise it rolls back all.
func RunDown(schema string, steps int) error {
	m, err := newMigrate(schema)
	if err != nil {
		return err
	}
	defer m.Close()

	if steps > 0 {
		if err := m.Steps(-steps); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("migrate down %d steps failed: %w", steps, err)
		}
	} else {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("migrate down failed: %w", err)
		}
	}

	return nil
}

// RunDownAll rolls back migrations for all schemas defined in config
func RunDownAll(steps int) error {
	for _, schema := range config.Config.GetSchemas() {
		fmt.Printf("Running down migrations for schema: %s\n", schema)
		if err := RunDown(schema, steps); err != nil {
			return fmt.Errorf("failed to run down migrations for schema %s: %w", schema, err)
		}
		fmt.Printf("Successfully rolled back migrations for schema: %s\n", schema)
	}
	return nil
}

// GetVersion returns the current migration version for the specified schema
func GetVersion(schema string) (uint, bool, error) {
	m, err := newMigrate(schema)
	if err != nil {
		return 0, false, err
	}
	defer m.Close()

	return m.Version()
}

// GetAllVersions returns the current migration versions for all schemas
func GetAllVersions() map[string]VersionInfo {
	result := make(map[string]VersionInfo)
	for _, schema := range config.Config.GetSchemas() {
		version, dirty, err := GetVersion(schema)
		result[schema] = VersionInfo{Version: version, Dirty: dirty, Error: err}
	}
	return result
}

// Force sets the migration version for the specified schema
func Force(version int, schema string) error {
	m, err := newMigrate(schema)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Force(version); err != nil {
		return fmt.Errorf("force migration failed: %w", err)
	}

	return nil
}
