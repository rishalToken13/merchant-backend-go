package app

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"token13/merchant-backend-go/internal/transport/http/handlers"
)

type API struct {
	Engine *gin.Engine
}

func NewAPI(log *slog.Logger, authH *handlers.AuthHandler) *API {
	r := gin.New()
	r.Use(gin.Recovery())

	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info("http_request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
		)
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	v1 := r.Group("/v1")
	authGroup := v1.Group("/auth")
	authGroup.POST("/register", authH.Register)
	authGroup.POST("/login", authH.Login)

	return &API{Engine: r}
}