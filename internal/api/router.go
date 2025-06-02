package api

import (
	"weatherapi/internal/api/handlers"
	"weatherapi/internal/mailer"

	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
)

func SetupRouter(db bun.IDB, sender mailer.EmailSender) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	h := handlers.NewHandler(db, sender)
	// API роутери
	r.GET("/api/weather", handlers.GetWeatherHandler())
	r.POST("/api/subscribe", h.SubscribeHandler)
	r.GET("/api/confirm", h.InvalidConfirmHandler)
	r.GET("/api/confirm/*tokenPath", h.ConfirmHandler)
	r.GET("/api/unsubscribe", h.InvalidUnsubscribeHandler)
	r.GET("/api/unsubscribe/*tokenPath", h.UnsubscribeHandler)

	return r
}
