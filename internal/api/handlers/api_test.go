package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"weatherapi/internal/api/handlers"
	dbpkg "weatherapi/internal/db"
	"weatherapi/internal/db/models"
	"weatherapi/internal/mailer"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func setupRouter(db bun.IDB, sender mailer.EmailSender) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	h := handlers.NewHandler(db, sender)

	r.POST("/api/subscribe", h.SubscribeHandler)
	r.GET("/api/confirm/:token", h.ConfirmHandler)
	r.GET("/api/unsubscribe/:token", h.UnsubscribeHandler)

	return r
}

func TestSubscribe_Success(t *testing.T) {
	// використайте in-memory DB або мок
	db := setupTestDB(t)
	mockSender := &mailer.MockSender{}
	router := setupRouter(db, mockSender)

	payload := map[string]string{
		"email":     "test@example.com",
		"city":      "Kyiv",
		"frequency": "daily",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/subscribe", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSubscribe_Conflict(t *testing.T) {
	db := setupTestDB(t)
	mockSender := &mailer.MockSender{}
	_ = insertTestSubscription(db, "test@example.com")

	router := setupRouter(db, mockSender)

	payload := map[string]string{
		"email":     "test@example.com",
		"city":      "Kyiv",
		"frequency": "daily",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/subscribe", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}
func TestConfirm_Success(t *testing.T) {
	db := setupTestDB(t)
	token := uuid.New().String()
	_ = insertTestSubscriptionWithToken(db, "test@example.com", token, t)

	mockSender := &mailer.MockSender{}
	router := setupRouter(db, mockSender)

	req, _ := http.NewRequest("GET", "/api/confirm/"+token, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
func TestUnsubscribe_Success(t *testing.T) {
	db := setupTestDB(t)
	token := uuid.New().String()
	_ = insertTestSubscriptionWithToken(db, "test@example.com", token, t)

	mockSender := &mailer.MockSender{}
	router := setupRouter(db, mockSender)

	req, _ := http.NewRequest("GET", "/api/unsubscribe/"+token, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
func insertTestSubscription(db bun.IDB, email string) error {
	sub := &models.Subscription{
		Email:     email,
		City:      "Kyiv",
		Frequency: "daily",
		Token:     uuid.New().String(),
		CreatedAt: time.Now(),
	}
	_, err := db.NewInsert().Model(sub).Exec(context.Background())
	return err
}

func insertTestSubscriptionWithToken(db bun.IDB, email, token string, t *testing.T) error {
	sub := &models.Subscription{
		Email:     email,
		City:      "Kyiv",
		Frequency: "daily",
		Token:     token,
		CreatedAt: time.Now(),
	}
	//_, err := db.NewInsert().Model(sub).Exec(context.Background())
	res, err := db.NewInsert().Model(sub).Exec(context.Background())
	t.Log("Inserted subscription with token:", token, "result:", res)
	return err
}

func setupTestDB(t *testing.T) *bun.DB {
	dsn := os.Getenv("TEST_DB_URL")
	if dsn == "" {
		t.Fatal("TEST_DB_URL not set")
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())
	ctx := context.Background()

	// Запускаємо міграції (в твоєму пакеті dbpkg)
	if err := dbpkg.RunMigrations(ctx, db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Очистка таблиці subscriptions
	_, err := db.ExecContext(ctx, `TRUNCATE TABLE subscriptions RESTART IDENTITY CASCADE;`)
	if err != nil {
		t.Fatalf("failed to truncate subscriptions table: %v", err)
	}

	return db
}
