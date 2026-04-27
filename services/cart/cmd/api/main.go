package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/services/cart/internal/api/controller"
	"github.com/lppduy/ecom-poc/services/cart/internal/config"
	"github.com/lppduy/ecom-poc/services/cart/internal/repository"
	"github.com/lppduy/ecom-poc/services/cart/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := repository.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	if err := repository.InitSchema(db); err != nil {
		log.Fatalf("failed to init schema: %v", err)
	}

	repo := repository.NewCartRepository(db)
	cartService := service.NewCartService(repo)
	cartController := controller.NewCartController(cartService, cfg.DefaultUserID)

	router := gin.Default()
	cartController.RegisterRoutes(router)

	addr := ":" + cfg.Port
	fmt.Printf("service started on %s\n", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
