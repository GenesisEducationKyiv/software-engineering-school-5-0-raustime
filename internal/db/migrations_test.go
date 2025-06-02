package db_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"weatherapi/internal/db"
	"weatherapi/internal/db/models"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func setupTestDB(t *testing.T) *bun.DB {
	dsn := os.Getenv("TEST_DB_URL")
	if dsn == "" {
		t.Fatal("TEST_DB_URL not set")
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	db := bun.NewDB(sqldb, pgdialect.New())
	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func TestRunMigrations(t *testing.T) {
	ctx := context.Background()
	dbConn := setupTestDB(t)

	err := db.RunMigrations(ctx, dbConn)
	assert.NoError(t, err)

	// Перевіряємо, чи таблиця створена (Count має повернути 0 записів)
	count, err := dbConn.NewSelect().
		Model((*models.Subscription)(nil)).
		Count(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}
