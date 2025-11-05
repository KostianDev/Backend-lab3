package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"bckndlab3/src/internal/config"
	"bckndlab3/src/internal/database"
	"bckndlab3/src/internal/http/handlers"
	"bckndlab3/src/internal/http/router"
	"bckndlab3/src/internal/migrations"
	"bckndlab3/src/internal/services"
	"bckndlab3/src/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to unwrap database handle: %v", err)
	}
	defer sqlDB.Close()

	if err := migrations.Run(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	gin.SetMode(cfg.GinMode)

	authService := storage.NewAuthService(db)
	accountService := storage.NewAccountService(db, cfg.AllowNegativeBalance)

	timeProvider := services.SystemTimeProvider{}

	authHandler := handlers.NewAuthHandler(authService)
	accountHandler := handlers.NewAccountHandler(accountService, timeProvider)

	engine := router.New(router.Dependencies{
		Auth:    authHandler,
		Account: accountHandler,
	})

	if err := engine.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatalf("failed to start HTTP server: %v", err)
	}
}
