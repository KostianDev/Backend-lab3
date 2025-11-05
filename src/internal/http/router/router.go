package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"bckndlab3/src/internal/http/handlers"
	"bckndlab3/src/internal/http/middleware"
)

// Dependencies groups handler dependencies for router setup.
type Dependencies struct {
	Auth    *handlers.AuthHandler
	Account *handlers.AccountHandler
}

// New creates and configures the HTTP router.
func New(deps Dependencies) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery(), middleware.ErrorHandler())

	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := engine.Group("/api/v1")

	auth := api.Group("/auth")
	deps.Auth.RegisterRoutes(auth)

	accounts := api.Group("/accounts")
	deps.Account.RegisterRoutes(accounts)

	return engine
}
