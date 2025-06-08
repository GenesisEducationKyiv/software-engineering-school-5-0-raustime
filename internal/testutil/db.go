package testutil

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"weatherapi/internal/db/migration"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// SetupTestDB creates a test database connection and runs migrations
func SetupTestDB(t *testing.T) *bun.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DB_URL")
	if dsn == "" {
		t.Fatal("TEST_DB_URL environment variable not set")
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	ctx := context.Background()

	// Test database connection
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("failed to ping test database: %v", err)
	}

	// Get migrations directory path
	migrationsDir := getMigrationsDir()

	// Run migrations
	migrationRunner := migration.NewRunner(db, migrationsDir)
	if err := migrationRunner.RunMigrations(ctx); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Clean up function to be called in test cleanup
	t.Cleanup(func() {
		cleanupTestDB(t, db)
	})

	return db
}

// CleanupTestDB truncates all tables for clean test state
func CleanupTestDB(t *testing.T, db *bun.DB) {
	t.Helper()
	cleanupTestDB(t, db)
}

// cleanupTestDB performs the actual cleanup
func cleanupTestDB(t *testing.T, db *bun.DB) {
	ctx := context.Background()

	// List of tables to truncate (add more as needed)
	tables := []string{
		"subscriptions",
		// Add other table names here
	}

	for _, table := range tables {
		query := `TRUNCATE TABLE ` + table + ` RESTART IDENTITY CASCADE;`
		if _, err := db.ExecContext(ctx, query); err != nil {
			t.Logf("Warning: failed to truncate table %s: %v", table, err)
		}
	}
}

// getMigrationsDir returns the path to migrations directory
func getMigrationsDir() string {
	// Get the current file's directory
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)

	// Go up to project root and find migrations directory
	projectRoot := filepath.Join(dir, "..", "..")
	migrationsDir := filepath.Join(projectRoot, "migrations")

	return migrationsDir
}

// CreateTestDBWithData creates a test DB and populates it with test data
func CreateTestDBWithData(t *testing.T) *bun.DB {
	t.Helper()

	db := SetupTestDB(t)

	// Add any common test data here
	// Example:
	// seedTestData(t, db)

	return db
}

// Optional: Seed test data function
func seedTestData(t *testing.T, db *bun.DB) {
	t.Helper()

	ctx := context.Background()

	// Example: Insert test subscriptions
	testSubscriptions := []map[string]interface{}{
		{
			"email":     "test1@example.com",
			"city":      "Kyiv",
			"frequency": "daily",
			"confirmed": true,
			"token":     "test_token_1",
		},
		{
			"email":     "test2@example.com",
			"city":      "Lviv",
			"frequency": "weekly",
			"confirmed": false,
			"token":     "test_token_2",
		},
	}

	for _, sub := range testSubscriptions {
		_, err := db.NewInsert().
			Model(&sub).
			Table("subscriptions").
			Exec(ctx)
		if err != nil {
			t.Fatalf("failed to seed test data: %v", err)
		}
	}
}
