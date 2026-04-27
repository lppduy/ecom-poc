package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/lppduy/ecom-poc/services/catalog/internal/api/controller"
	"github.com/lppduy/ecom-poc/services/catalog/internal/config"
	"github.com/lppduy/ecom-poc/services/catalog/internal/repository"
	"github.com/lppduy/ecom-poc/services/catalog/internal/service"
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

	repo := repository.NewProductRepository(db)
	if err := repo.SeedDefaultsIfEmpty(); err != nil {
		log.Fatalf("failed to seed products: %v", err)
	}

	productService := service.NewProductService(repo)
	productController := controller.NewProductController(productService)

	router := gin.Default()
	productController.RegisterRoutes(router)

	addr := ":" + cfg.Port
	fmt.Printf("service started on %s\n", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
