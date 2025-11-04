package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"bckndlab3/src/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	gin.SetMode(cfg.GinMode)

	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	if err := router.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}
}
