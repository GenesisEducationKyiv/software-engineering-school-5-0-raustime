package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"weatherapi/internal/api/handlers"
	"weatherapi/internal/db/models"
	"weatherapi/internal/mailer"
	"weatherapi/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

const testTemplateDir = "test_handler_templates"

// TestMain для налаштування тестового оточення для всіх handler тестів
func TestMain(m *testing.M) {
	log.Printf("DEBUG: TestMain started")

	// Встановлюємо тестові змінні оточення ПЕРЕД створенням шаблонів
	os.Setenv("APP_BASE_URL", "https://example.com")
	os.Setenv("TEMPLATE_DIR", testTemplateDir)

	log.Printf("DEBUG: Set TEMPLATE_DIR to: %s", os.Getenv("TEMPLATE_DIR"))

	// Setup: створюємо тестові шаблони для handlers
	if err := setupHandlerTestTemplates(); err != nil {
		fmt.Printf("Failed to setup handler test templates: %v\n", err)
		os.Exit(1)
	}

	// Оновлюємо TemplateDir в mailer пакеті - ВАЖЛИВО: робимо це після встановлення env var
	mailer.SetTemplateDir(testTemplateDir)

	log.Printf("DEBUG: After SetTemplateDir, mailer.TemplateDir should be: %s", testTemplateDir)

	// Запускаємо тести
	code := m.Run()

	// Cleanup
	cleanupHandlerTestTemplates()
	os.Unsetenv("TEMPLATE_DIR")
	os.Unsetenv("APP_BASE_URL")

	os.Exit(code)
}

