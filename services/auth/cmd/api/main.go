package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/services/auth/internal/api/controller"
	"github.com/lppduy/ecom-poc/services/auth/internal/api/routes"
	"github.com/lppduy/ecom-poc/services/auth/internal/config"
	"github.com/lppduy/ecom-poc/services/auth/internal/repository"
	"github.com/lppduy/ecom-poc/services/auth/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := repository.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("auth: connect postgres: %v", err)
	}

	userRepo := repository.NewGormUserRepository(db)
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret)
	authCtrl := controller.NewAuthController(authSvc, cfg.JWTSecret)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	routes.Register(r, authCtrl, cfg.JWTSecret)

	log.Printf("auth service listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("auth: run: %v", err)
	}
}
