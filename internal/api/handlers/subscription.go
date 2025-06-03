package handlers

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"weatherapi/internal/db/models"
	"weatherapi/internal/mailer"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

var ErrAlreadySubscribed = errors.New("email already subscribed")

type Handler struct {
	DB     bun.IDB
	Sender mailer.EmailSender
}

func NewHandler(db bun.IDB, sender mailer.EmailSender) *Handler {
	return &Handler{
		DB:     db,
		Sender: sender,
	}
}

func (h *Handler) SubscribeHandler(c *gin.Context) {

	var form struct {
		Email     string `form:"email" binding:"required,email"`
		City      string `form:"city" binding:"required"`
		Frequency string `form:"frequency" binding:"required,oneof=hourly daily"`
	}
	if err := c.ShouldBind(&form); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Check if already subscribed
	var existing models.Subscription
	err := h.DB.NewSelect().Model(&existing).Where("email = ?", form.Email).Scan(c)
	if err == nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	token := uuid.New().String()
	sub := &models.Subscription{
		Email:     form.Email,
		City:      form.City,
		Frequency: form.Frequency,
		Token:     token,
		CreatedAt: time.Now(),
	}

	if _, err := h.DB.NewInsert().Model(sub).Exec(c); err != nil {
		log.Printf("DB insert error: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Send confirmation email
	if err := mailer.SendConfirmationEmailWithSender(h.Sender, form.Email, token); err != nil {
		// Підписку вже створено, але повідомляємо про помилку з email
		log.Printf("Failed to send confirmation email to %s: %v", form.Email, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) InvalidConfirmHandler(c *gin.Context) {
	c.AbortWithStatus(http.StatusBadRequest)
	return
}

func (h *Handler) ConfirmHandler(c *gin.Context) {

	tokenWithSlash := c.Param("tokenPath")
	token := strings.TrimPrefix(tokenWithSlash, "/")

	if _, err := uuid.Parse(token); err != nil {
		c.AbortWithStatus(http.StatusBadRequest) // 400: Invalid token format
		return
	}

	var sub models.Subscription
	err := h.DB.NewSelect().Model(&sub).Where("token = ?", token).Scan(c)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	sub.Confirmed = true
	sub.ConfirmedAt = time.Now()
	if _, err := h.DB.NewUpdate().Model(&sub).WherePK().Exec(c); err != nil {
		log.Printf("Failed to confirm token %s: %v", token, err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	c.Status(http.StatusOK)
}

func (h *Handler) InvalidUnsubscribeHandler(c *gin.Context) {
	c.AbortWithStatus(http.StatusBadRequest)
	return
}

func (h *Handler) UnsubscribeHandler(c *gin.Context) {

	tokenWithSlash := c.Param("tokenPath")
	token := strings.TrimPrefix(tokenWithSlash, "/")

	if _, err := uuid.Parse(token); err != nil {
		c.AbortWithStatus(http.StatusBadRequest) // 400: Invalid token format
		return
	}

	res, err := h.DB.NewDelete().Model((*models.Subscription)(nil)).Where("token = ?", token).Exec(c)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	count, err := res.RowsAffected()
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if count == 0 {
		c.AbortWithStatus(http.StatusNotFound) // 404 Token not found
		return
	}

	c.Status(http.StatusOK)

}
