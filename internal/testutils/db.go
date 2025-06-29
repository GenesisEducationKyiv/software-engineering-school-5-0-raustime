package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"weatherapi/internal/db/migration"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// SetupTestDB creates a test database connection and runs migrations.
func SetupTestDB(t *testing.T) *bun.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DB_URL")
	if dsn == "" {
		t.Fatal("TEST_DB_URL environment variable not set")
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	ctx := context.Background()

	// Test database connection.
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("failed to ping test database: %v", err)
	}

	// Get migrations directory path.
	migrationsDir := getMigrationsDir()
	t.Logf("Resolved migrations path: %s", migrationsDir)
	// Run migrations.
	migrationRunner := migration.NewRunner(db, migrationsDir)
	if err := migrationRunner.RunMigrations(ctx); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Clean up function to be called in test cleanup.
	t.Cleanup(func() {
		cleanupTestDB(t, db)
	})

	return db
}

// CleanupTestDB truncates all tables for clean test state.
func CleanupTestDB(t *testing.T, db *bun.DB) {
	t.Helper()
	cleanupTestDB(t, db)
}

// cleanupTestDB performs the actual cleanup.
func cleanupTestDB(t *testing.T, db *bun.DB) {
	t.Helper()
	ctx := context.Background()

	// List of tables to truncate (add more as needed).
	tables := []string{
		"subscriptions",
	}

	for _, table := range tables {
		query := `TRUNCATE TABLE ` + table + ` RESTART IDENTITY CASCADE;`
		if _, err := db.ExecContext(ctx, query); err != nil {
			t.Logf("Warning: failed to truncate table %s: %v", table, err)
		}
	}
}

func getMigrationsDir() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	projectRoot := filepath.Join(dir, "..", "..")
	migrationsDir := filepath.Join(projectRoot, "migrations")

	fmt.Println("💡 Looking for migrations at:", migrationsDir)

	return migrationsDir
}

// CreateTestDBWithData creates a test DB and populates it with test data.
func CreateTestDBWithData(t *testing.T) *bun.DB {
	t.Helper()
	db := SetupTestDB(t)
	return db
}
