package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	inventorypb "github.com/lppduy/ecom-poc/gen/inventory"
	"github.com/lppduy/ecom-poc/services/inventory/internal/api/controller"
	"github.com/lppduy/ecom-poc/services/inventory/internal/api/routes"
	"github.com/lppduy/ecom-poc/services/inventory/internal/config"
	inventorygrpc "github.com/lppduy/ecom-poc/services/inventory/internal/grpc"
	"github.com/lppduy/ecom-poc/services/inventory/internal/repository"
	"github.com/lppduy/ecom-poc/services/inventory/internal/service"
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

	repo := repository.NewStockRepository(db)
	if err := repo.SeedDefaultsIfEmpty(); err != nil {
		log.Fatalf("failed to seed stock: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}
	flashSaleRepo := repository.NewFlashSaleRepository(redisClient)

	inventoryService := service.NewInventoryService(repo)
	inventoryController := controller.NewInventoryController(inventoryService, flashSaleRepo)

	// gRPC server (port 9084) for internal service-to-service calls
	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			log.Fatalf("inventory: grpc listen: %v", err)
		}
		grpcServer := grpc.NewServer()
		inventorypb.RegisterInventoryServiceServer(grpcServer, inventorygrpc.NewInventoryServer(inventoryService))
		log.Printf("inventory gRPC server listening on :%s", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("inventory: grpc serve: %v", err)
		}
	}()

	// HTTP server (port 8084) for external/admin endpoints
	router := gin.Default()
	routes.Register(router, inventoryController)

	addr := ":" + cfg.Port
	fmt.Printf("inventory HTTP server listening on %s\n", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
