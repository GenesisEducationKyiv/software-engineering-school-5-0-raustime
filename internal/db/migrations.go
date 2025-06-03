package db

import (
	"context"
	"fmt"
	"weatherapi/internal/db/models"

	"github.com/uptrace/bun"
)

func RunMigrations(ctx context.Context, db *bun.DB) error {
	// Спершу видаляємо послідовність, якщо існує
	_, err := db.ExecContext(ctx, `DROP SEQUENCE IF EXISTS subscriptions_id_seq CASCADE;`)
	if err != nil {
		return fmt.Errorf("failed to drop sequence: %w", err)
	}

	_, err = db.ExecContext(ctx, `DROP INDEX IF EXISTS pg_class_relname_nsp_index CASCADE;`)
	if err != nil {
		return fmt.Errorf("failed to drop ІNDEX: %w", err)
	}

	// Видаляємо таблицю subscriptions, якщо існує
	_, err = db.NewDropTable().
		Model((*models.Subscription)(nil)).
		IfExists().
		Cascade().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to drop table subscriptions: %w", err)
	}

	// Створюємо таблицю заново
	_, err = db.NewCreateTable().
		Model((*models.Subscription)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create table subscriptions: %w", err)
	}

	return nil
}
