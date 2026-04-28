package main

import (
	"context"
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

	redisClient := repository.NewRedisClient(cfg.RedisAddr)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}
	log.Printf("connected to redis at %s", cfg.RedisAddr)

	repo := repository.NewRedisCartRepository(redisClient)
	cartService := service.NewCartService(repo)
	cartController := controller.NewCartController(cartService)

	router := gin.Default()
	cartController.RegisterRoutes(router, cfg.JWTSecret)

	addr := ":" + cfg.Port
	fmt.Printf("service started on %s\n", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
