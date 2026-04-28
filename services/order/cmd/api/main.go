package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/lppduy/ecom-poc/services/order/internal/api/controller"
	"github.com/lppduy/ecom-poc/services/order/internal/api/routes"
	"github.com/lppduy/ecom-poc/services/order/internal/client"
	"github.com/lppduy/ecom-poc/services/order/internal/config"
	"github.com/lppduy/ecom-poc/services/order/internal/event"
	"github.com/lppduy/ecom-poc/services/order/internal/repository"
	"github.com/lppduy/ecom-poc/services/order/internal/service"
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

	brokers := strings.Split(cfg.KafkaBrokers, ",")
	publisher := event.NewKafkaPublisher(brokers)
	defer publisher.Close()

	outboxRepo := repository.NewOutboxRepository(db)
	event.StartRelay(context.Background(), outboxRepo, publisher)

	repo := repository.NewOrderRepository(db)
	cartClient := client.NewCartHTTPClient(cfg.CartBaseURL)
	inventoryClient := client.NewInventoryHTTPClient(cfg.InventoryBaseURL)
	orderService := service.NewOrderService(repo, cartClient, inventoryClient)
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("order: connect redis: %v", err)
	}

	orderController := controller.NewOrderController(orderService)

	router := gin.Default()
	routes.RegisterOrderRoutes(router, orderController, cfg.JWTSecret, rdb)

	addr := ":" + cfg.Port
	fmt.Printf("service started on %s\n", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
