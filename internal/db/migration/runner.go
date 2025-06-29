package migration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/uptrace/bun"
)

// Migration represents a single migration.
type Migration struct {
	Version string
	Name    string
	UpSQL   string
}

// Runner handles database migrations.
type Runner struct {
	db            *bun.DB
	migrationsDir string
}

// NewRunner creates a new migration runner.
func NewRunner(db *bun.DB, migrationsDir string) *Runner {
	return &Runner{
		db:            db,
		migrationsDir: migrationsDir,
	}
}

// RunMigrations executes all pending migrations.
func (r *Runner) RunMigrations(ctx context.Context) error {
	// Create migrations table if it doesn't exist.
	if err := r.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Load all migration files.
	migrations, err := r.loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Get applied migrations.
	appliedMigrations, err := r.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Apply pending migrations.
	for _, migration := range migrations {
		if _, applied := appliedMigrations[migration.Version]; applied {
			fmt.Printf("Migration %s already applied, skipping\n", migration.Version)
			continue
		}

		fmt.Printf("Applying migration %s: %s\n", migration.Version, migration.Name)
		if err := r.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
		}
	}

	return nil
}

// createMigrationsTable creates the migrations tracking table.
func (r *Runner) createMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			version VARCHAR PRIMARY KEY,
			name VARCHAR NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
		)
	`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *Runner) loadMigrations() ([]Migration, error) {
	var migrations []Migration

	err := filepath.WalkDir(r.migrationsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".up.sql") {
			return nil
		}

		// Extract version and name from filename.
		filename := d.Name()
		parts := strings.Split(filename, "_")
		if len(parts) < 2 {
			return fmt.Errorf("invalid migration filename format: %s", filename)
		}

		version := parts[0]
		name := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".up.sql")

		// Read up migration.
		upSQL, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read up migration file %s: %w", path, err)
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    name,
			UpSQL:   string(upSQL),
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort migrations by version.
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// getAppliedMigrations returns a map of applied migration versions.
func (r *Runner) getAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT version FROM migrations")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Failed to close rows: %v", err)
		}
	}()
	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// applyMigration applies a single migration.
func (r *Runner) applyMigration(ctx context.Context, migration Migration) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("tx rollback failed: %v", err)
		}
	}()

	// Execute migration SQL.
	if _, err := tx.ExecContext(ctx, migration.UpSQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO migrations (version, name) VALUES (?, ?)",
		migration.Version, migration.Name,
	)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}

func (r *Runner) Rollback(ctx context.Context) error {
	return fmt.Errorf("rollback is no longer supported")
}