func setupHandlerTestTemplates() error {
	log.Printf("DEBUG: Creating templates in: %s", testTemplateDir)

	// Видаляємо директорію якщо вона існує, щоб почати з чистого аркуша
	if err := os.RemoveAll(testTemplateDir); err != nil {
		log.Printf("DEBUG: Warning - could not remove existing template dir: %v", err)
	}

	if err := os.MkdirAll(testTemplateDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	confirmationTemplate := `<!DOCTYPE html>
<html>
<head><title>Confirm Subscription</title></head>
<body>
	<h1>Confirm your subscription</h1>
	<p>Click <a href="{{.ConfirmURL}}">here</a> to confirm</p>
</body>
</html>`

	confirmationPath := filepath.Join(testTemplateDir, "confirmation_email.html")
	if err := os.WriteFile(confirmationPath, []byte(confirmationTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write confirmation_email.html: %w", err)
	}
	log.Printf("DEBUG: Created template: %s", confirmationPath)

	weatherTemplate := `<!DOCTYPE html>
<html>
<head><title>Weather Update</title></head>
<body>
	<h1>Weather in {{.City}}</h1>
	<p>Temperature: {{.Temperature}}°C</p>
	<p>Description: {{.Description}}</p>
	<p><a href="{{.UnsubscribeURL}}">Unsubscribe</a></p>
</body>
</html>`

	weatherPath := filepath.Join(testTemplateDir, "weather_email.html")
	if err := os.WriteFile(weatherPath, []byte(weatherTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write weather_email.html: %w", err)
	}
	log.Printf("DEBUG: Created template: %s", weatherPath)

	// Перевіряємо, що файли дійсно створилися та доступні для читання
	if _, err := os.Stat(confirmationPath); os.IsNotExist(err) {
		return fmt.Errorf("confirmation template was not created: %s", confirmationPath)
	}
	if _, err := os.Stat(weatherPath); os.IsNotExist(err) {
		return fmt.Errorf("weather template was not created: %s", weatherPath)
	}

	// Додаткова перевірка - можемо прочитати файли
	if data, err := os.ReadFile(confirmationPath); err != nil {
		return fmt.Errorf("cannot read confirmation template: %w", err)
	} else {
		log.Printf("DEBUG: Confirmation template content length: %d", len(data))
	}

	if data, err := os.ReadFile(weatherPath); err != nil {
		return fmt.Errorf("cannot read weather template: %w", err)
	} else {
		log.Printf("DEBUG: Weather template content length: %d", len(data))
	}

	log.Printf("DEBUG: Templates setup completed successfully")
	return nil
}

func setupRouter(db bun.IDB, sender mailer.EmailSender) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	h := handlers.NewHandler(db, sender)

	r.POST("/api/subscribe", h.SubscribeHandler)
	r.GET("/api/confirm/:token", h.ConfirmHandler)
	r.GET("/api/unsubscribe/:token", h.UnsubscribeHandler)

	return r
}

func cleanupHandlerTestTemplates() {
	log.Printf("DEBUG: Cleaning up test templates")
	if err := os.RemoveAll(testTemplateDir); err != nil {
		log.Printf("DEBUG: Warning - failed to cleanup templates: %v", err)
	}
}

func TestSubscribe_Success(t *testing.T) {
	log.Printf("DEBUG: TestSubscribe_Success started")

	// Перевіряємо поточний стан
	log.Printf("DEBUG: Current working directory: %s", getCurrentDir())
	log.Printf("DEBUG: TEMPLATE_DIR env var: %s", os.Getenv("TEMPLATE_DIR"))

	// Перевіряємо, чи існують шаблони
	templatePath := filepath.Join(testTemplateDir, "confirmation_email.html")
	if _, err := os.Stat(templatePath); err != nil {
		t.Fatalf("Template file does not exist: %s, error: %v", templatePath, err)
	}
	log.Printf("DEBUG: Template file exists: %s", templatePath)

	// Використовуємо тестову БД
	db := testutil.SetupTestDB(t)
	mockSender := &mailer.MockSender{}

	// Замінюємо глобальний Email sender на наш мок
	originalEmail := mailer.Email
	mailer.Email = mockSender
	defer func() { mailer.Email = originalEmail }()

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

	log.Printf("DEBUG: About to call router.ServeHTTP")
	router.ServeHTTP(w, req)

	log.Printf("DEBUG: Response status: %d", w.Code)
	log.Printf("DEBUG: Response body: %s", w.Body.String())

	assert.Equal(t, http.StatusOK, w.Code)

	// Перевіряємо, що email було "відправлено"
	assert.Equal(t, "test@example.com", mockSender.LastTo)
	assert.Equal(t, "Confirm your subscription", mockSender.LastSubject)
	assert.Contains(t, mockSender.LastBody, "confirm")
	assert.Contains(t, mockSender.LastBody, "https://example.com/api/confirm/")
}

func TestSubscribe_InvalidEmail(t *testing.T) {
	db := testutil.SetupTestDB(t)
	mockSender := &mailer.MockSender{}
	router := setupRouter(db, mockSender)

	payload := map[string]string{
		"email":     "invalid-email",
		"city":      "Kyiv",
		"frequency": "daily",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/subscribe", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubscribe_MissingFields(t *testing.T) {
	db := testutil.SetupTestDB(t)
	mockSender := &mailer.MockSender{}
	router := setupRouter(db, mockSender)

	payload := map[string]string{
		"email": "test@example.com",
		// missing city and frequency
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/subscribe", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubscribe_InvalidFrequency(t *testing.T) {
	db := testutil.SetupTestDB(t)
	mockSender := &mailer.MockSender{}
	router := setupRouter(db, mockSender)

	payload := map[string]string{
		"email":     "test@example.com",
		"city":      "Kyiv",
		"frequency": "weekly", // invalid - should be hourly or daily
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/subscribe", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSubscribe_AlreadyExists(t *testing.T) {
	db := testutil.SetupTestDB(t)
	mockSender := &mailer.MockSender{}

	// Створюємо існуючу підписку
	existing := &models.Subscription{
		Email:     "test@example.com",
		City:      "Kyiv",
		Frequency: "daily",
		Token:     uuid.New().String(),
		CreatedAt: time.Now(),
	}
	_, err := db.NewInsert().Model(existing).Exec(context.Background())
	require.NoError(t, err)

	router := setupRouter(db, mockSender)

	payload := map[string]string{
		"email":     "test@example.com",
		"city":      "Lviv",
		"frequency": "hourly",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/subscribe", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestSubscribe_EmailSendError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	errorSender := &ErrorMockSender{err: fmt.Errorf("SMTP server unavailable")}

	router := setupRouter(db, errorSender)

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

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestConfirm_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	mockSender := &mailer.MockSender{}

	// Створюємо підписку для підтвердження
	token := uuid.New().String()
	sub := &models.Subscription{
		Email:     "test@example.com",
		City:      "Kyiv",
		Frequency: "daily",
		Token:     token,
		Confirmed: false,
		CreatedAt: time.Now(),
	}
	_, err := db.NewInsert().Model(sub).Exec(context.Background())
	require.NoError(t, err)

	router := setupRouter(db, mockSender)

	req, _ := http.NewRequest("GET", "/api/confirm/"+token, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Перевіряємо, що підписка підтверджена
	var updated models.Subscription
	err = db.NewSelect().Model(&updated).Where("token = ?", token).Scan(context.Background())
	require.NoError(t, err)
	assert.True(t, updated.Confirmed)
	assert.False(t, updated.ConfirmedAt.IsZero())
}

func TestConfirm_InvalidToken(t *testing.T) {
	db := testutil.SetupTestDB(t)
	mockSender := &mailer.MockSender{}
	router := setupRouter(db, mockSender)

	req, _ := http.NewRequest("GET", "/api/confirm/invalid-token", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConfirm_TokenNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	mockSender := &mailer.MockSender{}
	router := setupRouter(db, mockSender)

	nonExistentToken := uuid.New().String()
	req, _ := http.NewRequest("GET", "/api/confirm/"+nonExistentToken, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUnsubscribe_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	mockSender := &mailer.MockSender{}

	// Створюємо підписку
	token := uuid.New().String()
	sub := &models.Subscription{
		Email:     "test@example.com",
		City:      "Kyiv",
		Frequency: "daily",
		Token:     token,
		Confirmed: true,
		CreatedAt: time.Now(),
	}
	_, err := db.NewInsert().Model(sub).Exec(context.Background())
	require.NoError(t, err)

	router := setupRouter(db, mockSender)

	req, _ := http.NewRequest("GET", "/api/unsubscribe/"+token, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Перевіряємо, що підписка видалена
	var deleted models.Subscription
	err = db.NewSelect().Model(&deleted).Where("token = ?", token).Scan(context.Background())
	assert.Error(t, err) // Should not find the record
}

func TestUnsubscribe_InvalidToken(t *testing.T) {
	db := testutil.SetupTestDB(t)
	mockSender := &mailer.MockSender{}
	router := setupRouter(db, mockSender)

	req, _ := http.NewRequest("GET", "/api/unsubscribe/invalid-token", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUnsubscribe_TokenNotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	mockSender := &mailer.MockSender{}
	router := setupRouter(db, mockSender)

	nonExistentToken := uuid.New().String()
	req, _ := http.NewRequest("GET", "/api/unsubscribe/"+nonExistentToken, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// Допоміжна функція для отримання поточної директорії
func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}

// Допоміжний мок для тестування помилок
type ErrorMockSender struct {
	err error
}

func (e *ErrorMockSender) Send(to, subject, htmlBody string) error {
	return e.err
}
